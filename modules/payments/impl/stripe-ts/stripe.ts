import type {
  IPaymentsService,
  CreateCheckoutRequest,
  CheckoutSession,
  CreateCustomerRequest,
  Customer,
  Subscription,
  WebhookEvent,
} from '../../../contracts/ts/payments'

export interface StripeConfig {
  secretKey: string
  webhookSecret: string
}

/**
 * StripePaymentsService implements IPaymentsService using Stripe.
 */
export class StripePaymentsService implements IPaymentsService {
  constructor(private readonly config: StripeConfig) {}

  async createCheckoutSession(req: CreateCheckoutRequest): Promise<CheckoutSession> {
    // TODO: implement using stripe npm package
    // const session = await stripe.checkout.sessions.create({...})
    console.warn('[stripe-payments] stub: createCheckoutSession() not implemented')
    return { id: '', url: '', status: 'open' }
  }

  async getCheckoutSession(sessionId: string): Promise<CheckoutSession> {
    // TODO: implement using stripe.checkout.sessions.retrieve(sessionId)
    console.warn('[stripe-payments] stub: getCheckoutSession() not implemented')
    return { id: sessionId, url: '', status: 'expired' }
  }

  async createCustomer(req: CreateCustomerRequest): Promise<Customer> {
    // TODO: implement using stripe.customers.create() — check existing first
    console.warn('[stripe-payments] stub: createCustomer() not implemented')
    return { id: '', email: req.email }
  }

  async getCustomer(customerId: string): Promise<Customer> {
    // TODO: implement using stripe.customers.retrieve(customerId)
    console.warn('[stripe-payments] stub: getCustomer() not implemented')
    return { id: customerId, email: '' }
  }

  async getSubscription(subscriptionId: string): Promise<Subscription> {
    // TODO: implement using stripe.subscriptions.retrieve(subscriptionId)
    console.warn('[stripe-payments] stub: getSubscription() not implemented')
    return { id: subscriptionId, customerId: '', status: 'canceled', priceId: '', currentPeriodEnd: 0 }
  }

  async cancelSubscription(subscriptionId: string, atPeriodEnd = true): Promise<void> {
    // TODO: implement using stripe.subscriptions.cancel() or update(cancel_at_period_end)
    console.warn('[stripe-payments] stub: cancelSubscription() not implemented')
  }

  async constructWebhookEvent(payload: Buffer | string, signature: string): Promise<WebhookEvent> {
    // TODO: implement using stripe.webhooks.constructEvent(payload, signature, this.config.webhookSecret)
    console.warn('[stripe-payments] stub: constructWebhookEvent() not implemented')
    return { id: '', type: '', data: {} }
  }
}
