// Package admin implements AdminService by wrapping the AuthService.
package admin

import (
	"context"
	"fmt"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Service implements contracts.AdminService.
type Service struct {
	auth contracts.AuthService
}

// New creates a new admin service. auth is required.
func New(auth contracts.AuthService) *Service {
	return &Service{auth: auth}
}

// ListUsers returns a paginated list of all users.
func (s *Service) ListUsers(ctx context.Context, opts contracts.ListUsersOptions) (*contracts.UserList, error) {
	return s.auth.ListUsers(ctx, opts)
}

// SetUserRole updates a user's role.
func (s *Service) SetUserRole(ctx context.Context, userID string, role string) error {
	return s.auth.UpdateUserRole(ctx, userID, role)
}

// DeleteUser removes a user.
func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	return s.auth.DeleteUser(ctx, userID)
}

// GetDashboardStats returns basic dashboard statistics.
func (s *Service) GetDashboardStats(ctx context.Context) (*contracts.DashboardStats, error) {
	list, err := s.auth.ListUsers(ctx, contracts.ListUsersOptions{Limit: 1})
	if err != nil {
		return nil, fmt.Errorf("admin: get stats: %w", err)
	}
	return &contracts.DashboardStats{
		TotalUsers:    list.Total,
		RecentSignups: min(list.Total, 100),
	}, nil
}

// IsAdmin checks if a user has the admin role.
func (s *Service) IsAdmin(user *contracts.AuthUser) bool {
	return user.Role == "admin"
}
