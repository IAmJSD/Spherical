package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hako/durafmt"
	"github.com/jakemakesstuff/spherical/utils/random"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	ignoreNumber    uint64
	ignoreTable     = map[uint64]struct{}{}
	ignoreTableLock = sync.Mutex{}
)

// Get the ignore number start randomly. Ensures that it is VERY likely that different instances have different ID's.
// Only a random uint32 so we do not hit the top boundary and have no way to grow.
func init() {
	ignoreNumber = uint64(random.New().Uint32())
}

func getIgnoreId() uint64 {
	val := atomic.AddUint64(&ignoreNumber, 1)
	ignoreTableLock.Lock()
	ignoreTable[val] = struct{}{}
	ignoreTableLock.Unlock()
	return val
}

func isIgnoreId(val uint64) bool {
	ignoreTableLock.Lock()
	_, ok := ignoreTable[val]
	if ok {
		delete(ignoreTable, val)
	}
	ignoreTableLock.Unlock()
	return ok
}

type editPayload struct {
	ID        uint64          `json:"i"`
	TableName string          `json:"t"`
	Metadata  json.RawMessage `json:"m"`
}

func (e editPayload) dispatch(block bool) {
	// Get the table events.
	watchEventsLock.RLock()
	dispatchers := watchEvents[e.TableName]
	watchEventsLock.RUnlock()

	// Publish all the things.
	for _, v := range dispatchers {
		if block {
			// Blocking send. Used for handlers calling the EditAndPublish function.
			v <- e.Metadata
		} else {
			// Fast non-blocking send.
			select {
			case v <- e.Metadata:
			default:
			}
		}
	}
}

// EditAndPublish is used to push an edit on a table and publish the event. This will block until the events are published,
// so you can be assured any events are done by the time this is over.
func EditAndPublish(ctx context.Context, tableName string, f func(context.Context) error, metadata any) error {
	metadataMarshal, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	p := editPayload{
		ID:        getIgnoreId(),
		TableName: tableName,
		Metadata:  metadataMarshal,
	}
	b, _ := json.Marshal(p)
	if err := f(ctx); err != nil {
		return err
	}

	currentConnLock.RLock()
	redisConn := currentRedisConn
	currentConnLock.RUnlock()
	if err := redisConn.Publish(ctx, "table-update", b).Err(); err != nil {
		return err
	}
	p.dispatch(true)
	return nil
}

var (
	// Events related to watching a table.
	watchEvents     = map[string][]chan json.RawMessage{}
	watchEventsLock = sync.RWMutex{}

	// Events related to watching generic events.
	genericEvents         = map[string][]chan []byte{}
	genericEventsHandlers = map[string]struct{}{}
	genericEventsLock     = sync.RWMutex{}
)

// Starts the watch loop.
func watchLoop(ctx context.Context, client *redis.Client) {
	// Defines the backoff time.
	var backoffTime time.Duration

	// Outer for loop to handle re-subscriptions.
	for {
		// Get the subscriber.
		subscriber := client.Subscribe(ctx, "table-update")

		// Inner for loop to handle content.
		for {
			msg, err := subscriber.ReceiveMessage(ctx)
			if err != nil {
				// Check if the context was cancelled.
				if errors.Is(err, context.Canceled) {
					// Just return here. This was because it was re-initialised. This loop will be remade.
					return
				}

				// Calculate the backoff value.
				backoffTime += time.Millisecond * 10
				if backoffTime > time.Minute*10 {
					// After 10 mins, go back to the original polling speed.
					backoffTime = time.Millisecond * 10
				}

				// Log the error.
				_, _ = fmt.Fprintf(os.Stderr, "[db] Watch had an error: %v. Backing off for %s.",
					err, durafmt.Parse(backoffTime).String())

				// Sleep for the time specified, kill the subscriber, and then break.
				time.Sleep(backoffTime)
				_ = subscriber.Close()
				break
			}

			var p editPayload
			if err := json.Unmarshal([]byte(msg.Payload), &p); err == nil {
				// Make sure this isn't intentionally ignored.
				if !isIgnoreId(p.ID) {
					// If not, dispatch it non-blocking.
					p.dispatch(false)
				}
			} else {
				// Error with the payload.
				_, _ = fmt.Fprintf(os.Stderr, "[db] Watch failed to unmarshal json: %v", err)
			}
		}
	}
}

