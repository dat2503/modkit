/**
 * IErrorTrackingService reports errors and exceptions to a monitoring platform.
 * Must be initialized second, immediately after IObservabilityService.
 * All unhandled errors should be reported here before the process exits.
 */
export interface IErrorTrackingService {
  /**
   * Reports an error with optional contextual data.
   * Call this for unexpected errors that should alert on-call.
   */
  captureError(err: Error, opts?: CaptureOptions): Promise<void>;

  /**
   * Reports a non-error message (e.g. a warning or diagnostic event).
   */
  captureMessage(msg: string, level: ErrorLevel, opts?: CaptureOptions): Promise<void>;

  /**
   * Associates subsequent events with the given user identity.
   * Call this after authenticating a request.
   */
  setUser(user: ErrorUser): void;

  /**
   * Clears the current user association.
   * Call this after a request completes (or at logout).
   */
  clearUser(): void;

  /**
   * Waits for all pending events to be sent, up to the given timeout in ms.
   * Call this during graceful shutdown before exiting.
   */
  flush(timeoutMs?: number): Promise<void>;
}

/** Additional context for a captured event. */
export interface CaptureOptions {
  /** Key-value pairs for filtering events in the dashboard. */
  tags?: Record<string, string>;

  /** Arbitrary extra data attached to the event. */
  extra?: Record<string, unknown>;

  /** Overrides the grouping fingerprint (for custom grouping rules). */
  fingerprint?: string[];
}

/** Identifies the user associated with an error event. */
export interface ErrorUser {
  /** User's application ID. */
  id: string;

  /** User's email address. */
  email?: string;
}

/** Severity of a captured message. */
export type ErrorLevel = 'debug' | 'info' | 'warning' | 'error' | 'fatal';
