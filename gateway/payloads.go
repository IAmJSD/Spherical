package gateway

//go:generate go run generate_serializer.go

// HelloPayload is used to define a hello payload.
type HelloPayload struct {
	// Token is the token to use.
	Token string `msgpack:"token"`

	// CrossNode is used to define if the token is cross node. You can generally omit this if you are a client. This is
	// for nodes.
	CrossNode bool `msgpack:"cross_node,omitempty"`
}

// generate_serializer_case HelloPayload 0

// DisconnectPayload is used to define a disconnect payload. This is a special case because it is json.
type DisconnectPayload struct {
	// Reason is the reason for the disconnect.
	Reason string `json:"reason"`

	// Reconnect is used to define if the client should reconnect.
	Reconnect bool `json:"reconnect"`
}

// AcceptedPayload is used to define a payload sent after the connection is accepted but before the connection is "done".
// Spherical might be internally dialling other nodes at this point.
type AcceptedPayload struct {
	// HeartbeatInterval is the interval to send HeartbeatPayload at in milliseconds.
	HeartbeatInterval uint `msgpack:"heartbeat_interval"`
}

// generate_serializer_case AcceptedPayload 1

// HeartbeatPayload is used to define a heartbeat payload. The client or server that gets this should immediately
// respond with a heartbeat of the same ID.
type HeartbeatPayload struct {
	// ID is the ID of the heartbeat.
	ID string `msgpack:"id"`
}

// generate_serializer_case HeartbeatPayload 2

// JoinGuildPayload is used to define a join guild payload.
type JoinGuildPayload struct {
	// Hostname is the hostname where the guild is. Note that this should be blank if this is a cross node socket.
	Hostname string `msgpack:"hostname,omitempty"`

	// InviteCode is the invite code to use.
	InviteCode string `msgpack:"invite_code"`

	// ReplyID is the ID to reply with.
	ReplyID string `msgpack:"reply_id"`
}

// generate_serializer_case JoinGuildPayload 3

// ChannelType is the type of channel.
type ChannelType string

const (
	// ChannelTypeText is a text channel.
	ChannelTypeText ChannelType = "text"

	// ChannelTypeVoice is a voice channel.
	ChannelTypeVoice ChannelType = "voice"
)

// ChannelPermissions is used to define the permissions for a channel.
type ChannelPermissions uint64

// Channel is used to define a channel inside a guild.
type Channel struct {
	// ID is the ID of the channel.
	ID uint64 `msgpack:"id"`

	// Name is the name of the channel.
	Name string `msgpack:"name"`

	// Type is the type of the channel.
	Type ChannelType `msgpack:"type"`

	// Permissions is the permissions of the channel.
	Permissions ChannelPermissions `msgpack:"permissions"`
}

// Member is used to define the partial member object.
type Member struct {
	// ID is the ID of the member.
	ID uint64 `msgpack:"id"`

	// Hostname is the hostname of the member.
	Hostname string `msgpack:"hostname"`

	// TODO
}

// Guild is used to define any information about guilds.
type Guild struct {
	// ID is the ID of the guild.
	ID uint64 `msgpack:"id"`

	// Hostname is the hostname where the guild is. Note that this should be blank if this is a cross node socket.
	Hostname string `msgpack:"hostname,omitempty"`

	// Available is whether the guild is available. THe remainder of the struct is not filled if this is false.
	Available bool `msgpack:"available"`

	// Name is the name of the guild.
	Name string `msgpack:"name,omitempty"`

	// Icon is the icon of the guild.
	Icon string `msgpack:"icon,omitempty"`

	// Splash is the splash of the guild.
	Splash string `msgpack:"splash,omitempty"`

	// Channels is the channels of the guild.
	Channels []Channel `msgpack:"channels,omitempty"`

	// OwnerID is the ID of the owner of the guild.
	OwnerID uint64 `msgpack:"owner_id,omitempty"`

	// Members is the members of the guild.
	Members []*Member `msgpack:"members,omitempty"`
}

// ReadyPayload is used to define a ready payload.
type ReadyPayload struct {
	// AvailableGuilds is the guilds that are available.
	AvailableGuilds []Guild `msgpack:"available_guilds"`

	// UnavailableGuilds is the guilds that are unavailable.
	UnavailableGuilds []Guild `msgpack:"unavailable_guilds"`
}

// generate_serializer_case ReadyPayload 4
