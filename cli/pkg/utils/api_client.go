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

package utils

import (
	"crypto/tls"
	"net/http"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/cisco-open/kubei/api/client/client"
	"github.com/cisco-open/kubei/cli/pkg/config"
)

func NewHTTPClient(conf *config.Backend) *client.KubeClarityAPIs {
	var transport *httptransport.Runtime
	if conf.DisableTLS {
		transport = httptransport.New(conf.Host, client.DefaultBasePath, []string{"http"})
	} else if conf.InsecureSkipVerify {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // nolint: gosec
		transport = httptransport.NewWithClient(conf.Host, client.DefaultBasePath, []string{"https"},
			&http.Client{Transport: customTransport})
	} else {
		transport = httptransport.New(conf.Host, client.DefaultBasePath, []string{"https"})
	}
	apiClient := client.New(transport, strfmt.Default)
	return apiClient
}
