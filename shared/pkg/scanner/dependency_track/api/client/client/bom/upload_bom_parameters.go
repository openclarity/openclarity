// Code generated by go-swagger; DO NOT EDIT.

package bom

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// NewUploadBomParams creates a new UploadBomParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewUploadBomParams() *UploadBomParams {
	return &UploadBomParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewUploadBomParamsWithTimeout creates a new UploadBomParams object
// with the ability to set a timeout on a request.
func NewUploadBomParamsWithTimeout(timeout time.Duration) *UploadBomParams {
	return &UploadBomParams{
		timeout: timeout,
	}
}

// NewUploadBomParamsWithContext creates a new UploadBomParams object
// with the ability to set a context for a request.
func NewUploadBomParamsWithContext(ctx context.Context) *UploadBomParams {
	return &UploadBomParams{
		Context: ctx,
	}
}

// NewUploadBomParamsWithHTTPClient creates a new UploadBomParams object
// with the ability to set a custom HTTPClient for a request.
func NewUploadBomParamsWithHTTPClient(client *http.Client) *UploadBomParams {
	return &UploadBomParams{
		HTTPClient: client,
	}
}

/* UploadBomParams contains all the parameters to send to the API endpoint
   for the upload bom operation.

   Typically these are written to a http.Request.
*/
type UploadBomParams struct {

	// AutoCreate.
	AutoCreate *bool

	// Bom.
	Bom *string

	// Project.
	Project *string

	// ProjectName.
	ProjectName *string

	// ProjectVersion.
	ProjectVersion *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the upload bom params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *UploadBomParams) WithDefaults() *UploadBomParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the upload bom params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *UploadBomParams) SetDefaults() {
	var (
		autoCreateDefault = bool(false)
	)

	val := UploadBomParams{
		AutoCreate: &autoCreateDefault,
	}

	val.timeout = o.timeout
	val.Context = o.Context
	val.HTTPClient = o.HTTPClient
	*o = val
}

// WithTimeout adds the timeout to the upload bom params
func (o *UploadBomParams) WithTimeout(timeout time.Duration) *UploadBomParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the upload bom params
func (o *UploadBomParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the upload bom params
func (o *UploadBomParams) WithContext(ctx context.Context) *UploadBomParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the upload bom params
func (o *UploadBomParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the upload bom params
func (o *UploadBomParams) WithHTTPClient(client *http.Client) *UploadBomParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the upload bom params
func (o *UploadBomParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithAutoCreate adds the autoCreate to the upload bom params
func (o *UploadBomParams) WithAutoCreate(autoCreate *bool) *UploadBomParams {
	o.SetAutoCreate(autoCreate)
	return o
}

// SetAutoCreate adds the autoCreate to the upload bom params
func (o *UploadBomParams) SetAutoCreate(autoCreate *bool) {
	o.AutoCreate = autoCreate
}

// WithBom adds the bom to the upload bom params
func (o *UploadBomParams) WithBom(bom *string) *UploadBomParams {
	o.SetBom(bom)
	return o
}

// SetBom adds the bom to the upload bom params
func (o *UploadBomParams) SetBom(bom *string) {
	o.Bom = bom
}

// WithProject adds the project to the upload bom params
func (o *UploadBomParams) WithProject(project *string) *UploadBomParams {
	o.SetProject(project)
	return o
}

// SetProject adds the project to the upload bom params
func (o *UploadBomParams) SetProject(project *string) {
	o.Project = project
}

// WithProjectName adds the projectName to the upload bom params
func (o *UploadBomParams) WithProjectName(projectName *string) *UploadBomParams {
	o.SetProjectName(projectName)
	return o
}

// SetProjectName adds the projectName to the upload bom params
func (o *UploadBomParams) SetProjectName(projectName *string) {
	o.ProjectName = projectName
}

// WithProjectVersion adds the projectVersion to the upload bom params
func (o *UploadBomParams) WithProjectVersion(projectVersion *string) *UploadBomParams {
	o.SetProjectVersion(projectVersion)
	return o
}

// SetProjectVersion adds the projectVersion to the upload bom params
func (o *UploadBomParams) SetProjectVersion(projectVersion *string) {
	o.ProjectVersion = projectVersion
}

// WriteToRequest writes these params to a swagger request
func (o *UploadBomParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.AutoCreate != nil {

		// form param autoCreate
		var frAutoCreate bool
		if o.AutoCreate != nil {
			frAutoCreate = *o.AutoCreate
		}
		fAutoCreate := swag.FormatBool(frAutoCreate)
		if fAutoCreate != "" {
			if err := r.SetFormParam("autoCreate", fAutoCreate); err != nil {
				return err
			}
		}
	}

	if o.Bom != nil {

		// form param bom
		var frBom string
		if o.Bom != nil {
			frBom = *o.Bom
		}
		fBom := frBom
		if fBom != "" {
			if err := r.SetFormParam("bom", fBom); err != nil {
				return err
			}
		}
	}

	if o.Project != nil {

		// form param project
		var frProject string
		if o.Project != nil {
			frProject = *o.Project
		}
		fProject := frProject
		if fProject != "" {
			if err := r.SetFormParam("project", fProject); err != nil {
				return err
			}
		}
	}

	if o.ProjectName != nil {

		// form param projectName
		var frProjectName string
		if o.ProjectName != nil {
			frProjectName = *o.ProjectName
		}
		fProjectName := frProjectName
		if fProjectName != "" {
			if err := r.SetFormParam("projectName", fProjectName); err != nil {
				return err
			}
		}
	}

	if o.ProjectVersion != nil {

		// form param projectVersion
		var frProjectVersion string
		if o.ProjectVersion != nil {
			frProjectVersion = *o.ProjectVersion
		}
		fProjectVersion := frProjectVersion
		if fProjectVersion != "" {
			if err := r.SetFormParam("projectVersion", fProjectVersion); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}