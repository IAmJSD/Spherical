package application

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jakemakesstuff/spherical/public"
)

// Router is used to return this packages router.
func Router(dev bool) http.Handler {
	r := mux.NewRouter()

	r.PathPrefix("/setup").HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		http.Redirect(writer, req, "/", http.StatusPermanentRedirect)
	})

	r.PathPrefix("/").Handler(http.FileServer(http.FS(public.GetFS(dev))))

	return r
}
