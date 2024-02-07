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

package assetscanestimation

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

// getLatestAssetScanStats - for each family, find the latest AssetScan that has AssetScanStats of this family,
// and add the family stats to the returned aggregated AssetScanStats.
// nolint:cyclop
func (w *Watcher) getLatestAssetScanStats(ctx context.Context, asset *apitypes.Asset) apitypes.AssetScanStats {
	var stats apitypes.AssetScanStats

	filterTmpl := "asset/id eq '%s' and status/state eq 'Done' and (status/%s/errors eq null or length(status/%s/errors) eq 0) and scanFamiliesConfig/%s/enabled eq true"

	families := []string{"exploits", "sbom", "vulnerabilities", "malware", "misconfigurations", "rootkits", "secrets"}
	for _, family := range families {
		params := apitypes.GetAssetScansParams{
			Filter:  to.Ptr(fmt.Sprintf(filterTmpl, *asset.Id, family, family, family)),
			Top:     to.Ptr(1), // get the latest asset scan for this family
			OrderBy: to.Ptr("status/lastTransitionTime DESC"),
		}
		res, err := w.client.GetAssetScans(ctx, params)
		if err != nil {
			logrus.Errorf("Failed to get asset scans for %s. Omitting stats: %v", family, err)
			continue
		}

		if res.Items == nil || len(*res.Items) == 0 || (*res.Items)[0].Stats == nil {
			continue
		}
		assetScan := (*res.Items)[0]

		switch family {
		case "exploits":
			stats.Exploits = assetScan.Stats.Exploits
		case "sbom":
			stats.Sbom = assetScan.Stats.Sbom
		case "vulnerabilities":
			stats.Vulnerabilities = assetScan.Stats.Vulnerabilities
		case "malware":
			stats.Malware = assetScan.Stats.Malware
		case "misconfigurations":
			stats.Misconfigurations = assetScan.Stats.Misconfigurations
		case "rootkits":
			stats.Rootkits = assetScan.Stats.Rootkits
		case "secrets":
			stats.Secrets = assetScan.Stats.Secrets
		}
	}

	return stats
}
