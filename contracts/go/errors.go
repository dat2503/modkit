package contracts

import "context"

// ErrorTrackingService reports errors and exceptions to a monitoring platform.
// Must be initialized second, immediately after ObservabilityService.
// All unhandled errors should be reported here before the process exits.
type ErrorTrackingService interface {
	// CaptureError reports an error with optional contextual data.
	// Call this for unexpected errors that should alert on-call.
	CaptureError(ctx context.Context, err error, opts CaptureOptions) error

	// CaptureMessage reports a non-error message (e.g. a warning or diagnostic event).
	CaptureMessage(ctx context.Context, msg string, level ErrorLevel, opts CaptureOptions) error

	// SetUser associates subsequent events in this context with the given user.
	// Call this after authenticating a request.
	SetUser(ctx context.Context, user ErrorUser) context.Context

	// Flush waits for all pending events to be sent, up to the given timeout.
	// Call this during graceful shutdown before exiting.
	Flush(ctx context.Context) error
}

// CaptureOptions provides additional context for a captured event.
type CaptureOptions struct {
	// Tags are key-value pairs for filtering events in the dashboard.
	Tags map[string]string

	// Extra holds arbitrary extra data attached to the event.
	Extra map[string]any

	// Fingerprint overrides the grouping fingerprint (for custom grouping rules).
	Fingerprint []string
}

// ErrorUser identifies the user associated with an error event.
type ErrorUser struct {
	// ID is the user's application ID.
	ID string

	// Email is the user's email address.
	Email string
}

// ErrorLevel represents the severity of a captured message.
type ErrorLevel string

const (
	ErrorLevelDebug   ErrorLevel = "debug"
	ErrorLevelInfo    ErrorLevel = "info"
	ErrorLevelWarning ErrorLevel = "warning"
	ErrorLevelError   ErrorLevel = "error"
	ErrorLevelFatal   ErrorLevel = "fatal"
)
