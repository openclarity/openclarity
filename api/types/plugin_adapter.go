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

package types

import (
	"github.com/openclarity/vmclarity/core/to"
	plugintypes "github.com/openclarity/vmclarity/plugins/sdk-go/types"
)

// DefaultPluginAdapter is used to convert latest version Plugin API models to VMClarity.
var DefaultPluginAdapter PluginAdapter = &pluginAdapter{}

// PluginAdapter is responsible for converting Plugin security findings to
// low-level VMClarity findings.
type PluginAdapter interface {
	Result(data plugintypes.Result) ([]FindingInfo, error)
	Exploit(data plugintypes.Exploit) (*ExploitFindingInfo, error)
	InfoFinder(data plugintypes.InfoFinder) (*InfoFinderFindingInfo, error)
	Malware(data plugintypes.Malware) (*MalwareFindingInfo, error)
	Misconfiguration(data plugintypes.Misconfiguration) (*MisconfigurationFindingInfo, error)
	Package(data plugintypes.Package) (*PackageFindingInfo, error)
	Rootkit(data plugintypes.Rootkit) (*RootkitFindingInfo, error)
	Secret(data plugintypes.Secret) (*SecretFindingInfo, error)
	Vulnerability(data plugintypes.Vulnerability) (*VulnerabilityFindingInfo, error)
}

type pluginAdapter struct{}

//nolint:gocognit,cyclop
func (p pluginAdapter) Result(data plugintypes.Result) ([]FindingInfo, error) {
	var findings []FindingInfo

	// Convert exploits
	if exploits := data.Vmclarity.Exploits; exploits != nil {
		for _, exploit := range *exploits {
			exploit, err := p.Exploit(exploit)
			if err != nil {
				return nil, err
			}
			if exploit == nil {
				continue
			}

			var finding FindingInfo
			_ = finding.FromExploitFindingInfo(*exploit)
			findings = append(findings, finding)
		}
	}

	// Convert info finders
	if infoFinders := data.Vmclarity.InfoFinder; infoFinders != nil {
		for _, infoFinder := range *infoFinders {
			infoFinder, err := p.InfoFinder(infoFinder)
			if err != nil {
				return nil, err
			}
			if infoFinder == nil {
				continue
			}

			var finding FindingInfo
			_ = finding.FromInfoFinderFindingInfo(*infoFinder)
			findings = append(findings, finding)
		}
	}

	// Convert malwares
	if malwares := data.Vmclarity.Malware; malwares != nil {
		for _, malware := range *malwares {
			malware, err := p.Malware(malware)
			if err != nil {
				return nil, err
			}
			if malware == nil {
				continue
			}

			var finding FindingInfo
			_ = finding.FromMalwareFindingInfo(*malware)
			findings = append(findings, finding)
		}
	}

	// Convert misconfigurations
	if misconfigurations := data.Vmclarity.Misconfigurations; misconfigurations != nil {
		for _, misconfiguration := range *misconfigurations {
			misconfiguration, err := p.Misconfiguration(misconfiguration)
			if err != nil {
				return nil, err
			}
			if misconfiguration == nil {
				continue
			}

			var finding FindingInfo
			_ = finding.FromMisconfigurationFindingInfo(*misconfiguration)
			findings = append(findings, finding)
		}
	}

	// Convert packages
	if packages := data.Vmclarity.Packages; packages != nil {
		for _, pkg := range *packages {
			pkg, err := p.Package(pkg)
			if err != nil {
				return nil, err
			}
			if pkg == nil {
				continue
			}

			var finding FindingInfo
			_ = finding.FromPackageFindingInfo(*pkg)
			findings = append(findings, finding)
		}
	}

	// Convert rootkits
	if rootkits := data.Vmclarity.Rootkits; rootkits != nil {
		for _, rootkit := range *rootkits {
			rootkit, err := p.Rootkit(rootkit)
			if err != nil {
				return nil, err
			}
			if rootkit == nil {
				continue
			}

			var finding FindingInfo
			_ = finding.FromRootkitFindingInfo(*rootkit)
			findings = append(findings, finding)
		}
	}

	// Convert secrets
	if secrets := data.Vmclarity.Secrets; secrets != nil {
		for _, secret := range *secrets {
			secret, err := p.Secret(secret)
			if err != nil {
				return nil, err
			}
			if secret == nil {
				continue
			}

			var finding FindingInfo
			_ = finding.FromSecretFindingInfo(*secret)
			findings = append(findings, finding)
		}
	}

	return findings, nil
}

