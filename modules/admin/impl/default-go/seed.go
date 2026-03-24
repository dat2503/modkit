// Package admin contains the seed command for creating the default admin account.
//
// Build and run:
//
//	go run cmd/seed/main.go
package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Seed creates the default admin account using Better Auth's REST API.
func Seed() error {
	email := envOr("ADMIN_DEFAULT_EMAIL", "admin@localhost")
	password := envOr("ADMIN_DEFAULT_PASSWORD", "changeme")
	authURL := envOr("BETTER_AUTH_URL", "http://localhost:8080")
	secret := os.Getenv("BETTER_AUTH_SECRET")

	fmt.Printf("Creating admin account: %s\n", email)

	// 1. Sign up the admin user
	signupBody, _ := json.Marshal(map[string]string{
		"name": "Admin", "email": email, "password": password,
	})
	signupResp, err := http.Post(authURL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(signupBody))
	if err != nil {
		return fmt.Errorf("signup request failed: %w", err)
	}
	defer signupResp.Body.Close()

	if signupResp.StatusCode != http.StatusOK && signupResp.StatusCode != http.StatusConflict {
		return fmt.Errorf("signup returned %d", signupResp.StatusCode)
	}
	if signupResp.StatusCode == http.StatusConflict {
		fmt.Println("Admin account already exists, updating role...")
	}

	// 2. Get the user ID
	listURL := fmt.Sprintf("%s/api/auth/admin/list-users?searchField=email&searchValue=%s&limit=1", authURL, email)
	listReq, _ := http.NewRequest(http.MethodGet, listURL, nil)
	listReq.Header.Set("x-better-auth-secret", secret)
	listResp, err := http.DefaultClient.Do(listReq)
	if err != nil {
		return fmt.Errorf("list users request failed: %w", err)
	}
	defer listResp.Body.Close()

	var listData struct {
		Users []struct {
			ID string `json:"id"`
		} `json:"users"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listData); err != nil {
		return fmt.Errorf("decode list users: %w", err)
	}
	if len(listData.Users) == 0 {
		return fmt.Errorf("could not find admin user after signup")
	}
	userID := listData.Users[0].ID

	// 3. Set role to admin
	roleBody, _ := json.Marshal(map[string]string{"userId": userID, "role": "admin"})
	roleReq, _ := http.NewRequest(http.MethodPost, authURL+"/api/auth/admin/set-role", bytes.NewReader(roleBody))
	roleReq.Header.Set("Content-Type", "application/json")
	roleReq.Header.Set("x-better-auth-secret", secret)
	roleResp, err := http.DefaultClient.Do(roleReq)
	if err != nil {
		return fmt.Errorf("set role request failed: %w", err)
	}
	defer roleResp.Body.Close()

	if roleResp.StatusCode != http.StatusOK {
		return fmt.Errorf("set role returned %d", roleResp.StatusCode)
	}

	fmt.Printf("Admin account ready: %s (role: admin)\n", email)
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
