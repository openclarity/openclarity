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

import "github.com/spf13/viper"

const (
	AWSSubnetID          = "AWS_SUBNET_ID"
	AWSJobImageID        = "AWS_JOB_IMAGE_ID"
	AWSSecurityGroupID   = "AWS_SECURITY_GROUP_ID"
	defaultAWSJobImageID = "ami-0568773882d492fc8" // ubuntu server 22.04 LTS (HVM), SSD volume type
)

type Config struct {
	AmiID           string // image id of a scanner job
	SubnetID        string // the scanner's subnet ID
	SecurityGroupID string // the scanner's security group
}

func setConfigDefaults() {
	viper.SetDefault(AWSJobImageID, defaultAWSJobImageID)

	viper.AutomaticEnv()
}

func LoadConfig() *Config {
	setConfigDefaults()

	config := &Config{
		AmiID:           viper.GetString(AWSJobImageID),
		SubnetID:        viper.GetString(AWSSubnetID),
		SecurityGroupID: viper.GetString(AWSSecurityGroupID),
	}

	return config
}
