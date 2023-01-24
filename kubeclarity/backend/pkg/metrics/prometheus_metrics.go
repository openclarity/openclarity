// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package metrics

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/openclarity/kubeclarity/backend/pkg/database"

	"github.com/go-openapi/strfmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/api/server/restapi/operations"
)

const (
	// Prefix is to make metrics unique.
	Prefix string = "kubeclarity"
)

var (
	applicationCounter = promauto.NewGauge(prometheus.GaugeOpts{
		Name: Prefix + "_number_of_applications",
		Help: "The total number of applications",
	})
	resourceCounter = promauto.NewGauge(prometheus.GaugeOpts{
		Name: Prefix + "_number_of_resources",
		Help: "The total number of resources",
	})
	packageCounter = promauto.NewGauge(prometheus.GaugeOpts{
		Name: Prefix + "_number_of_packages",
		Help: "The total number of packages",
	})
	vulnerabilityCounter = promauto.NewGauge(prometheus.GaugeOpts{
		Name: Prefix + "_number_of_vulnerabilities_amount",
		Help: "The total number of vulnerabilities",
	})
	fixableVulnerabilityCounter = promauto.NewGauge(prometheus.GaugeOpts{
		Name: Prefix + "_number_of_fixable_vulnerabilities_amount",
		Help: "The total number of fixable vulnerabilities",
	})
	fixableVulnerability = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: Prefix + "_number_of_fixable_vulnerabilities",
		Help: "The number of fixable vulnerabilities per severity",
	}, []string{"vul_severity"})
	vulnerability = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: Prefix + "_number_of_vulnerabilities",
		Help: "The number of vulnerabilities per severity",
	}, []string{"vul_severity"})
	trendGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: Prefix + "_vulnerability_trend",
		Help: "Vulnerability trend in a 60 minute time window",
	}, []string{"vul_severity"})
	applicationGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: Prefix + "_application_vulnerability",
		Help: "Count of vulnerabilities per application, environment and severity",
	}, []string{"name", "environment", "vul_severity"})
)

type Server struct {
	refreshInterval int
	dbHandler       *database.Handler
}

func CreateMetrics(dbHandler *database.Handler, refreshInterval int) *Server {
	logrus.Infof("Adding /metrics endpoint to http server assumed started with healthz")
	http.Handle("/metrics", promhttp.Handler())
	return &Server{
		refreshInterval: refreshInterval,
		dbHandler:       dbHandler,
	}
}

func (s *Server) recordSummaryCounters() {
	if pkgCount, err := s.dbHandler.PackageTable().Count(nil); err != nil {
		logrus.Warnf("failed to get package count: %v", err)
	} else {
		packageCounter.Set(float64(pkgCount))
	}
	if appCount, err := s.dbHandler.ApplicationTable().Count(nil); err != nil {
		logrus.Warnf("failed to get application count: %v", err)
	} else {
		applicationCounter.Set(float64(appCount))
	}
	if resCount, err := s.dbHandler.ResourceTable().Count(nil); err != nil {
		logrus.Warnf("failed to get resource count: %v", err)
	} else {
		resourceCounter.Set(float64(resCount))
	}
}

func (s *Server) recordTrendCounters() {
	trends, err := s.dbHandler.NewVulnerabilityTable().
		GetNewVulnerabilitiesTrends(operations.GetDashboardTrendsVulnerabilitiesParams{
			StartTime: strfmt.DateTime(time.Now().Add(-60 * time.Minute)),
			EndTime:   strfmt.DateTime(time.Now()),
		})
	if err != nil {
		logrus.Warnf("failed get vulnernability trends: %v", err)
		return
	}
	for _, trend := range trends {
		for _, vul := range trend.NumOfVuls {
			trendGauge.WithLabelValues(string(vul.Severity)).Set(float64(vul.Count))
		}
	}
}

// Gauges for fixable vulnerabilities.
func (s *Server) recordFixableVulnerability() {
	fixableVulnerabilityCount, err := s.dbHandler.VulnerabilityTable().CountVulnerabilitiesWithFix()
	if err != nil {
		logrus.Warnf("failed to get count of fixable vulnerabilities: %v", err)
		return
	}

	var total uint32
	var totalWithFix uint32
	for _, fix := range fixableVulnerabilityCount {
		total += fix.CountTotal
		totalWithFix += fix.CountWithFix
		fixableVulnerability.WithLabelValues(string(fix.Severity)).Set(float64(fix.CountWithFix))
		vulnerability.WithLabelValues(string(fix.Severity)).Set(float64(fix.CountTotal))
	}
	fixableVulnerabilityCounter.Set(float64(totalWithFix))
	vulnerabilityCounter.Set(float64(total))
}

func (s *Server) recordApplicationVulnerabilities() {
	sortDir := string("ASC")

	applications, _, err := s.dbHandler.ApplicationTable().GetApplicationsAndTotal(
		database.GetApplicationsParams{
			GetApplicationsParams: operations.GetApplicationsParams{
				SortDir: &sortDir,
				SortKey: "applicationName",
			},
		})
	if err != nil {
		logrus.Warnf("failed get applications: %v", err)
		return
	}

	for _, application := range applications {
		for _, env := range strings.Split(application.Environments, "|") {
			if env != "" {
				applicationGauge.WithLabelValues(application.Name, env, "NEGLIGIBLE").Set(float64(application.TotalNegCount))
				applicationGauge.WithLabelValues(application.Name, env, "LOW").Set(float64(application.TotalLowCount))
				applicationGauge.WithLabelValues(application.Name, env, "MEDIUM").Set(float64(application.TotalMediumCount))
				applicationGauge.WithLabelValues(application.Name, env, "HIGH").Set(float64(application.TotalHighCount))
				applicationGauge.WithLabelValues(application.Name, env, "CRITICAL").Set(float64(application.TotalCriticalCount))
			}
		}
	}
}

func (s *Server) StartRecordingMetrics(ctx context.Context) {
	s.recordVulnerabilities()
	go func() {
		for {
			select {
			case <-ctx.Done():
				logrus.Info("received stop event")
				return
			case <-time.After(time.Duration(s.refreshInterval) * time.Second):
				s.recordVulnerabilities()
			}
		}
	}()
}

func (s *Server) recordVulnerabilities() {
	s.recordFixableVulnerability()
	s.recordSummaryCounters()
	s.recordApplicationVulnerabilities()
	s.recordTrendCounters()
}
