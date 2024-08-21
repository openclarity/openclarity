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

package aws

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/pricing"

	apitypes "github.com/openclarity/openclarity/api/types"
	"github.com/openclarity/openclarity/provider"
	"github.com/openclarity/openclarity/provider/aws/discoverer"
	"github.com/openclarity/openclarity/provider/aws/estimator"
	"github.com/openclarity/openclarity/provider/aws/estimator/scanestimation"
	"github.com/openclarity/openclarity/provider/aws/scanner"
)

var _ provider.Provider = &Provider{}

type Provider struct {
	*discoverer.Discoverer
	*scanner.Scanner
	*estimator.Estimator
}

func (p *Provider) Kind() apitypes.CloudProvider {
	return apitypes.AWS
}

func New(ctx context.Context) (provider.Provider, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration. Provider=AWS: %w", err)
	}

	if err = config.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate provider configuration. Provider=AWS: %w", err)
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	// Find architecture specific InstanceType for Scanner instance
	scannerInstanceType, ok := config.ScannerInstanceTypeMapping[config.ScannerInstanceArchitecture]
	if !ok {
		return nil, fmt.Errorf("failed to find instance type for architecture. Arch=%s", config.ScannerInstanceArchitecture)
	}

	// Find architecture specific AMI for Scanner instance
	scannerInstanceAMI, ok := config.ScannerInstanceAMIMapping[config.ScannerInstanceArchitecture]
	if !ok {
		return nil, fmt.Errorf("failed to find AMI for architecture. Arch=%s", config.ScannerInstanceArchitecture)
	}

	return &Provider{
		Discoverer: &discoverer.Discoverer{
			Ec2Client: ec2Client,
		},
		Scanner: &scanner.Scanner{
			Kind:                apitypes.AWS,
			ScannerRegion:       config.ScannerRegion,
			BlockDeviceName:     config.BlockDeviceName,
			ScannerInstanceType: scannerInstanceType,
			ScannerInstanceAMI:  scannerInstanceAMI,
			SecurityGroupID:     config.SecurityGroupID,
			SubnetID:            config.SubnetID,
			KeyPairName:         config.KeyPairName,
			Ec2Client:           ec2Client,
		},
		Estimator: &estimator.Estimator{
			ScannerRegion:       config.ScannerRegion,
			ScannerInstanceType: scannerInstanceType,
			ScanEstimator:       scanestimation.New(pricing.NewFromConfig(cfg), ec2Client),
			Ec2Client:           ec2Client,
		},
	}, nil
}
