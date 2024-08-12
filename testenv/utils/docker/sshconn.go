// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package docker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/openclarity/vmclarity/testenv/utils"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/client"
)

func ClientOptsWithSSHConn(_ context.Context, workDir string, keys *utils.SSHKeyPair, input *utils.SSHForwardInput) ([]client.Opt, error) {
	privateKeyFile := filepath.Join(workDir, "id_rsa")
	publicKeyFile := filepath.Join(workDir, "id_rsa.pub")
	if _, err := os.Stat(privateKeyFile); errors.Is(err, os.ErrNotExist) {
		if err := keys.Save(privateKeyFile, publicKeyFile); err != nil {
			return nil, fmt.Errorf("failed to save SSH keys to filesystem: %w", err)
		}
	}

	helper, err := connhelper.GetConnectionHelperWithSSHOpts(
		"ssh://"+input.User+"@"+input.HostAddressPort(),
		// Automatically add host key to known_hosts file
		[]string{
			"-o", "StrictHostKeyChecking=no",
			"-o", "ForwardAgent=no",
			"-i", privateKeyFile,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection helper: %w", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: helper.Dialer,
		},
	}

	var clientOpts []client.Opt
	clientOpts = append(clientOpts,
		client.WithHTTPClient(httpClient),
		client.WithHost(helper.Host),
		client.WithDialContext(helper.Dialer),
		client.WithAPIVersionNegotiation(),
	)

	return clientOpts, nil
}
