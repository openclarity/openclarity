// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
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

package scanconfig

import (
	"fmt"
	"time"

	"github.com/aptible/supercronic/cronexpr"

	apitypes "github.com/openclarity/vmclarity/api/types"
)

// The ScheduleWindow represents a timeframe defined by the start and end timestamps.
type ScheduleWindow struct {
	start time.Time
	end   time.Time
}

// Start returns the beginning of the ScheduleWindow.
func (w *ScheduleWindow) Start() time.Time {
	return w.start
}

// End returns the end of the ScheduleWindow.
func (w *ScheduleWindow) End() time.Time {
	return w.end
}

// In returns true if the t time is within the w ScheduleWindow, otherwise it returns false.
func (w *ScheduleWindow) In(t time.Time) bool {
	if t.Before(w.start) || t.After(w.end) {
		return false
	}

	return true
}

// Before returns true if w ScheduleWindow is before t time.Time, otherwise false.
func (w *ScheduleWindow) Before(t time.Time) bool {
	return w.end.Before(t)
}

// After returns true if w ScheduleWindow is after t time.Time, otherwise false.
func (w *ScheduleWindow) After(t time.Time) bool {
	return w.start.After(t)
}

// Next returns a new ScheduleWindow which represents the timeframe after w ScheduleWindow.
func (w ScheduleWindow) Next() *ScheduleWindow {
	return &ScheduleWindow{
		start: w.end,
		end:   w.end.Add(w.end.Sub(w.start)),
	}
}

// Prev returns a new ScheduleWindow which represents the timeframe before w ScheduleWindow.
func (w ScheduleWindow) Prev() *ScheduleWindow {
	return &ScheduleWindow{
		start: w.start.Add(-1 * w.end.Sub(w.start)),
		end:   w.start,
	}
}

func (w ScheduleWindow) String() string {
	return fmt.Sprintf("start: %s, end: %s", w.start.Format(time.RFC3339), w.end.Format(time.RFC3339))
}

// NewScheduleWindow returns a new ScheduleWindow representing a timeframe where now parameter defines the middle of the
// calculated timeframe.
//
//	+--------------------+
//	|   ScheduleWindow   |
//	+--------------------+
//	          ^
//	         now (middle)
//
//	<-------------------->
//	         size
//
// nolint:mnd
func NewScheduleWindow(now time.Time, size time.Duration) *ScheduleWindow {
	return &ScheduleWindow{
		start: now.Add(-1 * size / 2),
		end:   now.Add(size / 2),
	}
}

// OperationTime is a single point in time defined by the time parameter. Providing an optional cron parameter makes the
// OperationTime recurring in case the cron expression represents a recurring cadence.
type OperationTime struct {
	time time.Time
	cron *cronexpr.Expression
}

// Next returns a new OperationTime representing the next cadence in case cron parameter is provided,
// otherwise it returns itself.
func (o OperationTime) Next() *OperationTime {
	if o.cron != nil && !o.cron.Next(o.time).IsZero() {
		o.time = o.cron.Next(o.time)
	}

	return &o
}

// NextAfter returns a new OperationTime representing time which is after then the provided t time. It returns itself
// if t is zero or o has no cron set.
func (o OperationTime) NextAfter(t time.Time) *OperationTime {
	if t.IsZero() {
		return &o
	}

	if o.cron == nil {
		return &o
	}

	for t.After(o.Time()) {
		next := o.Next()
		if next.Time().Equal(o.Time()) {
			break
		}
		o = *next
	}

	return &o
}

// Time returns time.Time represented by the OperationTime.
func (o *OperationTime) Time() time.Time {
	return o.time
}

// IsRecurring returns true of cron is set and its expression defines a recurring cadence.
func (o *OperationTime) IsRecurring() bool {
	next := o.Next()
	return !o.Time().Equal(next.Time())
}

func (o OperationTime) String() string {
	return o.Time().Format(time.RFC3339)
}

// NewOperationTime returns a OperationTime created by using the provided t time and c cron expression.
func NewOperationTime(t time.Time, c *cronexpr.Expression) *OperationTime {
	// Check if c cron expression represents a single point in time which case it is used instead of t time.
	if c != nil {
		cronTime, ok := isCronPointInTime(c)
		if ok {
			t = cronTime
		}
	}

	return &OperationTime{
		time: t,
		cron: c,
	}
}

func isCronPointInTime(c *cronexpr.Expression) (time.Time, bool) {
	// NOTE: from.Add(1) is needed as `from` represents zero time and cronexpr returns zero time if it is provided
	//       with zero time as first parameter. Non-standard cron expressions (Quartz) may include year field
	//       representing time which might be in the past hence the `from` time is set to zero time.
	from := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	t := c.Next(from.Add(1))
	n := c.Next(t)
	if n.IsZero() || t.Equal(n) {
		return t, true
	}

	return from, false
}

