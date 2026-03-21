// Package sentry implements the ErrorTrackingService interface using Sentry.
package sentry

import (
	"context"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Config holds the configuration for the Sentry error tracking provider.
type Config struct {
	// DSN is the Sentry Data Source Name from your project settings.
	DSN string

	// Environment is the environment tag (e.g. "production", "staging").
	Environment string

	// TracesSampleRate is the fraction of transactions to capture (0.0–1.0).
	TracesSampleRate float64
}

// Service implements contracts.ErrorTrackingService using Sentry.
type Service struct {
	cfg Config
	// TODO: add sentry-go hub
}

// New creates a new Sentry error tracking service.
func New(cfg Config) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) CaptureError(ctx context.Context, err error, opts contracts.CaptureOptions) error {
	// TODO: implement using github.com/getsentry/sentry-go sentry.CaptureException(err)
	panic("not implemented")
}

func (s *Service) CaptureMessage(ctx context.Context, msg string, level contracts.ErrorLevel, opts contracts.CaptureOptions) error {
	// TODO: implement using sentry.CaptureMessage(msg) with level
	panic("not implemented")
}

func (s *Service) SetUser(ctx context.Context, user contracts.ErrorUser) context.Context {
	// TODO: implement using sentry.ConfigureScope(func(scope) { scope.SetUser(...) })
	panic("not implemented")
}

func (s *Service) Flush(ctx context.Context) error {
	// TODO: implement using sentry.Flush(timeout)
	panic("not implemented")
}
