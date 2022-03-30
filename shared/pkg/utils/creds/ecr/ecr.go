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

package ecr

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/containers/image/v5/docker/reference"
)

const ecrURL = "amazonaws.com"

type ECR struct{}

func (e *ECR) Name() string {
	return "ecr"
}

func (e *ECR) IsSupported(named reference.Named) bool {
	return strings.HasSuffix(reference.Domain(named), ecrURL)
}

func (e *ECR) GetCredentials(ctx context.Context, _ reference.Named) (username, password string, err error) {
	input := &ecr.GetAuthorizationTokenInput{}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	client := ecr.New(sess)

	result, err := client.GetAuthorizationTokenWithContext(ctx, input)
	if err != nil {
		return "", "", fmt.Errorf("failed to get authorization token: %w", err)
	}

	for _, data := range result.AuthorizationData {
		b, err := base64.StdEncoding.DecodeString(*data.AuthorizationToken)
		if err != nil {
			return "", "", fmt.Errorf("base64 decode failed: %w", err)
		}
		// e.g. AWS:eyJwYXlsb2...
		split := strings.SplitN(string(b), ":", 2) // nolint:gomnd
		if len(split) == 2 {                       // nolint:gomnd
			return split[0], split[1], nil
		}
	}

	return "", "", nil
}
