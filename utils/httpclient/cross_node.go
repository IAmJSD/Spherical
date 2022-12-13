package httpclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/jakemakesstuff/spherical/config"
	"github.com/jakemakesstuff/spherical/hashverifier"
	"github.com/jakemakesstuff/spherical/httproutes/application/auth"
	"github.com/jakemakesstuff/spherical/utils/globalstate"
)

// SendCrossNodeMessage is used to send a cross node message to another node.
func SendCrossNodeMessage(
	ctx context.Context, recvHostname, path string, data any, userData auth.UserData,
) (*http.Response, error) {
	// Make a new cross node message from ourselves.
	privKey := globalstate.Get("keyring").(*crypto.KeyRing)
	selfHostname := config.Config().Hostname
	message, err := hashverifier.NewCrossNodeMessage(data, selfHostname, privKey)
	if err != nil {
		return nil, err
	}

	// JSON encode the message.
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	// Redact things that are unneeded to do cross-node.
	userData.Hostname = ""
	userData.Email = ""
	userData.Confirmed = false

	// JSON encode the user data.
	userDataJSON, err := json.Marshal(userData)
	if err != nil {
		return nil, err
	}

	// Create the authorization header.
	authHeader := "cross-node " + base64.StdEncoding.EncodeToString(userDataJSON)

	// Validate all parts of the URL we are sending to.
	if recvHostname == selfHostname {
		return nil, errors.New("cannot send cross node message to ourselves")
	}
	u, err := url.Parse("https://" + recvHostname + path)
	if err != nil {
		return nil, err
	}
	finalUrl := u.String()

	// Send the request.
	r, err := http.NewRequestWithContext(ctx, "POST", finalUrl, bytes.NewReader(messageJSON))
	if err != nil {
		return nil, err
	}
	r.Header.Set("Authorization", authHeader)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Content-Length", strconv.Itoa(len(messageJSON)))
	r.Header.Set("Accept", "application/msgpack")
	return http.DefaultClient.Do(r)
}
