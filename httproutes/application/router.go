package application

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jakemakesstuff/spherical/assets"
)

// Router is used to return this packages router.
func Router() http.Handler {
	r := mux.NewRouter()

	r.PathPrefix("/assets/").Handler(http.StripPrefix(
		"/assets/", http.FileServer(http.FS(assets.CompilationResult.FS))))

	r.PathPrefix("/setup").HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		http.Redirect(writer, req, "/", http.StatusPermanentRedirect)
	})

	return r
}