func (p pluginAdapter) Exploit(data plugintypes.Exploit) (*ExploitFindingInfo, error) {
	return &ExploitFindingInfo{
		CveID:       data.CveID,
		Description: data.Description,
		Name:        data.Name,
		SourceDB:    data.SourceDB,
		Title:       data.Title,
		Urls:        data.Urls,
	}, nil
}

func (p pluginAdapter) InfoFinder(data plugintypes.InfoFinder) (*InfoFinderFindingInfo, error) {
	typeMapping := map[plugintypes.InfoFinderType]InfoType{
		plugintypes.InfoFinderTypeSSHAuthorizedKeyFingerprint: InfoTypeSSHAuthorizedKeyFingerprint,
		plugintypes.InfoFinderTypeSSHDaemonKeyFingerprint:     InfoTypeSSHDaemonKeyFingerprint,
		plugintypes.InfoFinderTypeSSHKnownHostFingerprint:     InfoTypeSSHKnownHostFingerprint,
		plugintypes.InfoFinderTypeSSHPrivateKeyFingerprint:    InfoTypeSSHPrivateKeyFingerprint,
		plugintypes.InfoFinderTypeUnknown:                     InfoTypeUNKNOWN,
	}

	tp := InfoTypeUNKNOWN
	if data.Type != nil {
		if t, ok := typeMapping[*data.Type]; ok {
			tp = t
		}
	}

	return &InfoFinderFindingInfo{
		Data:        data.Data,
		Path:        data.Path,
		ScannerName: to.Ptr(""),
		Type:        &tp,
	}, nil
}

func (p pluginAdapter) Malware(data plugintypes.Malware) (*MalwareFindingInfo, error) {
	return &MalwareFindingInfo{
		MalwareName: data.MalwareName,
		MalwareType: data.MalwareType,
		Path:        data.Path,
		RuleName:    data.RuleName,
	}, nil
}

func (p pluginAdapter) Misconfiguration(data plugintypes.Misconfiguration) (*MisconfigurationFindingInfo, error) {
	severityMapping := map[plugintypes.MisconfigurationSeverity]MisconfigurationSeverity{
		plugintypes.MisconfigurationSeverityHigh:   MisconfigurationHighSeverity,
		plugintypes.MisconfigurationSeverityMedium: MisconfigurationMediumSeverity,
		plugintypes.MisconfigurationSeverityLow:    MisconfigurationLowSeverity,
		plugintypes.MisconfigurationSeverityInfo:   MisconfigurationInfoSeverity,
	}

	severity := MisconfigurationInfoSeverity
	if data.Severity != nil {
		if s, ok := severityMapping[*data.Severity]; ok {
			severity = s
		}
	}

	return &MisconfigurationFindingInfo{
		Category:    data.Category,
		Description: data.Description,
		Id:          data.Id,
		Location:    data.Location,
		Message:     data.Message,
		Remediation: data.Remediation,
		// TODO(ramizpolic): Remove ScannerName property from Misconfiguration API.
		// TODO(ramizpolic): This data is available on higher Finding object.
		ScannerName: to.Ptr(""),
		Severity:    &severity,
	}, nil
}

func (p pluginAdapter) Package(data plugintypes.Package) (*PackageFindingInfo, error) {
	return &PackageFindingInfo{
		Cpes:     data.Cpes,
		Language: data.Language,
		Licenses: data.Licenses,
		Name:     data.Name,
		Purl:     data.Purl,
		Type:     data.Type,
		Version:  data.Version,
	}, nil
}

