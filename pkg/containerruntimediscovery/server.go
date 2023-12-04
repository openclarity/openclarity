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

package containerruntimediscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/log"
)

type ListImagesResponse struct {
	Images []models.ContainerImageInfo
}

type ListContainersResponse struct {
	Containers []models.ContainerInfo
}

type ContainerRuntimeDiscoveryServer struct {
	server *http.Server

	discoverer Discoverer
}

func NewContainerRuntimeDiscoveryServer(listenAddr string, discoverer Discoverer) *ContainerRuntimeDiscoveryServer {
	crds := &ContainerRuntimeDiscoveryServer{
		discoverer: discoverer,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/images", crds.ListImages)
	mux.HandleFunc("/containers", crds.ListContainers)

	crds.server = &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,  // nolint:gomnd
		IdleTimeout:       30 * time.Second, // nolint:gomnd
	}

	return crds
}

func (crds *ContainerRuntimeDiscoveryServer) Serve(ctx context.Context) {
	logger := log.GetLoggerFromContextOrDefault(ctx)
	go func() {
		if err := crds.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("image resolver server error: %v", err)
		}
	}()
}

func (crds *ContainerRuntimeDiscoveryServer) Shutdown(ctx context.Context) error {
	err := crds.server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}
	return nil
}

func (crds *ContainerRuntimeDiscoveryServer) ListImages(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/images" {
		http.NotFound(w, req)
		return
	}

	if req.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf("400 unsupported method %v", req.Method), http.StatusBadRequest)
		return
	}

	images, err := crds.discoverer.Images(req.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to discover images: %v", err), http.StatusInternalServerError)
		return
	}

	response := ListImagesResponse{
		Images: images,
	}
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (crds *ContainerRuntimeDiscoveryServer) ListContainers(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/containers" {
		http.NotFound(w, req)
		return
	}

	if req.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf("400 unsupported method %v", req.Method), http.StatusBadRequest)
		return
	}

	containers, err := crds.discoverer.Containers(req.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to discover containers: %v", err), http.StatusInternalServerError)
		return
	}

	response := ListContainersResponse{
		Containers: containers,
	}
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}
