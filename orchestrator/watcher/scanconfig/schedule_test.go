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
	"testing"
	"time"

	"github.com/aptible/supercronic/cronexpr"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

func TestOperationTime(t *testing.T) {
	tests := []struct {
		Name      string
		Time      time.Time
		AfterTime time.Time
		Cron      *cronexpr.Expression

		ExpectedTime      time.Time
		ExpectedNextTime  time.Time
		ExpectedAfterTime time.Time

		ExpectedRecurringMatcher types.GomegaMatcher
	}{
		{
			Name:                     "Operation time with no cron",
			Time:                     time.Date(2023, 5, 17, 10, 0, 0, 0, time.UTC),
			AfterTime:                time.Date(2023, 5, 17, 12, 0, 0, 0, time.UTC),
			Cron:                     nil,
			ExpectedTime:             time.Date(2023, 5, 17, 10, 0, 0, 0, time.UTC),
			ExpectedNextTime:         time.Date(2023, 5, 17, 10, 0, 0, 0, time.UTC),
			ExpectedAfterTime:        time.Date(2023, 5, 17, 10, 0, 0, 0, time.UTC),
			ExpectedRecurringMatcher: BeFalse(),
		},
		{
			Name:                     "Operation time with single point in time cron",
			Time:                     time.Date(2023, 5, 17, 10, 0, 0, 0, time.UTC),
			AfterTime:                time.Date(2023, 5, 17, 12, 0, 0, 0, time.UTC),
			Cron:                     cronexpr.MustParse("0 11 17 5 * 2024"),
			ExpectedTime:             time.Date(2024, 5, 17, 11, 0, 0, 0, time.UTC),
			ExpectedNextTime:         time.Date(2024, 5, 17, 11, 0, 0, 0, time.UTC),
			ExpectedAfterTime:        time.Date(2024, 5, 17, 11, 0, 0, 0, time.UTC),
			ExpectedRecurringMatcher: BeFalse(),
		},
		{
			Name:                     "Operation time with periodic cron - 1",
			Time:                     time.Date(2023, 5, 17, 10, 0, 0, 0, time.UTC),
			AfterTime:                time.Date(2023, 5, 18, 12, 0, 0, 0, time.UTC),
			Cron:                     cronexpr.MustParse("0 11 * * * *"),
			ExpectedTime:             time.Date(2023, 5, 17, 10, 0, 0, 0, time.UTC),
			ExpectedNextTime:         time.Date(2023, 5, 17, 11, 0, 0, 0, time.UTC),
			ExpectedAfterTime:        time.Date(2023, 5, 19, 11, 0, 0, 0, time.UTC),
			ExpectedRecurringMatcher: BeTrue(),
		},
		{
			Name:                     "Operation time with periodic cron - 2",
			Time:                     time.Date(2023, 5, 17, 10, 0, 0, 0, time.UTC),
			AfterTime:                time.Date(2023, 5, 18, 12, 0, 0, 0, time.UTC),
			Cron:                     cronexpr.MustParse("0 */8 * * * *"),
			ExpectedTime:             time.Date(2023, 5, 17, 10, 0, 0, 0, time.UTC),
			ExpectedNextTime:         time.Date(2023, 5, 17, 16, 0, 0, 0, time.UTC),
			ExpectedAfterTime:        time.Date(2023, 5, 18, 16, 0, 0, 0, time.UTC),
			ExpectedRecurringMatcher: BeTrue(),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			op := NewOperationTime(test.Time, test.Cron)

			g.Expect(op.Time()).Should(Equal(test.ExpectedTime))
			g.Expect(op.Next().Time()).Should(Equal(test.ExpectedNextTime))
			g.Expect(op.NextAfter(test.AfterTime).Time()).Should(Equal(test.ExpectedAfterTime))
			g.Expect(op.IsRecurring()).Should(test.ExpectedRecurringMatcher)
		})
	}
}

