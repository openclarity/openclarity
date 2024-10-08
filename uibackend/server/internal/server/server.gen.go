// Package server provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.3.1-0.20240915195924-0502e95d86bb DO NOT EDIT.
package server

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/runtime"
	. "github.com/openclarity/openclarity/uibackend/types"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get a list of findings impact for the dashboard.
	// (GET /dashboard/findingsImpact)
	GetDashboardFindingsImpact(ctx echo.Context) error
	// Get a list of finding trends for all finding types.
	// (GET /dashboard/findingsTrends)
	GetDashboardFindingsTrends(ctx echo.Context, params GetDashboardFindingsTrendsParams) error
	// Get a list of riskiest assets for the dashboard.
	// (GET /dashboard/riskiestAssets)
	GetDashboardRiskiestAssets(ctx echo.Context) error
	// Get a list of riskiest regions for the dashboard.
	// (GET /dashboard/riskiestRegions)
	GetDashboardRiskiestRegions(ctx echo.Context) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetDashboardFindingsImpact converts echo context to params.
func (w *ServerInterfaceWrapper) GetDashboardFindingsImpact(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetDashboardFindingsImpact(ctx)
	return err
}

// GetDashboardFindingsTrends converts echo context to params.
func (w *ServerInterfaceWrapper) GetDashboardFindingsTrends(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetDashboardFindingsTrendsParams
	// ------------- Required query parameter "startTime" -------------

	err = runtime.BindQueryParameter("form", true, true, "startTime", ctx.QueryParams(), &params.StartTime)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter startTime: %s", err))
	}

	// ------------- Required query parameter "endTime" -------------

	err = runtime.BindQueryParameter("form", true, true, "endTime", ctx.QueryParams(), &params.EndTime)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter endTime: %s", err))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetDashboardFindingsTrends(ctx, params)
	return err
}

// GetDashboardRiskiestAssets converts echo context to params.
func (w *ServerInterfaceWrapper) GetDashboardRiskiestAssets(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetDashboardRiskiestAssets(ctx)
	return err
}

// GetDashboardRiskiestRegions converts echo context to params.
func (w *ServerInterfaceWrapper) GetDashboardRiskiestRegions(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetDashboardRiskiestRegions(ctx)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/dashboard/findingsImpact", wrapper.GetDashboardFindingsImpact)
	router.GET(baseURL+"/dashboard/findingsTrends", wrapper.GetDashboardFindingsTrends)
	router.GET(baseURL+"/dashboard/riskiestAssets", wrapper.GetDashboardRiskiestAssets)
	router.GET(baseURL+"/dashboard/riskiestRegions", wrapper.GetDashboardRiskiestRegions)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
