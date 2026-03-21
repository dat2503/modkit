import type { IRealtimeService, RealtimeEvent, RealtimeConn, RealtimeHandler } from '../../../contracts/ts/realtime'
import type { ICacheService } from '../../../contracts/ts/cache'

export interface WebSocketConfig {
  maxConnectionsPerUser?: number
  pingIntervalSeconds?: number
}

/**
 * WebSocketRealtimeService implements IRealtimeService using native WebSockets with Redis pub/sub.
 */
export class WebSocketRealtimeService implements IRealtimeService {
  private readonly handlers = new Map<string, RealtimeHandler>()

  constructor(
    private readonly config: WebSocketConfig,
    private readonly cache: ICacheService,
  ) {}

  async publish(topic: string, payload: unknown): Promise<number> {
    // TODO: serialize payload, publish to Redis pub/sub channel for topic,
    // return count of connected clients that received it
    throw new Error('not implemented')
  }

  async publishToUser(userId: string, event: RealtimeEvent): Promise<void> {
    // TODO: look up connections for userId in cache, write event to each
    throw new Error('not implemented')
  }

  subscribe(topic: string, handler: RealtimeHandler): void {
    this.handlers.set(topic, handler)
  }

  async handleConnection(conn: RealtimeConn, userId: string): Promise<void> {
    // TODO: register conn in cache, set up message/close handlers, handle ping/pong,
    // deregister on disconnect
    throw new Error('not implemented')
  }

  async disconnect(userId: string): Promise<void> {
    // TODO: close all connections for userId, remove from cache
    throw new Error('not implemented')
  }

  async connectedUsers(): Promise<string[]> {
    // TODO: read connection registry from cache, return unique user IDs
    throw new Error('not implemented')
  }
}
