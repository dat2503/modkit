// Package stripe implements the PaymentsService interface using Stripe.
package stripe

import (
	"encoding/json"
	"fmt"

	"context"

	"github.com/stripe/stripe-go/v76/client"
	stripego "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"

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
	sc  *client.API
}

// New creates a new Stripe payments service.
func New(cfg Config) (*Service, error) {
	if cfg.SecretKey == "" {
		return nil, fmt.Errorf("stripe: SecretKey is required")
	}
	if cfg.WebhookSecret == "" {
		return nil, fmt.Errorf("stripe: WebhookSecret is required")
	}
	sc := &client.API{}
	sc.Init(cfg.SecretKey, nil)
	return &Service{cfg: cfg, sc: sc}, nil
}

func (s *Service) CreateCheckoutSession(_ context.Context, req contracts.CreateCheckoutRequest) (*contracts.CheckoutSession, error) {
	items := make([]*stripego.CheckoutSessionLineItemParams, len(req.LineItems))
	for i, item := range req.LineItems {
		items[i] = &stripego.CheckoutSessionLineItemParams{
			Price:    stripego.String(item.PriceID),
			Quantity: stripego.Int64(int64(item.Quantity)),
		}
	}
	params := &stripego.CheckoutSessionParams{
		Mode:       stripego.String(req.Mode),
		LineItems:  items,
		SuccessURL: stripego.String(req.SuccessURL),
		CancelURL:  stripego.String(req.CancelURL),
	}
	if req.CustomerID != "" {
		params.Customer = stripego.String(req.CustomerID)
	}
	if len(req.Metadata) > 0 {
		params.Metadata = req.Metadata
	}

	sess, err := s.sc.CheckoutSessions.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe: create checkout session: %w", err)
	}
	return &contracts.CheckoutSession{
		ID:         sess.ID,
		URL:        sess.URL,
		Status:     string(sess.Status),
		CustomerID: customerID(sess),
		Metadata:   sess.Metadata,
	}, nil
}

func (s *Service) GetCheckoutSession(_ context.Context, sessionID string) (*contracts.CheckoutSession, error) {
	sess, err := s.sc.CheckoutSessions.Get(sessionID, nil)
	if err != nil {
		return nil, fmt.Errorf("stripe: get checkout session %q: %w", sessionID, err)
	}
	return &contracts.CheckoutSession{
		ID:         sess.ID,
		URL:        sess.URL,
		Status:     string(sess.Status),
		CustomerID: customerID(sess),
		Metadata:   sess.Metadata,
	}, nil
}

func (s *Service) CreateCustomer(_ context.Context, req contracts.CreateCustomerRequest) (*contracts.Customer, error) {
	params := &stripego.CustomerParams{
		Email: stripego.String(req.Email),
		Name:  stripego.String(req.Name),
	}
	if len(req.Metadata) > 0 {
		params.Metadata = req.Metadata
	}
	cust, err := s.sc.Customers.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe: create customer: %w", err)
	}
	return &contracts.Customer{ID: cust.ID, Email: cust.Email, Name: cust.Name}, nil
}

func (s *Service) GetCustomer(_ context.Context, customerID string) (*contracts.Customer, error) {
	cust, err := s.sc.Customers.Get(customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("stripe: get customer %q: %w", customerID, err)
	}
	return &contracts.Customer{ID: cust.ID, Email: cust.Email, Name: cust.Name}, nil
}

func (s *Service) GetSubscription(_ context.Context, subscriptionID string) (*contracts.Subscription, error) {
	sub, err := s.sc.Subscriptions.Get(subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("stripe: get subscription %q: %w", subscriptionID, err)
	}
	priceID := ""
	if len(sub.Items.Data) > 0 {
		priceID = sub.Items.Data[0].Price.ID
	}
	return &contracts.Subscription{
		ID:               sub.ID,
		CustomerID:       sub.Customer.ID,
		Status:           string(sub.Status),
		PriceID:          priceID,
		CurrentPeriodEnd: sub.CurrentPeriodEnd,
	}, nil
}

func (s *Service) CancelSubscription(_ context.Context, subscriptionID string, atPeriodEnd bool) error {
	if atPeriodEnd {
		_, err := s.sc.Subscriptions.Update(subscriptionID, &stripego.SubscriptionParams{
			CancelAtPeriodEnd: stripego.Bool(true),
		})
		if err != nil {
			return fmt.Errorf("stripe: schedule subscription cancel %q: %w", subscriptionID, err)
		}
		return nil
	}
	_, err := s.sc.Subscriptions.Cancel(subscriptionID, nil)
	if err != nil {
		return fmt.Errorf("stripe: cancel subscription %q: %w", subscriptionID, err)
	}
	return nil
}

func (s *Service) ConstructWebhookEvent(payload []byte, signature string) (*contracts.WebhookEvent, error) {
	event, err := webhook.ConstructEvent(payload, signature, s.cfg.WebhookSecret)
	if err != nil {
		return nil, fmt.Errorf("stripe: construct webhook event: %w", err)
	}
	data, err := json.Marshal(event.Data.Object)
	if err != nil {
		return nil, fmt.Errorf("stripe: marshal webhook data: %w", err)
	}
	return &contracts.WebhookEvent{
		ID:   event.ID,
		Type: string(event.Type),
		Data: data,
	}, nil
}

// customerID safely extracts the customer ID from a checkout session,
// which may hold a string ID or an expanded Customer object.
func customerID(sess *stripego.CheckoutSession) string {
	if sess.Customer != nil {
		return sess.Customer.ID
	}
	return ""
}
