/**
 * IJobsService manages a background job queue for async processing and retries.
 * Use for any operation >500ms, any side effect that doesn't need to block the HTTP response,
 * or any operation requiring retry logic.
 * Requires the cache module as the queue backend.
 */
export interface IJobsService {
  /**
   * Adds a job to the queue for immediate processing.
   * payload must be JSON-serializable.
   */
  enqueue(jobType: string, payload: unknown, opts?: JobOptions): Promise<JobHandle>;

  /**
   * Adds a job to the queue to be processed after the given delay in milliseconds.
   */
  enqueueIn(jobType: string, payload: unknown, delayMs: number, opts?: JobOptions): Promise<JobHandle>;

  /**
   * Adds a job to the queue to be processed at a specific time.
   */
  enqueueAt(jobType: string, payload: unknown, processAt: Date, opts?: JobOptions): Promise<JobHandle>;

  /**
   * Registers a handler function for the given job type.
   * Handlers must be idempotent — they may be called more than once for the same job.
   */
  registerHandler(jobType: string, handler: JobHandler): void;

  /**
   * Begins processing jobs from the queue.
   * Returns a promise that resolves when the processor is stopped.
   */
  start(): Promise<void>;

  /**
   * Gracefully shuts down the job processor, waiting for in-flight jobs to complete.
   */
  stop(): Promise<void>;
}

/** Returned when a job is successfully enqueued. */
export interface JobHandle {
  /** Queue-assigned unique job ID. */
  id: string;

  /** The job type that was enqueued. */
  type: string;

  /** The queue the job was placed in. */
  queue: string;
}

/**
 * A function that processes a single job.
 * payload is the deserialized job payload.
 * Return void on success. Throw an error to trigger a retry.
 */
export type JobHandler = (payload: unknown) => Promise<void>;

/** Controls how a job behaves when enqueued. */
export interface JobOptions {
  /** Which queue to place the job in. Defaults to the default queue. */
  queue?: string;

  /** Maximum number of retry attempts on failure. Defaults to global config (typically 3). */
  maxRetries?: number;

  /**
   * Prevents duplicate jobs with the same key from being enqueued.
   * If a job with the same uniqueKey already exists, the new job is dropped.
   */
  uniqueKey?: string;
}
