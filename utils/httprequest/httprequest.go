package httprequest

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"

	"github.com/vmihailenco/msgpack/v5"
)

const twomb = int64(2 * 1024 * 1024)

// UnmarshalFromBody is used to unmarshal a request body into a value.
func UnmarshalFromBody(r *http.Request, v any) error {
	// Get the body.
	body, err := io.ReadAll(io.LimitReader(r.Body, twomb))
	defer r.Body.Close()
	if err != nil {
		return err
	}

	// Get the content type.
	ct := r.Header.Get("Content-Type")
	switch ct {
	case "application/xml", "text/xml":
		return xml.Unmarshal(body, v)
	case "application/msgpack", "application/x-msgpack":
		return msgpack.Unmarshal(body, v)
	}
	return json.Unmarshal(body, v)
}
