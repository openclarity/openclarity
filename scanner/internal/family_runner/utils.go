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

package family_runner // nolint:revive,stylecheck

import (
	"github.com/openclarity/vmclarity/scanner/families"
	exploits "github.com/openclarity/vmclarity/scanner/families/exploits/types"
	infofinder "github.com/openclarity/vmclarity/scanner/families/infofinder/types"
	malware "github.com/openclarity/vmclarity/scanner/families/malware/types"
	misconfiguration "github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	plugins "github.com/openclarity/vmclarity/scanner/families/plugins/types"
	rootkits "github.com/openclarity/vmclarity/scanner/families/rootkits/types"
	sbom "github.com/openclarity/vmclarity/scanner/families/sbom/types"
	secrets "github.com/openclarity/vmclarity/scanner/families/secrets/types"
	vulnerabilities "github.com/openclarity/vmclarity/scanner/families/vulnerabilities/types"
)

func getFamilyScanMetadata(result any) *families.ScanMetadata {
	switch r := result.(type) {
	case *exploits.Result:
		if r != nil {
			return &r.Metadata
		}
	case *infofinder.Result:
		if r != nil {
			return &r.Metadata
		}
	case *malware.Result:
		if r != nil {
			return &r.Metadata
		}
	case *misconfiguration.Result:
		if r != nil {
			return &r.Metadata
		}
	case *plugins.Result:
		if r != nil {
			return &r.Metadata
		}
	case *rootkits.Result:
		if r != nil {
			return &r.Metadata
		}
	case *sbom.Result:
		if r != nil {
			return &r.Metadata
		}
	case *secrets.Result:
		if r != nil {
			return &r.Metadata
		}
	case *vulnerabilities.Result:
		if r != nil {
			return &r.Metadata
		}
	}

	return nil
}
