// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package utils

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	DefaultSSHPort = 22
	SSHTimeout     = 2 * time.Minute
)

type SSHKeyPair struct {
	PublicKey  []byte
	PrivateKey []byte
}

// Save is responsible for writing the SSHKeyPair to the filesystem.
func (p *SSHKeyPair) Save(privKeyFile, pubKeyFile string) error {
	err := os.WriteFile(privKeyFile, p.PrivateKey, 0o600) //nolint:mnd
	if err != nil {
		return fmt.Errorf("failed to save private key file: %w", err)
	}

	err = os.WriteFile(pubKeyFile, p.PublicKey, 0o600) //nolint:gosec,mnd
	if err != nil {
		return fmt.Errorf("failed to save public key file: %w", err)
	}

	return nil
}

// Load is responsible for loading the ssh keypair from filesystem into SSHKeyPair.
func (p *SSHKeyPair) Load(privKeyFile, pubKeyFile string) error {
	var err error

	// Read the private key file.
	p.PrivateKey, err = os.ReadFile(privKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}

	// Read the public key file.
	p.PublicKey, err = os.ReadFile(pubKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	return nil
}

// GenerateSSHKeyPair generates a new SSH key pair.
func GenerateSSHKeyPair() (*SSHKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048) //nolint:mnd
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	var privateKeyData []byte
	b := bytes.NewBuffer(privateKeyData)

	if err = pem.Encode(b, privateKeyPEM); err != nil {
		return nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create public key: %w", err)
	}

	return &SSHKeyPair{
		PublicKey:  ssh.MarshalAuthorizedKey(publicKey),
		PrivateKey: b.Bytes(),
	}, nil
}

// Load SSH key-pair if provided, generate and save otherwise.
func LoadOrGenerateAndSaveSSHKeyPair(privKeyFile, pubKeyFile, workDir string) (*SSHKeyPair, error) {
	var err error
	sshKeyPair := &SSHKeyPair{}

	if privKeyFile != "" && pubKeyFile != "" {
		err = sshKeyPair.Load(privKeyFile, pubKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load ssh key pair: %w", err)
		}
	} else {
		sshKeyPair, err = GenerateSSHKeyPair()
		if err != nil {
			return nil, fmt.Errorf("failed to generate ssh key pair: %w", err)
		}

		privateKeyFile := filepath.Join(workDir, "id_rsa")
		publicKeyFile := filepath.Join(workDir, "id_rsa.pub")
		if _, err := os.Stat(privateKeyFile); errors.Is(err, os.ErrNotExist) {
			if err := sshKeyPair.Save(privateKeyFile, publicKeyFile); err != nil {
				return nil, fmt.Errorf("failed to save SSH keys to filesystem: %w", err)
			}
		}
	}

	return sshKeyPair, nil
}

type SSHForwardCallback func(ctx context.Context, err error) error

type SSHForwardInput struct {
	PrivateKey    []byte
	User          string
	Host          string
	Port          int
	LocalPort     int
	RemoteAddress string
	RemotePort    int

	Callback SSHForwardCallback
}

func (i *SSHForwardInput) HostAddressPort() string {
	return fmt.Sprintf("%s:%d", i.Host, i.Port)
}

func (i *SSHForwardInput) LocalAddressPort() string {
	return fmt.Sprintf("localhost:%d", i.LocalPort)
}

func (i *SSHForwardInput) RemoteAddressPort() string {
	if i.RemoteAddress == "" {
		i.RemoteAddress = "localhost"
	}

	return fmt.Sprintf("%s:%d", i.RemoteAddress, i.RemotePort)
}

type SSHPortForward struct {
	input        *SSHForwardInput
	clientConfig ssh.ClientConfig

	cancel context.CancelFunc
	stop   atomic.Bool
}

