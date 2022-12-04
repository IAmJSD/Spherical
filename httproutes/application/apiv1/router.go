package apiv1

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jakemakesstuff/spherical/gateway"
	"github.com/jakemakesstuff/spherical/httproutes/application/auth"
)

// Router is used to consume a router and add API V1 routes.
func Router(r *mux.Router) {
	// Defines routes relating to the gateway.
	r.Methods("POST").Path("/gateway/cross_node").Handler(auth.Middleware(http.HandlerFunc(crossNodeGwHn)))
	r.Path("/gateway").HandlerFunc(gateway.WebSocketHandler)
}
