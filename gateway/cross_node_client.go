package gateway

import (
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type crossNodeClient struct {
	ws *websocket.Conn
	r  payloadReader

	w     payloadWriter
	wLock sync.Mutex

	conn *connection

	readyResult   *crossNodeReady
	readyChans    []chan<- crossNodeReady
	readyChanLock sync.Mutex

	destructors     []func()
	destructorsLock sync.Mutex
}

type crossNodeReady struct {
	err     error
	payload *ReadyPayload
}

func (c crossNodeReady) unwrap() (*ReadyPayload, error) { return c.payload, c.err }

func (c *crossNodeClient) send(x any) error {
	val, err := serializePayload(x)
	if err != nil {
		return err
	}
	c.wLock.Lock()
	defer c.wLock.Unlock()
	_ = c.ws.SetWriteDeadline(time.Now().Add(clientHeartbeat * 2))
	return c.w.WriteMessage(websocket.BinaryMessage, val)
}

func newCrossNodeClient(addr, token string, conn *connection) (*crossNodeClient, error) {
	// Dial the websocket.
	ws, _, err := websocket.DefaultDialer.Dial(addr+"?compress=zlib-stream", nil)
	if err != nil {
		return nil, err
	}
	var re payloadReader = &cmpReader{origin: ws}
	var wr payloadWriter = &cmpWriter{origin: ws}

	// Make the cross node client.
	c := &crossNodeClient{
		ws:   ws,
		r:    re,
		w:    wr,
		conn: conn,
	}

	// Send the hello payload to the cross node client.
	err = c.send(&HelloPayload{
		Token:     token,
		CrossNode: true,
	})
	if err != nil {
		_ = c.ws.Close()
		return nil, err
	}

	// Wait for the accepted payload.
	_ = c.ws.SetReadDeadline(time.Now().Add(clientHeartbeat * 2))
	_, p, err := c.r.ReadMessage()
	if err != nil {
		_ = c.ws.Close()
		return nil, err
	}
	_, ok := parsePayload(p).(*AcceptedPayload)
	if !ok {
		_ = c.ws.Close()
		return nil, errors.New("accepted payload not received first")
	}

	// Start the read loop in a goroutine.
	go c.readLoop()

	// Return the client and hook the disconnect handler.
	conn.addDisconnector(c.disconnect)
	return c, nil
}

func (c *crossNodeClient) disconnect() {
	_ = c.ws.Close()
}

func (c *crossNodeClient) addReadyHandler(ch chan<- crossNodeReady) {
	c.readyChanLock.Lock()
	defer c.readyChanLock.Unlock()

	if c.readyResult != nil {
		// Just send the ready result we already have.
		select {
		case ch <- *c.readyResult:
		default:
		}
		return
	}

	c.readyChans = append(c.readyChans, ch)
}

func (c *crossNodeClient) readLoop() {
	for {
		// Read the websocket payload,
		_, p, err := c.r.ReadMessage()
		if err != nil {
			_ = c.ws.Close()
			c.sendReady(crossNodeReady{err: err})
			c.destructorsLock.Lock()
			destructors := c.destructors
			c.destructors = nil
			c.destructorsLock.Unlock()
			for _, v := range destructors {
				v()
			}
			return
		}

		// Parse the payload.
		switch val := parsePayload(p).(type) {
		case *ReadyPayload:
			c.sendReady(crossNodeReady{payload: val})
		case *HeartbeatPayload:
			_ = c.send(val)
		}
		// TODO
	}
}

func (c *crossNodeClient) sendReady(r crossNodeReady) {
	c.readyChanLock.Lock()
	defer c.readyChanLock.Unlock()

	if c.readyResult != nil {
		// Ready was already sent. Packet seems dodgy.
		return
	}

	for _, v := range c.readyChans {
		select {
		case v <- r:
		default:
		}
	}
	c.readyChans = nil
	c.readyResult = &r
}

func (c *crossNodeClient) addDestructor(f func()) {
	c.destructorsLock.Lock()
	c.destructors = append(c.destructors, f)
	c.destructorsLock.Unlock()
}
