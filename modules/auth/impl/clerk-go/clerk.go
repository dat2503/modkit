// Package clerk implements the AuthService interface using Clerk (clerk.com).
package clerk

import (
	"context"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Config holds the configuration for the Clerk auth provider.
type Config struct {
	// SecretKey is the Clerk secret key (sk_live_... or sk_test_...).
	SecretKey string

	// PublishableKey is the Clerk publishable key (pk_live_... or pk_test_...).
	PublishableKey string
}

// Service implements contracts.AuthService using Clerk.
type Service struct {
	cfg   Config
	cache contracts.CacheService
	// TODO: add clerk SDK client
}

// New creates a new Clerk auth service.
// cache is required for session storage.
func New(cfg Config, cache contracts.CacheService) *Service {
	return &Service{cfg: cfg, cache: cache}
}

// ValidateToken validates a Clerk session token and returns the authenticated user.
func (s *Service) ValidateToken(ctx context.Context, token string) (*contracts.AuthUser, error) {
	// TODO: implement using clerk-sdk-go
	// 1. Verify JWT signature using Clerk's JWKS endpoint
	// 2. Extract user ID from claims
	// 3. Optionally cache validated user to reduce Clerk API calls
	panic("not implemented")
}

// GetUser retrieves a user by their Clerk user ID.
func (s *Service) GetUser(ctx context.Context, userID string) (*contracts.AuthUser, error) {
	// TODO: implement using clerk-sdk-go users.Get(userID)
	panic("not implemented")
}

// ListUsers returns a paginated list of users from Clerk.
func (s *Service) ListUsers(ctx context.Context, opts contracts.ListUsersOptions) (*contracts.UserList, error) {
	// TODO: implement using clerk-sdk-go users.List(limit, offset)
	panic("not implemented")
}

// DeleteUser removes a user from Clerk.
func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	// TODO: implement using clerk-sdk-go users.Delete(userID)
	panic("not implemented")
}
