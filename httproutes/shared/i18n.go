package shared

import (
	"net/http"
	"strconv"

	"github.com/jakemakesstuff/spherical/i18n"
	"github.com/vmihailenco/msgpack/v5"
)

func i18nHn(w http.ResponseWriter, r *http.Request) {
	strings := i18n.GetAllFrontend(r)
	b, _ := msgpack.Marshal(strings)
	w.Header().Set("Content-Type", "application/msgpack")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.WriteHeader(200)
	_, _ = w.Write(b)
}
