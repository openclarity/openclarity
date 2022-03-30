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

package rest

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"wwwin-github.cisco.com/eti/scan-gazr/runtime_scan/api/server/models"
	"wwwin-github.cisco.com/eti/scan-gazr/runtime_scan/api/server/restapi"
	"wwwin-github.cisco.com/eti/scan-gazr/runtime_scan/api/server/restapi/operations"
)

type Server struct {
	server  *restapi.Server
	handler ScanResultsHandler
}

type ScanResultsHandler interface {
	HandleScanResults(params operations.PostScanScanUUIDResultsParams) error
	HandleScanContentAnalysis(params operations.PostScanScanUUIDContentAnalysisParams) error
	HandleCISDockerBenchmarkScanResults(params operations.PostScanScanUUIDCisDockerBenchmarkResultsParams) error
}

func CreateRESTServer(port int, handler ScanResultsHandler) (*Server, error) {
	s := &Server{
		handler: handler,
	}

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to load swagger spec: %v", err)
	}

	api := operations.NewKubeClarityRuntimeScanAPIsAPI(swaggerSpec)

	api.PostScanScanUUIDResultsHandler = operations.PostScanScanUUIDResultsHandlerFunc(func(params operations.PostScanScanUUIDResultsParams) middleware.Responder {
		return s.PostScanScanUUIDResults(params)
	})

	api.PostScanScanUUIDContentAnalysisHandler = operations.PostScanScanUUIDContentAnalysisHandlerFunc(func(params operations.PostScanScanUUIDContentAnalysisParams) middleware.Responder {
		return s.PostScanScanUUIDContentAnalysis(params)
	})

	api.PostScanScanUUIDCisDockerBenchmarkResultsHandler = operations.PostScanScanUUIDCisDockerBenchmarkResultsHandlerFunc(func(params operations.PostScanScanUUIDCisDockerBenchmarkResultsParams) middleware.Responder {
		return s.PostScanScanUUIDCISDockerBenchmarkResults(params)
	})

	server := restapi.NewServer(api)

	server.ConfigureFlags()
	server.ConfigureAPI()
	server.Port = port

	s.server = server

	return s, nil
}

func (s *Server) Start(errChan chan struct{}) {
	log.Infof("Starting runtime scan REST server")
	go func() {
		if err := s.server.Serve(); err != nil {
			log.Errorf("Failed to serve runtime scan REST server: %v", err)
			if errChan != nil {
				errChan <- struct{}{}
			}
		}
	}()
	log.Infof("Runtime scan REST server is running")
}

func (s *Server) Stop() {
	log.Infof("Stopping runtime scan REST server")
	if s.server != nil {
		if err := s.server.Shutdown(); err != nil {
			log.Errorf("Failed to shutdown runtime scan REST server: %v", err)
		}
	}
}

func (s *Server) PostScanScanUUIDResults(params operations.PostScanScanUUIDResultsParams) middleware.Responder {
	err := s.handler.HandleScanResults(params)
	if err != nil {
		log.Errorf("Failed to handle scan results: %v", err)
		return operations.NewPostScanScanUUIDResultsDefault(http.StatusInternalServerError).
			WithPayload(&models.APIResponse{
				Message: "Oops",
			})
	}

	return operations.NewPostScanScanUUIDResultsCreated()
}

func (s *Server) PostScanScanUUIDContentAnalysis(params operations.PostScanScanUUIDContentAnalysisParams) middleware.Responder {
	err := s.handler.HandleScanContentAnalysis(params)
	if err != nil {
		log.Errorf("Failed to handle scan content analysis: %v", err)
		return operations.NewPostScanScanUUIDContentAnalysisDefault(http.StatusInternalServerError).
			WithPayload(&models.APIResponse{
				Message: "Oops",
			})
	}

	return operations.NewPostScanScanUUIDContentAnalysisCreated()
}

func (s *Server) PostScanScanUUIDCISDockerBenchmarkResults(params operations.PostScanScanUUIDCisDockerBenchmarkResultsParams) middleware.Responder {
	err := s.handler.HandleCISDockerBenchmarkScanResults(params)
	if err != nil {
		log.Errorf("Failed to handle CIS docker benchmark scan results: %v", err)
		return operations.NewPostScanScanUUIDCisDockerBenchmarkResultsDefault(http.StatusInternalServerError).
			WithPayload(&models.APIResponse{
				Message: "Oops",
			})
	}

	return operations.NewPostScanScanUUIDCisDockerBenchmarkResultsCreated()
}
