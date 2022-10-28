package oobe

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

	r.Path("/").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("Hello World!"))
	})

	return r
}
