// Package resend implements the EmailService interface using Resend (resend.com).
package resend

import (
	"context"
	"fmt"

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
	cfg Config
	// TODO: add resend-go client
}

// New creates a new Resend email service.
func New(cfg Config) (*Service, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("resend: APIKey is required")
	}
	if cfg.FromDefault == "" {
		return nil, fmt.Errorf("resend: FromDefault is required")
	}
	return &Service{cfg: cfg}, nil
}

// Send sends a single transactional email via Resend.
func (s *Service) Send(ctx context.Context, msg contracts.EmailMessage) (*contracts.EmailResult, error) {
	// TODO: implement using github.com/resendlabs/resend-go
	// client := resend.NewClient(s.cfg.APIKey)
	// params := &resend.SendEmailRequest{To: msg.To, From: from, Subject: msg.Subject, Html: msg.Body.HTML, Text: msg.Body.Text}
	// resp, err := client.Emails.Send(params)
	panic("not implemented")
}

// SendBatch sends multiple emails in a single Resend API call.
func (s *Service) SendBatch(ctx context.Context, msgs []contracts.EmailMessage) ([]*contracts.EmailResult, error) {
	// TODO: implement using resend batch send API
	panic("not implemented")
}
