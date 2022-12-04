package gateway

import (
	"github.com/gorilla/websocket"
	"github.com/jakemakesstuff/spherical/httproutes/application/auth"
)

type connection struct {
	ws   *websocket.Conn
	r    payloadReader
	w    payloadWriter
	user *auth.UserData
}

// Starts everything to handle this connection.
func (c *connection) start() {
	// TODO
}
