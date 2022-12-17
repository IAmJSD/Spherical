package apiv1

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jakemakesstuff/spherical/db"
	"github.com/jakemakesstuff/spherical/errhandler"
	"github.com/jakemakesstuff/spherical/i18n"
	"github.com/jakemakesstuff/spherical/jobs"
	"github.com/jakemakesstuff/spherical/scheduler"
	"github.com/jakemakesstuff/spherical/utils/httperrors"
	"github.com/jakemakesstuff/spherical/utils/httprequest"
	"github.com/jakemakesstuff/spherical/utils/httpresponse"
)

type authInfo struct {
	Username string `json:"username" xml:"username" msgpack:"username"`
	Password string `json:"password" xml:"password" msgpack:"password"`
}

func authPasswordHn(w http.ResponseWriter, r *http.Request) {
	// Get the username and password from the body.
	var info authInfo
	err := httprequest.UnmarshalFromBody(r, &info)
	if err != nil {
		httperrors.Throw(w, r, httperrors.InvalidBody{
			Message: i18n.GetWithRequest(r, "httproutes/application/apiv1/password_auth:invalid_body"),
		})
		return
	}

	// Authenticate with the username and password.
	userId, err := db.AuthenticateUserByPassword(r.Context(), info.Username, info.Password)
	if err != nil {
		// Handle if this user does not support password authentication.
		if err == db.ErrNotPasswordAuth {
			httperrors.Throw(w, r, httperrors.NoPasswordAuthSupport{
				Message: i18n.GetWithRequest(r, "httproutes/application/apiv1/password_auth:no_password_auth"),
			})
			return
		}

		// Handle if the user does not provide valid credentials.
		if err == db.ErrInvalidCredentials {
			httperrors.Throw(w, r, httperrors.InvalidAuth{
				Message: i18n.GetWithRequest(r, "httproutes/application/apiv1/password_auth:invalid_credentials"),
			})
			return
		}

		// Check if this is a half-authentication case.
		he, ok := err.(httperrors.HalfAuthentication)
		if ok {
			httperrors.Throw(w, r, he)
			return
		}

		// This is a standard internal server error.
		errhandler.Process(err, "httproutes/application/apiv1/authPasswordHn", map[string]string{
			"where": "db.AuthenticateUserByPassword",
		})
		httperrors.Throw(w, r, httperrors.InternalServerError{})
		return
	}

	// Handle creating the UUID.
	id, err := uuid.NewRandom()
	if err != nil {
		errhandler.Process(err, "httproutes/application/apiv1/authPasswordHn", map[string]string{
			"where": "uuid.NewRandom",
		})
		httperrors.Throw(w, r, httperrors.InternalServerError{})
		return
	}
	idS := id.String()

	// Create the session delete job.
	jobId, err := jobs.TokenDeleteJob.Schedule(r.Context(), idS, scheduler.Metadata{
		Retries:        3,
		Timeout:        time.Second * 5,
		RefireDuration: time.Second,
	}, time.Hour*24*30)
	if err != nil {
		errhandler.Process(err, "httproutes/application/apiv1/authPasswordHn", map[string]string{
			"where": "jobs.TokenDeleteJob.Schedule",
		})
		httperrors.Throw(w, r, httperrors.InternalServerError{})
		return
	}

	// Create the session.
	err = db.CreateSession(r.Context(), idS, userId, jobId)
	if err != nil {
		errhandler.Process(err, "httproutes/application/apiv1/authPasswordHn", map[string]string{
			"where": "db.CreateSession",
		})
		httperrors.Throw(w, r, httperrors.InternalServerError{})
		return
	}

	// Return the session token.
	httpresponse.MarshalToAccept(w, r, http.StatusOK, map[string]string{
		"token": idS,
	})
}