func TestScanConfigSchedule(t *testing.T) {
	tests := []struct {
		Name       string
		ScanConfig *apitypes.ScanConfig
		Schedule   *ScheduleWindow

		ExpectedState        ScheduleState
		ExpectedErrorMatcher types.GomegaMatcher

		ExpectedOperationTimeMatcher  types.GomegaMatcher
		ExpectedScheduleWindowMatcher types.GomegaMatcher
	}{
		{
			Name: "Disabled field is true",
			ScanConfig: &apitypes.ScanConfig{
				Disabled: to.Ptr(true),
			},
			ExpectedState:                 ScheduleStateDisabled,
			ExpectedErrorMatcher:          Not(HaveOccurred()),
			ExpectedOperationTimeMatcher:  BeNil(),
			ExpectedScheduleWindowMatcher: BeNil(),
		},
		{
			Name:                          "Scheduled is nil",
			ScanConfig:                    &apitypes.ScanConfig{},
			ExpectedState:                 ScheduleStateUnscheduled,
			ExpectedErrorMatcher:          Not(HaveOccurred()),
			ExpectedOperationTimeMatcher:  BeNil(),
			ExpectedScheduleWindowMatcher: BeNil(),
		},
		{
			Name: "OperationTime and CronLine are nil",
			ScanConfig: &apitypes.ScanConfig{
				Scheduled: &apitypes.RuntimeScheduleScanConfig{},
			},
			ExpectedState:                 ScheduleStateUnscheduled,
			ExpectedErrorMatcher:          Not(HaveOccurred()),
			ExpectedOperationTimeMatcher:  BeNil(),
			ExpectedScheduleWindowMatcher: BeNil(),
		},
		{
			Name: "Only OperationTime is set in the present",
			ScanConfig: &apitypes.ScanConfig{
				Scheduled: &apitypes.RuntimeScheduleScanConfig{
					CronLine:      nil,
					OperationTime: to.Ptr(time.Now()),
				},
			},
			Schedule:                      NewScheduleWindow(time.Now(), 5*time.Minute),
			ExpectedState:                 ScheduleStateDue,
			ExpectedErrorMatcher:          Not(HaveOccurred()),
			ExpectedOperationTimeMatcher:  Not(BeNil()),
			ExpectedScheduleWindowMatcher: Not(BeNil()),
		},
		{
			Name: "OperationTime and CronLine with exact time are set to present",
			ScanConfig: &apitypes.ScanConfig{
				Scheduled: &apitypes.RuntimeScheduleScanConfig{
					CronLine:      to.Ptr("0 11 17 5 * 2023"),
					OperationTime: to.Ptr(time.Date(2023, 5, 17, 11, 0, 0, 0, time.UTC)),
				},
			},
			Schedule:                      NewScheduleWindow(time.Date(2023, 5, 17, 11, 0, 0, 0, time.UTC), 5*time.Minute),
			ExpectedState:                 ScheduleStateDue,
			ExpectedErrorMatcher:          Not(HaveOccurred()),
			ExpectedOperationTimeMatcher:  Not(BeNil()),
			ExpectedScheduleWindowMatcher: Not(BeNil()),
		},
		{
			Name: "OperationTime is set in the past and recurring CronLine is provided",
			ScanConfig: &apitypes.ScanConfig{
				Scheduled: &apitypes.RuntimeScheduleScanConfig{
					CronLine:      to.Ptr("0 2 * * *"),
					OperationTime: to.Ptr(time.Date(2023, 4, 17, 11, 0, 0, 0, time.UTC)),
				},
			},
			Schedule:                      NewScheduleWindow(time.Date(2023, 5, 17, 11, 0, 0, 0, time.UTC), 5*time.Minute),
			ExpectedState:                 ScheduleStateOverdue,
			ExpectedErrorMatcher:          Not(HaveOccurred()),
			ExpectedOperationTimeMatcher:  Not(BeNil()),
			ExpectedScheduleWindowMatcher: Not(BeNil()),
		},
		{
			Name: "OperationTime is set in the past and no CronLine is provided",
			ScanConfig: &apitypes.ScanConfig{
				Scheduled: &apitypes.RuntimeScheduleScanConfig{
					OperationTime: to.Ptr(time.Date(2023, 4, 17, 11, 0, 0, 0, time.UTC)),
				},
			},
			Schedule:                      NewScheduleWindow(time.Date(2023, 5, 17, 11, 0, 0, 0, time.UTC), 5*time.Minute),
			ExpectedState:                 ScheduleStateUnscheduled,
			ExpectedErrorMatcher:          Not(HaveOccurred()),
			ExpectedOperationTimeMatcher:  Not(BeNil()),
			ExpectedScheduleWindowMatcher: Not(BeNil()),
		},
		{
			Name: "OperationTime is set in the future and no CronLine is provided",
			ScanConfig: &apitypes.ScanConfig{
				Scheduled: &apitypes.RuntimeScheduleScanConfig{
					OperationTime: to.Ptr(time.Date(2023, 6, 17, 11, 0, 0, 0, time.UTC)),
				},
			},
			Schedule:                      NewScheduleWindow(time.Date(2023, 5, 17, 11, 0, 0, 0, time.UTC), 5*time.Minute),
			ExpectedState:                 ScheduleStateNotDue,
			ExpectedErrorMatcher:          Not(HaveOccurred()),
			ExpectedOperationTimeMatcher:  Not(BeNil()),
			ExpectedScheduleWindowMatcher: Not(BeNil()),
		},
		{
			Name: "OperationTime is set in the future and recurring CronLine is provided",
			ScanConfig: &apitypes.ScanConfig{
				Scheduled: &apitypes.RuntimeScheduleScanConfig{
					CronLine:      to.Ptr("0 2 * * *"),
					OperationTime: to.Ptr(time.Date(2023, 5, 17, 12, 0, 0, 0, time.UTC)),
				},
			},
			Schedule:                      NewScheduleWindow(time.Date(2023, 5, 17, 11, 0, 0, 0, time.UTC), 5*time.Minute),
			ExpectedState:                 ScheduleStateNotDue,
			ExpectedErrorMatcher:          Not(HaveOccurred()),
			ExpectedOperationTimeMatcher:  Not(BeNil()),
			ExpectedScheduleWindowMatcher: Not(BeNil()),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			s, err := NewScanConfigSchedule(test.ScanConfig, test.Schedule)
			g.Expect(err).Should(test.ExpectedErrorMatcher)
			g.Expect(s.State).Should(Equal(test.ExpectedState))
			g.Expect(s.OperationTime).Should(test.ExpectedOperationTimeMatcher)
			g.Expect(s.Window).Should(test.ExpectedScheduleWindowMatcher)
		})
	}
}

