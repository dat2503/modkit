# Payments Module

Payment processing for modkit projects using Stripe.

## Overview

The payments module wraps [Stripe](https://stripe.com) to provide checkout sessions, subscription management, and webhook event processing. It never stores card data — all PCI compliance is delegated to Stripe.

## Implementations

| Name | Label | Phase | Runtimes |
|------|-------|-------|---------|
| `stripe` | Stripe | MVP | Go, Bun |

## Setup

1. Create a Stripe account at [stripe.com](https://stripe.com)
2. Get your API keys from the Stripe Dashboard → Developers → API keys
3. Set env vars:
   ```
   STRIPE_SECRET_KEY=sk_test_...
   STRIPE_WEBHOOK_SECRET=whsec_...
   ```
4. Set up a webhook endpoint in Stripe Dashboard → Developers → Webhooks
   - URL: `https://yourapp.com/api/v1/webhooks/stripe`
   - Events: `checkout.session.completed`, `customer.subscription.deleted`, `customer.subscription.updated`

## Local webhook testing

Use the Stripe CLI to forward webhooks to your local server:
```bash
stripe listen --forward-to localhost:8080/api/v1/webhooks/stripe
```

## Configuration

See `config.schema.json` for all environment variables.

## Dependencies

- **observability** (optional) — traces checkout and webhook calls
- **error-tracking** (optional) — reports payment processing errors
