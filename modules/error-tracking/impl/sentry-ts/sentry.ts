import type { IErrorTrackingService, CaptureOptions, ErrorUser, ErrorLevel } from '../../../contracts/ts/errors'

export interface SentryConfig {
  dsn: string
  environment?: string
  tracesSampleRate?: number
}

/**
 * SentryErrorTrackingService implements IErrorTrackingService using Sentry.
 */
export class SentryErrorTrackingService implements IErrorTrackingService {
  constructor(private readonly config: SentryConfig) {}

  async captureError(err: Error, opts?: CaptureOptions): Promise<void> {
    // TODO: implement using @sentry/node Sentry.captureException(err)
    console.warn('[sentry] stub: captureError() not implemented')
  }

  async captureMessage(msg: string, level: ErrorLevel, opts?: CaptureOptions): Promise<void> {
    // TODO: implement using @sentry/node Sentry.captureMessage(msg, level)
    console.warn('[sentry] stub: captureMessage() not implemented')
  }

  setUser(user: ErrorUser): void {
    // TODO: implement using Sentry.setUser({ id, email })
    console.warn('[sentry] stub: setUser() not implemented')
  }

  clearUser(): void {
    // TODO: implement using Sentry.setUser(null)
    console.warn('[sentry] stub: clearUser() not implemented')
  }

  async flush(timeoutMs = 2000): Promise<void> {
    // TODO: implement using Sentry.flush(timeoutMs)
    console.warn('[sentry] stub: flush() not implemented')
  }
}
