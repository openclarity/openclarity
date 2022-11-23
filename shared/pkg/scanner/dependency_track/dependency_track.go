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

package dependency_track // nolint:revive,stylecheck

// nolint:staticcheck
import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/shared/pkg/config"
	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner/dependency_track/api/client/client"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner/dependency_track/api/client/client/bom"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner/dependency_track/api/client/client/project"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner/dependency_track/api/client/client/vulnerability"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner/dependency_track/api/client/models"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"
	utilsVul "github.com/openclarity/kubeclarity/shared/pkg/utils/vulnerability"
)

// nolint:gosec
const (
	ScannerName      = "dependency-track"
	apiKeyHeaderName = "X-Api-Key"
)

var referencesRegexp = regexp.MustCompile(`\[([^\[\]]*)]`)

type Scanner struct {
	logger     *log.Entry
	client     *client.DependencyTrackAPI
	config     config.DependencyTrackConfig
	resultChan chan job_manager.Result
}

func (s *Scanner) AuthenticateRequest(request runtime.ClientRequest, _ strfmt.Registry) error {
	if err := request.SetHeaderParam(apiKeyHeaderName, s.config.APIKey); err != nil {
		return fmt.Errorf("failed to set API key header: %v", err)
	}
	return nil
}

func New(c job_manager.IsConfig, logger *log.Entry, resultChan chan job_manager.Result) job_manager.Job {
	conf := c.(*config.Config) // nolint:forcetypeassert
	config := config.ConvertToDependencyTrackConfig(conf.Scanner, logger)
	return &Scanner{
		logger:     logger.Dup().WithField("scanner", ScannerName),
		config:     config,
		client:     newHTTPClient(config),
		resultChan: resultChan,
	}
}

func newHTTPClient(conf config.DependencyTrackConfig) *client.DependencyTrackAPI {
	var transport *httptransport.Runtime
	if conf.DisableTLS {
		transport = httptransport.New(conf.Host, client.DefaultBasePath, []string{"http"})
	} else if conf.InsecureSkipVerify {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()      // nolint:forcetypeassert
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // nolint: gosec
		transport = httptransport.NewWithClient(conf.Host, client.DefaultBasePath, []string{"https"},
			&http.Client{Transport: customTransport})
	} else {
		transport = httptransport.New(conf.Host, client.DefaultBasePath, []string{"https"})
	}
	apiClient := client.New(transport, strfmt.Default)
	return apiClient
}

func (s *Scanner) Run(sourceType utils.SourceType, source string) error {
	if sourceType != utils.SBOM {
		s.logger.Infof("Ignoring non SBOM input. type=%v", sourceType)
		s.resultChan <- &scanner.Results{
			Matches: nil, // empty results
			ScannerInfo: scanner.Info{
				Name: ScannerName,
			},
		}
		return nil
	}
	go s.run(source)

	return nil
}

