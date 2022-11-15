package oobe

import (
	"github.com/jakemakesstuff/spherical/html"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jakemakesstuff/spherical/public"
)

// Router is used to return this packages router.
func Router(dev bool) http.Handler {
	r := mux.NewRouter()

	r.Path("/").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/setup", http.StatusTemporaryRedirect)
	})

	r.Methods("GET").Path("/setup").HandlerFunc(html.Handler(
		"Setup", "Lets setup your new Spherical node!", true,
		nil, 200))

	r.PathPrefix("/").Handler(http.FileServer(http.FS(public.GetFS(dev))))

	return r
}
