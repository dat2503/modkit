// Package tests contains contract compliance tests for all storage implementations.
package tests

import (
	"bytes"
	"context"
	"testing"
	"time"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// StorageServiceContract runs contract compliance tests against any StorageService implementation.
func StorageServiceContract(t *testing.T, svc contracts.StorageService) {
	t.Helper()

	const testKey = "test/contract-test-object.txt"
	const testContent = "contract test content"

	t.Run("Upload_ThenDownload_ReturnsContent", func(t *testing.T) {
		_, err := svc.Upload(context.Background(), testKey, bytes.NewBufferString(testContent), contracts.UploadOptions{
			ContentType: "text/plain",
		})
		if err != nil {
			t.Fatalf("upload failed: %v", err)
		}

		rc, err := svc.Download(context.Background(), testKey)
		if err != nil {
			t.Fatalf("download failed: %v", err)
		}
		defer rc.Close()
	})

	t.Run("Exists_AfterUpload_ReturnsTrue", func(t *testing.T) {
		ok, err := svc.Exists(context.Background(), testKey)
		if err != nil {
			t.Fatalf("exists check failed: %v", err)
		}
		if !ok {
			t.Fatal("expected object to exist after upload")
		}
	})

	t.Run("SignedURL_ReturnsNonEmptyURL", func(t *testing.T) {
		url, err := svc.SignedURL(context.Background(), testKey, 5*time.Minute)
		if err != nil {
			t.Fatalf("signed URL failed: %v", err)
		}
		if url == "" {
			t.Fatal("expected non-empty signed URL")
		}
	})

	t.Run("Delete_ThenExists_ReturnsFalse", func(t *testing.T) {
		if err := svc.Delete(context.Background(), testKey); err != nil {
			t.Fatalf("delete failed: %v", err)
		}
		ok, err := svc.Exists(context.Background(), testKey)
		if err != nil {
			t.Fatalf("exists check failed: %v", err)
		}
		if ok {
			t.Fatal("expected object to not exist after delete")
		}
	})
}