func (s *Scanner) run(source string) {
	projectName := s.getProjectName()
	projectVersion := s.getProjectVersion()

	sbom, err := ioutil.ReadFile(source)
	if err != nil {
		s.reportError(fmt.Errorf("failed to read SBOM from file (%s): %w", source, err))
		return
	}

	sbomS := string(sbom)

	// create a project
	createProjectParams := project.NewCreateProjectParams().WithBody(&models.Project{
		Active:  true,
		Name:    projectName,
		Version: projectVersion,
	})

	s.logger.Infof("Creating Project. name=%v, version=%v", projectName, projectVersion)

	var projectUUID string
	projectCreated, err := s.client.Project.CreateProject(createProjectParams, s)
	if err != nil {
		// nolint:errorlint
		switch err.(type) {
		case *project.CreateProjectConflict:
			s.logger.Infof("Project already exists.")
			getProject, err := s.client.Project.GetProject1(project.NewGetProject1Params().WithName(projectName).WithVersion(projectVersion), s)
			if err != nil {
				s.reportError(fmt.Errorf("failed to get project: %w", err))
				return
			}
			s.logger.Debugf("Project already exists. payload=%+v", getProject.GetPayload())
			projectUUID = string(*getProject.GetPayload().UUID)
		default:
			s.reportError(fmt.Errorf("failed to create project: %w", err))
			return
		}
	} else {
		projectUUID = string(*projectCreated.GetPayload().UUID)
		s.logger.Infof("Project created. name=%v, version=%v, uuid=%v", projectName, projectVersion, projectUUID)
		s.logger.Debugf("Project created. payload=%+v", projectCreated.GetPayload())
	}

	defer func() {
		if s.config.ShouldDeleteProject {
			s.logger.Infof("Deleting project. project uuid=%v", projectUUID)
			if err := s.client.Project.DeleteProject(project.NewDeleteProjectParams().WithUUID(projectUUID), s); err != nil {
				s.logger.Warnf("Failed to delete project. uuid=%v: %v", projectUUID, err)
			}
		}
	}()

	uploadBomParams := bom.NewUploadBomParams().
		WithProject(&projectUUID).
		WithBom(&sbomS)
	if _, err := s.client.Bom.UploadBom(uploadBomParams, s); err != nil {
		s.reportError(fmt.Errorf("failed to upload bom: %w", err))
		return
	}

	s.logger.Infof("SBOM was uploaded successfully. project uuid=%v", projectUUID)

	vulnerabilitiesByProject, err := s.getVulnerabilitiesWithRetry(projectUUID)
	if err != nil {
		s.reportError(fmt.Errorf("failed to get vulnerabilities result: %w", err))
		return
	}

	s.logger.Infof("Vulnerabilities by project was found. project uuid=%v, total vulnerabilities=%v",
		projectUUID, vulnerabilitiesByProject.XTotalCount)

	s.logger.Infof("Sending successful results")
	s.resultChan <- s.createResults(vulnerabilitiesByProject.GetPayload())
}

func (s *Scanner) getProjectVersion() string {
	projectVersion := s.config.ProjectVersion
	if projectVersion == "" {
		projectVersion = uuid.NewV4().String()
	}
	return projectVersion
}

func (s *Scanner) getProjectName() string {
	projectName := s.config.ProjectName
	if projectName == "" {
		projectName = uuid.NewV4().String()
	}
	return projectName
}

func (s *Scanner) getVulnerabilitiesWithRetry(projectUUID string) (*vulnerability.GetVulnerabilitiesByProjectOK, error) {
	var err error
	var vulnerabilitiesByProject *vulnerability.GetVulnerabilitiesByProjectOK

	vulnerabilitiesByProjectParams := vulnerability.NewGetVulnerabilitiesByProjectParams().WithUUID(projectUUID)

	foundVulnerabilitiesTotalCount := int64(0)
	s.logger.Infof("Trying to get vulnerabilities result. project uuid=%v", projectUUID)
	for i := 0; i < s.config.FetchVulnerabilitiesRetryCount; i++ {
		time.Sleep(s.config.FetchVulnerabilitiesRetrySleep)

		vulnerabilitiesByProject, err = s.client.Vulnerability.GetVulnerabilitiesByProject(vulnerabilitiesByProjectParams, s)
		if err != nil {
			s.logger.Infof("failed to get vulnerabilities by project: %v", err)
		} else if vulnerabilitiesByProject.XTotalCount == 0 {
			s.logger.Infof("Empty vulnerabilities results - retrying...")
		} else if foundVulnerabilitiesTotalCount != vulnerabilitiesByProject.XTotalCount {
			foundVulnerabilitiesTotalCount = vulnerabilitiesByProject.XTotalCount
			s.logger.Infof("Got %d vulnerabilities - verifying that scanning was done...", foundVulnerabilitiesTotalCount)
		} else if foundVulnerabilitiesTotalCount == vulnerabilitiesByProject.XTotalCount {
			// we got multiple time the same total count - we can assume scanning was done
			return vulnerabilitiesByProject, nil
		}
	}

	return nil, fmt.Errorf("failed to get vulnerabilities by project: %w", err)
}

func (s *Scanner) reportError(err error) {
	res := &scanner.Results{
		Error: err,
	}

	s.logger.Error(res.Error)
	s.resultChan <- res
}

