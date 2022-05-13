package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/api/server/models"
	"github.com/openclarity/kubeclarity/backend/pkg/database"
	"github.com/openclarity/kubeclarity/backend/pkg/runtime_scanner"
)

type Scheduler struct {
	stopChan chan struct{}
	// Send scan requests through here.
	scanChan  chan *runtime_scanner.ScanConfig
	dbHandler database.Database
}

type State struct {
	SchedulerParams *SchedulerParams
}

type SchedulerParams struct {
	Namespaces                    []string
	CisDockerBenchmarkScanEnabled bool
	// interval in seconds
	IntervalSec int64
	StartTime   time.Time
	SingleScan  bool
}

const (
	ByDaysScheduleScanConfig  = "ByDaysScheduleScanConfig"
	ByHoursScheduleScanConfig = "ByHoursScheduleScanConfig"
	SingleScheduleScanConfig  = "SingleScheduleScanConfig"
	WeeklyScheduleScanConfig  = "WeeklyScheduleScanConfig"
)

func CreateScheduler(scanChan chan *runtime_scanner.ScanConfig, dbHandler database.Database) *Scheduler {
	return &Scheduler{
		stopChan:  make(chan struct{}),
		scanChan:  scanChan,
		dbHandler: dbHandler,
	}
}

func (s *Scheduler) Init() {
	// read last schedule scan config from db
	sched, err := s.dbHandler.SchedulerTable().Get()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info("No scheduler config")
			return
		}
		log.Errorf("Failed to get scheduler table: %v", err)
		return
	}
	scanConfig := models.RuntimeScheduleScanConfig{}
	if err := json.Unmarshal([]byte(sched.Config), &scanConfig); err != nil {
		log.Errorf("Failed to unmarshal scheduler config: %v", err)
		return
	}
	startTime, err := time.Parse(time.RFC3339, sched.NextScanTime)
	if err != nil {
		log.Errorf("Failed to parse nextScanTime: %v. %v", sched.NextScanTime, err)
		return
	}

	// start schedule scan
	s.Schedule(&SchedulerParams{
		Namespaces:                    scanConfig.Namespaces,
		CisDockerBenchmarkScanEnabled: scanConfig.CisDockerBenchmarkScanEnabled,
		IntervalSec:                   sched.Interval,
		StartTime:                     startTime,
		SingleScan:                    scanConfig.ScanConfigType().ScheduleScanConfigType() == SingleScheduleScanConfig,
	})
}

func (s *Scheduler) Schedule(params *SchedulerParams) {
	// Clear
	close(s.stopChan)
	s.stopChan = make(chan struct{})

	startsAtSec := getStartsAtSec(time.Now().UTC(), params.StartTime, time.Duration(params.IntervalSec))

	go s.spin(params, startsAtSec)
}

func getNextScanTime(timeNow, currentScanTime time.Time, intervalSec time.Duration) time.Time {
	if currentScanTime.Before(timeNow) {
		timePassedSec := timeNow.Sub(currentScanTime) / time.Second
		remainingInterval := timePassedSec % intervalSec
		if remainingInterval == 0 {
			currentScanTime = timeNow
		} else {
			currentScanTime = timeNow.Add((intervalSec - remainingInterval) * time.Second)
		}
	}
	return currentScanTime
}

func getStartsAtSec(timeNow time.Time, startTime time.Time, intervalSec time.Duration) time.Duration {
	nextScanTime := getNextScanTime(timeNow, startTime, intervalSec)

	startsAt := nextScanTime.Sub(timeNow)

	return startsAt / time.Second
}

func (s *Scheduler) spin(params *SchedulerParams, startsAtSec time.Duration) {
	log.Errorf("Starting a new schedule scan. interval: %v, start time: %v, start in(sec): %v, namespaces: %v, cisDockerBenchmarkScanEnabled: %v",
		params.IntervalSec, params.StartTime, startsAtSec, params.Namespaces, params.CisDockerBenchmarkScanEnabled)
	singleScan := params.SingleScan
	interval := time.Duration(params.IntervalSec)

	timer := time.NewTimer(startsAtSec * time.Second)
	select {
	case <-s.stopChan:
		return
	case <-timer.C:
		go func() {
			ticker := time.NewTicker(interval * time.Second)
			defer ticker.Stop()
			for {
				if err := s.sendScan(params); err != nil {
					log.Errorf("Failed to send scan: %v", err)
				}
				if singleScan {
					return
				}
				select {
				case <-ticker.C:
				case <-s.stopChan:
					log.Debugf("Received a stop signal...")
					return
				}
			}
		}()
	}
}

func (s *Scheduler) sendScan(params *SchedulerParams) error {
	scanConfig := &runtime_scanner.ScanConfig{
		ScanType:                      models.ScanTypeSCHEDULE,
		CisDockerBenchmarkScanEnabled: params.CisDockerBenchmarkScanEnabled,
		Namespaces:                    params.Namespaces,
	}
	select {
	case s.scanChan <- scanConfig:
	default:
		return fmt.Errorf("failed to send scan config to channel")
	}
	// update next scan time.
	nextScanTime := time.Now().Add(time.Duration(params.IntervalSec)).UTC().Format(time.RFC3339)
	if err := s.dbHandler.SchedulerTable().UpdateNextScanTime(nextScanTime); err != nil {
		return fmt.Errorf("UpdateNextScanTime failed: %v", err)
	}
	return nil
}
