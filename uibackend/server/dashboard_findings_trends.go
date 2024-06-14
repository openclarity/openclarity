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

package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/uibackend/types"
)

const (
	numOfTimePoints = 10
)

func (s *ServerImpl) GetDashboardFindingsTrends(ctx echo.Context, params types.GetDashboardFindingsTrendsParams) error {
	reqCtx := ctx.Request().Context()
	if err := validateParams(params); err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("Request params are not valid: %v", err))
	}

	times := createTimes(params)

	findingTypes := types.GetFindingTypes()
	errs := make(chan error, len(findingTypes))
	findingsTrendsChan := make(chan types.FindingTrends, len(findingTypes))

	var wg sync.WaitGroup
	for _, findingType := range findingTypes {
		ft := findingType
		wg.Add(1)
		go func() {
			defer wg.Done()
			trends, err := s.getFindingTrendsForFindingType(reqCtx, ft, times)
			if err != nil {
				errs <- fmt.Errorf("failed to get %s trends: %w", ft, err)
				return
			}
			findingsTrendsChan <- trends
		}()
	}
	wg.Wait()
	close(errs)
	close(findingsTrendsChan)

	var err error
	for e := range errs {
		if e != nil {
			err = errors.Join(err, e)
		}
	}
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, err.Error())
	}

	var findingsTrends types.FindingsTrends
	for findingTrends := range findingsTrendsChan {
		findingsTrends = append(findingsTrends, findingTrends)
	}

	return sendResponse(ctx, http.StatusOK, findingsTrends)
}

func validateParams(params types.GetDashboardFindingsTrendsParams) error {
	if !params.StartTime.Before(params.EndTime) {
		return errors.New("start time must be before end time")
	}

	return nil
}

// createTimes returns a slice of points in time between endTime and startTime.
// the finding trends will be reported based on the amount of active findings in each time.
func createTimes(params types.GetDashboardFindingsTrendsParams) []time.Time {
	times := make([]time.Time, numOfTimePoints)
	timeBetweenPoints := params.EndTime.Sub(params.StartTime) / numOfTimePoints
	t := params.EndTime
	for i := numOfTimePoints - 1; i >= 0; i-- {
		times[i] = t
		t = t.Add(-timeBetweenPoints)
	}
	return times
}

func (s *ServerImpl) getFindingTrendsForFindingType(ctx context.Context, findingType types.FindingType, times []time.Time) (types.FindingTrends, error) {
	trends := make([]types.FindingTrend, len(times))
	for i, point := range times {
		trend, err := s.getFindingTrendPerPoint(ctx, findingType, point)
		if err != nil {
			return types.FindingTrends{}, fmt.Errorf("failed to get finding trend: %w", err)
		}
		trends[i] = trend
	}

	return types.FindingTrends{
		FindingType: &findingType,
		Trends:      &trends,
	}, nil
}

func (s *ServerImpl) getFindingTrendPerPoint(ctx context.Context, findingType types.FindingType, point time.Time) (types.FindingTrend, error) {
	// Count total findings for the given finding type that was active during the given time point.
	findings, err := s.Client.GetAssetFindings(ctx, apitypes.GetAssetFindingsParams{
		Count: to.Ptr(true),
		Filter: to.Ptr(fmt.Sprintf(
			"finding/findingInfo/objectType eq '%s' and firstSeen le %v and (invalidatedOn eq null or invalidatedOn gt %v)",
			getObjectType(findingType), point.Format(time.RFC3339), point.Format(time.RFC3339))),
		// Select the smallest amount of data to return in items, we only care about the count.
		Select: to.Ptr("id"),
		Top:    to.Ptr(0),
	})
	if err != nil {
		return types.FindingTrend{}, fmt.Errorf("failed to get findings for the given point: %w", err)
	}

	return types.FindingTrend{
		Count: findings.Count,
		Time:  &point,
	}, nil
}

func getObjectType(findingType types.FindingType) string {
	switch findingType {
	case types.EXPLOIT:
		return "Exploit"
	case types.MALWARE:
		return "Malware"
	case types.MISCONFIGURATION:
		return "Misconfiguration"
	case types.PACKAGE:
		return "Package"
	case types.ROOTKIT:
		return "Rootkit"
	case types.SECRET:
		return "Secret"
	case types.VULNERABILITY:
		return "Vulnerability"
	}

	// Should not happen.
	panic("unsupported object type")
}
