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

package main

import (
	"context"
	"flag"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	provider_service "github.com/openclarity/vmclarity/provider/external/utils/proto"
)

type Provider struct {
	provider_service.UnimplementedProviderServer
	ReadyToScan bool
}

func (p *Provider) DiscoverAssets(_ context.Context, _ *provider_service.DiscoverAssetsParams) (*provider_service.DiscoverAssetsResult, error) {
	return &provider_service.DiscoverAssetsResult{
		Assets: []*provider_service.Asset{
			{
				AssetType: &provider_service.Asset_Vminfo{
					Vminfo: &provider_service.VMInfo{
						Id:           "Id",
						Location:     "Location",
						Image:        "Image",
						InstanceType: "InstanceType",
						Platform:     "Platform",
						Tags: []*provider_service.Tag{
							{
								Key: "key",
								Val: "val",
							},
						},
						LaunchTime: timestamppb.New(time.Now()),
					},
				},
			},
		},
	}, nil
}

func (p *Provider) RemoveAssetScan(context.Context, *provider_service.RemoveAssetScanParams) (*provider_service.RemoveAssetScanResult, error) {
	// clean all resources that were created during RunAssetScan
	return &provider_service.RemoveAssetScanResult{}, nil
}

const retryAfterSec = 30

func (p *Provider) RunAssetScan(context.Context, *provider_service.RunAssetScanParams) (*provider_service.RunAssetScanResult, error) {
	// Create all resources needed in order to start the scan.
	// It can be spinning up a VM or snapshotting a volume.
	// It should be non-blocking and idempotent.

	if !p.ReadyToScan {
		// flip the ready to scan bit, so next time we're getting called, ERR_NONE will be returned.
		p.ReadyToScan = true

		// Tell VMClarity that resources are not ready and RunAssetScan should be called again in 30 seconds from now.
		return &provider_service.RunAssetScanResult{
			Err: &provider_service.Error{ErrorType: &provider_service.Error_ErrRetry{
				ErrRetry: &provider_service.ErrorRetryable{
					Err:   "not all resources are ready for scanning",
					After: retryAfterSec,
				},
			}},
		}, nil
	}

	// In case there was some fatal error that can not be recovered, a fatal error should be returned:
	//return &provider_service.RunAssetScanResult{
	//	Err: &provider_service.Error{ErrorType: &provider_service.Error_ErrFatal{
	//		ErrFatal: &provider_service.ErrorFatal{
	//			Err: "failed to scan due to fatal error",
	//		},
	//	}},
	//}, nil

	// when all the resource creation is done and ready, an error type of ErrorType_ERR_NONE should be return.
	return &provider_service.RunAssetScanResult{
		Err: &provider_service.Error{ErrorType: &provider_service.Error_ErrNone{
			ErrNone: &provider_service.ErrorNone{},
		}},
	}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", "localhost:24230")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	provider_service.RegisterProviderServer(grpcServer, &Provider{})
	log.Infof("listening.....")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