func (f *SSHPortForward) Start(ctx context.Context) error {
	ctx, f.cancel = context.WithCancel(ctx)

	// Dial the remote server.
	client, err := f.DialWithTimeout(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for the SSH server to be ready: %w", err)
	}

	// Listen on local port.
	listener, err := net.Listen("tcp", f.input.LocalAddressPort())
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// TODO: wait until the BannerCallBack is invoked

	go func() {
		var callbackErr error
		defer client.Close()
		defer listener.Close()

		for callbackErr == nil && !f.stop.Load() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			func() {
				// Accept local connection.
				local, err := listener.Accept()
				if err != nil {
					callbackErr = f.input.Callback(ctx, fmt.Errorf("local SSH tunnel error: %w", err))
				}
				defer local.Close()

				// Dial remote server.
				remote, err := client.Dial("tcp", f.input.RemoteAddressPort())
				if err != nil {
					callbackErr = f.input.Callback(ctx, fmt.Errorf("remote SSH tunnel error: %w", err))
				}
				defer remote.Close()

				wg := sync.WaitGroup{}
				wg.Add(1)
				go func() {
					defer wg.Done()

					_, err := io.Copy(local, remote)
					if err != nil {
						callbackErr = f.input.Callback(ctx, fmt.Errorf("failed to copy data from remote to local: %w", err))
					}
				}()

				wg.Add(1)
				go func() {
					defer wg.Done()

					_, err := io.Copy(remote, local)
					if err != nil {
						callbackErr = f.input.Callback(ctx, fmt.Errorf("failed to copy data from remote to local: %w", err))
					}
				}()

				wg.Wait()
			}()
		}
	}()

	return nil
}

func (f *SSHPortForward) Stop() {
	f.stop.Store(true)

	f.cancel()
}

func (f *SSHPortForward) DialWithTimeout(ctx context.Context) (*ssh.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, SSHTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for SSH server to be ready: %w", ctx.Err())
		default:
			conn, err := ssh.Dial("tcp", f.input.HostAddressPort(), &f.clientConfig)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			return conn, nil
		}
	}
}

func NewSSHPortForward(input *SSHForwardInput) (*SSHPortForward, error) {
	if input.Callback == nil {
		input.Callback = func(ctx context.Context, err error) error {
			logger := GetLoggerFromContextOrDiscard(ctx)
			logger.Error(err)

			return err
		}
	}

	// Create Signer from private key.
	signer, err := ssh.ParsePrivateKey(input.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Create SSH client config.
	config := ssh.ClientConfig{
		User: input.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback:   ssh.InsecureIgnoreHostKey(), //nolint:gosec
		HostKeyAlgorithms: []string{ssh.KeyAlgoED25519},
	}

	return &SSHPortForward{
		input:        input,
		clientConfig: config,
	}, nil
}

type SSHJournalctlInput struct {
	PrivateKey []byte
	PublicKey  []byte
	User       string
	Host       string
	WorkDir    string
	Service    string
}

// GetServiceLogs retrieves log entries stored in journal by systemd via ssh.
func GetServiceLogs(input *SSHJournalctlInput, startTime time.Time, stdout, stderr io.Writer) error {
	keys := &SSHKeyPair{
		PrivateKey: input.PrivateKey,
		PublicKey:  input.PublicKey,
	}
	privateKeyFile := filepath.Join(input.WorkDir, "id_rsa")
	publicKeyFile := filepath.Join(input.WorkDir, "id_rsa.pub")
	if _, err := os.Stat(privateKeyFile); errors.Is(err, os.ErrNotExist) {
		if err := keys.Save(privateKeyFile, publicKeyFile); err != nil {
			return fmt.Errorf("failed to save SSH keys to filesystem: %w", err)
		}
	}

	args := []string{
		input.User + "@" + input.Host,
		"-i", privateKeyFile,
		"journalctl",
		"-u", input.Service + ".service",
		"--since", startTime.UTC().Format("2006-01-02\\ 15:04:05"),
	}
	cmd := exec.Command("ssh", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run journalctl command via ssh: %w", err)
	}

	return nil
}
