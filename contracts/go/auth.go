// Package contracts defines the module interfaces for the modkit registry.
// All application code must depend on these interfaces, never on concrete implementations.
package contracts

import "context"

// AuthService handles user authentication and session management.
// It validates tokens issued by a third-party auth provider (e.g. Clerk).
// Never implement password storage — always delegate to the provider.
type AuthService interface {
	// ValidateToken validates a JWT or session token and returns the authenticated user.
	// Returns ErrUnauthorized if the token is invalid or expired.
	ValidateToken(ctx context.Context, token string) (*AuthUser, error)

	// GetUser retrieves a user by their provider-assigned ID.
	// Returns ErrNotFound if the user does not exist.
	GetUser(ctx context.Context, userID string) (*AuthUser, error)

	// ListUsers returns a paginated list of users.
	ListUsers(ctx context.Context, opts ListUsersOptions) (*UserList, error)

	// DeleteUser removes a user from the auth provider.
	// This is typically called when a user requests account deletion.
	DeleteUser(ctx context.Context, userID string) error
}

// AuthUser represents an authenticated user returned from the auth provider.
type AuthUser struct {
	// ID is the provider-assigned unique identifier (e.g. "user_2abc123" for Clerk).
	ID string

	// Email is the user's primary email address.
	Email string

	// Name is the user's display name.
	Name string

	// AvatarURL is the URL of the user's profile picture. May be empty.
	AvatarURL string

	// Metadata holds arbitrary key-value pairs set on the user in the auth provider.
	Metadata map[string]string
}

// ListUsersOptions controls pagination for ListUsers.
type ListUsersOptions struct {
	// Limit is the maximum number of users to return. Defaults to 20, max 100.
	Limit int

	// Offset is the number of users to skip (for page-based pagination).
	Offset int
}

// UserList is the result of a ListUsers call.
type UserList struct {
	// Users is the slice of users for this page.
	Users []*AuthUser

	// Total is the total number of users across all pages.
	Total int
}
