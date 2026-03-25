import type { IJobsService, JobHandle, JobHandler, JobOptions } from '../../../contracts/ts/jobs'

export interface BullMQConfig {
  /** Redis URL — reuse the cache module's REDIS_URL. */
  redisUrl: string
  concurrency?: number
  maxRetries?: number
  defaultQueue?: string
}

/**
 * BullMQJobsService implements IJobsService using BullMQ (Bun only).
 * Uses Redis as the queue backend — reuse the cache module's Redis instance.
 */
export class BullMQJobsService implements IJobsService {
  private readonly handlers = new Map<string, JobHandler>()

  constructor(private readonly config: BullMQConfig) {}

  async enqueue(jobType: string, payload: unknown, opts?: JobOptions): Promise<JobHandle> {
    // TODO: implement using bullmq Queue.add(jobType, payload, opts)
    console.warn('[bullmq-jobs] stub: enqueue() not implemented')
    return { id: '', type: jobType, queue: this.config.defaultQueue ?? 'default' }
  }

  async enqueueIn(jobType: string, payload: unknown, delayMs: number, opts?: JobOptions): Promise<JobHandle> {
    // TODO: implement using bullmq Queue.add with { delay: delayMs }
    console.warn('[bullmq-jobs] stub: enqueueIn() not implemented')
    return { id: '', type: jobType, queue: this.config.defaultQueue ?? 'default' }
  }

  async enqueueAt(jobType: string, payload: unknown, processAt: Date, opts?: JobOptions): Promise<JobHandle> {
    const delayMs = processAt.getTime() - Date.now()
    return this.enqueueIn(jobType, payload, Math.max(0, delayMs), opts)
  }

  registerHandler(jobType: string, handler: JobHandler): void {
    this.handlers.set(jobType, handler)
  }

  async start(): Promise<void> {
    // TODO: implement using bullmq Worker with registered handlers
    console.warn('[bullmq-jobs] stub: start() not implemented')
  }

  async stop(): Promise<void> {
    // TODO: implement graceful Worker close
    console.warn('[bullmq-jobs] stub: stop() not implemented')
  }
}
