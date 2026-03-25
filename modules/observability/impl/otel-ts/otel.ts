import type { IObservabilityService, Span, SpanContext, LogLevel } from '../../../contracts/ts/observability'

export interface OtelConfig {
  serviceName: string
  endpoint?: string
  headers?: Record<string, string>
  logLevel?: LogLevel
}

/**
 * OtelObservabilityService implements IObservabilityService using OpenTelemetry.
 */
export class OtelObservabilityService implements IObservabilityService {
  constructor(private readonly config: OtelConfig) {}

  startSpan(operationName: string, parentCtx?: SpanContext): Span {
    // TODO: implement using @opentelemetry/api tracer.startSpan(operationName, {}, context)
    console.warn('[otel] stub: startSpan() not implemented')
    return { end() {}, setAttribute() {}, recordError() {}, context() { return {} } }
  }

  log(level: LogLevel, msg: string, fields?: Record<string, unknown>): void {
    // TODO: implement using @opentelemetry/api-logs logger
    console.warn('[otel] stub: log() not implemented')
  }

  recordMetric(name: string, value: number, labels?: Record<string, string>): void {
    // TODO: implement using @opentelemetry/api meter
    console.warn('[otel] stub: recordMetric() not implemented')
  }

  async shutdown(): Promise<void> {
    // TODO: implement graceful shutdown of SDK
    console.warn('[otel] stub: shutdown() not implemented')
  }
}
