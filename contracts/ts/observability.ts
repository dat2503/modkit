/**
 * IObservabilityService provides distributed tracing, structured logging, and metrics.
 * Must be initialized first, before all other modules.
 * Wrap all module calls and HTTP handlers with spans for end-to-end tracing.
 */
export interface IObservabilityService {
  /**
   * Starts a new trace span with the given operation name.
   * Always call span.end() when the operation is complete (use try/finally).
   */
  startSpan(operationName: string, parentCtx?: SpanContext): Span;

  /**
   * Emits a structured log entry at the given level.
   */
  log(level: LogLevel, msg: string, fields?: Record<string, unknown>): void;

  /**
   * Records a numeric metric value with optional labels.
   */
  recordMetric(name: string, value: number, labels?: Record<string, string>): void;

  /**
   * Flushes all pending telemetry and releases resources.
   * Call this during graceful shutdown.
   */
  shutdown(): Promise<void>;
}

/** Represents a single unit of work in a distributed trace. */
export interface Span {
  /** Marks the span as complete. Always call this — prefer try/finally. */
  end(): void;

  /** Attaches a key-value attribute to this span. */
  setAttribute(key: string, value: unknown): void;

  /** Records an error on this span and marks it as failed. */
  recordError(err: Error): void;

  /** Returns a context object that can be passed to startSpan to create a child span. */
  context(): SpanContext;
}

/** Opaque context that carries span propagation information. */
export type SpanContext = Record<string, unknown>;

/** Severity of a log entry. */
export type LogLevel = 'debug' | 'info' | 'warn' | 'error';
