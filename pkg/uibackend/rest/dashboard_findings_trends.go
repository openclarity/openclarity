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

package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"

	backendmodels "github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
	"github.com/openclarity/vmclarity/pkg/uibackend/api/models"
)

const (
	numOfTimePoints = 10
)

func (s *ServerImpl) GetDashboardFindingsTrends(ctx echo.Context, params models.GetDashboardFindingsTrendsParams) error {
	reqCtx := ctx.Request().Context()
	if err := validateParams(params); err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("Request params are not valid: %v", err))
	}

	times := createTimes(params)

	findingTypes := models.GetFindingTypes()
	errs := make(chan error, len(findingTypes))
	findingsTrendsChan := make(chan models.FindingTrends, len(findingTypes))

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

	var findingsTrends models.FindingsTrends
	for findingTrends := range findingsTrendsChan {
		findingsTrends = append(findingsTrends, findingTrends)
	}

	return sendResponse(ctx, http.StatusOK, findingsTrends)
}

func validateParams(params models.GetDashboardFindingsTrendsParams) error {
	if !params.StartTime.Before(params.EndTime) {
		return fmt.Errorf("start time must be before end time")
	}

	return nil
}

// createTimes returns a slice of points in time between endTime and startTime.
// the finding trends will be reported based on the amount of active findings in each time.
func createTimes(params models.GetDashboardFindingsTrendsParams) []time.Time {
	times := make([]time.Time, numOfTimePoints)
	timeBetweenPoints := params.EndTime.Sub(params.StartTime) / numOfTimePoints
	time := params.EndTime
	for i := numOfTimePoints - 1; i >= 0; i-- {
		times[i] = time
		time = time.Add(-timeBetweenPoints)
	}
	return times
}

func (s *ServerImpl) getFindingTrendsForFindingType(ctx context.Context, findingType models.FindingType, times []time.Time) (models.FindingTrends, error) {
	trends := make([]models.FindingTrend, len(times))
	for i, point := range times {
		trend, err := s.getFindingTrendPerPoint(ctx, findingType, point)
		if err != nil {
			return models.FindingTrends{}, fmt.Errorf("failed to get finding trend: %w", err)
		}
		trends[i] = trend
	}

	return models.FindingTrends{
		FindingType: &findingType,
		Trends:      &trends,
	}, nil
}

func (s *ServerImpl) getFindingTrendPerPoint(ctx context.Context, findingType models.FindingType, point time.Time) (models.FindingTrend, error) {
	// Count total findings for the given finding type that was active during the given time point.
	findings, err := s.BackendClient.GetFindings(ctx, backendmodels.GetFindingsParams{
		Count: utils.PointerTo(true),
		Filter: utils.PointerTo(fmt.Sprintf(
			"findingInfo/objectType eq '%s' and foundOn le %v and (invalidatedOn eq null or invalidatedOn gt %v)",
			getObjectType(findingType), point.Format(time.RFC3339), point.Format(time.RFC3339))),
		// Select the smallest amount of data to return in items, we only care about the count.
		Select: utils.PointerTo("id"),
		Top:    utils.PointerTo(0),
	})
	if err != nil {
		return models.FindingTrend{}, fmt.Errorf("failed to get findings for the given point: %w", err)
	}

	return models.FindingTrend{
		Count: findings.Count,
		Time:  &point,
	}, nil
}

func getObjectType(findingType models.FindingType) string {
	switch findingType {
	case models.EXPLOIT:
		return "Exploit"
	case models.MALWARE:
		return "Malware"
	case models.MISCONFIGURATION:
		return "Misconfiguration"
	case models.PACKAGE:
		return "Package"
	case models.ROOTKIT:
		return "Rootkit"
	case models.SECRET:
		return "Secret"
	case models.VULNERABILITY:
		return "Vulnerability"
	}

	// Should not happen.
	panic("unsupported object type")
}
