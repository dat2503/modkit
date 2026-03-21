package contracts

import (
	"context"
	"time"
)

// JobsService manages a background job queue for async processing and retries.
// Use for any operation >500ms, any side effect that doesn't need to block the HTTP response,
// or any operation requiring retry logic.
// Requires the cache module as the queue backend.
type JobsService interface {
	// Enqueue adds a job to the queue for immediate processing.
	// payload must be JSON-serializable.
	Enqueue(ctx context.Context, jobType string, payload any) (*JobHandle, error)

	// EnqueueIn adds a job to the queue to be processed after the given delay.
	EnqueueIn(ctx context.Context, jobType string, payload any, delay time.Duration) (*JobHandle, error)

	// EnqueueAt adds a job to the queue to be processed at a specific time.
	EnqueueAt(ctx context.Context, jobType string, payload any, processAt time.Time) (*JobHandle, error)

	// RegisterHandler registers a handler function for the given job type.
	// The handler receives the raw JSON payload; unmarshal it into the expected type.
	// Handlers must be idempotent — they may be called more than once for the same job.
	RegisterHandler(jobType string, handler JobHandler) error

	// Start begins processing jobs from the queue.
	// This is a blocking call — run it in a goroutine.
	Start(ctx context.Context) error

	// Stop gracefully shuts down the job processor, waiting for in-flight jobs to complete.
	Stop(ctx context.Context) error
}

// JobHandle is returned when a job is successfully enqueued.
type JobHandle struct {
	// ID is the queue-assigned unique job ID.
	ID string

	// Type is the job type that was enqueued.
	Type string

	// Queue is the queue the job was placed in.
	Queue string
}

// JobHandler is a function that processes a single job.
// ctx carries the request context. payload is the raw JSON bytes for the job.
// Return nil on success. Return an error to trigger a retry.
type JobHandler func(ctx context.Context, payload []byte) error

// JobOptions controls how a job behaves when enqueued.
type JobOptions struct {
	// Queue specifies which queue to place the job in.
	// If empty, uses the default queue.
	Queue string

	// MaxRetries is the maximum number of retry attempts on failure.
	// Defaults to the global configured value (typically 3).
	MaxRetries int

	// UniqueKey prevents duplicate jobs with the same key from being enqueued.
	// If a job with the same UniqueKey already exists in the queue, the new job is dropped.
	UniqueKey string
}