func isCronPeriodic(c *cronexpr.Expression) bool {
	// NOTE: `from.Add(1)` is needed as `from` represents zero time and cronexpr returns empty list if it is provided
	//       with zero time as first parameter. Non-standard cron expressions (Quartz) may include year field
	//       representing time range which might be in the past hence the `from` time is set to zero time.
	from := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	var numOfCalculatedTimes uint = 2

	return len(c.NextN(from.Add(1), numOfCalculatedTimes)) == int(numOfCalculatedTimes)
}

type ScheduleState int8

const (
	// ScheduleStateDisabled means the ScanConfig is disabled no Scan needs to be scheduled for it.
	ScheduleStateDisabled ScheduleState = iota
	// ScheduleStateUnscheduled means the ScanConfig does not define a schedule.
	ScheduleStateUnscheduled
	// ScheduleStateNotDue means the OperationTime for ScanConfig is in the future based on the current ScheduleWindow.
	ScheduleStateNotDue
	// ScheduleStateDue means the OperationTime for ScanConfig is in the current ScheduleWindow.
	ScheduleStateDue
	// ScheduleStateOverdue means the OperationTime for ScanConfig is in the past compared to
	// the current ScheduleWindow, but it has recurring schedule, therefore the new OperationTime needs
	// to be calculated from the cron expression.
	ScheduleStateOverdue
)

func (s ScheduleState) String() string {
	switch s {
	case ScheduleStateDisabled:
		return "Disabled"
	case ScheduleStateUnscheduled:
		return "Unscheduled"
	case ScheduleStateNotDue:
		return "NotDue"
	case ScheduleStateDue:
		return "Due"
	case ScheduleStateOverdue:
		return "OverDue"
	default:
		return "Unknown"
	}
}

// ScanConfigSchedule defines the state of scheduling based on the OperationTime and the ScheduleWindow where former is
// calculated using the ScanConfig schedule.
type ScanConfigSchedule struct {
	State         ScheduleState
	OperationTime *OperationTime
	Window        *ScheduleWindow
}

func (w ScanConfigSchedule) String() string {
	return fmt.Sprintf("state: %s, operation time: [%s], schedule window: [%s]", w.State, w.OperationTime, w.Window)
}

// NewScanConfigSchedule returns a ScanConfigSchedule using the provided scanConfig apitypes.ScanConfig and
// window ScheduleWindow.
// nolint:cyclop
func NewScanConfigSchedule(scanConfig *apitypes.ScanConfig, window *ScheduleWindow) (*ScanConfigSchedule, error) {
	if scanConfig.Disabled != nil && *scanConfig.Disabled {
		return &ScanConfigSchedule{
			State:  ScheduleStateDisabled,
			Window: window,
		}, nil
	}

	if scanConfig.Scheduled == nil || (scanConfig.Scheduled.CronLine == nil && scanConfig.Scheduled.OperationTime == nil) {
		return &ScanConfigSchedule{
			State:  ScheduleStateUnscheduled,
			Window: window,
		}, nil
	}

	var cronExpr *cronexpr.Expression
	var err error
	if scanConfig.Scheduled.CronLine != nil {
		cronExpr, err = cronexpr.Parse(*scanConfig.Scheduled.CronLine)
		if err != nil {
			return nil, fmt.Errorf("failed to parse cron expression %s: %w", *scanConfig.Scheduled.CronLine, err)
		}
	}

	var oTime time.Time
	if scanConfig.Scheduled.OperationTime != nil {
		oTime = *scanConfig.Scheduled.OperationTime
	}

	operationTime := NewOperationTime(oTime, cronExpr)

	if window.In(operationTime.Time()) {
		return &ScanConfigSchedule{
			State:         ScheduleStateDue,
			OperationTime: operationTime,
			Window:        window,
		}, nil
	}

	// Check whether the operationTime is before the window meaning that it is in the past compared to the window.
	if window.After(operationTime.Time()) {
		// If the operationTime is not a recurring that means it represents a one time event in the past,
		// therefore it is done.
		if !operationTime.IsRecurring() {
			return &ScanConfigSchedule{
				State:         ScheduleStateUnscheduled,
				OperationTime: operationTime,
				Window:        window,
			}, nil
		}
		// The operationTime is recurring which means we unexpectedly skipped some iterations therefore
		// it needs to be scheduled for the next ScheduleWindow.
		return &ScanConfigSchedule{
			State:         ScheduleStateOverdue,
			OperationTime: operationTime,
			Window:        window,
		}, nil
	}

	return &ScanConfigSchedule{
		State:         ScheduleStateNotDue,
		OperationTime: operationTime,
		Window:        window,
	}, nil
}