// AddWatchEvent is used to add an event to watch a table where the events are published. T should be the type that
// is the same as the JSON.
func AddWatchEvent[T any](tableName string, f func(T)) {
	ch := make(chan json.RawMessage)
	go func() {
		for {
			ev, ok := <-ch
			if !ok {
				return
			}
			var content T
			if err := json.Unmarshal(ev, &content); err == nil {
				f(content)
			}
		}
	}()
	watchEventsLock.Lock()
	watchEvents[tableName] = append(watchEvents[tableName], ch)
	watchEventsLock.Unlock()
}

type genericPayload struct {
	ID   uint64 `msgpack:"i"`
	Data []byte `msgpack:"d"`
}

// Watches the specified event name (with the "generic-event:" prefix) and calls the channels when it is received.
func watchEvent(ctx context.Context, eventName string, client *redis.Client) {
	// Prepend the prefix.
	channelName := eventName
	eventName = "generic-event:" + eventName

	// Defines the backoff time.
	var backoffTime time.Duration

	// Register that we are handling this.
	genericEventsLock.Lock()
	_, ok := genericEventsHandlers[channelName]
	if ok {
		// Raced.
		genericEventsLock.Unlock()
		return
	}
	genericEventsHandlers[channelName] = struct{}{}
	genericEventsLock.Unlock()
	defer func() {
		genericEventsLock.Lock()
		delete(genericEventsHandlers, channelName)
		genericEventsLock.Unlock()
	}()

	// Outer for loop to handle re-subscriptions.
	var closer func() error
	defer func() {
		if closer != nil {
			closer()
		}
	}()
	for {
		// Get the subscriber.
		subscriber := client.Subscribe(ctx, eventName)
		closer = subscriber.Close

		// Inner for loop to handle content.
		for {
			msg, err := subscriber.ReceiveMessage(ctx)
			if err != nil {
				// Check if the context was cancelled.
				if errors.Is(err, context.Canceled) {
					// Just return here. This was because it was re-initialised. This loop will be remade.
					return
				}

				// Calculate the backoff value.
				backoffTime += time.Millisecond * 10
				if backoffTime > time.Minute*10 {
					// After 10 mins, go back to the original polling speed.
					backoffTime = time.Millisecond * 10
				}

				// Log the error.
				_, _ = fmt.Fprintf(os.Stderr, "[db] Watch event %s had an error: %v. Backing off for %s.",
					err, durafmt.Parse(backoffTime).String())

				// Sleep for the time specified, kill the subscriber, and then break.
				time.Sleep(backoffTime)
				_ = subscriber.Close()
				break
			}

			var p genericPayload
			if err := msgpack.Unmarshal([]byte(msg.Payload), &p); err == nil {
				// Make sure this isn't intentionally ignored.
				if !isIgnoreId(p.ID) {
					// If not, dispatch it non-blocking.
					genericEventsLock.RLock()
					channels := genericEvents[channelName]
					genericEventsLock.RUnlock()
					if channels == nil {
						// Channels no longer exist.
						return
					}
					for _, ch := range channels {
						select {
						case ch <- p.Data:
						default:
						}
					}
				}
			} else {
				// Error with the payload.
				_, _ = fmt.Fprintf(
					os.Stderr, "[db] Watch failed to unmarshal msgpack for event %s: %v", eventName, err)
			}
		}
	}
}

// AddGenericEventHandler is used to add an event handler for an event.
func AddGenericEventHandler(eventName string, f func([]byte)) (deleter func()) {
	ch := make(chan []byte)
	go func() {
		for {
			ev, ok := <-ch
			if !ok {
				return
			}
			f(ev)
		}
	}()
	genericEventsLock.Lock()
	genericEvents[eventName] = append(genericEvents[eventName], ch)
	if _, ok := genericEventsHandlers[eventName]; !ok {
		currentConnLock.RLock()
		redisConn := currentRedisConn
		currentConnLock.RUnlock()
		go watchEvent(context.Background(), eventName, redisConn)
	}
	genericEventsLock.Unlock()
	return func() {
		genericEventsLock.Lock()
		defer genericEventsLock.Unlock()
		for i, c := range genericEvents[eventName] {
			if c == ch {
				genericEvents[eventName] = append(genericEvents[eventName][:i], genericEvents[eventName][i+1:]...)
				break
			}
		}
		if len(genericEvents[eventName]) == 0 {
			delete(genericEvents, eventName)
		}
	}
}

// PublishGenericEvent is used to publish a generic event.
func PublishGenericEvent(eventName string, data []byte) error {
	// Get the ID.
	id := getIgnoreId()

	// Marshal the payload.
	payload, err := msgpack.Marshal(genericPayload{
		ID:   id,
		Data: data,
	})
	if err != nil {
		return err
	}

	// Publish the event.
	return currentRedisConn.Publish(context.Background(), "generic-event:"+eventName, payload).Err()
}
