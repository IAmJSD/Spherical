package oobe

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jakemakesstuff/spherical/public"
)

// Router is used to return this packages router.
func Router(dev bool) http.Handler {
	r := mux.NewRouter()

	r.Path("/").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("Hello World!"))
	})

	r.PathPrefix("/").Handler(http.FileServer(public.GetFS(dev)))

	return r
}
