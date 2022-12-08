// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/openclarity/kubeclarity/api/server/models"
)

// PutRuntimeScanStopCreatedCode is the HTTP code returned for type PutRuntimeScanStopCreated
const PutRuntimeScanStopCreatedCode int = 201

/*PutRuntimeScanStopCreated Success

swagger:response putRuntimeScanStopCreated
*/
type PutRuntimeScanStopCreated struct {

	/*
	  In: Body
	*/
	Payload interface{} `json:"body,omitempty"`
}

// NewPutRuntimeScanStopCreated creates PutRuntimeScanStopCreated with default headers values
func NewPutRuntimeScanStopCreated() *PutRuntimeScanStopCreated {

	return &PutRuntimeScanStopCreated{}
}

// WithPayload adds the payload to the put runtime scan stop created response
func (o *PutRuntimeScanStopCreated) WithPayload(payload interface{}) *PutRuntimeScanStopCreated {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put runtime scan stop created response
func (o *PutRuntimeScanStopCreated) SetPayload(payload interface{}) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutRuntimeScanStopCreated) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(201)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// PutRuntimeScanStopBadRequestCode is the HTTP code returned for type PutRuntimeScanStopBadRequest
const PutRuntimeScanStopBadRequestCode int = 400

/*PutRuntimeScanStopBadRequest Scan failed to stop

swagger:response putRuntimeScanStopBadRequest
*/
type PutRuntimeScanStopBadRequest struct {

	/*
	  In: Body
	*/
	Payload string `json:"body,omitempty"`
}

// NewPutRuntimeScanStopBadRequest creates PutRuntimeScanStopBadRequest with default headers values
func NewPutRuntimeScanStopBadRequest() *PutRuntimeScanStopBadRequest {

	return &PutRuntimeScanStopBadRequest{}
}

// WithPayload adds the payload to the put runtime scan stop bad request response
func (o *PutRuntimeScanStopBadRequest) WithPayload(payload string) *PutRuntimeScanStopBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put runtime scan stop bad request response
func (o *PutRuntimeScanStopBadRequest) SetPayload(payload string) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutRuntimeScanStopBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

/*PutRuntimeScanStopDefault unknown error

swagger:response putRuntimeScanStopDefault
*/
type PutRuntimeScanStopDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.APIResponse `json:"body,omitempty"`
}

// NewPutRuntimeScanStopDefault creates PutRuntimeScanStopDefault with default headers values
func NewPutRuntimeScanStopDefault(code int) *PutRuntimeScanStopDefault {
	if code <= 0 {
		code = 500
	}

	return &PutRuntimeScanStopDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the put runtime scan stop default response
func (o *PutRuntimeScanStopDefault) WithStatusCode(code int) *PutRuntimeScanStopDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the put runtime scan stop default response
func (o *PutRuntimeScanStopDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the put runtime scan stop default response
func (o *PutRuntimeScanStopDefault) WithPayload(payload *models.APIResponse) *PutRuntimeScanStopDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put runtime scan stop default response
func (o *PutRuntimeScanStopDefault) SetPayload(payload *models.APIResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutRuntimeScanStopDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}