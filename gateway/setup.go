package gateway

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jakemakesstuff/spherical/config"
	"github.com/jakemakesstuff/spherical/db"
	"github.com/jakemakesstuff/spherical/httproutes/application/auth"
	"github.com/vmihailenco/msgpack/v5"
)

type payloadReader interface {
	ReadMessage() (messageType int, p []byte, err error)
}

type payloadWriter interface {
	WriteMessage(messageType int, data []byte) error
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  10000,
	WriteBufferSize: 10000,
}

type cmpReader struct {
	origin payloadReader
	buf    bytes.Buffer
	zr     io.Reader
	out    []byte
}

func (c *cmpReader) ReadMessage() (messageType int, p []byte, err error) {
	messageType, p, err = c.origin.ReadMessage()
	if err != nil {
		return
	}

	if len(p) != 0 {
		c.buf.Reset()
		_, _ = c.buf.Write(p)
		if c.zr == nil {
			c.zr, _ = zlib.NewReader(&c.buf)
		}
		if c.out == nil {
			c.out = make([]byte, 10000)
		}
		var n int
		n, err = c.zr.Read(c.out)
		if err != nil {
			return
		}
		p = c.out[:n]
		cpy := make([]byte, n)
		copy(cpy, p)
		p = cpy
	}

	return
}

type cmpWriter struct {
	origin payloadWriter
	buf    bytes.Buffer
	w      io.Writer
}

func (c *cmpWriter) WriteMessage(messageType int, data []byte) error {
	if len(data) == 0 {
		return c.origin.WriteMessage(messageType, data)
	}
	if c.w == nil {
		c.w = zlib.NewWriter(&c.buf)
	}
	_, err := c.w.Write(data)
	if err != nil {
		return err
	}
	tmp := c.buf.Bytes()
	defer c.buf.Reset()
	return c.origin.WriteMessage(messageType, tmp)
}

func disconnectWs(w *websocket.Conn, p *DisconnectPayload, closeCode int) {
	b, _ := json.Marshal(p)
	_ = w.WriteControl(
		websocket.CloseMessage, websocket.FormatCloseMessage(closeCode, string(b)),
		time.Now().Add(5*time.Second))
	_ = w.Close()
}

// WebSocketHandler is used to handle websocket connections.
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the query params contain zlib-stream.
	compress := strings.ToLower(r.URL.Query().Get("compress")) == "zlib-stream"

	// Upgrade the connection.
	ws, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		return
	}
	var re payloadReader = ws
	var wr payloadWriter = ws
	if compress {
		// Handle making the compressed reader and writer.
		re = &cmpReader{origin: ws}
		wr = &cmpWriter{origin: ws}
	}

	// Handle getting the first hello payload.
	_ = ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, p, err := re.ReadMessage()
	if err != nil {
		_ = ws.Close()
		return
	}

	// Handle reading the hello payload.
	payload, ok := parsePayload(p).(*HelloPayload)
	if !ok {
		disconnectWs(ws, &DisconnectPayload{
			Reason:    "hello payload was not sent first",
			Reconnect: false,
		}, websocket.CloseInvalidFramePayloadData)
		return
	}

	// Get a temp context since both cases need it.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Handle processing the auth token.
	var user auth.UserData
	if payload.CrossNode {
		// Get the cross node token from Redis.
		userB, err := db.GetCrossNodeToken(ctx, payload.Token)
		if err != nil {
			disconnectWs(ws, &DisconnectPayload{
				Reason:    "invalid cross node token",
				Reconnect: true,
			}, websocket.CloseInvalidFramePayloadData)
			return
		}
		err = msgpack.Unmarshal(userB, &user)
		if err != nil {
			disconnectWs(ws, &DisconnectPayload{
				Reason:    "internal server error",
				Reconnect: true,
			}, websocket.CloseInternalServerErr)
			return
		}
	} else {
		// Get the user data from the token.
		err = db.ScanUserFromToken(ctx, payload.Token, &user)
		if err != nil {
			disconnectWs(ws, &DisconnectPayload{
				Reason:    "invalid token",
				Reconnect: false,
			}, websocket.CloseInvalidFramePayloadData)
			return
		}
		user.Hostname = config.Config().Hostname
	}

	// We are authenticated! Spin up the connection!
	conn := &connection{
		ws:   ws,
		r:    re,
		w:    wr,
		user: &user,
	}
	conn.start()
}
