// Package resend implements the EmailService interface using Resend (resend.com).
package resend

import (
	"context"
	"fmt"

	resendclient "github.com/resendlabs/resend-go"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Config holds the configuration for the Resend email provider.
type Config struct {
	// APIKey is the Resend API key (re_...).
	APIKey string

	// FromDefault is the default from address used when EmailMessage.From is empty.
	FromDefault string
}

// Service implements contracts.EmailService using Resend.
type Service struct {
	cfg    Config
	client *resendclient.Client
}

// New creates a new Resend email service.
func New(cfg Config) (*Service, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("resend: APIKey is required")
	}
	if cfg.FromDefault == "" {
		return nil, fmt.Errorf("resend: FromDefault is required")
	}
	return &Service{cfg: cfg, client: resendclient.NewClient(cfg.APIKey)}, nil
}

// Send sends a single transactional email via Resend.
func (s *Service) Send(ctx context.Context, msg contracts.EmailMessage) (*contracts.EmailResult, error) {
	from := msg.From
	if from == "" {
		from = s.cfg.FromDefault
	}
	params := &resendclient.SendEmailRequest{
		From:    from,
		To:      msg.To,
		Subject: msg.Subject,
		Html:    msg.Body.HTML,
		Text:    msg.Body.Text,
		ReplyTo: msg.ReplyTo,
	}
	resp, err := s.client.Emails.Send(params)
	if err != nil {
		return nil, fmt.Errorf("resend: send: %w", err)
	}
	return &contracts.EmailResult{MessageID: resp.Id}, nil
}

// SendBatch sends multiple emails sequentially via Resend.
// Resend's batch API is used when available; falls back to sequential sends.
func (s *Service) SendBatch(ctx context.Context, msgs []contracts.EmailMessage) ([]*contracts.EmailResult, error) {
	results := make([]*contracts.EmailResult, 0, len(msgs))
	for i, msg := range msgs {
		result, err := s.Send(ctx, msg)
		if err != nil {
			return results, fmt.Errorf("resend: send batch[%d]: %w", i, err)
		}
		results = append(results, result)
	}
	return results, nil
}
