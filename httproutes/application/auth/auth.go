package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jakemakesstuff/spherical/config"
	"github.com/jakemakesstuff/spherical/db"
	"github.com/jakemakesstuff/spherical/errhandler"
	"github.com/jakemakesstuff/spherical/hashverifier"
	"github.com/jakemakesstuff/spherical/i18n"
	"github.com/jakemakesstuff/spherical/jobs"
	"github.com/jakemakesstuff/spherical/scheduler"
	"github.com/jakemakesstuff/spherical/utils/httperrors"
)

// UserData defines any user data which was injected during the middleware.
type UserData struct {
	// Username defines the username of the user.
	Username string `json:"username" validate:"username,required,min=1"`

	// UserID defines the user ID of the user.
	UserID uint64 `json:"user_id" validate:"user_id,required"`

	// Flags is used to define the flags of a user.
	Flags uint64 `json:"flags" validate:"flags,required"`

	// AvatarURL is used to define the avatar URL of the user.
	AvatarURL string `json:"avatar_url" validate:"url,required"`

	// Hostname is used to define the hostname of the node making this request.
	Hostname string `json:"hostname,omitempty"`

	// Email is used to define the e-mail address of the user. Is blank if not this node.
	Email string `json:"email,omitempty"`

	// Confirmed is used to define if the e-mail address is confirmed. Is false if not this node.
	Confirmed bool `json:"confirmed"`
}

// SameNode defines if the user data is coming from this node.
func (u *UserData) SameNode() bool { return u.Email != "" }

var validate = validator.New()

type closeStealer struct {
	io.Reader
	io.Closer
}

// Try and get the user data from the request. Returns the context (possibly with "user" added) and a
// bool saying if the user data was added.
func getUserData(r *http.Request) (*http.Request, bool, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// No auth header present.
		return r, false, nil
	}
	authHeaderSplit := strings.Split(authHeader, " ")

	authHeaderSplit[0] = strings.ToLower(authHeaderSplit[0])
	if authHeaderSplit[0] == "bearer" {
		// Expect there to be a token.
		if len(authHeaderSplit) == 1 {
			return r, false, nil
		}
		token := authHeaderSplit[1]

		// Try and scan the user data from the DB.
		var user UserData
		err := db.ScanUserFromToken(r.Context(), token, &user)
		present := err == nil
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
		}
		if present {
			// Keep the token alive if possible.
			_, _ = jobs.TokenExtendJob.ScheduleImmediately(r.Context(), jobs.TokenExtendEvent{
				Token: token,
				In:    time.Hour * 24 * 30,
			}, scheduler.Metadata{
				Retries:        2,
				Timeout:        time.Second * 5,
				RefireDuration: time.Second,
			})
		}
		user.Hostname = config.Config().Hostname
		r.WithContext(context.WithValue(r.Context(), "user", &user))
		return r, present, err
	}

	if authHeaderSplit[0] == "cross-node" {
		// Expect there to be a JSON blob.
		if len(authHeaderSplit) == 1 {
			return r, false, nil
		}
		blob := authHeaderSplit[1]

		// Unmarshal the blob.
		var user UserData
		b, err := base64.StdEncoding.DecodeString(blob)
		if err != nil {
			// Part of the header was invalid.
			return r, false, nil
		}
		err = json.Unmarshal(b, &user)
		if err != nil {
			// Part of the header was invalid.
			return r, false, nil
		}
		err = validate.Struct(user)
		if err != nil {
			// Header was not valid data.
			return r, false, nil
		}

		// Consume the body to a maximum of 2MB.
		b, err = io.ReadAll(io.LimitReader(r.Body, 2e+6))
		if err != nil {
			// Return invalid here - doesn't really matter if we hit the error bit.
			return r, false, nil
		}

		// Get the cross-node message.
		var msg hashverifier.CrossNodeMessage
		err = json.Unmarshal(b, &msg)
		if err != nil {
			// The body is not valid JSON. Cannot be a valid message.
			return r, false, nil
		}
		c := config.Config()
		if c.Consensus == 0 {
			c.Consensus = 3
		}
		err = msg.Validate(r.Context(), c.Consensus, c.Informers, c.Trusted)
		if err != nil {
			// Validation failed.
			if _, ok := err.(hashverifier.UserFacingError); ok {
				err = nil
			}
			return r, false, err
		}

		// Set the requests body to the inner JSON of the cross node message.
		r.Body = closeStealer{
			Reader: bytes.NewReader(msg.Data),
			Closer: r.Body,
		}

		// Set the hostname to the cross node data.
		user.Hostname = msg.Hostname
		user.Email = ""
		user.Confirmed = false
	}

	return r, false, nil
}

// Middleware is used to authenticate requests. The request MUST be authenticated to continue.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the authorization header.
		r, present, err := getUserData(r)
		if err != nil {
			// Some other error happened.
			errhandler.Process(err, "httproutes/application/auth", map[string]string{})
			httperrors.Throw(w, r, httperrors.InternalServerError{})
			return
		}
		if !present {
			// Invalid authentication information.
			httperrors.Throw(w, r, httperrors.InvalidAuth{
				Message: i18n.GetWithRequest(r, "httproutes/application/auth/auth:bad_auth_header"),
			})
			return
		}

		// Call the next handler.
		next.ServeHTTP(w, r)
	})
}
