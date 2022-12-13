package httperrors

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strconv"

	"github.com/vmihailenco/msgpack/v5"
)

// Throw is used to throw an HTTP error in the best way possible.
func Throw(w http.ResponseWriter, r *http.Request, err HTTPError) {
	// Get the status and body.
	status := err.status()
	body := map[string]any{
		"code": string(err.errorCode()),
		"body": err,
	}

	// Marshal to the best content type based on the accept header. Default to json if unsuccessful.
	accept := r.Header.Get("Accept")
	var b []byte
	switch accept {
	case "application/xml", "text/xml":
		b, _ = xml.Marshal(body)
		w.Header().Set("Content-Type", "application/xml")
	case "application/msgpack", "application/x-msgpack":
		b, _ = msgpack.Marshal(body)
		w.Header().Set("Content-Type", "application/msgpack")
	default:
		// Marshal to json.
		b, _ = json.Marshal(body)
		w.Header().Set("Content-Type", "application/json")
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.WriteHeader(status)
	_, _ = w.Write(b)
}
