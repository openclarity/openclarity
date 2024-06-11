// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sshtopology

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/scanner/families/infofinder/types"
	familiesutils "github.com/openclarity/vmclarity/scanner/families/utils"
	"github.com/openclarity/vmclarity/scanner/job_manager"
	"github.com/openclarity/vmclarity/scanner/utils"
)

const ScannerName = "sshTopology"

type Scanner struct {
	name       string
	logger     *log.Entry
	config     types.SSHTopologyConfig
	resultChan chan job_manager.Result
}

func New(_ string, c job_manager.IsConfig, logger *log.Entry, resultChan chan job_manager.Result) job_manager.Job {
	conf := c.(types.ScannersConfig) // nolint:forcetypeassert
	return &Scanner{
		name:       ScannerName,
		logger:     logger.Dup().WithField("scanner", ScannerName),
		config:     conf.SSHTopology,
		resultChan: resultChan,
	}
}

// nolint:cyclop,gocognit
func (s *Scanner) Run(ctx context.Context, sourceType utils.SourceType, userInput string) error {
	go func(ctx context.Context) {
		s.logger.Debugf("Running with input=%v and source type=%v", userInput, sourceType)
		retResults := types.ScannerResult{
			ScannerName: ScannerName,
		}

		// Validate this is an input type supported by the scanner,
		// otherwise return skipped.
		if !s.isValidInputType(sourceType) {
			s.sendResults(retResults, nil)
			return
		}

		fsPath, cleanup, err := familiesutils.ConvertInputToFilesystem(ctx, sourceType, userInput)
		if err != nil {
			s.sendResults(retResults, fmt.Errorf("failed to convert input to filesystem: %w", err))
			return
		}
		defer cleanup()

		var errs []error
		homeUserDirs := getHomeUserDirs(fsPath)
		s.logger.Debugf("Found home user dirs %+v", homeUserDirs)

		errorsChan := make(chan error)
		fingerprintsChan := make(chan []types.Info)

		var chanWg sync.WaitGroup
		chanWg.Add(1)
		go func() {
			defer chanWg.Done()
			for fingerprints := range fingerprintsChan {
				retResults.Infos = append(retResults.Infos, fingerprints...)
			}
		}()

		chanWg.Add(1)
		go func() {
			defer chanWg.Done()
			for e := range errorsChan {
				errs = append(errs, e)
			}
		}()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if sshDaemonKeysFingerprints, err := s.getSSHDaemonKeysFingerprints(fsPath); err != nil {
				errorsChan <- fmt.Errorf("failed to get ssh daemon keys: %w", err)
			} else {
				fingerprintsChan <- sshDaemonKeysFingerprints
			}
		}()

		for i := range homeUserDirs {
			dir := homeUserDirs[i]

			wg.Add(1)
			go func() {
				defer wg.Done()
				if sshPrivateKeysFingerprints, err := s.getSSHPrivateKeysFingerprints(dir); err != nil {
					errorsChan <- fmt.Errorf("failed to get ssh private keys: %w", err)
				} else {
					fingerprintsChan <- sshPrivateKeysFingerprints
				}
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				if sshAuthorizedKeysFingerprints, err := s.getSSHAuthorizedKeysFingerprints(dir); err != nil {
					errorsChan <- fmt.Errorf("failed to get ssh authorized keys: %w", err)
				} else {
					fingerprintsChan <- sshAuthorizedKeysFingerprints
				}
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				if sshKnownHostsFingerprints, err := s.getSSHKnownHostsFingerprints(dir); err != nil {
					errorsChan <- fmt.Errorf("failed to get ssh known hosts: %w", err)
				} else {
					fingerprintsChan <- sshKnownHostsFingerprints
				}
			}()
		}

		wg.Wait()
		close(errorsChan)
		close(fingerprintsChan)
		chanWg.Wait()

		retErr := errors.Join(errs...)
		if retErr != nil {
			retResults.Error = retErr
		}

		if len(retResults.Infos) > 0 && retResults.Error != nil {
			// Since we have findings, we want to share what we've got and only print the errors here.
			// Maybe we need to support to send both errors and findings in a higher level.
			s.logger.Error(retResults.Error)
			retResults.Error = nil
		}

		s.sendResults(retResults, nil)
	}(ctx)

	return nil
}

func getHomeUserDirs(rootDir string) []string {
	var dirs []string

	// Set root home if exists.
	rootHome := path.Join(rootDir, "root")
	if _, err := os.Stat(rootHome); err == nil {
		dirs = append(dirs, rootHome)
	}

	homeDirPath := path.Join(rootDir, "home")
	files, _ := os.ReadDir(homeDirPath)

	for _, f := range files {
		if f.IsDir() {
			dirs = append(dirs, path.Join(homeDirPath, f.Name()))
		}
	}

	return dirs
}

