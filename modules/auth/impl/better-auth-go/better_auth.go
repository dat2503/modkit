// Package betterauth implements AuthService using Better Auth's REST API.
// Better Auth (https://www.better-auth.com) runs as part of your Bun/TS backend.
// The Go implementation calls Better Auth's admin REST endpoints over HTTP.
package betterauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Config holds the configuration for the Better Auth provider.
type Config struct {
	// BaseURL is the URL where the Better Auth server is running.
	// Example: "http://localhost:3000" or "https://api.myapp.com"
	BaseURL string

	// Secret is the BETTER_AUTH_SECRET used to sign tokens.
	// Must match the secret configured in the Better Auth server.
	Secret string
}

// Service implements contracts.AuthService using Better Auth's REST API.
type Service struct {
	cfg    Config
	cache  contracts.CacheService
	client *http.Client
}

// New creates a new Better Auth service.
// cache is required for session storage.
func New(cfg Config, cache contracts.CacheService) (*Service, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("better-auth: BaseURL is required")
	}
	if cfg.Secret == "" {
		return nil, fmt.Errorf("better-auth: Secret is required")
	}
	return &Service{
		cfg:   cfg,
		cache: cache,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// ValidateToken verifies a Better Auth session token and returns the authenticated user.
func (s *Service) ValidateToken(ctx context.Context, token string) (*contracts.AuthUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.cfg.BaseURL+"/api/auth/get-session", nil)
	if err != nil {
		return nil, fmt.Errorf("better-auth: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("better-auth: get session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("better-auth: invalid or expired token")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("better-auth: get session returned %d", resp.StatusCode)
	}

	var result struct {
		User struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
			Image string `json:"image"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("better-auth: decode session: %w", err)
	}
	if result.User.ID == "" {
		return nil, fmt.Errorf("better-auth: session has no user")
	}

	return &contracts.AuthUser{
		ID:        result.User.ID,
		Email:     result.User.Email,
		Name:      result.User.Name,
		AvatarURL: result.User.Image,
	}, nil
}

// GetUser retrieves a user by ID using the Better Auth admin API.
func (s *Service) GetUser(ctx context.Context, userID string) (*contracts.AuthUser, error) {
	url := fmt.Sprintf("%s/api/auth/admin/list-users?searchField=id&searchValue=%s&limit=1", s.cfg.BaseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("better-auth: build request: %w", err)
	}
	req.Header.Set("x-better-auth-secret", s.cfg.Secret)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("better-auth: get user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("better-auth: user %q not found", userID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("better-auth: get user returned %d", resp.StatusCode)
	}

	var result struct {
		Users []baUser `json:"users"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("better-auth: decode user: %w", err)
	}
	if len(result.Users) == 0 {
		return nil, fmt.Errorf("better-auth: user %q not found", userID)
	}

	return baUserToContract(result.Users[0]), nil
}

// ListUsers returns a paginated list of users using the Better Auth admin API.
func (s *Service) ListUsers(ctx context.Context, opts contracts.ListUsersOptions) (*contracts.UserList, error) {
	limit := opts.Limit
	if limit == 0 {
		limit = 20
	}
	url := fmt.Sprintf("%s/api/auth/admin/list-users?limit=%d&offset=%d", s.cfg.BaseURL, limit, opts.Offset)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("better-auth: build request: %w", err)
	}
	req.Header.Set("x-better-auth-secret", s.cfg.Secret)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("better-auth: list users: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("better-auth: list users returned %d", resp.StatusCode)
	}

	var result struct {
		Users []baUser `json:"users"`
		Total int      `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("better-auth: decode users: %w", err)
	}

	users := make([]*contracts.AuthUser, len(result.Users))
	for i, u := range result.Users {
		users[i] = baUserToContract(u)
	}
	return &contracts.UserList{Users: users, Total: result.Total}, nil
}

// DeleteUser removes a user using the Better Auth admin API.
func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	body, _ := json.Marshal(map[string]string{"userId": userID})
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, s.cfg.BaseURL+"/api/auth/admin/remove-user", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("better-auth: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-better-auth-secret", s.cfg.Secret)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("better-auth: delete user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("better-auth: delete user returned %d", resp.StatusCode)
	}
	return nil
}

type baUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

func baUserToContract(u baUser) *contracts.AuthUser {
	return &contracts.AuthUser{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		AvatarURL: u.Image,
	}
}
