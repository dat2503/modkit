/**
 * IRealtimeService manages real-time bidirectional communication via WebSockets.
 * Publish events from the server to connected clients; clients subscribe to event types.
 * Requires the auth and cache modules.
 */
export interface IRealtimeService {
  /**
   * Broadcasts an event to all clients subscribed to the given topic.
   * Returns the number of clients that received the event.
   */
  publish(topic: string, payload: unknown): Promise<number>;

  /**
   * Broadcasts an event to all connections for a specific user.
   * Use for user-specific notifications (e.g. "your invoice was paid").
   */
  publishToUser(userId: string, event: RealtimeEvent): Promise<void>;

  /**
   * Registers a handler for incoming messages on the given topic.
   * Used for bidirectional scenarios where the client sends messages to the server.
   */
  subscribe(topic: string, handler: RealtimeHandler): void;

  /**
   * Upgrades an HTTP connection to WebSocket and registers the client.
   * Call this from your WebSocket upgrade handler.
   */
  handleConnection(conn: RealtimeConn, userId: string): Promise<void>;

  /**
   * Closes all connections for the given user.
   */
  disconnect(userId: string): Promise<void>;

  /**
   * Returns the set of user IDs with active connections.
   */
  connectedUsers(): Promise<string[]>;
}

/** A message sent to a connected client. */
export interface RealtimeEvent {
  /** Event type string (e.g. "invoice.paid", "message.new"). */
  type: string;

  /** Event data — will be JSON-encoded and sent to the client. */
  payload: unknown;
}

/** Abstracts a raw WebSocket connection. */
export interface RealtimeConn {
  /** Sends data as a WebSocket text message. */
  send(data: string): void;

  /** Registers a callback for incoming messages. */
  onMessage(handler: (data: string) => void): void;

  /** Registers a callback for when the connection closes. */
  onClose(handler: () => void): void;

  /** Closes the WebSocket connection. */
  close(): void;
}

/** Processes incoming messages from clients. */
export type RealtimeHandler = (userId: string, payload: unknown) => Promise<void>;
