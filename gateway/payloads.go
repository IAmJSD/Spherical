package gateway

//go:generate go run generate_serializer.go

// HelloPayload is used to define a hello payload.
type HelloPayload struct {
	// Token is the token to use.
	Token string `msgpack:"token"`

	// CrossNode is used to define if the token is cross node.
	CrossNode bool `msgpack:"cross_node"`
}

// generate_serializer_case HelloPayload 0

// DisconnectPayload is used to define a disconnect payload. This is a special case because it is json.
type DisconnectPayload struct {
	// Reason is the reason for the disconnect.
	Reason string `json:"reason"`

	// Reconnect is used to define if the client should reconnect.
	Reconnect bool `json:"reconnect"`
}
