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

	// Defines routes relating to the node.
	r.Methods("GET").Path("/node").HandlerFunc(nodeInfoHn)

	// Defines authentication related routes.
	r.Methods("POST").Path("/auth/password").HandlerFunc(authPasswordHn)
	//r.Methods("POST").Path("/auth/step/mfa").HandlerFunc(authExtraMFAHn)
}
