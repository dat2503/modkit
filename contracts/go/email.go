package contracts

import "context"

// EmailService sends transactional emails triggered by user actions.
// Do NOT use this for marketing or bulk email — use a dedicated marketing platform instead.
type EmailService interface {
	// Send sends a single transactional email.
	// Returns the provider-assigned message ID on success.
	Send(ctx context.Context, msg EmailMessage) (*EmailResult, error)

	// SendBatch sends multiple emails in a single API call.
	// Providers may impose limits on batch size — check provider docs.
	SendBatch(ctx context.Context, msgs []EmailMessage) ([]*EmailResult, error)
}

// EmailMessage is a single email to send.
type EmailMessage struct {
	// To is the list of recipient email addresses.
	To []string

	// From is the sender address. Must be a verified domain/address in your provider.
	From string

	// ReplyTo is the optional reply-to address.
	ReplyTo string

	// Subject is the email subject line.
	Subject string

	// Body holds the email content. Provide both HTML and Text for best deliverability.
	Body EmailBody

	// Headers holds additional email headers (e.g. List-Unsubscribe).
	Headers map[string]string

	// Tags hold provider-specific tags for analytics/filtering.
	Tags map[string]string
}

// EmailBody holds the content of an email in multiple formats.
type EmailBody struct {
	// HTML is the HTML version of the email body.
	HTML string

	// Text is the plain-text version. Always provide this as a fallback.
	Text string
}

// EmailResult is the result of a successful Send call.
type EmailResult struct {
	// MessageID is the provider-assigned message ID. Store this for delivery tracking.
	MessageID string
}
