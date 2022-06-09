package rest

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/api/server/models"
	"github.com/openclarity/kubeclarity/api/server/restapi/operations"
	"github.com/openclarity/kubeclarity/backend/pkg/database"
)

func (s *Server) GetCISDockerBenchmarkResults(params operations.GetCisdockerbenchmarkresultsIDParams) middleware.Responder {
	cisDockerBenchmarkResults, total, err := s.dbHandler.CISDockerBenchmarkResultTable().GetCISDockerBenchmarkResultsAndTotal(params)
	if err != nil {
		log.Error(err)
		return operations.NewGetCisdockerbenchmarkresultsIDDefault(http.StatusInternalServerError).
			WithPayload(
				&models.APIResponse{
					Message: "Oops",
				},
			)
	}

	log.Debugf("GetCISDockerBenchmarkResults controller was invoked. "+
		"params=%+v, vulnerabilitiesView=%+v, total=%+v", params, cisDockerBenchmarkResults, total)

	cisDockerBenchmarkAssessments := make([]*models.CISDockerBenchmarkAssessment, len(cisDockerBenchmarkResults))
	for i := range cisDockerBenchmarkResults {
		cisDockerBenchmarkAssessments[i] = database.CISDockerBenchmarkResultFromDB(&cisDockerBenchmarkResults[i])
	}

	return operations.NewGetCisdockerbenchmarkresultsIDOK().WithPayload(
		&operations.GetCisdockerbenchmarkresultsIDOKBody{
			Items: cisDockerBenchmarkAssessments,
			Total: &total,
		})
}
