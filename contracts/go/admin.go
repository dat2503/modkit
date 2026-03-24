package contracts

import "context"

// AdminService provides admin-only operations: user management and dashboard stats.
// Requires the auth module — wraps AuthService for role-based operations.
type AdminService interface {
	// ListUsers lists all users with pagination.
	ListUsers(ctx context.Context, opts ListUsersOptions) (*UserList, error)

	// SetUserRole updates a user's role (e.g. "admin", "user").
	SetUserRole(ctx context.Context, userID string, role string) error

	// DeleteUser deletes a user by ID.
	DeleteUser(ctx context.Context, userID string) error

	// GetDashboardStats returns dashboard statistics.
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)

	// IsAdmin checks if a user has the admin role.
	IsAdmin(user *AuthUser) bool
}

// DashboardStats contains admin dashboard statistics.
type DashboardStats struct {
	TotalUsers    int
	RecentSignups int
}
