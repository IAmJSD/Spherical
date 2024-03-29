package oobe

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jakemakesstuff/spherical/html"
	"github.com/jakemakesstuff/spherical/httproutes/shared"
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

	r.Methods("GET", "POST").Path("/install/state").HandlerFunc(installState)

	shared.Router(r, dev)

	return r
}
