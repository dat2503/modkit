// Package stripe implements the PaymentsService interface using Stripe.
package stripe

import (
	"context"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Config holds the configuration for the Stripe payments provider.
type Config struct {
	// SecretKey is the Stripe secret key (sk_live_... or sk_test_...).
	SecretKey string

	// WebhookSecret is the Stripe webhook signing secret (whsec_...).
	WebhookSecret string
}

// Service implements contracts.PaymentsService using Stripe.
type Service struct {
	cfg Config
	// TODO: add stripe-go client
}

// New creates a new Stripe payments service.
func New(cfg Config) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) CreateCheckoutSession(ctx context.Context, req contracts.CreateCheckoutRequest) (*contracts.CheckoutSession, error) {
	// TODO: implement using stripe-go stripe.CheckoutSession.New()
	panic("not implemented")
}

func (s *Service) GetCheckoutSession(ctx context.Context, sessionID string) (*contracts.CheckoutSession, error) {
	// TODO: implement using stripe-go stripe.CheckoutSession.Get()
	panic("not implemented")
}

func (s *Service) CreateCustomer(ctx context.Context, req contracts.CreateCustomerRequest) (*contracts.Customer, error) {
	// TODO: implement using stripe-go stripe.Customer.New() — check for existing first
	panic("not implemented")
}

func (s *Service) GetCustomer(ctx context.Context, customerID string) (*contracts.Customer, error) {
	// TODO: implement using stripe-go stripe.Customer.Get()
	panic("not implemented")
}

func (s *Service) GetSubscription(ctx context.Context, subscriptionID string) (*contracts.Subscription, error) {
	// TODO: implement using stripe-go stripe.Subscription.Get()
	panic("not implemented")
}

func (s *Service) CancelSubscription(ctx context.Context, subscriptionID string, atPeriodEnd bool) error {
	// TODO: implement using stripe-go stripe.Subscription.Cancel() or Update(cancel_at_period_end)
	panic("not implemented")
}

func (s *Service) ConstructWebhookEvent(payload []byte, signature string) (*contracts.WebhookEvent, error) {
	// TODO: implement using webhook.ConstructEvent(payload, signature, s.cfg.WebhookSecret)
	panic("not implemented")
}
