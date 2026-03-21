// Package sentry implements ErrorTrackingService using github.com/getsentry/sentry-go.
package sentry

import (
	"context"
	"fmt"
	"time"

	gosentry "github.com/getsentry/sentry-go"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Config holds the configuration for the Sentry error tracking provider.
type Config struct {
	// SentryDSN is the Sentry Data Source Name from your project settings.
	SentryDSN string

	// Environment is the environment tag (e.g. "production", "staging", "development").
	Environment string

	// TracesSampleRate is the fraction of transactions to capture (0.0–1.0).
	TracesSampleRate float64
}

// Service implements contracts.ErrorTrackingService using Sentry.
type Service struct {
	cfg Config
}

// New initialises the Sentry SDK. In non-production environments a failed init
// logs a warning and continues rather than aborting startup.
func New(cfg Config) (*Service, error) {
	err := gosentry.Init(gosentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.Environment,
		TracesSampleRate: cfg.TracesSampleRate,
	})
	if err != nil {
		if cfg.Environment == "production" {
			return nil, fmt.Errorf("sentry init: %w", err)
		}
		// Non-fatal in dev/staging — a dummy DSN is common during local development.
		fmt.Printf("warning: sentry init failed (non-fatal in %s): %v\n", cfg.Environment, err)
	}
	return &Service{cfg: cfg}, nil
}

// CaptureError reports an error to Sentry.
func (s *Service) CaptureError(ctx context.Context, err error, opts contracts.CaptureOptions) error {
	hub := gosentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = gosentry.CurrentHub()
	}
	hub.WithScope(func(scope *gosentry.Scope) {
		for k, v := range opts.Tags {
			scope.SetTag(k, v)
		}
		for k, v := range opts.Extra {
			scope.SetExtra(k, v)
		}
		if len(opts.Fingerprint) > 0 {
			scope.SetFingerprint(opts.Fingerprint)
		}
		hub.CaptureException(err)
	})
	return nil
}

// CaptureMessage reports a message to Sentry.
func (s *Service) CaptureMessage(ctx context.Context, msg string, level contracts.ErrorLevel, opts contracts.CaptureOptions) error {
	hub := gosentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = gosentry.CurrentHub()
	}
	hub.WithScope(func(scope *gosentry.Scope) {
		scope.SetLevel(toSentryLevel(level))
		for k, v := range opts.Tags {
			scope.SetTag(k, v)
		}
		hub.CaptureMessage(msg)
	})
	return nil
}

// SetUser attaches user context to the current Sentry scope.
func (s *Service) SetUser(ctx context.Context, user contracts.ErrorUser) context.Context {
	hub := gosentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = gosentry.CurrentHub().Clone()
		ctx = gosentry.SetHubOnContext(ctx, hub)
	}
	hub.ConfigureScope(func(scope *gosentry.Scope) {
		scope.SetUser(gosentry.User{
			ID:    user.ID,
			Email: user.Email,
		})
	})
	return ctx
}

// Flush waits for pending events to be sent to Sentry.
func (s *Service) Flush(_ context.Context) error {
	gosentry.Flush(2 * time.Second)
	return nil
}

func toSentryLevel(level contracts.ErrorLevel) gosentry.Level {
	switch level {
	case contracts.ErrorLevelDebug:
		return gosentry.LevelDebug
	case contracts.ErrorLevelInfo:
		return gosentry.LevelInfo
	case contracts.ErrorLevelWarning:
		return gosentry.LevelWarning
	case contracts.ErrorLevelFatal:
		return gosentry.LevelFatal
	default:
		return gosentry.LevelError
	}
}
