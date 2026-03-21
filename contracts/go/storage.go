package contracts

import (
	"context"
	"io"
	"time"
)

// StorageService manages object storage for files, images, and generated assets.
// All implementations must be compatible with S3-style object storage semantics.
type StorageService interface {
	// Upload stores an object at the given key. The key is the full object path (e.g. "avatars/user-123.jpg").
	// Returns the storage URL of the uploaded object.
	Upload(ctx context.Context, key string, r io.Reader, opts UploadOptions) (*UploadResult, error)

	// Download retrieves an object by key. The caller is responsible for closing the returned reader.
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes an object by key.
	Delete(ctx context.Context, key string) error

	// SignedURL generates a pre-signed URL for temporary access to a private object.
	// The URL expires after the given duration.
	SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)

	// PublicURL returns the public URL for a publicly accessible object.
	// Returns an error if the object is not publicly accessible.
	PublicURL(ctx context.Context, key string) (string, error)

	// Exists checks whether an object with the given key exists.
	Exists(ctx context.Context, key string) (bool, error)
}

// UploadOptions controls how an object is stored.
type UploadOptions struct {
	// ContentType is the MIME type of the object (e.g. "image/jpeg").
	// If empty, the provider may attempt to detect it.
	ContentType string

	// Public marks the object as publicly readable.
	// Use for assets that should be served without authentication.
	Public bool

	// Metadata holds arbitrary key-value pairs stored with the object.
	Metadata map[string]string
}

// UploadResult is returned after a successful upload.
type UploadResult struct {
	// Key is the object key as stored.
	Key string

	// URL is the direct URL to the object (may require signed access if private).
	URL string

	// Size is the number of bytes written.
	Size int64
}
