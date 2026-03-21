# Storage Module — Agent Instructions

## When to use

Include this module when:
- App allows file uploads (documents, images, attachments)
- App generates downloadable files (PDF invoices, CSV exports, reports)
- App stores user-generated media (avatars, photos, videos)

Do NOT use for:
- Database blobs — store files in object storage, store the URL in the database
- Temporary in-memory buffers — use standard library for that

## How to wire

### Go

1. Import `StorageService` from `contracts/go/storage.go`
2. Initialize in bootstrap:
   ```go
   storageSvc := s3.New(s3.Config{
       Bucket:          cfg.Storage.Bucket,
       Region:          cfg.Storage.Region,
       AccessKeyID:     cfg.Storage.AccessKeyID,
       SecretAccessKey: cfg.Storage.SecretAccessKey,
       Endpoint:        cfg.Storage.Endpoint, // empty for AWS, set for R2/MinIO
   })
   ```
3. Inject into handlers that upload/download files

### Bun (TypeScript)

1. Import `IStorageService` from `contracts/ts/storage.ts`
2. Initialize in bootstrap:
   ```typescript
   const storage = new S3StorageService({
     bucket: config.storage.bucket,
     region: config.storage.region,
     accessKeyId: config.storage.accessKeyId,
     secretAccessKey: config.storage.secretAccessKey,
     endpoint: config.storage.endpoint,
   })
   ```

## Key naming conventions

Use structured key paths. Never use flat keys:
```
avatars/{userId}.{ext}
invoices/{invoiceId}/invoice.pdf
exports/{userId}/{timestamp}-export.csv
uploads/{userId}/{uuid}.{ext}
```

## Upload pattern

```go
// In your handler:
file, header, _ := r.FormFile("file")
defer file.Close()

ext := filepath.Ext(header.Filename)
key := fmt.Sprintf("uploads/%s/%s%s", userID, uuid.New().String(), ext)

result, err := storage.Upload(ctx, key, file, contracts.UploadOptions{
    ContentType: header.Header.Get("Content-Type"),
    Public:      false,
})
// Store result.Key in the database
```

## Signed URL pattern (private files)

```go
// Generate a 1-hour download link:
url, err := storage.SignedURL(ctx, object.Key, 1*time.Hour)
// Return url to the client — never expose the raw S3 key
```

## Cloudflare R2 (recommended for cost)

R2 is S3-compatible with zero egress costs. To use R2 instead of AWS S3:
```
STORAGE_ENDPOINT=https://<account-id>.r2.cloudflarestorage.com
STORAGE_REGION=auto
```

## Required env vars

```
STORAGE_PROVIDER=s3
STORAGE_BUCKET=my-app-files
STORAGE_REGION=us-east-1
STORAGE_ACCESS_KEY_ID=AKIA...        # sensitive
STORAGE_SECRET_ACCESS_KEY=...        # sensitive
STORAGE_ENDPOINT=                    # empty for AWS, set for R2/MinIO
STORAGE_PUBLIC_BASE_URL=https://assets.yourapp.com  # optional
```

## Do NOT

- Expose raw S3 presigned URLs with long expiry (use ≤1 hour)
- Store `STORAGE_SECRET_ACCESS_KEY` anywhere other than env vars
- Use the bucket as a CDN origin — use CloudFront or Cloudflare in front
- Store file content in the database as base64 blobs
