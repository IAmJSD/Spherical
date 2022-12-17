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

	// HTTPCodeHalfAuthenticated is used when the user is half authenticated.
	HTTPCodeHalfAuthenticated = HTTPCode("half_authenticated")

	// HTTPCodeNoPasswordAuthSupport is used when the user does not support password authentication.
	HTTPCodeNoPasswordAuthSupport = HTTPCode("no_password_auth_support")

	// HTTPCodeInvalidBody is used when the body is invalid.
	HTTPCodeInvalidBody = HTTPCode("invalid_body")
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

// HalfAuthentication is thrown when the user is only partially authenticated. This is used for 2FA.
type HalfAuthentication struct {
	// SupportedMethods is the supported authentication methods.
	SupportedMethods []string `json:"supported_methods" xml:"supported_methods" msgpack:"supported_methods"`

	// HalfToken is the half token that can be used to complete the authentication.
	HalfToken string `json:"half_token" xml:"half_token" msgpack:"half_token"`
}

// Error implements the error interface.
func (h HalfAuthentication) Error() string { return "user half authenticated" }

func (h HalfAuthentication) status() int         { return http.StatusUnauthorized }
func (h HalfAuthentication) errorCode() HTTPCode { return HTTPCodeHalfAuthenticated }

// NoPasswordAuthSupport is thrown when the user does not support password authentication.
type NoPasswordAuthSupport struct {
	Message string `json:"message" xml:"message" msgpack:"message"`
}

func (n NoPasswordAuthSupport) status() int         { return http.StatusUnauthorized }
func (n NoPasswordAuthSupport) errorCode() HTTPCode { return HTTPCodeNoPasswordAuthSupport }

// InvalidBody is thrown when the body is invalid.
type InvalidBody struct {
	Message string `json:"message" xml:"message" msgpack:"message"`
}

func (i InvalidBody) status() int         { return http.StatusBadRequest }
func (i InvalidBody) errorCode() HTTPCode { return HTTPCodeInvalidBody }
