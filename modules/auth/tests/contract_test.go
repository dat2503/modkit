// Package tests contains contract compliance tests for all auth implementations.
// Each implementation must pass these tests to be considered registry-ready.
package tests

import (
	"context"
	"testing"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// AuthServiceContract runs all contract compliance tests against any AuthService implementation.
// Use this in each implementation's test file:
//
//	func TestClerkContract(t *testing.T) {
//	    svc := clerk.New(testConfig, testCache)
//	    AuthServiceContract(t, svc)
//	}
func AuthServiceContract(t *testing.T, svc contracts.AuthService) {
	t.Helper()

	t.Run("ValidateToken_InvalidToken_ReturnsError", func(t *testing.T) {
		_, err := svc.ValidateToken(context.Background(), "invalid-token")
		if err == nil {
			t.Fatal("expected error for invalid token, got nil")
		}
	})

	t.Run("GetUser_NonexistentUser_ReturnsError", func(t *testing.T) {
		_, err := svc.GetUser(context.Background(), "user_nonexistent_000")
		if err == nil {
			t.Fatal("expected error for nonexistent user, got nil")
		}
	})

	t.Run("ListUsers_ReturnsValidList", func(t *testing.T) {
		list, err := svc.ListUsers(context.Background(), contracts.ListUsersOptions{Limit: 10})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if list == nil {
			t.Fatal("expected non-nil UserList")
		}
		if list.Total < 0 {
			t.Fatalf("expected non-negative total, got %d", list.Total)
		}
	})
}
