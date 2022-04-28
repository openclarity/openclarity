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

package database

import (
	"strconv"
	"time"

	faker "github.com/bxcodec/faker/v3"
	log "github.com/sirupsen/logrus"

	"github.com/cisco-open/kubei/api/server/models"
	"github.com/cisco-open/kubei/backend/pkg/types"
	"github.com/cisco-open/kubei/shared/pkg/utils/slice"
)

var fakeAnalyzers = map[int][]string{
	0: {"syft", "gomod"},
	1: {"syft"},
	2: {"gomod"},
}

func createFakeApplication(scanTime time.Time) *Application {
	var app Application

	if err := faker.FakeData(&app); err != nil {
		panic(err)
	}
	app.ID = app.Name

	for i := 0; i < 2; i++ {
		shouldCreateCISDockerBenchmarkResults := true
		if i == 0 {
			shouldCreateCISDockerBenchmarkResults = false
		}
		app.Resources = append(app.Resources, createFakeResource(scanTime, shouldCreateCISDockerBenchmarkResults))
	}

	return &app
}

func createFakeResource(scanTime time.Time, shouldCreateCISDockerBenchmarkResults bool) Resource {
	var res Resource

	if err := faker.FakeData(&res); err != nil {
		panic(err)
	}
	res.ID = res.Hash

	for i := 0; i < 3; i++ {
		res.Packages = append(res.Packages, createFakePackage(scanTime))
		if shouldCreateCISDockerBenchmarkResults {
			res.CISDockerBenchmarkChecks = append(res.CISDockerBenchmarkChecks, createFakeCISDockerBenchmarkResult())
		}
	}

	return res
}

func createFakeCISDockerBenchmarkResult() CISDockerBenchmarkCheck {
	var res CISDockerBenchmarkCheck
	if err := faker.FakeData(&res); err != nil {
		panic(err)
	}
	res.ID = res.Code
	return res
}

type fakeCvssAttributes struct {
	shouldSetCVSS, cvssShouldMatchVulSeverity bool
}

var fakeCvssAttributesMap = map[int]fakeCvssAttributes{
	0: {
		shouldSetCVSS: false,
	},
	1: {
		shouldSetCVSS:              true,
		cvssShouldMatchVulSeverity: false,
	},
	2: {
		shouldSetCVSS:              true,
		cvssShouldMatchVulSeverity: true,
	},
}

func createFakePackage(scanTime time.Time) Package {
	var pkg Package
	if err := faker.FakeData(&pkg); err != nil {
		panic(err)
	}
	pkg.ID = pkg.Name + "." + pkg.Version
	for i := 0; i < 4; i++ {
		attributes := fakeCvssAttributesMap[i%len(fakeCvssAttributesMap)]
		pkg.Vulnerabilities = append(pkg.Vulnerabilities, createFakeVulnerability(scanTime,
			attributes.shouldSetCVSS, attributes.cvssShouldMatchVulSeverity))
	}

	return pkg
}

func createFakeVulnerability(scanTime time.Time, shouldSetCVSS, cvssShouldMatchVulSeverity bool) Vulnerability {
	var vul Vulnerability
	if err := faker.FakeData(&vul); err != nil {
		panic(err)
	}
	vul.ID = vul.Name
	vul.ScannedAt = scanTime
	if shouldSetCVSS {
		baseScore := 8.8
		if cvssShouldMatchVulSeverity {
			baseScore = createFakeBaseScoreFromVulSeverity(vul.Severity)
		}
		fakeCVSS := createFakeCVSS(baseScore)
		vul.CVSS = CreateCVSSString(fakeCVSS)
		vul.CVSSBaseScore = fakeCVSS.GetBaseScore()
		vul.CVSSSeverity = int(ModelsVulnerabilitySeverityToInt[fakeCVSS.GetCVSSSeverity()])
	}
	return vul
}

// nolint:gomnd
func createFakeBaseScoreFromVulSeverity(severity int) float64 {
	/*
		https://nvd.nist.gov/vuln-metrics/cvss
		CVSS v3.0 Ratings
			Severity	Base Score Range
			None		0.0
			Low			0.1-3.9
			Medium		4.0-6.9
			High		7.0-8.9
			Critical	9.0-10.0
	*/
	switch SeverityIntToString[Severity(severity)] {
	case models.VulnerabilitySeverityNEGLIGIBLE:
		return 0
	case models.VulnerabilitySeverityLOW:
		return 1.1
	case models.VulnerabilitySeverityMEDIUM:
		return 4.5
	case models.VulnerabilitySeverityHIGH:
		return 7.5
	case models.VulnerabilitySeverityCRITICAL:
		return 9.5
	}
	return 0
}

// nolint:gomnd
func createFakeCVSS(baseScore float64) *types.CVSS {
	return &types.CVSS{
		CvssV3Metrics: &types.CVSSV3Metrics{
			BaseScore:           baseScore,
			ExploitabilityScore: 2.2,
			ImpactScore:         3.3,
		},
		CvssV3Vector: &types.CVSSV3Vector{
			AttackComplexity:   types.AttackComplexityHIGH,
			AttackVector:       types.AttackVectorNETWORK,
			Availability:       types.AvailabilityLOW,
			Confidentiality:    types.ConfidentialityHIGH,
			Integrity:          types.IntegrityHIGH,
			PrivilegesRequired: types.PrivilegesRequiredHIGH,
			Scope:              types.ScopeCHANGED,
			UserInteraction:    types.UserInteractionNONE,
			Vector:             "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H",
		},
	}
}

func createFakeVulnerabilityTrend(vulID string, t time.Time) *NewVulnerability {
	var vul NewVulnerability
	if err := faker.FakeData(&vul); err != nil {
		panic(err)
	}
	vul.AddedAt = t
	vul.VulID = vulID
	return &vul
}

func (db *Handler) CreateFakeData() {
	for i := 0; i < 5; i++ {
		app := createFakeApplication(time.Now().Add(time.Duration(i) * time.Second))

		fixVersions := map[PkgVulID]string{}

		for _, resource := range app.Resources {
			for _, p := range resource.Packages {
				for i, vulnerability := range p.Vulnerabilities {
					// half of vuls will have fix version
					if i%2 == 0 {
						fixVersions[CreatePkgVulID(p.ID, vulnerability.ID)] = strconv.Itoa(i)
					}
					// create fake new vulnerabilities
					newVul := createFakeVulnerabilityTrend(vulnerability.ID, vulnerability.ScannedAt)
					if err := db.NewVulnerabilityTable().Create(newVul); err != nil {
						log.Fatalf("failed to create new vulnerability: %v", err)
					}
				}
			}
		}

		analyzers := map[ResourcePkgID][]string{}

		for i, resource := range app.Resources {
			var resourceAnalyzers []string
			for i, p := range resource.Packages {
				fa := fakeAnalyzers[i%3]
				analyzers[CreateResourcePkgID(resource.ID, p.ID)] = fa
				resourceAnalyzers = append(resourceAnalyzers, fa...)
			}
			app.Resources[i].WithAnalyzers(slice.RemoveStringDuplicates(resourceAnalyzers))
		}

		params := &TransactionParams{
			FixVersions: fixVersions,
			Analyzers:   analyzers,
		}

		if err := db.ApplicationTable().Create(app, params); err != nil {
			panic(err.Error())
		}
	}
}
