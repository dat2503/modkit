package contracts

import "context"

// PaymentsService handles payment processing for one-time and recurring payments.
// Never store card data — always delegate to the payment provider.
type PaymentsService interface {
	// CreateCheckoutSession creates a hosted checkout page for the given items.
	// Returns a URL to redirect the user to for payment.
	CreateCheckoutSession(ctx context.Context, req CreateCheckoutRequest) (*CheckoutSession, error)

	// GetCheckoutSession retrieves an existing checkout session by ID.
	GetCheckoutSession(ctx context.Context, sessionID string) (*CheckoutSession, error)

	// CreateCustomer creates or retrieves a customer record in the payment provider.
	// Idempotent: if a customer with the given email already exists, returns the existing one.
	CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*Customer, error)

	// GetCustomer retrieves a customer by their provider-assigned ID.
	GetCustomer(ctx context.Context, customerID string) (*Customer, error)

	// GetSubscription retrieves an active subscription by ID.
	GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error)

	// CancelSubscription cancels an active subscription.
	// If atPeriodEnd is true, the subscription remains active until the end of the current period.
	CancelSubscription(ctx context.Context, subscriptionID string, atPeriodEnd bool) error

	// ConstructWebhookEvent validates and parses an incoming webhook payload.
	// Returns an error if the signature is invalid.
	// Use this in your webhook handler before processing any event.
	ConstructWebhookEvent(payload []byte, signature string) (*WebhookEvent, error)
}

// CreateCheckoutRequest describes a checkout session to create.
type CreateCheckoutRequest struct {
	// CustomerID is the provider-assigned customer ID. Optional — creates an anonymous session if empty.
	CustomerID string

	// LineItems is the list of items to charge.
	LineItems []LineItem

	// Mode is the checkout mode: "payment" for one-time, "subscription" for recurring.
	Mode string

	// SuccessURL is the URL to redirect to after successful payment.
	SuccessURL string

	// CancelURL is the URL to redirect to if the user cancels.
	CancelURL string

	// Metadata holds arbitrary key-value pairs attached to this checkout session.
	Metadata map[string]string
}

// LineItem is a single item in a checkout session.
type LineItem struct {
	// PriceID is the provider-assigned price ID (e.g. Stripe Price ID).
	PriceID string

	// Quantity is the number of units.
	Quantity int
}

// CheckoutSession represents a hosted checkout page.
type CheckoutSession struct {
	// ID is the provider-assigned session ID.
	ID string

	// URL is the hosted checkout URL to redirect the user to.
	URL string

	// Status is the current status: "open", "complete", or "expired".
	Status string

	// CustomerID is the associated customer ID. May be empty for anonymous checkouts.
	CustomerID string

	// Metadata holds the metadata attached when the session was created.
	Metadata map[string]string
}

// CreateCustomerRequest describes a customer to create.
type CreateCustomerRequest struct {
	// Email is the customer's email address.
	Email string

	// Name is the customer's display name.
	Name string

	// Metadata holds arbitrary key-value pairs attached to this customer.
	Metadata map[string]string
}

// Customer represents a customer in the payment provider.
type Customer struct {
	// ID is the provider-assigned customer ID.
	ID string

	// Email is the customer's email address.
	Email string

	// Name is the customer's display name.
	Name string
}

// Subscription represents an active recurring subscription.
type Subscription struct {
	// ID is the provider-assigned subscription ID.
	ID string

	// CustomerID is the associated customer ID.
	CustomerID string

	// Status is the subscription status: "active", "canceled", "past_due", etc.
	Status string

	// PriceID is the price this subscription is for.
	PriceID string

	// CurrentPeriodEnd is the Unix timestamp when the current billing period ends.
	CurrentPeriodEnd int64
}

// WebhookEvent is a parsed and verified incoming webhook event.
type WebhookEvent struct {
	// ID is the provider-assigned event ID (use for idempotency).
	ID string

	// Type is the event type (e.g. "checkout.session.completed").
	Type string

	// Data holds the event payload as a raw JSON byte slice.
	// Unmarshal into the appropriate type based on Type.
	Data []byte
}
