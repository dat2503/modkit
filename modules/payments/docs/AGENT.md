# Payments Module — Agent Instructions

## When to use

Include this module when:
- The app processes payments (one-time or recurring)
- The app has a subscription/freemium model
- Users purchase anything within the app

Do NOT use when:
- The app is fully free with no monetization
- Payments are handled outside the app (e.g. manual invoicing)

## How to wire

### Go

1. Import `PaymentsService` from `contracts/go/payments.go`
2. Initialize in bootstrap:
   ```go
   paymentsSvc := stripe.New(stripe.Config{
       SecretKey:     cfg.Payments.SecretKey,
       WebhookSecret: cfg.Payments.WebhookSecret,
   })
   ```
3. Inject into handlers that need to create checkouts or manage subscriptions
4. Register a webhook endpoint — Stripe requires this for subscription lifecycle:
   ```go
   router.Post("/api/v1/webhooks/stripe", stripeWebhookHandler(paymentsSvc))
   ```

### Bun (TypeScript)

1. Import `IPaymentsService` from `contracts/ts/payments.ts`
2. Initialize in bootstrap:
   ```typescript
   const payments = new StripePaymentsService({
     secretKey: config.payments.secretKey,
     webhookSecret: config.payments.webhookSecret,
   })
   ```
3. Register webhook endpoint:
   ```typescript
   app.post('/api/v1/webhooks/stripe', stripeWebhookHandler(payments))
   ```

## Checkout flow

```
Client → POST /api/v1/checkout → createCheckoutSession() → redirect to Stripe URL
Stripe → user pays → redirect to successUrl
Stripe → POST /api/v1/webhooks/stripe → checkout.session.completed event
Backend → fulfill order / activate subscription
```

## Webhook handler pattern (Go)

```go
func stripeWebhookHandler(svc contracts.PaymentsService) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body)
        event, err := svc.ConstructWebhookEvent(body, r.Header.Get("Stripe-Signature"))
        if err != nil {
            writeError(w, http.StatusBadRequest, "invalid signature")
            return
        }
        switch event.Type {
        case "checkout.session.completed":
            // fulfill order
        case "customer.subscription.deleted":
            // deactivate subscription
        }
        w.WriteHeader(http.StatusOK)
    }
}
```

## Idempotency

Stripe may deliver the same webhook event more than once. Use `event.ID` as an
idempotency key — store processed event IDs and skip duplicates.

## Required env vars

```
PAYMENTS_PROVIDER=stripe
STRIPE_SECRET_KEY=sk_test_...        # sensitive
STRIPE_WEBHOOK_SECRET=whsec_...      # sensitive
STRIPE_PUBLISHABLE_KEY=pk_test_...   # frontend only
```

## Do NOT

- Store card numbers or CVVs — Stripe handles all PCI compliance
- Process webhook events without verifying the signature
- Use the secret key on the frontend — publishable key only
- Hardcode price IDs — store them in config or the database
