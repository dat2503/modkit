import type { IStorageService, UploadOptions, UploadResult } from '../../../contracts/ts/storage'

export interface S3Config {
  bucket: string
  region: string
  accessKeyId: string
  secretAccessKey: string
  /** Custom endpoint for S3-compatible providers (Cloudflare R2, MinIO). Empty for AWS S3. */
  endpoint?: string
  publicBaseUrl?: string
}

/**
 * S3StorageService implements IStorageService using AWS S3 (or any S3-compatible provider).
 */
export class S3StorageService implements IStorageService {
  constructor(private readonly config: S3Config) {}

  async upload(key: string, data: Buffer | ReadableStream, opts?: UploadOptions): Promise<UploadResult> {
    // TODO: implement using @aws-sdk/client-s3 PutObjectCommand
    throw new Error('not implemented')
  }

  async download(key: string): Promise<Buffer> {
    // TODO: implement using @aws-sdk/client-s3 GetObjectCommand
    throw new Error('not implemented')
  }

  async delete(key: string): Promise<void> {
    // TODO: implement using @aws-sdk/client-s3 DeleteObjectCommand
    throw new Error('not implemented')
  }

  async signedUrl(key: string, expirySeconds: number): Promise<string> {
    // TODO: implement using @aws-sdk/s3-request-presigner getSignedUrl
    throw new Error('not implemented')
  }

  async publicUrl(key: string): Promise<string> {
    // TODO: return this.config.publicBaseUrl + '/' + key if configured
    throw new Error('not implemented')
  }

  async exists(key: string): Promise<boolean> {
    // TODO: implement using @aws-sdk/client-s3 HeadObjectCommand
    throw new Error('not implemented')
  }
}
