package gateway

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jakemakesstuff/spherical/config"
	"github.com/jakemakesstuff/spherical/db"
	"github.com/jakemakesstuff/spherical/errhandler"
	"github.com/jakemakesstuff/spherical/httproutes/application/auth"
	"github.com/jakemakesstuff/spherical/utils/helpers"
	"github.com/jakemakesstuff/spherical/utils/httpclient"
	"github.com/vmihailenco/msgpack/v5"
)

type connection struct {
	ws *websocket.Conn

	w     payloadWriter
	wLock sync.Mutex

	user     *auth.UserData
	userLock sync.RWMutex

	crossNodeClients map[string]*crossNodeClient
	crossNodeLock    sync.RWMutex

	activeHeartbeats     map[string]*time.Timer
	activeHeartbeatsLock sync.Mutex

	disconnectors     func()
	disconnectorsLock sync.Mutex
}

func (c *connection) addDisconnector(f func()) {
	c.disconnectorsLock.Lock()
	defer c.disconnectorsLock.Unlock()
	oldFunc := c.disconnectors
	c.disconnectors = func() {
		if oldFunc != nil {
			// There was a function before it. We should remove it.
			oldFunc()
		}
		f()
	}
}

const clientHeartbeat = time.Second * 2

func (c *connection) send(x any) error {
	val, err := serializePayload(x)
	if err != nil {
		return err
	}
	c.wLock.Lock()
	defer c.wLock.Unlock()
	_ = c.ws.SetWriteDeadline(time.Now().Add(clientHeartbeat * 2))
	return c.w.WriteMessage(websocket.BinaryMessage, val)
}

// Handles the cross node connection spawning and trying to get guilds in the context window.\
func (c *connection) spawnCrossNodeConn(
	ctx context.Context, hostname string, guildIds []uint64,
) (ready *ReadyPayload, stillAlive bool, err error) {
	// Get a cross node gateway token.
	c.userLock.RLock()
	user := c.user
	c.userLock.RUnlock()
	var res *http.Response
	res, err = httpclient.SendCrossNodeMessage(
		ctx, hostname, "/api/v1/gateway/cross_node",
		map[string]bool{"r": true}, *user)
	if err != nil {
		return
	}

	// Get the token.
	if res.StatusCode != 200 {
		err = errors.New("returned status code " + strconv.Itoa(res.StatusCode) + " from cross node gateway")
		return
	}
	var b []byte
	b, err = io.ReadAll(io.LimitReader(res.Body, 10000))
	_ = res.Body.Close()
	if err != nil {
		return
	}

	// Unmarshal the token.
	var token string
	err = msgpack.Unmarshal(b, &token)
	if err != nil {
		return
	}

	// Connect to the cross node gateway.
	var cc *crossNodeClient
	cc, err = newCrossNodeClient("wss://"+hostname+"/api/v1/gateway", token, c)
	if err != nil {
		// Failed to connect to the cross node gateway.
		return
	}

	// Add the cross node client to the map.
	c.crossNodeLock.Lock()
	if c.crossNodeClients == nil {
		c.crossNodeClients = map[string]*crossNodeClient{}
	}
	c.crossNodeClients[hostname] = cc
	cc.addDestructor(func() {
		c.crossNodeLock.Lock()
		delete(c.crossNodeClients, hostname)
		c.crossNodeLock.Unlock()
	})
	c.crossNodeLock.Unlock()

	// Add the ready handler and wait to either be timed out or for it to be done.
	ch := make(chan crossNodeReady, 1)
	cc.addReadyHandler(ch)
	select {
	case <-ctx.Done():
		// We got timed out. Go ahead and send that this is unavailable.
		err = errors.New("guilds timed out")
		stillAlive = true
	case x := <-ch:
		// Handle reading the result and processing any errors.
		if ready, err = x.unwrap(); err != nil {
			// Failed to connect to the websocket.
			return
		}

		// Process the ready payload.
		if ready.UnavailableGuilds == nil {
			ready.UnavailableGuilds = []Guild{}
		}
		if ready.AvailableGuilds == nil {
			ready.AvailableGuilds = []Guild{}
		}
		unavailable := make([]Guild, 0, len(ready.UnavailableGuilds))
		for _, v := range ready.UnavailableGuilds {
			if helpers.SliceIncludes(guildIds, v.ID) {
				unavailable = append(unavailable, Guild{
					ID:        v.ID,
					Hostname:  hostname,
					Available: false,
				})
			}
		}
		ready.UnavailableGuilds = unavailable
		available := make([]Guild, 0, len(ready.AvailableGuilds))
		for _, v := range ready.AvailableGuilds {
			if helpers.SliceIncludes(guildIds, v.ID) {
				// Ensure the hostname is what we expect and then append it.
				v.Hostname = hostname
				available = append(available, v)
			}
		}
		ready.AvailableGuilds = available

		// Return here.
		return
	}

	// Start a goroutine to wait for the ready payload to finally be sent.
	go func() {
		// Wait for the goroutine with no timeout.
		x := <-ch
		ready, err := x.unwrap()

		if err != nil {
			// Start the backoff and then return.
			// TODO
			return
		}

		// Send all compatible available guilds as update payloads.
		for _, v := range ready.AvailableGuilds {
			if helpers.SliceIncludes(guildIds, v.ID) {
				// Ensure the hostname is what we expect and then send it.
				v.Hostname = hostname
				err = c.send(GuildUpdatePayload{v})
				if err != nil {
					// Failure to send means the rest won't work.
					return
				}
			}
		}
	}()

	// Send the error.
	return
}

