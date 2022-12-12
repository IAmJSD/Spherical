package gateway

import (
	"context"
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
)

type connection struct {
	ws *websocket.Conn
	r  payloadReader

	w     payloadWriter
	wLock sync.Mutex

	user     *auth.UserData
	userLock sync.RWMutex

	activeHeartbeats     map[string]*time.Timer
	activeHeartbeatsLock sync.Mutex
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
) (guilds []Guild, err error) {
	// TODO
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
			guilds, err := c.spawnCrossNodeConn(ctx, hostname, guildIds)
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				// Append guilds to available.
				availableGuilds = append(availableGuilds, guilds...)
			} else {
				// Append guilds to unavailable.
				guilds = make([]Guild, len(guildIds))
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
func (c *connection) readLoop() {
	for {
		// Set a read timeout.
		_ = c.ws.SetReadDeadline(time.Now().Add(clientHeartbeat * 2))

		// Get the packet.
		_, p, err := c.r.ReadMessage()
		if err != nil {
			_ = c.ws.Close()
			return
		}

		// Deserialize the packet.
		switch val := parsePayload(p).(type) {
		case *HeartbeatPayload:
			// Handle the active heartbeat ticker.
			c.activeHeartbeatsLock.Lock()
			if timer, ok := c.activeHeartbeats[val.ID]; ok {
				timer.Stop()
				delete(c.activeHeartbeats, val.ID)
			}
			c.activeHeartbeatsLock.Unlock()
			// TODO
		}
	}
}

// Starts everything to handle this connection.
func (c *connection) start() {
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
		_ = c.ws.Close()
		return
	}

	// Start the heartbeat loop.
	go c.heartbeatLoop()

	// Subscribe to any events wanted for this socket.
	// c.subscribeToEvents() // TODO

	// Start the read loop.
	c.readLoop()
}
