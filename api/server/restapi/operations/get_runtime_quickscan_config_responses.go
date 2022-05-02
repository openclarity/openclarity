// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/cisco-open/kubei/api/server/models"
)

// GetRuntimeQuickscanConfigOKCode is the HTTP code returned for type GetRuntimeQuickscanConfigOK
const GetRuntimeQuickscanConfigOKCode int = 200

/*GetRuntimeQuickscanConfigOK Success

swagger:response getRuntimeQuickscanConfigOK
*/
type GetRuntimeQuickscanConfigOK struct {

	/*
	  In: Body
	*/
	Payload *models.RuntimeQuickScanConfig `json:"body,omitempty"`
}

// NewGetRuntimeQuickscanConfigOK creates GetRuntimeQuickscanConfigOK with default headers values
func NewGetRuntimeQuickscanConfigOK() *GetRuntimeQuickscanConfigOK {

	return &GetRuntimeQuickscanConfigOK{}
}

// WithPayload adds the payload to the get runtime quickscan config o k response
func (o *GetRuntimeQuickscanConfigOK) WithPayload(payload *models.RuntimeQuickScanConfig) *GetRuntimeQuickscanConfigOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get runtime quickscan config o k response
func (o *GetRuntimeQuickscanConfigOK) SetPayload(payload *models.RuntimeQuickScanConfig) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetRuntimeQuickscanConfigOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetRuntimeQuickscanConfigDefault unknown error

swagger:response getRuntimeQuickscanConfigDefault
*/
type GetRuntimeQuickscanConfigDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.APIResponse `json:"body,omitempty"`
}

// NewGetRuntimeQuickscanConfigDefault creates GetRuntimeQuickscanConfigDefault with default headers values
func NewGetRuntimeQuickscanConfigDefault(code int) *GetRuntimeQuickscanConfigDefault {
	if code <= 0 {
		code = 500
	}

	return &GetRuntimeQuickscanConfigDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get runtime quickscan config default response
func (o *GetRuntimeQuickscanConfigDefault) WithStatusCode(code int) *GetRuntimeQuickscanConfigDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get runtime quickscan config default response
func (o *GetRuntimeQuickscanConfigDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get runtime quickscan config default response
func (o *GetRuntimeQuickscanConfigDefault) WithPayload(payload *models.APIResponse) *GetRuntimeQuickscanConfigDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get runtime quickscan config default response
func (o *GetRuntimeQuickscanConfigDefault) SetPayload(payload *models.APIResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetRuntimeQuickscanConfigDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}