package shared

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jakemakesstuff/spherical/public"
)

// Router is used to consume a router and add shared routes.
func Router(r *mux.Router, dev bool) {
	r.Path("/api/internal/i18n").Methods("GET").HandlerFunc(i18nHn)

	r.PathPrefix("/").Handler(http.FileServer(http.FS(public.GetFS(dev))))
}
