package gateway

import (
	"sync"

	"github.com/gorilla/websocket"
)

type crossNodeClient struct {
	ws *websocket.Conn

	conn *connection

	readyResult   *crossNodeReady
	readyChans    []chan<- crossNodeReady
	readyChanLock sync.Mutex
}

type crossNodeReady struct {
	err     error
	payload *ReadyPayload
}

func (c crossNodeReady) unwrap() (*ReadyPayload, error) { return c.payload, c.err }

func newCrossNodeClient(addr, token string, conn *connection) *crossNodeClient {
	c := &crossNodeClient{
		ws:   nil,
		conn: nil,
	}
	conn.addDisconnector(c.disconnect)
	return c
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