func TestScheduleWindow(t *testing.T) {
	tests := []struct {
		Name     string
		FromTime time.Time
		Size     time.Duration
		InTime   time.Time

		ExpectedToBeInWindow bool
		ExpectedStart        time.Time
		ExpectedEnd          time.Time

		ExpectedNextStart time.Time
		ExpectedNextEnd   time.Time
	}{
		{
			Name:                 "Schedule window",
			FromTime:             time.Date(2023, 5, 17, 10, 10, 0, 0, time.UTC),
			Size:                 6 * time.Minute,
			InTime:               time.Date(2023, 5, 17, 10, 11, 0, 0, time.UTC),
			ExpectedToBeInWindow: true,
			ExpectedStart:        time.Date(2023, 5, 17, 10, 7, 0, 0, time.UTC),
			ExpectedEnd:          time.Date(2023, 5, 17, 10, 13, 0, 0, time.UTC),
			ExpectedNextStart:    time.Date(2023, 5, 17, 10, 13, 0, 0, time.UTC),
			ExpectedNextEnd:      time.Date(2023, 5, 17, 10, 19, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			w := NewScheduleWindow(test.FromTime, test.Size)

			g.Expect(w.In(test.FromTime)).Should(Equal(test.ExpectedToBeInWindow))
			g.Expect(w.Start()).Should(Equal(test.ExpectedStart))
			g.Expect(w.End()).Should(Equal(test.ExpectedEnd))
			g.Expect(w.Next().Start()).Should(Equal(test.ExpectedNextStart))
			g.Expect(w.Next().End()).Should(Equal(test.ExpectedNextEnd))
		})
	}
}

func TestCronHelpers(t *testing.T) {
	tests := []struct {
		Name     string
		FromTime time.Time
		Cron     *cronexpr.Expression

		ExpectedNextTime             time.Time
		ExpectedIsPointInTimeMatcher types.GomegaMatcher
		ExpectedCronTime             time.Time
		ExpectedIsRecurringMatcher   types.GomegaMatcher
	}{
		{
			Name:                         "Single point in time in future",
			FromTime:                     time.Date(2023, 5, 17, 10, 10, 0, 0, time.UTC),
			Cron:                         cronexpr.MustParse("0 11 17 5 * 2024"),
			ExpectedNextTime:             time.Date(2024, 5, 17, 11, 0, 0, 0, time.UTC),
			ExpectedIsPointInTimeMatcher: BeTrue(),
			ExpectedCronTime:             time.Date(2024, 5, 17, 11, 0, 0, 0, time.UTC),
			ExpectedIsRecurringMatcher:   BeFalse(),
		},
		{
			Name:                         "Single point in time in past",
			FromTime:                     time.Date(2023, 5, 17, 10, 10, 0, 0, time.UTC),
			Cron:                         cronexpr.MustParse("0 11 17 5 * 2022"),
			ExpectedNextTime:             time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
			ExpectedIsPointInTimeMatcher: BeTrue(),
			ExpectedCronTime:             time.Date(2022, 5, 17, 11, 0, 0, 0, time.UTC),
			ExpectedIsRecurringMatcher:   BeFalse(),
		},
		{
			Name:                         "Recurring",
			FromTime:                     time.Date(2023, 5, 17, 10, 10, 0, 0, time.UTC),
			Cron:                         cronexpr.MustParse("0 */2 * * *"),
			ExpectedNextTime:             time.Date(2023, 5, 17, 12, 0, 0, 0, time.UTC),
			ExpectedIsPointInTimeMatcher: BeFalse(),
			ExpectedCronTime:             time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
			ExpectedIsRecurringMatcher:   BeTrue(),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			next := test.Cron.Next(test.FromTime)
			cronTime, ok := isCronPointInTime(test.Cron)
			recurring := isCronPeriodic(test.Cron)

			g.Expect(next).Should(Equal(test.ExpectedNextTime))
			g.Expect(ok).Should(test.ExpectedIsPointInTimeMatcher)
			g.Expect(cronTime).Should(Equal(test.ExpectedCronTime))
			g.Expect(recurring).Should(test.ExpectedIsRecurringMatcher)
		})
	}
}
