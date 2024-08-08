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

package scanner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/exploits"
	"github.com/openclarity/vmclarity/scanner/families/infofinder"
	"github.com/openclarity/vmclarity/scanner/families/malware"
	"github.com/openclarity/vmclarity/scanner/families/misconfiguration"
	"github.com/openclarity/vmclarity/scanner/families/plugins"
	"github.com/openclarity/vmclarity/scanner/families/rootkits"
	"github.com/openclarity/vmclarity/scanner/families/sbom"
	"github.com/openclarity/vmclarity/scanner/families/secrets"
	"github.com/openclarity/vmclarity/scanner/families/utils"
	"github.com/openclarity/vmclarity/scanner/families/vulnerabilities"
	"github.com/openclarity/vmclarity/scanner/internal/family_runner"
	"github.com/openclarity/vmclarity/utils/fsutils/containerrootfs"
	"github.com/openclarity/vmclarity/workflow"
	workflowtypes "github.com/openclarity/vmclarity/workflow/types"
)

const (
	shutdownGracePeriod = 2 * time.Second
)

type Scanner struct {
	config *Config
	tasks  []workflowtypes.Task[workflowParams]
}

func New(config *Config) *Scanner {
	if config == nil {
		config = &Config{}
	}

	manager := &Scanner{
		config: config,
	}

	// Analyzers
	if config.SBOM.Enabled {
		manager.tasks = append(manager.tasks, newFamilyTask("sbom", sbom.New(config.SBOM)))
	}

	// Scanners
	if config.Vulnerabilities.Enabled {
		// must run after SBOM to support the case when configured to use the output from sbom
		var deps []string
		if config.SBOM.Enabled {
			deps = append(deps, "sbom")
		}

		manager.tasks = append(manager.tasks, newFamilyTask("vulnerabilities", vulnerabilities.New(config.Vulnerabilities), deps...))
	}
	if config.Secrets.Enabled {
		manager.tasks = append(manager.tasks, newFamilyTask("secrets", secrets.New(config.Secrets)))
	}
	if config.Rootkits.Enabled {
		manager.tasks = append(manager.tasks, newFamilyTask("rootkits", rootkits.New(config.Rootkits)))
	}
	if config.Malware.Enabled {
		manager.tasks = append(manager.tasks, newFamilyTask("malware", malware.New(config.Malware)))
	}
	if config.Misconfiguration.Enabled {
		manager.tasks = append(manager.tasks, newFamilyTask("misconfiguration", misconfiguration.New(config.Misconfiguration)))
	}
	if config.InfoFinder.Enabled {
		manager.tasks = append(manager.tasks, newFamilyTask("infofinder", infofinder.New(config.InfoFinder)))
	}
	if config.Plugins.Enabled {
		manager.tasks = append(manager.tasks, newFamilyTask("plugins", plugins.New(config.Plugins)))
	}

	// Enrichers.
	if config.Exploits.Enabled {
		// must run after Vulnerabilities to support the case when configured to use the output from Vulnerabilities
		var deps []string
		if config.Vulnerabilities.Enabled {
			deps = append(deps, "vulnerabilities")
		}

		manager.tasks = append(manager.tasks, newFamilyTask("exploits", exploits.New(config.Exploits), deps...))
	}

	return manager
}

func (m *Scanner) Run(ctx context.Context, notifier families.FamilyNotifier) []error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	// Register container cache
	utils.ContainerRootfsCache = containerrootfs.NewCache()
	defer func() {
		err := utils.ContainerRootfsCache.CleanupAll()
		if err != nil {
			logger.WithError(err).Errorf("Failed to cleanup all cached container rootfs files")
		}
	}()

	// Create an error channel to send/receive all processing errors to/on
	errCh := make(chan error)

	// Run task processor in the background so that we can properly subscribe and
	// listen for errors.
	go func() {
		defer close(errCh)

		// Create families processor
		processor, err := workflow.New[workflowParams, workflowtypes.Task[workflowParams]](m.tasks)
		if err != nil {
			errCh <- fmt.Errorf("failed to create families processor: %w", err)
			return
		}

		// Run families processor and wait until all family workflow tasks have
		// completed. Same workflow parameters are passed to all family runners.
		err = processor.Run(ctx, workflowParams{
			Notifier: notifier,
			Store:    families.NewResultStore(),
			ErrCh:    errCh,
		})
		if err != nil {
			// Grace period to allow families to shut down properly in the event
			// of context cancellation.
			time.Sleep(shutdownGracePeriod)
			errCh <- fmt.Errorf("failed to run families processor: %w", err)
			return
		}
	}()

	var errs []error
	var oneOrMoreFamiliesFailed bool

	// Listen for all processing errors
	for err := range errCh {
		if err == nil {
			continue
		}

		// If the received processing error is due to a family failure, skip it as we
		// have already logged it. Otherwise, collect the error as it might be due to
		// some internal error that we need to know more about.
		var ferr *family_runner.FamilyFailedError
		if errors.As(err, &ferr) {
			oneOrMoreFamiliesFailed = true
		} else {
			errs = append(errs, err)
		}
	}

	// Mark the whole scan failed in case one of the families failed
	if oneOrMoreFamiliesFailed {
		errs = append(errs, errors.New("at least one family failed to run"))
	}

	return errs
}

// workflowParams defines parameters for familyRunner workflow tasks.
type workflowParams struct {
	Notifier families.FamilyNotifier
	Store    families.ResultStore
	ErrCh    chan<- error
}

// FIXME(ramizpolic): This is an experimental feature available to track changes
// between sequential and parallel execution of the families logic. Remove this
// once all the changes regarding speed and stability have been tested and
// verified.

var (
	familiesTaskMutex    sync.Mutex
	_, familiesAsyncMode = os.LookupEnv("VMCLARITY_FAMILIES_RUN_ASYNC")
)

// newFamilyTask returns a workflow task that runs a specific family.
func newFamilyTask[T any](name string, family families.Family[T], deps ...string) workflowtypes.Task[workflowParams] {
	return workflowtypes.Task[workflowParams]{
		Name: name,
		Deps: deps,
		Fn: func(ctx context.Context, params workflowParams) error {
			// FIXME(ramizpolic): Remove once fully tested
			if !familiesAsyncMode {
				familiesTaskMutex.Lock()
				defer familiesTaskMutex.Unlock()
			}

			// Execute family and forward collected errors
			errs := family_runner.New(family).Run(ctx, params.Notifier, params.Store)
			for _, err := range errs {
				params.ErrCh <- err
			}

			return nil
		},
	}
}
