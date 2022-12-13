package httpresponse

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strconv"

	"github.com/vmihailenco/msgpack/v5"
)

// MarshalToAccept is used to marshal a value to the accept header type.
func MarshalToAccept(w http.ResponseWriter, r *http.Request, status int, body any) {
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
