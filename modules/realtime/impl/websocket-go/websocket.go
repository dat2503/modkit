// Package websocket implements the RealtimeService interface using native WebSockets.
package websocket

import (
	"context"
	"log"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Config holds the configuration for the WebSocket realtime provider.
type Config struct {
	// MaxConnectionsPerUser limits concurrent connections per user. Default: 5.
	MaxConnectionsPerUser int

	// PingIntervalSeconds is the WebSocket ping interval to detect dead connections. Default: 30.
	PingIntervalSeconds int
}

// Service implements contracts.RealtimeService using native WebSockets with Redis pub/sub.
type Service struct {
	cfg   Config
	cache contracts.CacheService
	// TODO: add connection registry and Redis pub/sub subscriber
}

// New creates a new WebSocket realtime service.
func New(cfg Config, cache contracts.CacheService) *Service {
	if cfg.MaxConnectionsPerUser == 0 {
		cfg.MaxConnectionsPerUser = 5
	}
	if cfg.PingIntervalSeconds == 0 {
		cfg.PingIntervalSeconds = 30
	}
	return &Service{cfg: cfg, cache: cache}
}

func (s *Service) Publish(ctx context.Context, topic string, payload any) (int, error) {
	// TODO: serialize payload, publish to Redis pub/sub channel for topic,
	// return count of connected clients that received it
	log.Printf("[websocket-realtime] stub: Publish() not implemented")
	return 0, nil
}

func (s *Service) PublishToUser(ctx context.Context, userID string, event contracts.RealtimeEvent) error {
	// TODO: look up connections for userID in cache, write event to each
	log.Printf("[websocket-realtime] stub: PublishToUser() not implemented")
	return nil
}

func (s *Service) Subscribe(topic string, handler contracts.RealtimeHandler) error {
	// TODO: register handler for incoming client messages on topic
	log.Printf("[websocket-realtime] stub: Subscribe() not implemented")
	return nil
}

func (s *Service) HandleConnection(ctx context.Context, conn contracts.RealtimeConn, userID string) error {
	// TODO: register conn in cache, start read/write loops, handle ping/pong,
	// deregister on disconnect
	log.Printf("[websocket-realtime] stub: HandleConnection() not implemented")
	return nil
}

func (s *Service) Disconnect(ctx context.Context, userID string) error {
	// TODO: close all connections for userID, remove from cache
	log.Printf("[websocket-realtime] stub: Disconnect() not implemented")
	return nil
}

func (s *Service) ConnectedUsers(ctx context.Context) ([]string, error) {
	// TODO: read connection registry from cache, return unique user IDs
	log.Printf("[websocket-realtime] stub: ConnectedUsers() not implemented")
	return nil, nil
}
