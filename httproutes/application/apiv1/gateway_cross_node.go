package apiv1

import (
	"net/http"

	"github.com/jakemakesstuff/spherical/db"
	"github.com/jakemakesstuff/spherical/errhandler"
	"github.com/jakemakesstuff/spherical/httproutes/application/auth"
	"github.com/jakemakesstuff/spherical/i18n"
	"github.com/jakemakesstuff/spherical/utils/httperrors"
	"github.com/jakemakesstuff/spherical/utils/httpresponse"
	"github.com/vmihailenco/msgpack/v5"
)

func crossNodeGwHn(w http.ResponseWriter, r *http.Request) {
	// Close the body on end of request. Drain gang!
	defer r.Body.Close()

	// Get the user.
	user := r.Context().Value("user").(*auth.UserData)
	if user.SameNode() {
		// We should error for users if they are on the same node.
		httperrors.Throw(w, r, httperrors.SameNode{
			Message: i18n.GetWithRequest(r, "httproutes/application/apiv1/gateway_cross_node:same_node"),
		})
		return
	}

	// Marshal into msgpack.
	b, _ := msgpack.Marshal(user)

	// Get the cross node token.
	token, err := db.BuildCrossNodeToken(r.Context(), b)
	if err != nil {
		// We should error if we can't build the token.
		httperrors.Throw(w, r, httperrors.InternalServerError{})
		errhandler.Process(err, "httproutes/application/apiv1/gateway_cross_node", map[string]string{})
		return
	}

	// Write the token to the request.
	httpresponse.MarshalToAccept(w, r, http.StatusOK, token)
}
