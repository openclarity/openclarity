package metrics

import (
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/openclarity/kubeclarity/api/server/restapi/operations"
	"github.com/openclarity/kubeclarity/backend/pkg/database"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

const (
	// Prefix is to make metrics unique
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
		Name: Prefix + "_number_of_vulnerabilities_total",
		Help: "The total number of vulnerabilities",
	})
	fixableVulnerabilityCounter = promauto.NewGauge(prometheus.GaugeOpts{
		Name: Prefix + "_number_of_fixable_vulnerabilities_total",
		Help: "The total number of fixable vulnerabilities",
	})
	fixableVulnerability = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: Prefix + "_number_of_fixable_vulnerabilities",
		Help: "The number of fixable vulnerabilities per severity",
	}, []string{"severity"})
	vulnerability = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: Prefix + "_number_of_vulnerabilities",
		Help: "The number of vulnerabilities per severity",
	}, []string{"severity"})
	trendGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: Prefix + "_vulnerability_trend",
		Help: "Incoming vulnerabilities within the last hour",
	}, []string{"severity"})
	applicationGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: Prefix + "_application_vulnerability",
		Help: "Count of vulnerabilities per application and level",
	}, []string{"name", "environment", "severity"})
)

type Server struct {
	refreshInterval int
	dbHandler       *database.Handler
}

func ProduceMetrics(dbHandler *database.Handler, refreshInterval int) *Server {
	return &Server{
		refreshInterval: refreshInterval,
		dbHandler:       dbHandler,
	}
}

func init() {
	logrus.Infof("Adding /metrics endpoint to http server assumed started with healthz")
	http.Handle("/metrics", promhttp.Handler())
}

func (s *Server) recordSummaryCounters() {
	pkgCount, _ := s.dbHandler.PackageTable().Count(nil)
	packageCounter.Set(float64(pkgCount))
	appCount, _ := s.dbHandler.ApplicationTable().Count(nil)
	applicationCounter.Set(float64(appCount))
	resCount, _ := s.dbHandler.ResourceTable().Count(nil)
	resourceCounter.Set(float64(resCount))
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

// Gauges for fixable vulnerabilities
func (s *Server) recordFixableVulnerability() {
	fixableVulnerabilityCount, _ := s.dbHandler.VulnerabilityTable().CountVulnerabilitiesWithFix()
	var total uint32 = 0
	var totalWithFix uint32 = 0
	for _, fix := range fixableVulnerabilityCount {
		total += fix.CountTotal
		totalWithFix += fix.CountWithFix
		fixableVulnerability.WithLabelValues(fmt.Sprintf("%s", fix.Severity)).Set(float64(fix.CountWithFix))
		vulnerability.WithLabelValues(fmt.Sprintf("%s", fix.Severity)).Set(float64(fix.CountTotal))
	}
	fixableVulnerabilityCounter.Set(float64(totalWithFix))
	vulnerabilityCounter.Set(float64(total))
}

func (s *Server) recordApplicationVulnerabilities() {
	sortDir := string("ASC")

	var applications, _, err = s.dbHandler.ApplicationTable().GetApplicationsAndTotal(
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

func (s *Server) StartRecordingMetrics() {
	go func() {
		for {
			s.recordFixableVulnerability()
			s.recordSummaryCounters()
			s.recordApplicationVulnerabilities()
			s.recordTrendCounters()

			time.Sleep(time.Duration(s.refreshInterval) * time.Second)
		}
	}()
}
