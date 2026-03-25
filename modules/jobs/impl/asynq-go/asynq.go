// Package asynq implements the JobsService interface using Asynq (Go only).
// Asynq uses Redis as its queue backend — reuse the Redis URL from the cache module.
package asynq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

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
	redisOpt asynq.RedisConnOpt
	client   *asynq.Client
	server   *asynq.Server
	mux      *asynq.ServeMux
}

// New creates a new Asynq jobs service.
func New(cfg Config) (*Service, error) {
	if cfg.RedisURL == "" {
		return nil, fmt.Errorf("asynq: RedisURL is required")
	}
	if cfg.Concurrency == 0 {
		cfg.Concurrency = 10
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.DefaultQueue == "" {
		cfg.DefaultQueue = "default"
	}

	redisOpt, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("asynq: parse redis URL: %w", err)
	}

	client := asynq.NewClient(redisOpt)
	mux := asynq.NewServeMux()

	return &Service{
		cfg:      cfg,
		redisOpt: redisOpt,
		client:   client,
		mux:      mux,
	}, nil
}

func (s *Service) Enqueue(ctx context.Context, jobType string, payload any) (*contracts.JobHandle, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("asynq: marshal payload for %q: %w", jobType, err)
	}
	task := asynq.NewTask(jobType, data)
	info, err := s.client.EnqueueContext(ctx, task,
		asynq.Queue(s.cfg.DefaultQueue),
		asynq.MaxRetry(s.cfg.MaxRetries),
	)
	if err != nil {
		return nil, fmt.Errorf("asynq: enqueue %q: %w", jobType, err)
	}
	return &contracts.JobHandle{ID: info.ID, Type: info.Type, Queue: info.Queue}, nil
}

func (s *Service) EnqueueIn(ctx context.Context, jobType string, payload any, delay time.Duration) (*contracts.JobHandle, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("asynq: marshal payload for %q: %w", jobType, err)
	}
	task := asynq.NewTask(jobType, data)
	info, err := s.client.EnqueueContext(ctx, task,
		asynq.Queue(s.cfg.DefaultQueue),
		asynq.MaxRetry(s.cfg.MaxRetries),
		asynq.ProcessIn(delay),
	)
	if err != nil {
		return nil, fmt.Errorf("asynq: enqueue-in %q: %w", jobType, err)
	}
	return &contracts.JobHandle{ID: info.ID, Type: info.Type, Queue: info.Queue}, nil
}

func (s *Service) EnqueueAt(ctx context.Context, jobType string, payload any, processAt time.Time) (*contracts.JobHandle, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("asynq: marshal payload for %q: %w", jobType, err)
	}
	task := asynq.NewTask(jobType, data)
	info, err := s.client.EnqueueContext(ctx, task,
		asynq.Queue(s.cfg.DefaultQueue),
		asynq.MaxRetry(s.cfg.MaxRetries),
		asynq.ProcessAt(processAt),
	)
	if err != nil {
		return nil, fmt.Errorf("asynq: enqueue-at %q: %w", jobType, err)
	}
	return &contracts.JobHandle{ID: info.ID, Type: info.Type, Queue: info.Queue}, nil
}

func (s *Service) RegisterHandler(jobType string, handler contracts.JobHandler) error {
	s.mux.HandleFunc(jobType, func(ctx context.Context, t *asynq.Task) error {
		return handler(ctx, t.Payload())
	})
	return nil
}

func (s *Service) Start(ctx context.Context) error {
	srv := asynq.NewServer(s.redisOpt, asynq.Config{
		Concurrency: s.cfg.Concurrency,
		Queues:      map[string]int{s.cfg.DefaultQueue: 1},
	})
	s.server = srv
	if err := srv.Start(s.mux); err != nil {
		return fmt.Errorf("asynq: start server: %w", err)
	}
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	if s.server != nil {
		s.server.Shutdown()
	}
	if err := s.client.Close(); err != nil {
		return fmt.Errorf("asynq: close client: %w", err)
	}
	return nil
}
