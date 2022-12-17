package gateway

// Code generated by generate_serializer.go. DO NOT EDIT.

import (
	"encoding/binary"
	"errors"

	"github.com/vmihailenco/msgpack/v5"
)

// Parses the payload. Returns nil if the payload is invalid.
func parsePayload(b []byte) any {
	// Get the uint16 from the first 2 bytes.
	if len(b) < 2 {
		return nil
	}
	id := binary.BigEndian.Uint16(b)
	b = b[2:]

	// Switch on the payload.
	switch id {
	case 0:
		// Parse the HelloPayload payload.
		p := &HelloPayload{}
		err := msgpack.Unmarshal(b, p)
		if err != nil {
			return nil
		}
		return p
	case 1:
		// Parse the AcceptedPayload payload.
		p := &AcceptedPayload{}
		err := msgpack.Unmarshal(b, p)
		if err != nil {
			return nil
		}
		return p
	case 2:
		// Parse the HeartbeatPayload payload.
		p := &HeartbeatPayload{}
		err := msgpack.Unmarshal(b, p)
		if err != nil {
			return nil
		}
		return p
	case 3:
		// Parse the JoinGuildPayload payload.
		p := &JoinGuildPayload{}
		err := msgpack.Unmarshal(b, p)
		if err != nil {
			return nil
		}
		return p
	case 4:
		// Parse the ReadyPayload payload.
		p := &ReadyPayload{}
		err := msgpack.Unmarshal(b, p)
		if err != nil {
			return nil
		}
		return p
	case 5:
		// Parse the GuildUpdatePayload payload.
		p := &GuildUpdatePayload{}
		err := msgpack.Unmarshal(b, p)
		if err != nil {
			return nil
		}
		return p
	}
	return nil
}

// Serializes the payload. Returns an error if the payload is invalid.
func serializePayload(p any) ([]byte, error) {
	// Switch on the payload.
	var id uint16
	switch p.(type) {
	case *AcceptedPayload, AcceptedPayload:
		id = 1
	case *GuildUpdatePayload, GuildUpdatePayload:
		id = 5
	case *HeartbeatPayload, HeartbeatPayload:
		id = 2
	case *HelloPayload, HelloPayload:
		id = 0
	case *JoinGuildPayload, JoinGuildPayload:
		id = 3
	case *ReadyPayload, ReadyPayload:
		id = 4
	default:
		return nil, errors.New("invalid payload")
	}

	// Marshal the payload.
	b, err := msgpack.Marshal(p)
	if err != nil {
		return nil, err
	}

	// Prepend the id.
	b2 := make([]byte, 2+len(b))
	binary.BigEndian.PutUint16(b2, id)
	copy(b2[2:], b)
	return b2, nil
}
