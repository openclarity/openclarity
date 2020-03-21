package webapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Portshift-Admin/klar/forwarding"
	log "github.com/sirupsen/logrus"
	"html/template"
	"kubei/pkg/common"
	"kubei/pkg/orchestrator"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const htmlFileName = "view.html"
const htmlPath = "/usr/local/portshift/" + htmlFileName

//noinspection GoUnusedGlobalVariable
var templates = template.Must(template.ParseFiles(htmlPath))

type Webapp struct {
	kubeiOrchestrator        *orchestrator.Orchestrator
	goExecutionLock          *sync.Mutex
	executionConfig          *common.ExecutionConfiguration
	batchCompletedScansCount *int32
	showGoMsg                bool
	showGoWarning            bool
	checkShowGoWarning       bool
	lastScannedNamespace     string
	existingData             []*common.ContextualVulnerability
	numOfResults             int32
	scanIssuesMessages       *[]string
}

func (wa *Webapp) calculateTotals(vulnerabilities []*common.ExtendedContextualVulnerability) (int, int, int) {
	totalCritical := 0
	totalHigh := 0
	totalDefcon1 := 0
	for _, vul := range vulnerabilities {
		switch strings.ToUpper(vul.Vulnerability.Severity) {
		case "DEFCON1":
			totalDefcon1++
		case "CRITICAL":
			totalCritical++
		case "HIGH":
			totalHigh++
		}
	}
	return totalCritical, totalHigh, totalDefcon1
}

func (wa *Webapp) getListeningPort() string {
	listeningPort := os.Getenv("LISTENING_PORT")
	if listeningPort == "" {
		return "8080"
	}
	_, err := strconv.Atoi(listeningPort)
	if err != nil {
		log.Infof("LISTENING_PORT is invalid. defaulting to 8080")
		return "8080"
	}
	return listeningPort
}

func (wa *Webapp) getSeverityFromString(severity string) int {
	switch strings.ToUpper(severity) {
	case "DEFCON1":
		return 0
	case "CRITICAL":
		return 1
	case "HIGH":
		return 2
	case "MEDIUM":
		return 3
	case "LOW":
		return 4
	case "NEGLIGIBLE":
		return 5
	case "UNKNOWN":
		return 6
	default:
		panic(fmt.Sprintf("invalid severity %v", severity))
	}
}

// sort by severity, if equals sort by name
func (wa *Webapp) sortVulnerabilities(data []*common.ExtendedContextualVulnerability) []*common.ExtendedContextualVulnerability {
	sort.Slice(data[:], func(i, j int) bool {
		left := wa.getSeverityFromString(data[i].Vulnerability.Severity)
		right := wa.getSeverityFromString(data[j].Vulnerability.Severity)
		if left == right {
			return data[i].Pod < data[j].Pod
		}
		return left < right
	})

	return data
}

func (wa *Webapp) extendVulnerabilitiesContext(data []*common.ContextualVulnerability) []*common.ExtendedContextualVulnerability {
	var extendedContextualVulnerabilities []*common.ExtendedContextualVulnerability
	for _, v := range data {
		image := v.Image
		vulnerability := v.Vulnerability

		imageK8ExtendedContexts := wa.kubeiOrchestrator.ImageK8ExtendedContextMap[common.ContainerImageName(image)]
		for _, imageK8ExtendedContext := range imageK8ExtendedContexts {
			extendedContextualVulnerability := common.ExtendedContextualVulnerability{
				Vulnerability: vulnerability,
				Pod:           imageK8ExtendedContext.Pod,
				Container:     imageK8ExtendedContext.Container,
				Image:         image,
				Namespace:     imageK8ExtendedContext.Namespace,
			}
			extendedContextualVulnerabilities = append(extendedContextualVulnerabilities, &extendedContextualVulnerability)
		}
	}

	wa.sortVulnerabilities(extendedContextualVulnerabilities)
	return extendedContextualVulnerabilities
}

func (wa *Webapp) readBodyData(req *http.Request, w http.ResponseWriter) *forwarding.ImageVulnerabilities {
	decoder := json.NewDecoder(req.Body)
	var bodyData *forwarding.ImageVulnerabilities
	err := decoder.Decode(&bodyData)
	if err != nil {
		log.Errorf("Error reading body: %v", err)
		http.Error(w, "Failed to save", http.StatusBadRequest)
		return nil
	}

	defer req.Body.Close()

	return bodyData
}

func (wa *Webapp) appendLists(originalList []*common.ContextualVulnerability, toAppendCandidates *forwarding.ImageVulnerabilities) []*common.ContextualVulnerability {
	for _, vul := range toAppendCandidates.Vulnerabilities {
		candidate := &common.ContextualVulnerability{
			Vulnerability: vul,
			Image:         toAppendCandidates.Image,
		}
		originalList = wa.appendVulnerabilityIfMissing(originalList, candidate)
	}

	return originalList
}

func (wa *Webapp) appendVulnerabilityIfMissing(vs []*common.ContextualVulnerability, candidate *common.ContextualVulnerability) []*common.ContextualVulnerability {
	for _, ele := range vs {
		if ele.Vulnerability.Name == candidate.Vulnerability.Name &&
			ele.Image == candidate.Image {
			return vs
		}
	}
	vs = append(vs, candidate)
	return vs
}

func (wa *Webapp) handleGoMsg() {
	if wa.executionConfig.TargetNamespace == "" {
		wa.lastScannedNamespace = "all namespaces"
	} else {
		wa.lastScannedNamespace = "namespace " + wa.executionConfig.TargetNamespace
	}
	wa.showGoMsg = true
	go func() {
		time.Sleep(5 * time.Second)
		wa.showGoMsg = false
	}()
}