func (s *Scanner) getSSHDaemonKeysFingerprints(rootPath string) ([]types.Info, error) {
	sshDaemonConfigDir := path.Join(rootPath, "/etc/ssh")

	// Check daemon config directory exists, some setups might not have ssh
	// installed.
	_, err := os.Stat(sshDaemonConfigDir)
	if errors.Is(err, os.ErrNotExist) {
		return []types.Info{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("unexpected error checking %s exists: %w", sshDaemonConfigDir, err)
	}

	paths, err := s.getPrivateKeysPaths(sshDaemonConfigDir, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get private keys paths: %w", err)
	}
	s.logger.Debugf("Found ssh daemon private keys paths %+v", paths)

	fingerprints, err := s.getFingerprints(paths, types.SSHDaemonKeyFingerprint)
	if err != nil {
		return nil, fmt.Errorf("failed to get ssh daemon private keys fingerprints: %w", err)
	}
	s.logger.Debugf("Found ssh daemon private keys fingerprints %+v", fingerprints)

	return fingerprints, nil
}

func (s *Scanner) getSSHPrivateKeysFingerprints(homeUserDir string) ([]types.Info, error) {
	paths, err := s.getPrivateKeysPaths(homeUserDir, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get private keys paths: %w", err)
	}
	s.logger.Debugf("Found ssh private keys paths %+v", paths)

	infos, err := s.getFingerprints(paths, types.SSHPrivateKeyFingerprint)
	if err != nil {
		return nil, fmt.Errorf("failed to get ssh private keys fingerprints: %w", err)
	}
	s.logger.Debugf("Found ssh private keys fingerprints %+v", infos)

	return infos, nil
}

func (s *Scanner) getSSHAuthorizedKeysFingerprints(homeUserDir string) ([]types.Info, error) {
	infos, err := s.getFingerprints([]string{path.Join(homeUserDir, ".ssh/authorized_keys")}, types.SSHAuthorizedKeyFingerprint)
	if err != nil {
		return nil, fmt.Errorf("failed to get ssh authorized keys fingerprints: %w", err)
	}
	s.logger.Debugf("Found ssh authorized keys fingerprints %+v", infos)

	return infos, nil
}

func (s *Scanner) getSSHKnownHostsFingerprints(homeUserDir string) ([]types.Info, error) {
	infos, err := s.getFingerprints([]string{path.Join(homeUserDir, ".ssh/known_hosts")}, types.SSHKnownHostFingerprint)
	if err != nil {
		return nil, fmt.Errorf("failed to get ssh known hosts fingerprints: %w", err)
	}
	s.logger.Debugf("Found ssh known hosts fingerprints %+v", infos)

	return infos, nil
}

func (s *Scanner) getFingerprints(paths []string, infoType types.InfoType) ([]types.Info, error) {
	var infos []types.Info

	for _, p := range paths {
		_, err := os.Stat(p)
		if os.IsNotExist(err) {
			s.logger.Debugf("File (%v) does not exist.", p)
			continue
		} else if err != nil {
			return nil, fmt.Errorf("failed to check file: %w", err)
		}

		var output []byte
		if output, err = s.executeSSHKeyGenFingerprintCommand("sha256", p); err != nil {
			return nil, fmt.Errorf("failed to execute ssh-keygen command: %w", err)
		}

		infos = append(infos, parseSSHKeyGenFingerprintCommandOutput(string(output), infoType, p)...)
	}

	return infos, nil
}

func (s *Scanner) getPrivateKeysPaths(rootPath string, recursive bool) ([]string, error) {
	var paths []string
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if path != rootPath && !recursive {
				return filepath.SkipDir
			}
			return nil
		}

		isPrivateKeyFile, err := isPrivateKey(path)
		if err != nil {
			s.logger.Errorf("failed to verify if file (%v) is private key file - skipping: %v", path, err)
			return nil
		}

		if isPrivateKeyFile {
			paths = append(paths, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walks the file tree rooted at %v: %w", rootPath, err)
	}

	return paths, nil
}

func isPrivateKey(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Splits on newlines by default.
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		// We only need to look at the first line to find PEM private keys.
		return strings.Contains(scanner.Text(), "PRIVATE KEY"), nil
	}

	if err = scanner.Err(); err != nil {
		return false, fmt.Errorf("failed to scan file: %w", err)
	}

	return false, nil
}

func parseSSHKeyGenFingerprintCommandOutput(output string, infoType types.InfoType, path string) []types.Info {
	lines := strings.Split(output, "\n")
	infos := make([]types.Info, 0, len(lines))
	for i := range lines {
		if lines[i] == "" {
			continue
		}
		infos = append(infos, types.Info{
			Type: infoType,
			Path: path,
			Data: lines[i],
		})
	}
	return infos
}

func (s *Scanner) executeSSHKeyGenFingerprintCommand(hashAlgo string, filePath string) ([]byte, error) {
	args := []string{
		"-E",
		hashAlgo,
		"-l",
		"-f",
		filePath,
	}
	cmd := exec.Command("ssh-keygen", args...)
	s.logger.Infof("Running command: %v", cmd.String())
	output, err := utils.RunCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to run command: %w", err)
	}

	return output, nil
}

func (s *Scanner) isValidInputType(sourceType utils.SourceType) bool {
	switch sourceType {
	case utils.ROOTFS, utils.IMAGE, utils.DOCKERARCHIVE, utils.OCIARCHIVE, utils.OCIDIR:
		return true
	case utils.DIR, utils.FILE, utils.SBOM:
		s.logger.Infof("Source type %v is not supported for %s, skipping.", ScannerName, sourceType)
	default:
		s.logger.Infof("Unknown source type %v, skipping.", sourceType)
	}
	return false
}

func (s *Scanner) sendResults(results types.ScannerResult, err error) {
	if err != nil {
		s.logger.Error(err)
		results.Error = err
	}
	select {
	case s.resultChan <- &results:
	default:
		s.logger.Error("Failed to send results on channel")
	}
}
