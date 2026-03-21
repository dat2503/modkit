// Package asynq implements the JobsService interface using Asynq (Go only).
// Asynq uses Redis as its queue backend — reuse the Redis URL from the cache module.
package asynq

import (
	"context"
	"time"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Config holds the configuration for the Asynq job queue.
type Config struct {
	// RedisURL is the Redis connection URL — reuse the cache module's REDIS_URL.
	RedisURL string

	// Concurrency is the number of jobs to process concurrently. Default: 10.
	Concurrency int

	// MaxRetries is the maximum retry attempts per job. Default: 3.
	MaxRetries int

	// DefaultQueue is the default queue name. Default: "default".
	DefaultQueue string
}

// Service implements contracts.JobsService using Asynq.
type Service struct {
	cfg      Config
	handlers map[string]contracts.JobHandler
	// TODO: add asynq client and server
}

// New creates a new Asynq jobs service.
func New(cfg Config) *Service {
	if cfg.Concurrency == 0 {
		cfg.Concurrency = 10
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.DefaultQueue == "" {
		cfg.DefaultQueue = "default"
	}
	return &Service{cfg: cfg, handlers: make(map[string]contracts.JobHandler)}
}

func (s *Service) Enqueue(ctx context.Context, jobType string, payload any) (*contracts.JobHandle, error) {
	// TODO: implement using github.com/hibiken/asynq client.Enqueue(asynq.NewTask(jobType, jsonPayload))
	panic("not implemented")
}

func (s *Service) EnqueueIn(ctx context.Context, jobType string, payload any, delay time.Duration) (*contracts.JobHandle, error) {
	// TODO: implement using asynq.ProcessIn(delay) option
	panic("not implemented")
}

func (s *Service) EnqueueAt(ctx context.Context, jobType string, payload any, processAt time.Time) (*contracts.JobHandle, error) {
	// TODO: implement using asynq.ProcessAt(processAt) option
	panic("not implemented")
}

func (s *Service) RegisterHandler(jobType string, handler contracts.JobHandler) error {
	s.handlers[jobType] = handler
	return nil
}

func (s *Service) Start(ctx context.Context) error {
	// TODO: implement asynq.NewServer + ServeMux with registered handlers
	panic("not implemented")
}

func (s *Service) Stop(ctx context.Context) error {
	// TODO: implement graceful server shutdown
	panic("not implemented")
}
