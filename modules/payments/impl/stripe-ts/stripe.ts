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
    throw new Error('not implemented')
  }

  async getCheckoutSession(sessionId: string): Promise<CheckoutSession> {
    // TODO: implement using stripe.checkout.sessions.retrieve(sessionId)
    throw new Error('not implemented')
  }

  async createCustomer(req: CreateCustomerRequest): Promise<Customer> {
    // TODO: implement using stripe.customers.create() — check existing first
    throw new Error('not implemented')
  }

  async getCustomer(customerId: string): Promise<Customer> {
    // TODO: implement using stripe.customers.retrieve(customerId)
    throw new Error('not implemented')
  }

  async getSubscription(subscriptionId: string): Promise<Subscription> {
    // TODO: implement using stripe.subscriptions.retrieve(subscriptionId)
    throw new Error('not implemented')
  }

  async cancelSubscription(subscriptionId: string, atPeriodEnd = true): Promise<void> {
    // TODO: implement using stripe.subscriptions.cancel() or update(cancel_at_period_end)
    throw new Error('not implemented')
  }

  async constructWebhookEvent(payload: Buffer | string, signature: string): Promise<WebhookEvent> {
    // TODO: implement using stripe.webhooks.constructEvent(payload, signature, this.config.webhookSecret)
    throw new Error('not implemented')
  }
}
