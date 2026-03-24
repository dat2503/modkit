// Package clerk implements AuthService using github.com/clerk/clerk-sdk-go/v2.
package clerk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkjwt "github.com/clerk/clerk-sdk-go/v2/jwt"
	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Config holds the configuration for the Clerk auth provider.
type Config struct {
	// SecretKey is the Clerk secret key (sk_live_... or sk_test_...).
	SecretKey string

	// PublishableKey is the Clerk publishable key (pk_live_... or pk_test_...).
	PublishableKey string

	// WebhookSecret is used to verify Clerk webhook payloads (optional).
	WebhookSecret string
}

// Service implements contracts.AuthService using Clerk.
type Service struct {
	cfg   Config
	cache contracts.CacheService
}

// New creates a new Clerk auth service.
// cache is required for session storage.
func New(cfg Config, cache contracts.CacheService) (*Service, error) {
	if cfg.SecretKey == "" {
		return nil, fmt.Errorf("clerk: SecretKey is required")
	}
	clerk.SetKey(cfg.SecretKey)
	return &Service{cfg: cfg, cache: cache}, nil
}

// ValidateToken verifies a Clerk session JWT and returns the authenticated user.
func (s *Service) ValidateToken(ctx context.Context, token string) (*contracts.AuthUser, error) {
	claims, err := clerkjwt.Verify(ctx, &clerkjwt.VerifyParams{
		Token: token,
	})
	if err != nil {
		return nil, fmt.Errorf("clerk: invalid token: %w", err)
	}

	return &contracts.AuthUser{
		ID: claims.Subject,
	}, nil
}

// GetUser retrieves a Clerk user by ID.
func (s *Service) GetUser(ctx context.Context, userID string) (*contracts.AuthUser, error) {
	u, err := clerkuser.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("clerk: get user: %w", err)
	}
	return clerkUserToContract(u), nil
}

// ListUsers returns a paginated list of users from Clerk.
func (s *Service) ListUsers(ctx context.Context, opts contracts.ListUsersOptions) (*contracts.UserList, error) {
	limit := int64(opts.Limit)
	offset := int64(opts.Offset)
	if limit == 0 {
		limit = 10
	}

	list, err := clerkuser.List(ctx, &clerkuser.ListParams{
		ListParams: clerk.ListParams{
			Limit:  &limit,
			Offset: &offset,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("clerk: list users: %w", err)
	}

	users := make([]*contracts.AuthUser, len(list.Users))
	for i, u := range list.Users {
		users[i] = clerkUserToContract(u)
	}
	return &contracts.UserList{Users: users, Total: int(list.TotalCount)}, nil
}

// DeleteUser removes a user from Clerk.
func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	_, err := clerkuser.Delete(ctx, userID)
	if err != nil {
		return fmt.Errorf("clerk: delete user: %w", err)
	}
	return nil
}

// UpdateUserRole updates a user's role in Clerk via public metadata.
func (s *Service) UpdateUserRole(ctx context.Context, userID string, role string) error {
	_, err := clerkuser.UpdateMetadata(ctx, userID, &clerkuser.UpdateMetadataParams{
		PublicMetadata: &json.RawMessage{},
	})
	// NOTE: Clerk stores role in publicMetadata.role — the actual
	// UpdateMetadata call should set {"role": role} in PublicMetadata.
	// This is a stub that needs the proper Clerk SDK metadata update.
	if err != nil {
		return fmt.Errorf("clerk: update user role: %w", err)
	}
	return nil
}

func clerkUserToContract(u *clerk.User) *contracts.AuthUser {
	au := &contracts.AuthUser{ID: u.ID, Role: "user"}
	if u.ImageURL != nil {
		au.AvatarURL = *u.ImageURL
	}
	if u.PublicMetadata != nil {
		var meta map[string]interface{}
		if err := json.Unmarshal(*u.PublicMetadata, &meta); err == nil {
			if r, ok := meta["role"].(string); ok {
				au.Role = r
			}
		}
	}
	if u.FirstName != nil && u.LastName != nil {
		au.Name = *u.FirstName + " " + *u.LastName
	} else if u.FirstName != nil {
		au.Name = *u.FirstName
	}
	for _, e := range u.EmailAddresses {
		if u.PrimaryEmailAddressID != nil && e.ID == *u.PrimaryEmailAddressID {
			au.Email = e.EmailAddress
			break
		}
	}
	return au
}
