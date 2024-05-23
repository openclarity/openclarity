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

package utils

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

func EC2FiltersFromEC2Tags(tags []ec2types.Tag) []ec2types.Filter {
	filters := make([]ec2types.Filter, 0, len(tags))
	for _, tag := range tags {
		var name string
		var value []string

		if tag.Key == nil || *tag.Key == "" {
			continue
		}
		name = "tag:" + *tag.Key

		if tag.Value != nil {
			value = []string{*tag.Value}
		}

		filters = append(filters, ec2types.Filter{
			Name:   to.Ptr(name),
			Values: value,
		})
	}

	return filters
}

func EC2FiltersFromTags(tags []apitypes.Tag) []ec2types.Filter {
	filters := make([]ec2types.Filter, 0, len(tags))
	for _, tag := range tags {
		var name string
		var value []string

		if tag.Key == "" {
			continue
		}

		name = "tag:" + tag.Key
		if tag.Value != "" {
			value = []string{tag.Value}
		}

		filters = append(filters, ec2types.Filter{
			Name:   to.Ptr(name),
			Values: value,
		})
	}

	return filters
}

func GetTagsFromECTags(tags []ec2types.Tag) []apitypes.Tag {
	if len(tags) == 0 {
		return nil
	}

	ret := make([]apitypes.Tag, len(tags))
	for i, tag := range tags {
		ret[i] = apitypes.Tag{
			Key:   *tag.Key,
			Value: *tag.Value,
		}
	}
	return ret
}

func GetInstanceState(result *ec2.DescribeInstancesOutput, instanceID string) ec2types.InstanceStateName {
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if strings.Compare(*instance.InstanceId, instanceID) == 0 {
				if instance.State != nil {
					return instance.State.Name
				}
			}
		}
	}
	return ec2types.InstanceStateNamePending
}
