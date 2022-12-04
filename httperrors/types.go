package httperrors

import (
	"encoding/xml"
	"net/http"

	"github.com/vmihailenco/msgpack/v5"
)

// HTTPCode is used to define the HTTP error code.
type HTTPCode string

const (
	// HTTPCodeInvalidAuth is used when the authentication is invalid.
	HTTPCodeInvalidAuth = HTTPCode("invalid_auth")

	// HTTPCodeInternalServerError is used when there is an internal server error.
	HTTPCodeInternalServerError = HTTPCode("internal_server_error")

	// HTTPCodeSameNode is an error that is used when the user is on the same node for a cross-node specific request.
	HTTPCodeSameNode = HTTPCode("same_node")
)

// HTTPError is used to define a interface that represents an HTTP error.
type HTTPError interface {
	status() int
	errorCode() HTTPCode
}

// InvalidAuth is used to define an invalid authentication error.
type InvalidAuth struct {
	Message string `json:"message" xml:"message" msgpack:"message"`
}

func (i InvalidAuth) status() int { return http.StatusUnauthorized }

func (i InvalidAuth) errorCode() HTTPCode { return HTTPCodeInvalidAuth }

// InternalServerError is used to define an internal server error.
type InternalServerError struct{}

func (i InternalServerError) status() int { return http.StatusInternalServerError }

func (i InternalServerError) errorCode() HTTPCode { return HTTPCodeInternalServerError }

// MarshalJSON is used to marshal the error to JSON.
func (i InternalServerError) MarshalJSON() ([]byte, error) {
	return []byte(`{"message":"Internal Server Error"}`), nil
}

// MarshalXML is used to marshal the error to XML.
func (i InternalServerError) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(struct {
		Message string `xml:"message"`
	}{
		Message: "Internal Server Error",
	}, start)
}

// EncodeMsgpack is used to marshal the error to Msgpack.
func (i InternalServerError) EncodeMsgpack(e *msgpack.Encoder) error {
	return e.Encode(map[string]string{
		"message": "Internal Server Error",
	})
}

// SameNode is used to define an error that is used when the user is on the same node for a cross-node specific request.
type SameNode struct {
	Message string `json:"message" xml:"message" msgpack:"message"`
}

func (i SameNode) status() int         { return http.StatusBadRequest }
func (i SameNode) errorCode() HTTPCode { return HTTPCodeSameNode }