/****************************************************** HANDLERS ******************************************************/

func (wa *Webapp) addHandler(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt32(wa.batchCompletedScansCount, 1)
	atomic.AddInt32(&wa.numOfResults, 1)
	wa.kubeiOrchestrator.DataUpdateLock.Lock()
	defer wa.kubeiOrchestrator.DataUpdateLock.Unlock()

	newImageVulnerabilities := wa.readBodyData(r, w)
	if newImageVulnerabilities == nil {
		log.Debugf("Received an 'add' request with no body...")
		return
	}
	log.Debugf("Received an 'add' request for image %s", newImageVulnerabilities.Image)
	if !newImageVulnerabilities.Success {
		*wa.scanIssuesMessages = common.AppendStringIfMissing(*wa.scanIssuesMessages, "Scan of image "+newImageVulnerabilities.Image+" has failed! See container logs for more info.")
	} else {
		log.Debugf("Scan of image %s is done!", newImageVulnerabilities.Image)
	}

	wa.existingData = wa.appendLists(wa.existingData, newImageVulnerabilities)

	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(wa.existingData)

	if err != nil {
		log.Errorf("Error encoding DB: %v", err)
		http.Error(w, "Failed to save", http.StatusInternalServerError)
		return
	}

	log.Debug("Successfully added new vulnerabilities!")

	http.Redirect(w, r, "/view", http.StatusSeeOther)
}

func (wa *Webapp) viewHandler(w http.ResponseWriter, _ *http.Request) {
	log.Debug("Received a 'view' request...")

	extendedContextualVulnerabilities := wa.extendVulnerabilitiesContext(wa.existingData)
	totalCritical, totalHigh, totalDefcon1 := wa.calculateTotals(extendedContextualVulnerabilities)
	wa.numOfResults = int32(len(extendedContextualVulnerabilities))
	viewData := &common.ViewData{
		Vulnerabilities:      extendedContextualVulnerabilities,
		Total:                len(extendedContextualVulnerabilities),
		TotalDefcon1:         totalDefcon1,
		TotalCritical:        totalCritical,
		TotalHigh:            totalHigh,
		ShowGoMsg:            wa.showGoMsg,
		ShowGoWarning:        wa.showGoWarning,
		LastScannedNamespace: wa.lastScannedNamespace,
	}

	if wa.checkShowGoWarning {
		wa.showGoWarning = int(atomic.LoadInt32(&wa.numOfResults)) != 0
	}

	err := templates.ExecuteTemplate(w, htmlFileName, viewData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (wa *Webapp) clearHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Received a 'clear' request...")

	wa.clearData()

	http.Redirect(w, r, "/view", http.StatusSeeOther)
}

func (wa *Webapp) goVerifyHandler(w http.ResponseWriter, r *http.Request) {
	wa.showGoWarning = int(atomic.LoadInt32(&wa.numOfResults)) != 0
	if wa.showGoWarning {
		wa.checkShowGoWarning = true
		http.Redirect(w, r, "/view", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/go/run", http.StatusSeeOther)
	}
}

func (wa *Webapp) clearData() {
	wa.kubeiOrchestrator.DataUpdateLock.Lock()
	defer wa.kubeiOrchestrator.DataUpdateLock.Unlock()

	wa.existingData = []*common.ContextualVulnerability{}
	wa.kubeiOrchestrator.ImageK8ExtendedContextMap = make(common.ImageK8ExtendedContextMap)
	wa.numOfResults = 0
}

func (wa *Webapp) goCancelHandler(w http.ResponseWriter, r *http.Request) {
	wa.checkShowGoWarning = false
	wa.showGoWarning = false
	http.Redirect(w, r, "/view", http.StatusSeeOther)
}

func (wa *Webapp) goRunHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Received a 'go' request...")
	wa.clearData()

	wa.goExecutionLock.Lock()

	defer wa.goExecutionLock.Unlock()

	wa.checkShowGoWarning = false

	wa.handleGoMsg()

	go wa.kubeiOrchestrator.Scan()

	http.Redirect(w, r, "/go/cancel", http.StatusSeeOther)
}

/******************************************************* PUBLIC *******************************************************/

func Init(executionConfiguration *common.ExecutionConfiguration) *Webapp {
	var batchCompletedScansCount int32 = 0
	var scanIssuesMessages []string
	goExecutionLock := sync.Mutex{}
	dataUpdateLock := sync.Mutex{}
	var imageK8ExtendedContextMap = make(common.ImageK8ExtendedContextMap)
	kubeiOrchestrator := orchestrator.Init(executionConfiguration, &dataUpdateLock, imageK8ExtendedContextMap, &scanIssuesMessages, &batchCompletedScansCount)
	return &Webapp{
		kubeiOrchestrator:        kubeiOrchestrator,
		goExecutionLock:          &goExecutionLock,
		executionConfig:          executionConfiguration,
		batchCompletedScansCount: &batchCompletedScansCount,
		showGoMsg:                false,
		showGoWarning:            false,
		checkShowGoWarning:       false,
		lastScannedNamespace:     "",
		existingData:             nil,
		numOfResults:             0,
		scanIssuesMessages:       &scanIssuesMessages,
	}
}

func (wa *Webapp) Run() {
	log.Infof("RUNNING...")
	http.HandleFunc("/add/", wa.addHandler)
	http.HandleFunc("/view/", wa.viewHandler)
	http.HandleFunc("/clear/", wa.clearHandler)
	http.HandleFunc("/go/run/", wa.goRunHandler)
	http.HandleFunc("/go/verify/", wa.goVerifyHandler)
	http.HandleFunc("/go/cancel/", wa.goCancelHandler)
	log.Fatal(http.ListenAndServe(":"+wa.getListeningPort(), nil))
}