func (s *Scanner) createResults(vulnerabilities []*models.Vulnerability) *scanner.Results {
	// TODO:
	// distro := getDistro(doc)
	matches := make([]scanner.Match, len(vulnerabilities))
	for i := range vulnerabilities {
		vul := vulnerabilities[i]
		for j := range vul.Components {
			component := vul.Components[j]

			matches[i] = scanner.Match{
				Vulnerability: scanner.Vulnerability{
					ID:          vul.VulnID,
					Description: vul.Description,
					Links:       getLinks(vul.References),
					// Distro:      distro, // missing - can we get it from purl?
					CVSS:     getCVSS(vul),
					Fix:      getFix(vul),
					Severity: strings.ToUpper(vul.Severity),
					Package: scanner.Package{
						Name:    component.Name,
						Version: getVersion(component),
						// TODO: can take it from purl
						// nginx image example "pkg:deb/debian/apt@1.8.2.3?arch=amd64"
						// go folder example "pkg:golang/github.com/dgrijalva/jwt-go@v3.2.0%20incompatible"
						// Type:     string(vul.Artifact.Type),
						// Language: string(vul.Artifact.Language),
						Licenses: getLicense(component),
						CPEs:     getCPE(component),
						PURL:     unescapePurl(component.Purl),
					},
					LayerID: "", // TODO: missing
					Path:    "", // TODO: missing
				},
			}
		}
	}

	return &scanner.Results{
		Matches: matches,
		ScannerInfo: scanner.Info{
			Name: ScannerName,
		},
		Source: scanner.Source{
			// The Type and Name of Source is populated after the merge,
			// because the input of dependency-track is an SBOM which contains the values of them.
			Type: "",
			Name: "",
		},
	}
}

func unescapePurl(purl string) string {
	// We need to unescape purl since dependency-track returned escaped "+" in the purl
	//	example: pkg:deb/debian/tar@1.30%20dfsg-6?arch=amd64
	// 	expected: pkg:deb/debian/tar@1.30+dfsg-6?arch=amd64
	// NOTE: Decided to use replace since both url.PathUnescape() and url.QueryUnescape()
	//       didn't unescape it as expected, got space instead of "+"
	return strings.ReplaceAll(purl, "%20", "+")
}

func getVersion(component *models.Component) string {
	if component.Version == nil {
		return ""
	}
	return *component.Version
}

func getCPE(component *models.Component) []string {
	ret := make([]string, 0)
	if component.Cpe != nil {
		ret = append(ret, *component.Cpe)
	}
	return ret
}

func getLicense(component *models.Component) []string {
	ret := make([]string, 0)
	if component.License != nil {
		ret = append(ret, *component.License)
	}
	return ret
}

func getFix(vul *models.Vulnerability) scanner.Fix {
	fix := scanner.Fix{
		State: "not-fixed",
	}

	if vul.PatchedVersions != "" {
		fix.Versions = append(fix.Versions, vul.PatchedVersions)
		fix.State = "fixed"
	}

	return fix
}

func getLinks(references string) []string {
	links := make([]string, 0)
	subMatchAll := referencesRegexp.FindAllString(references, -1)
	for _, link := range subMatchAll {
		link = strings.Trim(link, "[")
		link = strings.Trim(link, "]")
		links = append(links, link)
	}
	return links
}

func getCVSS(vul *models.Vulnerability) []scanner.CVSS {
	var ret []scanner.CVSS

	if vul.CvssV2Vector != "" {
		ret = append(ret, scanner.CVSS{
			Version: "2.0",
			Vector:  removeParentheses(vul.CvssV2Vector),
			Metrics: scanner.CvssMetrics{
				BaseScore:           vul.CvssV2BaseScore,
				ExploitabilityScore: &vul.CvssV2ExploitabilitySubScore,
				ImpactScore:         &vul.CvssV2ImpactSubScore,
			},
		})
	}
	if vul.CvssV3Vector != "" {
		ret = append(ret, scanner.CVSS{
			Version: utilsVul.GetCVSSV3VersionFromVector(vul.CvssV3Vector),
			Vector:  vul.CvssV3Vector,
			Metrics: scanner.CvssMetrics{
				BaseScore:           vul.CvssV3BaseScore,
				ExploitabilityScore: &vul.CvssV3ExploitabilitySubScore,
				ImpactScore:         &vul.CvssV3ImpactSubScore,
			},
		})
	}

	return ret
}

func removeParentheses(vector string) string {
	return strings.TrimSuffix(strings.TrimPrefix(vector, "("), ")")
}
