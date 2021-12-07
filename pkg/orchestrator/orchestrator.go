package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	dockle_types "github.com/Portshift/dockle/pkg/types"
	"github.com/Portshift/klar/forwarding"
	"github.com/Portshift/kubei/pkg/config"
	"github.com/Portshift/kubei/pkg/scanner"
	"github.com/Portshift/kubei/pkg/types"
)

type Orchestrator struct {
	scanner   *scanner.Scanner
	config    *config.Config
	clientset kubernetes.Interface
	server    *http.Server
	sync.Mutex
}

type VulnerabilitiesScanner interface {
	Start() error
	Scan(scanConfig *config.ScanConfig) error
	ScanProgress() types.ScanProgress
	Results() *types.ScanResults
	Clear()
	Stop()
}

func Create(config *config.Config, clientset kubernetes.Interface) *Orchestrator {
	o := &Orchestrator{
		scanner:   scanner.CreateScanner(config, clientset),
		config:    config,
		clientset: clientset,
		server:    &http.Server{Addr: ":" + config.ResultListenPort},
		Mutex:     sync.Mutex{},
	}

	http.HandleFunc("/result/", o.resultHttpHandler)
	http.HandleFunc("/dockerfileScanResult/", o.dockerfileResultHttpHandler)

	return o
}

func readResultBodyData(req *http.Request) (*forwarding.ImageVulnerabilities, error) {
	decoder := json.NewDecoder(req.Body)
	var bodyData *forwarding.ImageVulnerabilities
	err := decoder.Decode(&bodyData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode result: %v", err)
	}

	return bodyData, nil
}

func (o *Orchestrator) resultHttpHandler(w http.ResponseWriter, r *http.Request) {
	result, err := readResultBodyData(r)
	if err != nil || result == nil {
		log.Errorf("Invalid result. err=%v, result=%+v", err, result)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Debugf("Result was received. image=%+v, success=%+v, scanUUID=%+v",
		result.Image, result.Success, result.ScanUUID)

	err = o.getScanner().HandleVulnerabilitiesResult(result)
	if err != nil {
		log.Errorf("Failed to handle result. err=%v, result=%+v", err, result)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Debugf("Result was added successfully. image=%+v", result.Image)
	w.WriteHeader(http.StatusAccepted)
}

func readDockerfileScanResultBodyData(req *http.Request) (*dockle_types.ImageAssessment, error) {
	decoder := json.NewDecoder(req.Body)
	var bodyData *dockle_types.ImageAssessment
	err := decoder.Decode(&bodyData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode result: %v", err)
	}

	return bodyData, nil
}

func (o *Orchestrator) dockerfileResultHttpHandler(w http.ResponseWriter, r *http.Request) {
	result, err := readDockerfileScanResultBodyData(r)
	if err != nil || result == nil {
		log.Errorf("Invalid result. err=%v, result=%+v", err, result)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Debugf("Dockerfile scan result was received. image=%+v, success=%+v, scanUUID=%+v",
		result.Image, result.Success, result.ScanUUID)

	err = o.getScanner().HandleDockerfileResult(result)
	if err != nil {
		log.Errorf("Failed to handle result. err=%v, result=%+v", err, result)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Debugf("Result was added successfully. image=%+v", result.Image)
	w.WriteHeader(http.StatusAccepted)
}

func (o *Orchestrator) Start() error {
	// Start result server
	log.Infof("Starting Orchestrator server")

	return o.server.ListenAndServe()
}

func (o *Orchestrator) Stop() {
	o.Clear()

	log.Infof("Stopping Orchestrator server")
	if o.server != nil {
		if err := o.server.Shutdown(context.Background()); err != nil {
			log.Errorf("Failed to shutdown server: %v", err)
		}
	}
}

func (o *Orchestrator) Scan(scanConfig *config.ScanConfig) error {
	if err := o.getScanner().Scan(scanConfig); err != nil {
		return err
	}

	return nil
}

func (o *Orchestrator) ScanProgress() types.ScanProgress {
	return o.getScanner().ScanProgress()
}

func (o *Orchestrator) Results() *types.ScanResults {
	return o.getScanner().Results()
}

func (o *Orchestrator) Clear() {
	o.Lock()
	defer o.Unlock()

	log.Infof("Clearing Orchestrator")
	o.scanner.Clear()
	o.scanner = scanner.CreateScanner(o.config, o.clientset)
}

func (o *Orchestrator) getScanner() *scanner.Scanner {
	o.Lock()
	defer o.Unlock()

	return o.scanner
}