// Handles fetching guilds and initiating cross node connections.
func (c *connection) handleGuildFetching(ctx context.Context) (availableGuilds, unavailableGuilds []Guild, err error) {
	// Get the members guild ID's.
	guildIds, err := db.GetMemberGuilds(ctx, c.user.Hostname, c.user.UserID)
	if err != nil {
		return
	}

	// Get our own hostname.
	selfHostname := config.Config().Hostname
	var guildQueries []uint64
	var guildHostnames map[string][]uint64
	for _, guildFrag := range guildIds {
		// Get the relevant bits.
		s := strings.SplitN(guildFrag, "@", 2)
		guildId, _ := strconv.ParseUint(s[0], 10, 64)
		hostname := s[1]

		// Check if we're on the same node.
		if hostname == selfHostname {
			// Append it to guildQueries and we will manage later.
			guildQueries = append(guildQueries, guildId)
			continue
		} else {
			// Add to the map slice.
			guildHostnames[hostname] = append(guildHostnames[hostname], guildId)
		}
	}

	// If guildQueries is not nil, we need to handle it.
	if guildQueries != nil {
		scanners, err := db.GetGuildScanners(ctx, guildQueries)
		if err != nil {
			return nil, nil, err
		}
		for _, scanner := range scanners {
			g := Guild{}
			err = scanner(&g)
			if err != nil {
				return nil, nil, err
			}
			availableGuilds = append(availableGuilds, g)
		}
	}

	// Handle all cross node stuff.
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for hostname, guildIds := range guildHostnames {
		go func(hostname string, guildIds []uint64) {
			defer wg.Done()
			ready, stillAlive, err := c.spawnCrossNodeConn(ctx, hostname, guildIds)
			if !stillAlive {
				// TODO: handle backing off.
			}
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				// Append guilds to available and unavailable.
				availableGuilds = append(availableGuilds, ready.AvailableGuilds...)
				unavailableGuilds = append(unavailableGuilds, ready.UnavailableGuilds...)
			} else {
				// Append guilds to unavailable.
				guilds := make([]Guild, len(guildIds))
				for i, guildId := range guildIds {
					guilds[i] = Guild{
						ID:       guildId,
						Hostname: hostname,
					}
				}
				unavailableGuilds = append(unavailableGuilds, guilds...)
			}
		}(hostname, guildIds)
	}
	wg.Wait()

	// Return the guilds.
	return
}

// Handles looping for heartbeats and making sure they come back in time.
func (c *connection) heartbeatLoop() {
	for {
		// Sleep for the heartbeat wait.
		time.Sleep(clientHeartbeat)

		// Start a delayed destroyer after the message is sent.
		id := uuid.NewString()
		err := c.send(HeartbeatPayload{ID: id})
		if err != nil {
			// Just return here.
			return
		}
		timer := time.AfterFunc(clientHeartbeat*2, func() {
			// Close the connection.
			disconnectWs(c.ws, &DisconnectPayload{
				Reason:    "connection timed out",
				Reconnect: true,
			}, websocket.CloseGoingAway)
		})
		c.activeHeartbeatsLock.Lock()
		if c.activeHeartbeats == nil {
			c.activeHeartbeats = map[string]*time.Timer{}
		}
		c.activeHeartbeats[id] = timer
		c.activeHeartbeatsLock.Unlock()
	}
}

// Starts a read loop.
func (c *connection) readLoop(r payloadReader) {
	for {
		// Set a read timeout.
		_ = c.ws.SetReadDeadline(time.Now().Add(clientHeartbeat * 2))

		// Get the packet.
		_, p, err := r.ReadMessage()
		if err != nil {
			c.disconnectorsLock.Lock()
			d := c.disconnectors
			c.disconnectorsLock.Unlock()
			if d != nil {
				d()
			}
			_ = c.ws.Close()
			return
		}

		// Deserialize the packet.
		switch val := parsePayload(p).(type) {
		case *HeartbeatPayload:
			// Handle the active heartbeat ticker.
			c.activeHeartbeatsLock.Lock()
			if c.activeHeartbeats != nil {
				if timer, ok := c.activeHeartbeats[val.ID]; ok {
					timer.Stop()
					delete(c.activeHeartbeats, val.ID)
				}
			}
			c.activeHeartbeatsLock.Unlock()
		}
		// TODO
	}
}

// Starts everything to handle this connection.
func (c *connection) start(re payloadReader) {
	// Send a accepted payload message.
	err := c.send(AcceptedPayload{
		HeartbeatInterval: uint(clientHeartbeat.Milliseconds()),
	})
	if err != nil {
		_ = c.ws.Close()
		return
	}

	// Make a context to setup everything.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to every node possible and get all guilds.
	availableGuilds, unavailableGuilds, err := c.handleGuildFetching(ctx)
	if err != nil {
		// Handle disconnecting with an error.
		errhandler.Process(err, "gateway/connection", map[string]string{
			"action": "handleGuildFetching",
		})
		disconnectWs(c.ws, &DisconnectPayload{
			Reason:    "internal server error",
			Reconnect: false,
		}, websocket.CloseInternalServerErr)
		return
	}
	if availableGuilds == nil {
		availableGuilds = []Guild{}
	}
	if unavailableGuilds == nil {
		unavailableGuilds = []Guild{}
	}

	// Send the ready payload.
	err = c.send(ReadyPayload{
		AvailableGuilds:   availableGuilds,
		UnavailableGuilds: unavailableGuilds,
	})
	if err != nil {
		c.disconnectorsLock.Lock()
		d := c.disconnectors
		c.disconnectorsLock.Unlock()
		if d != nil {
			d()
		}
		_ = c.ws.Close()
		return
	}

	// Start the heartbeat loop.
	go c.heartbeatLoop()

	// Subscribe to any events wanted for this socket.
	// c.subscribeToEvents() // TODO

	// Start the read loop.
	c.readLoop(re)
}
