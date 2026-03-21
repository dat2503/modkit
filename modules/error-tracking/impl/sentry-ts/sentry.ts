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
    throw new Error('not implemented')
  }

  async captureMessage(msg: string, level: ErrorLevel, opts?: CaptureOptions): Promise<void> {
    // TODO: implement using @sentry/node Sentry.captureMessage(msg, level)
    throw new Error('not implemented')
  }

  setUser(user: ErrorUser): void {
    // TODO: implement using Sentry.setUser({ id, email })
    throw new Error('not implemented')
  }

  clearUser(): void {
    // TODO: implement using Sentry.setUser(null)
    throw new Error('not implemented')
  }

  async flush(timeoutMs = 2000): Promise<void> {
    // TODO: implement using Sentry.flush(timeoutMs)
    throw new Error('not implemented')
  }
}