func (p pluginAdapter) Rootkit(data plugintypes.Rootkit) (*RootkitFindingInfo, error) {
	typeMapping := map[plugintypes.RootkitType]RootkitType{
		plugintypes.RootkitTypeApplication: RootkitTypeAPPLICATION,
		plugintypes.RootkitTypeFirmware:    RootkitTypeFIRMWARE,
		plugintypes.RootkitTypeKernel:      RootkitTypeKERNEL,
		plugintypes.RootkitTypeMemory:      RootkitTypeMEMORY,
		plugintypes.RootkitTypeUnknown:     RootkitTypeUNKNOWN,
	}

	tp := RootkitTypeUNKNOWN
	if data.RootkitType != nil {
		if t, ok := typeMapping[*data.RootkitType]; ok {
			tp = t
		}
	}

	return &RootkitFindingInfo{
		Message:     data.Message,
		RootkitName: data.RootkitName,
		RootkitType: &tp,
	}, nil
}

func (p pluginAdapter) Secret(data plugintypes.Secret) (*SecretFindingInfo, error) {
	return &SecretFindingInfo{
		Description: data.Description,
		EndColumn:   data.EndColumn,
		EndLine:     data.EndLine,
		FilePath:    data.FilePath,
		Fingerprint: data.Fingerprint,
		StartColumn: data.StartColumn,
		StartLine:   data.StartLine,
	}, nil
}

func (p pluginAdapter) Vulnerability(data plugintypes.Vulnerability) (*VulnerabilityFindingInfo, error) {
	cvss := []VulnerabilityCvss{}
	if data.Cvss != nil {
		for _, c := range *data.Cvss {
			cvss = append(cvss, VulnerabilityCvss{
				Metrics: &VulnerabilityCvssMetrics{
					BaseScore:           c.BaseScore,
					ExploitabilityScore: c.ExploitabilityScore,
					ImpactScore:         c.ImpactScore,
				},
				Vector:  c.Vector,
				Version: c.Version,
			})
		}
	}

	distro := &VulnerabilityDistro{}
	if data.Distro != nil {
		distro.IDLike = data.Distro.IDLike
		distro.Name = data.Distro.Name
		distro.Version = data.Distro.Version
	}

	fix := &VulnerabilityFix{}
	if data.Fix != nil {
		fix.State = data.Fix.State
		fix.Versions = data.Fix.Versions
	}

	pkg := &Package{}
	if data.Package == nil {
		pkg.Cpes = data.Package.Cpes
		pkg.Language = data.Package.Language
		pkg.Licenses = data.Package.Licenses
		pkg.Name = data.Package.Name
		pkg.Purl = data.Package.Purl
		pkg.Type = data.Package.Type
		pkg.Version = data.Package.Version
	}

	severityMapping := map[plugintypes.VulnerabilitySeverity]VulnerabilitySeverity{
		plugintypes.VulnerabilitySeverityCritical:   CRITICAL,
		plugintypes.VulnerabilitySeverityHigh:       HIGH,
		plugintypes.VulnerabilitySeverityLow:        LOW,
		plugintypes.VulnerabilitySeverityMedium:     MEDIUM,
		plugintypes.VulnerabilitySeverityNegligible: NEGLIGIBLE,
	}

	severity := NEGLIGIBLE
	if data.Severity != nil {
		if s, ok := severityMapping[*data.Severity]; ok {
			severity = s
		}
	}

	return &VulnerabilityFindingInfo{
		Cvss:              &cvss,
		Description:       data.Description,
		Distro:            distro,
		Fix:               fix,
		LayerId:           data.LayerId,
		Links:             data.Links,
		Package:           pkg,
		Path:              data.Path,
		Severity:          &severity,
		VulnerabilityName: data.VulnerabilityName,
	}, nil
}
