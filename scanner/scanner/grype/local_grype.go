// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package grype

import (
	"context"
	"errors"
	"fmt"

	"github.com/anchore/clio"
	"github.com/anchore/grype/grype"
	"github.com/anchore/grype/grype/db"
	"github.com/anchore/grype/grype/grypeerr"
	"github.com/anchore/grype/grype/matcher"
	"github.com/anchore/grype/grype/matcher/dotnet"
	"github.com/anchore/grype/grype/matcher/golang"
	"github.com/anchore/grype/grype/matcher/java"
	"github.com/anchore/grype/grype/matcher/javascript"
	"github.com/anchore/grype/grype/matcher/python"
	"github.com/anchore/grype/grype/matcher/ruby"
	"github.com/anchore/grype/grype/matcher/stock"
	"github.com/anchore/grype/grype/pkg"
	grype_models "github.com/anchore/grype/grype/presenter/models"
	"github.com/anchore/grype/grype/store"
	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/cataloging"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/scanner/utils"

	"github.com/openclarity/vmclarity/scanner/config"
	"github.com/openclarity/vmclarity/scanner/job_manager"
	"github.com/openclarity/vmclarity/scanner/utils/sbom"
)

const (
	// From https://github.com/anchore/grype/blob/v0.50.1/internal/config/datasources.go#L10
	defaultMavenBaseURL = "https://search.maven.org/solrsearch/select"
)

type LocalScanner struct {
	logger     *log.Entry
	config     config.LocalGrypeConfigEx
	resultChan chan job_manager.Result
	localImage bool
}

func newLocalScanner(conf *config.Config, logger *log.Entry, resultChan chan job_manager.Result) job_manager.Job {
	return &LocalScanner{
		logger:     logger.Dup().WithField("scanner", ScannerName).WithField("mode", "local"),
		config:     config.ConvertToLocalGrypeConfig(conf.Scanner, conf.Registry),
		resultChan: resultChan,
		localImage: conf.LocalImageScan,
	}
}

func (s *LocalScanner) Run(ctx context.Context, sourceType utils.SourceType, userInput string) error {
	go s.run(sourceType, userInput)

	return nil
}

func (s *LocalScanner) run(sourceType utils.SourceType, userInput string) {
	// TODO: make `loading DB` and `gathering packages` work in parallel
	// https://github.com/anchore/grype/blob/7e8ee40996ba3a4defb5e887ab0177d99cd0e663/cmd/root.go#L240

	dbConfig := db.Config{
		DBRootDir:           s.config.DBRootDir,
		ListingURL:          s.config.ListingURL,
		ValidateByHashOnGet: false,
		MaxAllowedBuiltAge:  s.config.MaxAllowedBuiltAge,
		ListingFileTimeout:  s.config.ListingFileTimeout,
		UpdateTimeout:       s.config.UpdateTimeout,
	}
	s.logger.Infof("Loading DB. update=%v", s.config.UpdateDB)

	vulnerabilityStore, dbStatus, _, err := grype.LoadVulnerabilityDB(dbConfig, s.config.UpdateDB)

	if err = validateDBLoad(err, dbStatus); err != nil {
		ReportError(s.resultChan, fmt.Errorf("failed to load vulnerability DB: %w", err), s.logger)
		return
	}

	var hash string
	var metadata map[string]string
	origInput := userInput
	if sourceType == utils.SBOM {
		bom, err := sbom.NewCycloneDX(userInput)
		if err != nil {
			ReportError(s.resultChan, fmt.Errorf("failed to create CycloneDX SBOM: %w", err), s.logger)
			return
		}

		origInput = bom.GetTargetNameFromSBOM()
		metadata = bom.GetMetadataFromSBOM()
		hash, err = bom.GetHashFromSBOM()
		if err != nil {
			ReportError(s.resultChan, fmt.Errorf("failed to get original hash from SBOM: %w", err), s.logger)
			return
		}
	}

	source := utils.CreateSource(sourceType, s.localImage)
	s.logger.Infof("Gathering packages for source %s", source)
	providerConfig := pkg.ProviderConfig{
		SyftProviderConfig: pkg.SyftProviderConfig{
			SBOMOptions: syft.DefaultCreateSBOMConfig().
				WithSearchConfig(cataloging.DefaultSearchConfig().WithScope(s.config.Scope)),
			RegistryOptions: s.config.RegistryOptions,
		},
	}

	packages, grypeContext, _, err := pkg.Provide(source+":"+userInput, providerConfig)
	if err != nil {
		ReportError(s.resultChan, fmt.Errorf("failed to analyze packages: %w", err), s.logger)
		return
	}

	s.logger.Infof("Found %d packages", len(packages))

	vulnerabilityMatcher := createVulnerabilityMatcher(vulnerabilityStore)
	allMatches, ignoredMatches, err := vulnerabilityMatcher.FindMatches(packages, grypeContext)
	// We can ignore ErrAboveSeverityThreshold since we are not setting the FailSeverity on the matcher.
	if err != nil && !errors.Is(err, grypeerr.ErrAboveSeverityThreshold) {
		ReportError(s.resultChan, fmt.Errorf("failed to find vulnerabilities: %w", err), s.logger)
		return
	}

	s.logger.Infof("Found %d vulnerabilities", len(allMatches.Sorted()))
	id := clio.Identification{}
	doc, err := grype_models.NewDocument(id, packages, grypeContext, *allMatches, ignoredMatches, vulnerabilityStore.MetadataProvider, nil, dbStatus)
	if err != nil {
		ReportError(s.resultChan, fmt.Errorf("failed to create document: %w", err), s.logger)
		return
	}

	s.logger.Infof("Sending successful results")
	s.resultChan <- CreateResults(doc, origInput, ScannerName, hash, metadata)
}

func createVulnerabilityMatcher(store *store.Store) *grype.VulnerabilityMatcher {
	matchers := matcher.NewDefaultMatchers(matcher.Config{
		Java: java.MatcherConfig{
			ExternalSearchConfig: java.ExternalSearchConfig{
				// Disable searching maven external source (this is the default for grype CLI too)
				SearchMavenUpstream: false,
				MavenBaseURL:        defaultMavenBaseURL,
			},
			UseCPEs: true,
		},
		Ruby: ruby.MatcherConfig{
			UseCPEs: true,
		},
		Python: python.MatcherConfig{
			UseCPEs: true,
		},
		Dotnet: dotnet.MatcherConfig{
			UseCPEs: true,
		},
		Javascript: javascript.MatcherConfig{
			UseCPEs: true,
		},
		Golang: golang.MatcherConfig{
			UseCPEs: true,
		},
		Stock: stock.MatcherConfig{
			UseCPEs: true,
		},
	})
	return &grype.VulnerabilityMatcher{
		Store:          *store,
		Matchers:       matchers,
		NormalizeByCVE: true,
	}
}

func validateDBLoad(loadErr error, status *db.Status) error {
	if loadErr != nil {
		return fmt.Errorf("failed to load vulnerability db: %w", loadErr)
	}
	if status == nil {
		return errors.New("unable to determine DB status")
	}
	if status.Err != nil {
		return fmt.Errorf("db could not be loaded: %w", status.Err)
	}
	return nil
}
