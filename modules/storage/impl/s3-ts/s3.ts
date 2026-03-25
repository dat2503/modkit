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
    console.warn('[s3-storage] stub: upload() not implemented')
    return { key, url: '', size: 0 }
  }

  async download(key: string): Promise<Buffer> {
    // TODO: implement using @aws-sdk/client-s3 GetObjectCommand
    console.warn('[s3-storage] stub: download() not implemented')
    return Buffer.alloc(0)
  }

  async delete(key: string): Promise<void> {
    // TODO: implement using @aws-sdk/client-s3 DeleteObjectCommand
    console.warn('[s3-storage] stub: delete() not implemented')
  }

  async signedUrl(key: string, expirySeconds: number): Promise<string> {
    // TODO: implement using @aws-sdk/s3-request-presigner getSignedUrl
    console.warn('[s3-storage] stub: signedUrl() not implemented')
    return ''
  }

  async publicUrl(key: string): Promise<string> {
    // TODO: return this.config.publicBaseUrl + '/' + key if configured
    console.warn('[s3-storage] stub: publicUrl() not implemented')
    return ''
  }

  async exists(key: string): Promise<boolean> {
    // TODO: implement using @aws-sdk/client-s3 HeadObjectCommand
    console.warn('[s3-storage] stub: exists() not implemented')
    return false
  }
}
