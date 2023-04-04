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

package rest

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
)

const (
	backgroundRecalculationInterval = 15 * time.Minute
)

type ServerImpl struct {
	BackendClient *backendclient.BackendClient
	findingsImpactData
}

func CreateUIBackedServer(client *backendclient.BackendClient) *ServerImpl {
	return &ServerImpl{
		BackendClient: client,
		findingsImpactData: findingsImpactData{
			findingsImpactFetchedChannel: make(chan struct{}),
		},
	}
}

func (s *ServerImpl) StartBackgroundProcessing(ctx context.Context) {
	go func() {
		s.runBackgroundRecalculation(ctx)
		for {
			select {
			case <-time.After(backgroundRecalculationInterval):
				s.runBackgroundRecalculation(ctx)
			case <-ctx.Done():
				log.Infof("Stop background recalculation")
				return
			}
		}
	}()
}

func (s *ServerImpl) runBackgroundRecalculation(ctx context.Context) {
	log.Infof("Background recalculation started...")
	s.recalculateFindingsImpact(ctx)
	log.Infof("Background recalculation ended...")
}
