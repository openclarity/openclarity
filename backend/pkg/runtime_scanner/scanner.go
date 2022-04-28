package runtime_scanner

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"github.com/openclarity/kubeclarity/api/server/models"
	runtime_scan_config "github.com/openclarity/kubeclarity/runtime_scan/pkg/config"
	"github.com/openclarity/kubeclarity/runtime_scan/pkg/orchestrator"
	"github.com/openclarity/kubeclarity/runtime_scan/pkg/types"
	_types "github.com/openclarity/kubeclarity/runtime_scan/pkg/types"
)

type Scanner interface {
	GetScanConfig() *ScanConfig
	ScanProgress() types.ScanProgress
	GetScannedNamespaces() []string
	GetLastScanStartTime() time.Time
	GetLastScanEndTime() time.Time
	Start(stopChan chan struct{})
	Stop()
	StopCurrentScan()
	Clear()
}

type RuntimeScanner struct {
	vulnerabilitiesScanner orchestrator.VulnerabilitiesScanner
	// Scan config used in the latest runtime scan.
	lastScanConfig    *ScanConfig
	lastScanStartTime time.Time
	lastScanEndTime   time.Time
	// List of latest scanned namespaces.
	scannedNamespaces      []string // nolint:structcheck
	stopCurrentScanChan    chan struct{}
	stopRuntimeScannerChan chan struct{}
	// Scan results will be sent through this channel.
	resultsChan chan *_types.ScanResults
	// New scan requests are coming through this channel.
	scanChan chan *ScanConfig
	lock     sync.RWMutex
}

func (s *RuntimeScanner) ScanProgress() _types.ScanProgress {
	return s.vulnerabilitiesScanner.ScanProgress()
}

type ScanType string

type ScanConfig struct {
	ScanType                      models.ScanType
	CisDockerBenchmarkScanEnabled bool
	Namespaces                    []string
}

func CreateRuntimeScanner(scanner orchestrator.VulnerabilitiesScanner, scanChan chan *ScanConfig, resultsChan chan *_types.ScanResults) Scanner {
	return &RuntimeScanner{
		vulnerabilitiesScanner: scanner,
		stopCurrentScanChan:    make(chan struct{}),
		stopRuntimeScannerChan: make(chan struct{}),
		scanChan:               scanChan,
		lock:                   sync.RWMutex{},
		resultsChan:            resultsChan,
		lastScanConfig:         &ScanConfig{},
	}
}

func (s *RuntimeScanner) GetLastScanStartTime() time.Time {
	return s.lastScanStartTime
}

func (s *RuntimeScanner) GetLastScanEndTime() time.Time {
	return s.lastScanEndTime
}

func (s *RuntimeScanner) GetScannedNamespaces() []string {
	return s.scannedNamespaces
}

func (s *RuntimeScanner) Clear() {
	s.vulnerabilitiesScanner.Clear()
}

func (s *RuntimeScanner) Start(stopChan chan struct{}) {
	s.vulnerabilitiesScanner.Start(stopChan)
	go func() {
		for {
			select {
			case <-stopChan:
				log.Info("Runtime scanner received stop event")
				return
			case scanConfig := <-s.scanChan:
				scanConfigB, _ := json.Marshal(scanConfig)
				log.Errorf("Received a new scan config: %s", scanConfigB)
				if err := s.startScan(scanConfig); err != nil {
					log.Errorf("Failed to start scan: %v", err)
					continue
				}
				s.scannedNamespaces = scanConfig.Namespaces
			}
		}
	}()
}

func (s *RuntimeScanner) Stop() {
	close(s.stopRuntimeScannerChan)
}

func (s *RuntimeScanner) GetScanConfig() *ScanConfig {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.lastScanConfig
}

func (s *RuntimeScanner) setScanConfig(scanConfig *ScanConfig) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.lastScanConfig = scanConfig
}

func (s *RuntimeScanner) startScan(scanConfig *ScanConfig) error {
	lastScanTime, err := time.Parse(time.RFC3339, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return err
	}
	s.lock.Lock()
	s.lastScanStartTime = lastScanTime
	stop := s.stopCurrentScanChan
	namespaces := scanConfig.Namespaces
	s.lock.Unlock()

	s.vulnerabilitiesScanner.Clear()

	// Save on state the config in the current runtime scan.
	s.setScanConfig(scanConfig)

	// need to create scan done channel for every new scan
	done := make(chan struct{})

	if len(namespaces) == 0 {
		// Empty namespaces list should scan all namespaces.
		namespaces = []string{corev1.NamespaceAll}
	}

	err = s.vulnerabilitiesScanner.Scan(&runtime_scan_config.ScanConfig{
		MaxScanParallelism:           10, // nolint:gomnd
		TargetNamespaces:             namespaces,
		IgnoredNamespaces:            nil,
		JobResultTimeout:             10 * time.Minute, // nolint:gomnd
		DeleteJobPolicy:              runtime_scan_config.DeleteJobPolicySuccessful,
		ShouldScanCISDockerBenchmark: scanConfig.CisDockerBenchmarkScanEnabled,
	}, done)
	if err != nil {
		return fmt.Errorf("failed to start scan: %v", err)
	}

	go func() {
		select {
		case <-done:
			s.lastScanEndTime, err = time.Parse(time.RFC3339, time.Now().UTC().Format(time.RFC3339))
			if err != nil {
				log.Errorf("Failed to parse time: %v", err)
			}
			results := s.vulnerabilitiesScanner.Results()
			select {
			case s.resultsChan <- results:
			default:
				log.Error("Failed to send results to channel")
			}
		case <-stop:
			log.Infof("Received a stop signal, not waiting for results")
		}
	}()

	return nil
}

func (s *RuntimeScanner) StopCurrentScan() {
	s.lock.Lock()
	close(s.stopCurrentScanChan)
	s.stopCurrentScanChan = make(chan struct{})
	s.scannedNamespaces = []string{}
	s.lock.Unlock()
}
