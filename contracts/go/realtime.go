package contracts

import "context"

// RealtimeService manages real-time bidirectional communication via WebSockets.
// Publish events from the server to connected clients; clients subscribe to event types.
// Requires the auth and cache modules.
type RealtimeService interface {
	// Publish broadcasts an event to all clients subscribed to the given topic.
	// payload must be JSON-serializable.
	// Returns the number of clients that received the event.
	Publish(ctx context.Context, topic string, payload any) (int, error)

	// PublishToUser broadcasts an event to all connections for a specific user.
	// Use this to send user-specific notifications (e.g. "your invoice was paid").
	PublishToUser(ctx context.Context, userID string, event RealtimeEvent) error

	// Subscribe registers a handler for incoming messages on the given topic.
	// Used for bidirectional scenarios (client sends messages to server).
	Subscribe(topic string, handler RealtimeHandler) error

	// HandleConnection upgrades an HTTP connection to WebSocket and registers the client.
	// Call this from your WebSocket upgrade handler.
	// conn is the raw WebSocket connection (implementation-specific type cast internally).
	HandleConnection(ctx context.Context, conn RealtimeConn, userID string) error

	// Disconnect closes all connections for the given user.
	Disconnect(ctx context.Context, userID string) error

	// ConnectedUsers returns the set of user IDs with active connections.
	ConnectedUsers(ctx context.Context) ([]string, error)
}

// RealtimeEvent is a message sent to a connected client.
type RealtimeEvent struct {
	// Type is the event type string (e.g. "invoice.paid", "message.new").
	Type string

	// Payload is the event data — will be JSON-encoded and sent to the client.
	Payload any
}

// RealtimeConn abstracts a raw WebSocket connection.
// Implementations cast this to their underlying connection type.
type RealtimeConn interface {
	// WriteJSON encodes v as JSON and sends it as a WebSocket message.
	WriteJSON(v any) error

	// ReadJSON reads the next WebSocket message and decodes it into v.
	ReadJSON(v any) error

	// Close closes the WebSocket connection.
	Close() error
}

// RealtimeHandler processes incoming messages from clients.
// topic is the subscribed topic, userID is the sender, payload is the raw JSON message.
type RealtimeHandler func(ctx context.Context, userID string, payload []byte) error
