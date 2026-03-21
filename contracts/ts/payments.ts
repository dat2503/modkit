/**
 * PaymentsService handles payment processing for one-time and recurring payments.
 * Never store card data — always delegate to the payment provider.
 */
export interface IPaymentsService {
  /**
   * Creates a hosted checkout page for the given items.
   * Returns the checkout session including the URL to redirect the user to.
   */
  createCheckoutSession(req: CreateCheckoutRequest): Promise<CheckoutSession>;

  /**
   * Retrieves an existing checkout session by ID.
   */
  getCheckoutSession(sessionId: string): Promise<CheckoutSession>;

  /**
   * Creates or retrieves a customer record in the payment provider.
   * Idempotent: if a customer with the given email already exists, returns the existing one.
   */
  createCustomer(req: CreateCustomerRequest): Promise<Customer>;

  /**
   * Retrieves a customer by their provider-assigned ID.
   */
  getCustomer(customerId: string): Promise<Customer>;

  /**
   * Retrieves an active subscription by ID.
   */
  getSubscription(subscriptionId: string): Promise<Subscription>;

  /**
   * Cancels an active subscription.
   * If atPeriodEnd is true, the subscription remains active until the current period ends.
   */
  cancelSubscription(subscriptionId: string, atPeriodEnd?: boolean): Promise<void>;

  /**
   * Validates and parses an incoming webhook payload.
   * Throws PaymentsError if the signature is invalid.
   * Always call this before processing any webhook event.
   */
  constructWebhookEvent(payload: Buffer | string, signature: string): Promise<WebhookEvent>;
}

/** Describes a checkout session to create. */
export interface CreateCheckoutRequest {
  /** Provider-assigned customer ID. Optional — creates anonymous session if omitted. */
  customerId?: string;

  /** List of items to charge. */
  lineItems: LineItem[];

  /** Checkout mode: "payment" for one-time, "subscription" for recurring. */
  mode: 'payment' | 'subscription';

  /** URL to redirect to after successful payment. */
  successUrl: string;

  /** URL to redirect to if the user cancels. */
  cancelUrl: string;

  /** Arbitrary key-value pairs attached to this checkout session. */
  metadata?: Record<string, string>;
}

/** A single item in a checkout session. */
export interface LineItem {
  /** Provider-assigned price ID (e.g. Stripe Price ID). */
  priceId: string;

  /** Number of units. */
  quantity: number;
}

/** Represents a hosted checkout page. */
export interface CheckoutSession {
  /** Provider-assigned session ID. */
  id: string;

  /** Hosted checkout URL to redirect the user to. */
  url: string;

  /** Current status: "open" | "complete" | "expired". */
  status: 'open' | 'complete' | 'expired';

  /** Associated customer ID. May be undefined for anonymous checkouts. */
  customerId?: string;

  /** Metadata attached when the session was created. */
  metadata?: Record<string, string>;
}

/** Describes a customer to create. */
export interface CreateCustomerRequest {
  /** Customer's email address. */
  email: string;

  /** Customer's display name. */
  name?: string;

  /** Arbitrary key-value pairs attached to this customer. */
  metadata?: Record<string, string>;
}

/** Represents a customer in the payment provider. */
export interface Customer {
  /** Provider-assigned customer ID. */
  id: string;
  email: string;
  name?: string;
}

/** Represents an active recurring subscription. */
export interface Subscription {
  /** Provider-assigned subscription ID. */
  id: string;
  customerId: string;

  /** Status: "active" | "canceled" | "past_due" | "trialing" | etc. */
  status: string;

  /** Price this subscription is for. */
  priceId: string;

  /** Unix timestamp when the current billing period ends. */
  currentPeriodEnd: number;
}

/** A parsed and verified incoming webhook event. */
export interface WebhookEvent {
  /** Provider-assigned event ID. Use for idempotency. */
  id: string;

  /** Event type (e.g. "checkout.session.completed"). */
  type: string;

  /** Raw event data. Cast to the appropriate type based on `type`. */
  data: unknown;
}
