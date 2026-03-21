/**
 * IStorageService manages object storage for files, images, and generated assets.
 * All implementations must be compatible with S3-style object storage semantics.
 */
export interface IStorageService {
  /**
   * Stores an object at the given key.
   * The key is the full object path (e.g. "avatars/user-123.jpg").
   */
  upload(key: string, data: Buffer | ReadableStream, opts?: UploadOptions): Promise<UploadResult>;

  /**
   * Retrieves an object by key as a Buffer.
   * Throws StorageError with code NOT_FOUND if the key does not exist.
   */
  download(key: string): Promise<Buffer>;

  /**
   * Removes an object by key. Returns without error if the key did not exist.
   */
  delete(key: string): Promise<void>;

  /**
   * Generates a pre-signed URL for temporary access to a private object.
   * The URL expires after the given number of seconds.
   */
  signedUrl(key: string, expirySeconds: number): Promise<string>;

  /**
   * Returns the public URL for a publicly accessible object.
   * Throws StorageError if the object is not publicly accessible.
   */
  publicUrl(key: string): Promise<string>;

  /**
   * Checks whether an object with the given key exists.
   */
  exists(key: string): Promise<boolean>;
}

/** Controls how an object is stored. */
export interface UploadOptions {
  /** MIME type of the object (e.g. "image/jpeg"). */
  contentType?: string;

  /** Marks the object as publicly readable. */
  public?: boolean;

  /** Arbitrary key-value pairs stored with the object. */
  metadata?: Record<string, string>;
}

/** Returned after a successful upload. */
export interface UploadResult {
  /** The object key as stored. */
  key: string;

  /** Direct URL to the object (may require signed access if private). */
  url: string;

  /** Number of bytes written. */
  size: number;
}
