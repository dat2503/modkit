/**
 * IEmailService sends transactional emails triggered by user actions.
 * Do NOT use this for marketing or bulk email — use a dedicated platform instead.
 */
export interface IEmailService {
  /**
   * Sends a single transactional email.
   * Returns the provider-assigned message ID on success.
   */
  send(msg: EmailMessage): Promise<EmailResult>;

  /**
   * Sends multiple emails in a single API call.
   * Providers may impose limits on batch size — check provider docs.
   */
  sendBatch(msgs: EmailMessage[]): Promise<EmailResult[]>;
}

/** A single email to send. */
export interface EmailMessage {
  /** List of recipient email addresses. */
  to: string[];

  /** Sender address. Must be a verified domain/address in your provider. */
  from: string;

  /** Optional reply-to address. */
  replyTo?: string;

  /** Email subject line. */
  subject: string;

  /** Email content. Provide both html and text for best deliverability. */
  body: EmailBody;

  /** Additional email headers (e.g. List-Unsubscribe). */
  headers?: Record<string, string>;

  /** Provider-specific tags for analytics/filtering. */
  tags?: Record<string, string>;
}

/** Holds the content of an email in multiple formats. */
export interface EmailBody {
  /** HTML version of the email body. */
  html?: string;

  /** Plain-text version. Always provide this as a fallback. */
  text: string;
}

/** Result of a successful send call. */
export interface EmailResult {
  /** Provider-assigned message ID. Store this for delivery tracking. */
  messageId: string;
}
