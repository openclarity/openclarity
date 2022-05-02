// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/cisco-open/kubei/api/server/restapi/operations"
)

//go:generate swagger generate server --target ../../server --name KubeClarityAPIs --spec ../../swagger.yaml --principal interface{}

func configureFlags(api *operations.KubeClarityAPIsAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.KubeClarityAPIsAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	if api.DeleteApplicationsIDHandler == nil {
		api.DeleteApplicationsIDHandler = operations.DeleteApplicationsIDHandlerFunc(func(params operations.DeleteApplicationsIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.DeleteApplicationsID has not yet been implemented")
		})
	}
	if api.GetApplicationResourcesHandler == nil {
		api.GetApplicationResourcesHandler = operations.GetApplicationResourcesHandlerFunc(func(params operations.GetApplicationResourcesParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetApplicationResources has not yet been implemented")
		})
	}
	if api.GetApplicationResourcesIDHandler == nil {
		api.GetApplicationResourcesIDHandler = operations.GetApplicationResourcesIDHandlerFunc(func(params operations.GetApplicationResourcesIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetApplicationResourcesID has not yet been implemented")
		})
	}
	if api.GetApplicationsHandler == nil {
		api.GetApplicationsHandler = operations.GetApplicationsHandlerFunc(func(params operations.GetApplicationsParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetApplications has not yet been implemented")
		})
	}
	if api.GetApplicationsIDHandler == nil {
		api.GetApplicationsIDHandler = operations.GetApplicationsIDHandlerFunc(func(params operations.GetApplicationsIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetApplicationsID has not yet been implemented")
		})
	}
	if api.GetDashboardCountersHandler == nil {
		api.GetDashboardCountersHandler = operations.GetDashboardCountersHandlerFunc(func(params operations.GetDashboardCountersParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetDashboardCounters has not yet been implemented")
		})
	}
	if api.GetDashboardMostVulnerableHandler == nil {
		api.GetDashboardMostVulnerableHandler = operations.GetDashboardMostVulnerableHandlerFunc(func(params operations.GetDashboardMostVulnerableParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetDashboardMostVulnerable has not yet been implemented")
		})
	}
	if api.GetDashboardPackagesPerLanguageHandler == nil {
		api.GetDashboardPackagesPerLanguageHandler = operations.GetDashboardPackagesPerLanguageHandlerFunc(func(params operations.GetDashboardPackagesPerLanguageParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetDashboardPackagesPerLanguage has not yet been implemented")
		})
	}
	if api.GetDashboardPackagesPerLicenseHandler == nil {
		api.GetDashboardPackagesPerLicenseHandler = operations.GetDashboardPackagesPerLicenseHandlerFunc(func(params operations.GetDashboardPackagesPerLicenseParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetDashboardPackagesPerLicense has not yet been implemented")
		})
	}
	if api.GetDashboardTrendsVulnerabilitiesHandler == nil {
		api.GetDashboardTrendsVulnerabilitiesHandler = operations.GetDashboardTrendsVulnerabilitiesHandlerFunc(func(params operations.GetDashboardTrendsVulnerabilitiesParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetDashboardTrendsVulnerabilities has not yet been implemented")
		})
	}
	if api.GetDashboardVulnerabilitiesWithFixHandler == nil {
		api.GetDashboardVulnerabilitiesWithFixHandler = operations.GetDashboardVulnerabilitiesWithFixHandlerFunc(func(params operations.GetDashboardVulnerabilitiesWithFixParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetDashboardVulnerabilitiesWithFix has not yet been implemented")
		})
	}
	if api.GetNamespacesHandler == nil {
		api.GetNamespacesHandler = operations.GetNamespacesHandlerFunc(func(params operations.GetNamespacesParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetNamespaces has not yet been implemented")
		})
	}
	if api.GetPackagesHandler == nil {
		api.GetPackagesHandler = operations.GetPackagesHandlerFunc(func(params operations.GetPackagesParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetPackages has not yet been implemented")
		})
	}
	if api.GetPackagesIDHandler == nil {
		api.GetPackagesIDHandler = operations.GetPackagesIDHandlerFunc(func(params operations.GetPackagesIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetPackagesID has not yet been implemented")
		})
	}
	if api.GetPackagesIDApplicationResourcesHandler == nil {
		api.GetPackagesIDApplicationResourcesHandler = operations.GetPackagesIDApplicationResourcesHandlerFunc(func(params operations.GetPackagesIDApplicationResourcesParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetPackagesIDApplicationResources has not yet been implemented")
		})
	}
	if api.GetRuntimeScanProgressHandler == nil {
		api.GetRuntimeScanProgressHandler = operations.GetRuntimeScanProgressHandlerFunc(func(params operations.GetRuntimeScanProgressParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetRuntimeScanProgress has not yet been implemented")
		})
	}
	if api.GetRuntimeScanResultsHandler == nil {
		api.GetRuntimeScanResultsHandler = operations.GetRuntimeScanResultsHandlerFunc(func(params operations.GetRuntimeScanResultsParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetRuntimeScanResults has not yet been implemented")
		})
	}
	if api.GetRuntimeScheduleScanConfigHandler == nil {
		api.GetRuntimeScheduleScanConfigHandler = operations.GetRuntimeScheduleScanConfigHandlerFunc(func(params operations.GetRuntimeScheduleScanConfigParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetRuntimeScheduleScanConfig has not yet been implemented")
		})
	}
	if api.GetVulnerabilitiesHandler == nil {
		api.GetVulnerabilitiesHandler = operations.GetVulnerabilitiesHandlerFunc(func(params operations.GetVulnerabilitiesParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetVulnerabilities has not yet been implemented")
		})
	}
	if api.GetVulnerabilitiesVulIDPkgIDHandler == nil {
		api.GetVulnerabilitiesVulIDPkgIDHandler = operations.GetVulnerabilitiesVulIDPkgIDHandlerFunc(func(params operations.GetVulnerabilitiesVulIDPkgIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetVulnerabilitiesVulIDPkgID has not yet been implemented")
		})
	}
	if api.PostApplicationsHandler == nil {
		api.PostApplicationsHandler = operations.PostApplicationsHandlerFunc(func(params operations.PostApplicationsParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PostApplications has not yet been implemented")
		})
	}
	if api.PostApplicationsContentAnalysisIDHandler == nil {
		api.PostApplicationsContentAnalysisIDHandler = operations.PostApplicationsContentAnalysisIDHandlerFunc(func(params operations.PostApplicationsContentAnalysisIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PostApplicationsContentAnalysisID has not yet been implemented")
		})
	}
	if api.PostApplicationsVulnerabilityScanIDHandler == nil {
		api.PostApplicationsVulnerabilityScanIDHandler = operations.PostApplicationsVulnerabilityScanIDHandlerFunc(func(params operations.PostApplicationsVulnerabilityScanIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PostApplicationsVulnerabilityScanID has not yet been implemented")
		})
	}
	if api.PutApplicationsIDHandler == nil {
		api.PutApplicationsIDHandler = operations.PutApplicationsIDHandlerFunc(func(params operations.PutApplicationsIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PutApplicationsID has not yet been implemented")
		})
	}
	if api.PutRuntimeScanStartHandler == nil {
		api.PutRuntimeScanStartHandler = operations.PutRuntimeScanStartHandlerFunc(func(params operations.PutRuntimeScanStartParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PutRuntimeScanStart has not yet been implemented")
		})
	}
	if api.PutRuntimeScanStopHandler == nil {
		api.PutRuntimeScanStopHandler = operations.PutRuntimeScanStopHandlerFunc(func(params operations.PutRuntimeScanStopParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PutRuntimeScanStop has not yet been implemented")
		})
	}
	if api.PutRuntimeScheduleScanConfigHandler == nil {
		api.PutRuntimeScheduleScanConfigHandler = operations.PutRuntimeScheduleScanConfigHandlerFunc(func(params operations.PutRuntimeScheduleScanConfigParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PutRuntimeScheduleScanConfig has not yet been implemented")
		})
	}
	if api.PutRuntimeScheduleScanStartHandler == nil {
		api.PutRuntimeScheduleScanStartHandler = operations.PutRuntimeScheduleScanStartHandlerFunc(func(params operations.PutRuntimeScheduleScanStartParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PutRuntimeScheduleScanStart has not yet been implemented")
		})
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
