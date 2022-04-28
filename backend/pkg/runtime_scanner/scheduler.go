package runtime_scanner

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/api/server/models"
	"github.com/openclarity/kubeclarity/backend/pkg/database"
)

type Scheduler struct {
	stopChan chan struct{}
	// Send scan requests through here.
	scanChan  chan *ScanConfig
	dbHandler database.Database
}

type State struct {
	SchedulerParams *SchedulerParams
}

type SchedulerParams struct {
	Namespaces                    []string
	CisDockerBenchmarkScanEnabled bool
	Interval                      int64
	StartTime                     time.Time
	SingleScan bool
}


const (
	ByDaysScheduleScanConfig  = "ByDaysScheduleScanConfig"
	ByHoursScheduleScanConfig = "ByHoursScheduleScanConfig"
	SingleScheduleScanConfig  = "SingleScheduleScanConfig"
	WeeklyScheduleScanConfig  = "WeeklyScheduleScanConfig"
)

func CreateScheduler(scanChan chan *ScanConfig, dbHandler database.Database) *Scheduler {
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
	startTime, err := calculateNextScanTimeOnStart(time.Now().UTC(), sched)
	if err != nil {
		log.Errorf("Failed to calculate next scan start time: %v", err)
		return
	}
	scanConfig := models.RuntimeScheduleScanConfig{}
	if err := json.Unmarshal([]byte(sched.Config), &scanConfig); err != nil {
		return
	}

	// start schedule scan
	s.Schedule(&SchedulerParams{
		Namespaces:                    scanConfig.Namespaces,
		CisDockerBenchmarkScanEnabled: scanConfig.CisDockerBenchmarkScanEnabled,
		Interval:                      sched.Interval,
		StartTime:                     startTime,
		SingleScan: scanConfig.ScanConfigType().ScheduleScanConfigType() == SingleScheduleScanConfig,
	})
}

// If startTime is after timeNow, next scan time will be startTime.
// If startTime is before timeNow (startTime already passed), need to calculate the next scan time base on the interval.
// we take the number of seconds since startTime to timeNow, and perform a modolu in order to get the time left from now to the the next scan.
func calculateNextScanTimeOnStart(timeNow time.Time, s *database.Scheduler) (time.Time, error) {
	var nextScanTime time.Time
	interval := time.Duration(s.Interval)

	// if first scan didn't started yet
	if s.LastScanTime == "" {
		startTime, err := time.Parse(time.RFC3339, s.StartTime)
		if err != nil {
			return time.Time{}, err
		}
		// if first scan start time has passed, need to add interval until start time is after or equal to time now
		for startTime.Before(timeNow) {
			startTime = startTime.Add(interval)
		}
		return startTime, nil
	}

	// if first scan already happened, then lastScanTime will be set.
	// need to
	lastScanTime, err := time.Parse(time.RFC3339, s.LastScanTime)
	if err != nil {
		return time.Time{}, err
	}

	timePassedSinceLastScan := timeNow.Sub(lastScanTime)
	timePassedSinceLastScan = timePassedSinceLastScan / time.Second
	timePassedSinceLastScan = timePassedSinceLastScan % interval
	nextScanTime = timeNow.Add((interval - timePassedSinceLastScan) * time.Second)
	return nextScanTime, nil
}

func (s *Scheduler) Schedule(params *SchedulerParams) {
	// Clear
	close(s.stopChan)
	s.stopChan = make(chan struct{})

	startsIn := getStartsIn(params.StartTime)

	go s.spin(params, startsIn)

	if err := s.dbHandler.SchedulerTable().UpdateStartTime(time.Now().UTC().Format(time.RFC3339)); err != nil {
		log.Errorf("Failed to update start time: %v", err)
	}
	if err := s.dbHandler.SchedulerTable().UpdateInterval(params.Interval); err != nil {
		log.Errorf("Failed to update interval: %v", err)
	}
}

func getStartsIn(startTime time.Time) time.Duration {
	startsIn := startTime.Sub(time.Now().UTC())

	if startsIn < 0 {
		startsIn = 0
	}
	return startsIn
}

func (s *Scheduler) spin(params *SchedulerParams, startsIn time.Duration) {
	log.Errorf("Starting a new schedule scan. interval: %v, start time: %v, start in(sec): %v, namespaces: %v, cisDockerBenchmarkScanEnabled: %v",
		params.Interval, params.StartTime, startsIn, params.Namespaces, params.CisDockerBenchmarkScanEnabled)
	singleScan := params.SingleScan
	interval := time.Duration(params.Interval) * time.Second

	timer := time.NewTimer(startsIn)
	select {
	case <-s.stopChan:
		return
	case <-timer.C:
		go func() {
			ticker := time.NewTicker(interval)
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
					return
				}
			}
		}()
	}
}

func (s *Scheduler) sendScan(params *SchedulerParams) error {
	scanConfig := &ScanConfig{
		ScanType:                      models.ScanTypeSCHEDULE,
		CisDockerBenchmarkScanEnabled: params.CisDockerBenchmarkScanEnabled,
		Namespaces:                    params.Namespaces,
	}
	select {
	case s.scanChan <- scanConfig:
	default:
		return fmt.Errorf("failed to send scan config to channel")
	}
	// update last scan time.
	if err := s.dbHandler.SchedulerTable().UpdateLastScanTime(); err != nil {
		return fmt.Errorf("UpdateLastScanTime failed: %v", err)
	}
	return nil
}