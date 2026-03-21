/**
 * ICacheService provides a key-value cache for sessions, rate limiting, and hot data.
 * Required by the auth module (session storage) and jobs module (queue backend).
 */
export interface ICacheService {
  /**
   * Retrieves the raw bytes for a key.
   * Returns undefined if the key does not exist or has expired.
   */
  get(key: string): Promise<Buffer | undefined>;

  /**
   * Stores a value with an optional TTL in seconds.
   * If ttlSeconds is 0 or omitted, the key does not expire.
   */
  set(key: string, value: Buffer | string, ttlSeconds?: number): Promise<void>;

  /**
   * Removes a key. Returns without error if the key did not exist.
   */
  delete(key: string): Promise<void>;

  /**
   * Checks whether a key exists without retrieving its value.
   */
  exists(key: string): Promise<boolean>;

  /**
   * Atomically increments an integer value by delta (default: 1).
   * Creates the key with value delta if it does not exist.
   * Returns the new value after incrementing.
   */
  increment(key: string, delta?: number): Promise<number>;

  /**
   * Sets a key only if it does not already exist (atomic SET NX).
   * Returns true if the key was set, false if it already existed.
   * Use for distributed locks and deduplication.
   */
  setNX(key: string, value: Buffer | string, ttlSeconds?: number): Promise<boolean>;

  /**
   * Updates the TTL for an existing key.
   * Throws CacheError with code NOT_FOUND if the key does not exist.
   */
  expire(key: string, ttlSeconds: number): Promise<void>;

  /**
   * Deletes all keys matching the given glob pattern.
   * Use with care — can be slow on large keyspaces.
   */
  flushPattern(pattern: string): Promise<void>;
}
