package webapp

import (
	"fmt"
	"github.com/Portshift/klar/clair"
	"github.com/Portshift/kubei/pkg/config"
	"github.com/Portshift/kubei/pkg/orchestrator"
	"github.com/Portshift/kubei/pkg/types"
	k8s_utils "github.com/Portshift/kubei/pkg/utils/k8s"
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

//noinspection GoUnusedGlobalVariable
var templates = template.Must(template.ParseFiles(htmlPath))

type Webapp struct {
	orchestrator         *orchestrator.Orchestrator
	executionConfig      *config.Config
	scanConfig           *config.ScanConfig
	showGoMsg            bool
	showGoWarning        bool // Warning that previous results will be lost
	checkShowGoWarning   bool
	lastScannedNamespace string
	sync.Mutex
}

type extendedContextualMetadata struct {
	Pod       string `json:"pod"`
	Container string `json:"container"`
	Image     string `json:"image"`
	Namespace string `json:"namespace"`
	Succeeded bool   `json:"succeeded"`
}

type extendedContextualVulnerability struct {
	extendedContextualMetadata
	Vulnerability *clair.Vulnerability `json:"vulnerability"`
}

type viewData struct {
	Vulnerabilities      []*extendedContextualVulnerability `json:"vulnerabilities,omitempty"`
	Total                int                                `json:"total"`
	TotalDefcon1         int                                `json:"totalDefcon1"`
	TotalCritical        int                                `json:"totalCritical"`
	TotalHigh            int                                `json:"totalHigh"`
	ShowGoMsg            bool                               `json:"showGoMsg"`
	ShowGoWarning        bool                               `json:"ShowGoWarning"`
	LastScannedNamespace string                             `json:"lastScannedNamespace"`
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

func (wa *Webapp) calculateTotals(vulnerabilities []*extendedContextualVulnerability) (totalCritical, totalHigh, totalDefcon1 int) {
	for _, vul := range vulnerabilities {
		if vul.Vulnerability == nil {
			continue
		}
		switch strings.ToUpper(vul.Vulnerability.Severity) {
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

// sort by severity, if equals or no vulnerability sort by name
func sortVulnerabilities(data []*extendedContextualVulnerability) []*extendedContextualVulnerability {
	sort.Slice(data[:], func(i, j int) bool {
		if data[i].Vulnerability == nil || data[j].Vulnerability == nil {
			return data[i].Pod < data[j].Pod
		}

		left := getSeverityFromString(data[i].Vulnerability.Severity)
		right := getSeverityFromString(data[j].Vulnerability.Severity)
		if left == right {
			return data[i].Pod < data[j].Pod
		}
		return left < right
	})

	return data
}

func (wa *Webapp) convertImageScanResults(results []*types.ImageScanResult) []*extendedContextualVulnerability {
	var extendedContextualVulnerabilities []*extendedContextualVulnerability
	severityThreshold := getSeverityFromString(wa.scanConfig.SeverityThreshold)
	for _, result := range results {
		metadata := extendedContextualMetadata{
			Pod:       result.PodName,
			Container: result.ContainerName,
			Image:     result.ImageName,
			Namespace: result.PodNamespace,
			Succeeded: result.Success,
		}
		// show failed scan
		if !result.Success {
			extendedContextualVulnerabilities = append(extendedContextualVulnerabilities, &extendedContextualVulnerability{
				extendedContextualMetadata: metadata,
			})
		} else {
			for _, vulnerability := range result.Vulnerabilities {
				if getSeverityFromString(vulnerability.Severity) > severityThreshold {
					log.Debugf("Vulnerability severity below threshold. image=%+v, vulnerability=%+v, threshold=%+v",
						metadata.Image, vulnerability, wa.scanConfig.SeverityThreshold)
					continue
				}
				extendedContextualVulnerabilities = append(extendedContextualVulnerabilities, &extendedContextualVulnerability{
					extendedContextualMetadata: metadata,
					Vulnerability:              vulnerability,
				})
			}
		}
	}

	sortedResults := sortVulnerabilities(extendedContextualVulnerabilities)

	return sortedResults
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
	extendedContextualVulnerabilities := wa.convertImageScanResults(results.ImageScanResults)
	totalCritical, totalHigh, totalDefcon1 := wa.calculateTotals(extendedContextualVulnerabilities)

	if wa.checkShowGoWarning {
		wa.showGoWarning = results.Progress.ImagesCompletedToScan != 0
	}

	err := templates.ExecuteTemplate(w, htmlFileName, &viewData{
		Vulnerabilities:      extendedContextualVulnerabilities,
		Total:                len(extendedContextualVulnerabilities),
		TotalDefcon1:         totalDefcon1,
		TotalCritical:        totalCritical,
		TotalHigh:            totalHigh,
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
