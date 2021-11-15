package webapp

import (
	"fmt"
	dockle_config "github.com/Portshift/dockle/config"
	dockle_writer "github.com/Portshift/dockle/pkg/report"
	dockle_types "github.com/Portshift/dockle/pkg/types"
	"github.com/Portshift/kubei/pkg/config"
	"github.com/Portshift/kubei/pkg/orchestrator"
	"github.com/Portshift/kubei/pkg/types"
	k8s_utils "github.com/Portshift/kubei/pkg/utils/k8s"
	grype_models "github.com/anchore/grype/grype/presenter/models"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const htmlFileName = "view.html"
const htmlPath = "/app/" + htmlFileName

type Webapp struct {
	orchestrator         *orchestrator.Orchestrator
	executionConfig      *config.Config
	scanConfig           *config.ScanConfig
	showGoMsg            bool
	showGoWarning        bool // Warning that previous results will be lost
	checkShowGoWarning   bool
	lastScannedNamespace string
	template             *template.Template
	sync.Mutex
}

type containerInfo struct {
	Container string `json:"container"`
	Image     string `json:"image"`
	Pod       string `json:"pod"`
	Namespace string `json:"namespace"`
	Succeeded bool   `json:"succeeded"`
}

type containerVulnerability struct {
	containerInfo
	Vulnerability *grype_models.Match `json:"vulnerability"`
}

type containerDockerfileVulnerability struct {
	containerInfo
	DockerfileVulnerability *dockerfileVulnerability `json:"dockerfileVulnerability"`
}

type dockerfileVulnerability struct {
	Code        string `json:"code"`
	Level       string `json:"level"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type viewVulnerabilities struct {
	Vulnerabilities      []*containerVulnerability `json:"vulnerabilities,omitempty"`
	Total                int                       `json:"total"`
	TotalDefcon1         int                       `json:"totalDefcon1"`
	TotalCritical        int                       `json:"totalCritical"`
	TotalHigh            int                       `json:"totalHigh"`
}

type viewDockerfileVulnerabilities struct {
	DockerfileVulnerabilities []*containerDockerfileVulnerability `json:"dockerfileVulnerabilities,omitempty"`
	Total                     int                                 `json:"total"`
	TotalFatal                int                                 `json:"totalFatal"`
	TotalWarn                 int                                 `json:"totalWarn"`
	TotalInfo                 int                                 `json:"totalInfo"`
}

type viewData struct {
	Vulnerabilities           *viewVulnerabilities           `json:"vulnerabilities,omitempty"`
	DockerfileVulnerabilities *viewDockerfileVulnerabilities `json:"dockerfileVulnerabilities,omitempty"`
	ShowGoMsg                 bool                           `json:"showGoMsg"`
	ShowGoWarning             bool                           `json:"ShowGoWarning"`
	LastScannedNamespace      string                         `json:"lastScannedNamespace"`
}

const (
	defcon1Vulnerability    = "DEFCON1"
	criticalVulnerability   = "CRITICAL"
	highVulnerability       = "HIGH"
	mediumVulnerability     = "MEDIUM"
	lowVulnerability        = "LOW"
	negligibleVulnerability = "NEGLIGIBLE"
	unknownVulnerability    = "UNKNOWN"
)

func calculateVulnerabilitiesTotals(vulnerabilities []*containerVulnerability) (totalCritical, totalHigh, totalDefcon1 int) {
	for _, vul := range vulnerabilities {
		if vul.Vulnerability == nil {
			continue
		}
		switch strings.ToUpper(vul.Vulnerability.Vulnerability.Severity) {
		case defcon1Vulnerability:
			totalDefcon1++
		case criticalVulnerability:
			totalCritical++
		case highVulnerability:
			totalHigh++
		}
	}
	return
}

func calculateDockerfileVulnerabilitiesTotals(dockerfileVulnerabilities []*containerDockerfileVulnerability) (totalFatal, totalWarn, totalinfo int) {
	for _, vul := range dockerfileVulnerabilities {
		if vul.DockerfileVulnerability == nil {
			continue
		}
		switch dockle_config.ExitLevelMap[vul.DockerfileVulnerability.Level] {
		case dockle_types.FatalLevel:
			totalFatal++
		case dockle_types.WarnLevel:
			totalWarn++
		case dockle_types.InfoLevel:
			totalinfo++
		}
	}
	return
}

func getSeverityFromString(severity string) int {
	switch strings.ToUpper(severity) {
	case defcon1Vulnerability:
		return 0
	case criticalVulnerability:
		return 1
	case highVulnerability:
		return 2
	case mediumVulnerability:
		return 3
	case lowVulnerability:
		return 4
	case negligibleVulnerability:
		return 5
	case unknownVulnerability:
		return 6
	default:
		log.Warnf("invalid severity %v", severity)
		return 6
	}
}

func formatDockerfileDescription(assessments dockle_types.AssessmentSlice) string {
	ret := ""
	for _, assessment := range assessments {
		ret = ret + fmt.Sprintf("%s, ", assessment.Desc)
	}

	ret = strings.TrimSuffix(ret, ", ")

	return ret
}

// sort by severity, if equals or no vulnerability sort by name
func sortVulnerabilities(data []*containerVulnerability) []*containerVulnerability {
	sort.Slice(data[:], func(i, j int) bool {
		if data[i].Vulnerability == nil || data[j].Vulnerability == nil {
			return data[i].Pod < data[j].Pod
		}

		left := getSeverityFromString(data[i].Vulnerability.Vulnerability.Severity)
		right := getSeverityFromString(data[j].Vulnerability.Vulnerability.Severity)
		if left == right {
			return data[i].Pod < data[j].Pod
		}
		return left < right
	})

	return data
}

// sort by severity, if equals or no dockerfile vulnerability sort by name
func sortDockerfileVulnerabilities(data []*containerDockerfileVulnerability) []*containerDockerfileVulnerability {
	sort.Slice(data[:], func(i, j int) bool {
		if data[i].DockerfileVulnerability == nil || data[j].DockerfileVulnerability == nil {
			return data[i].Pod < data[j].Pod
		}

		left := dockle_config.ExitLevelMap[data[i].DockerfileVulnerability.Level]
		right := dockle_config.ExitLevelMap[data[j].DockerfileVulnerability.Level]
		if left == right {
			return data[i].Pod < data[j].Pod
		}
		return left > right
	})

	return data
}


func (wa *Webapp) convertImageScanResults(results []*types.ImageScanResult) ([]*containerVulnerability, []*containerDockerfileVulnerability) {
	var containerVulnerabilities []*containerVulnerability
	var containerDockerfileVulnerabilities []*containerDockerfileVulnerability

	severityThreshold := getSeverityFromString(wa.scanConfig.SeverityThreshold)
	for _, result := range results {
		metadata := containerInfo{
			Pod:       result.PodName,
			Container: result.ContainerName,
			Image:     result.ImageName,
			Namespace: result.PodNamespace,
			Succeeded: result.Success,
		}
		// show failed scan
		if !result.Success {
			containerVulnerabilities = append(containerVulnerabilities, &containerVulnerability{
				containerInfo: metadata,
			})
		} else {
			for _, vulnerability := range result.Vulnerabilities.Matches {
				if getSeverityFromString(vulnerability.Vulnerability.Severity) > severityThreshold {
					log.Debugf("Vulnerability severity below threshold. image=%+v, vulnerability=%+v, threshold=%+v",
						metadata.Image, vulnerability, wa.scanConfig.SeverityThreshold)
					continue
				}
				containerVulnerabilities = append(containerVulnerabilities, &containerVulnerability{
					containerInfo: metadata,
					Vulnerability: &vulnerability,
				})
			}
			for _, dfVulnerability := range result.DockerfileScanResults {
				containerDockerfileVulnerabilities = append(containerDockerfileVulnerabilities, &containerDockerfileVulnerability{
					containerInfo: metadata,
					DockerfileVulnerability: &dockerfileVulnerability{
						Code:        dfVulnerability.Code,
						Level:       dockle_writer.AlertLabels[dfVulnerability.Level],
						Title:       dockle_types.TitleMap[dfVulnerability.Code],
						Description: formatDockerfileDescription(dfVulnerability.Assessments),
					},

				})
			}
		}
	}

	sortedVulnerabilities := sortVulnerabilities(containerVulnerabilities)
	sortedDockerfileVulnerabilities := sortDockerfileVulnerabilities(containerDockerfileVulnerabilities)


	return sortedVulnerabilities, sortedDockerfileVulnerabilities
}

func (wa *Webapp) handleGoMsg() {
	if wa.scanConfig.TargetNamespace == "" {
		wa.lastScannedNamespace = "all namespaces"
	} else {
		wa.lastScannedNamespace = "namespace " + wa.scanConfig.TargetNamespace
	}
	wa.showGoMsg = true
	go func() {
		time.Sleep(5 * time.Second)
		wa.showGoMsg = false
	}()
}

/****************************************************** HANDLERS ******************************************************/

func (wa *Webapp) viewHandler(w http.ResponseWriter, _ *http.Request) {
	log.Debug("Received a 'view' request...")

	results := wa.orchestrator.Results()
	vulnerabilities, dfVulnerabilities := wa.convertImageScanResults(results.ImageScanResults)
	totalCritical, totalHigh, totalDefcon1 := calculateVulnerabilitiesTotals(vulnerabilities)
	totalFatal, totalWarn, totalInfo := calculateDockerfileVulnerabilitiesTotals(dfVulnerabilities)

	if wa.checkShowGoWarning {
		wa.showGoWarning = results.Progress.ImagesCompletedToScan != 0
	}

	err := wa.template.ExecuteTemplate(w, htmlFileName, &viewData{
		Vulnerabilities:      &viewVulnerabilities{
			Vulnerabilities:      vulnerabilities,
			Total:                len(vulnerabilities),
			TotalDefcon1:         totalDefcon1,
			TotalCritical:        totalCritical,
			TotalHigh:            totalHigh,
		},
		DockerfileVulnerabilities: &viewDockerfileVulnerabilities{
			DockerfileVulnerabilities: dfVulnerabilities,
			Total:                     len(dfVulnerabilities),
			TotalFatal:                totalFatal,
			TotalWarn:                 totalWarn,
			TotalInfo:                 totalInfo,
		},

		ShowGoMsg:            wa.showGoMsg,
		ShowGoWarning:        wa.showGoWarning,
		LastScannedNamespace: wa.lastScannedNamespace,
	})
	if err != nil {
		log.Errorf("Failed to execute template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (wa *Webapp) clearHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Received a 'clear' request...")

	wa.clearData()

	http.Redirect(w, r, "/view", http.StatusSeeOther)
}

func (wa *Webapp) goVerifyHandler(w http.ResponseWriter, r *http.Request) {
	wa.showGoWarning = wa.orchestrator.ScanProgress().ImagesCompletedToScan != 0
	if wa.showGoWarning {
		wa.checkShowGoWarning = true
		http.Redirect(w, r, "/view", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/go/run", http.StatusSeeOther)
	}
}

func (wa *Webapp) clearData() {
	wa.orchestrator.Clear()
}

func (wa *Webapp) goCancelHandler(w http.ResponseWriter, r *http.Request) {
	wa.checkShowGoWarning = false
	wa.showGoWarning = false
	http.Redirect(w, r, "/view", http.StatusSeeOther)
}

func (wa *Webapp) goRunHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Received a 'go' request...")

	wa.Lock()

	defer wa.Unlock()

	wa.checkShowGoWarning = false

	wa.handleGoMsg()

	wa.orchestrator.Clear()
	err := wa.orchestrator.Scan(wa.scanConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		http.Redirect(w, r, "/go/cancel", http.StatusSeeOther)
	}
}

/******************************************************* PUBLIC *******************************************************/

func Init(config *config.Config, scanConfig *config.ScanConfig) (*Webapp, error) {
	clientset, err := k8s_utils.CreateClientset("")
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return &Webapp{
		orchestrator:         orchestrator.Create(config, clientset),
		executionConfig:      config,
		scanConfig:           scanConfig,
		showGoMsg:            false,
		showGoWarning:        false,
		checkShowGoWarning:   false,
		lastScannedNamespace: "",
		template:             template.Must(template.ParseFiles(htmlPath)),
		Mutex:                sync.Mutex{},
	}, nil
}

func (wa *Webapp) Run() {
	errChannel := make(chan error, 2)
	go func() {
		if err := wa.orchestrator.Start(); err != nil && err != http.ErrServerClosed {
			errChannel <- fmt.Errorf("failed to start Orchestrator: %v", err)
		}
	}()

	http.HandleFunc("/view/", wa.viewHandler)
	http.HandleFunc("/clear/", wa.clearHandler)
	http.HandleFunc("/go/run/", wa.goRunHandler)
	http.HandleFunc("/go/verify/", wa.goVerifyHandler)
	http.HandleFunc("/go/cancel/", wa.goCancelHandler)
	go func() {
		if err := http.ListenAndServe(":"+wa.executionConfig.WebappPort, nil); err != nil {
			errChannel <- fmt.Errorf("failed to start Webapp: %v", err)
		}
	}()

	log.Infof("Webapp is running")

	err := <-errChannel
	log.Fatal(err)
}
