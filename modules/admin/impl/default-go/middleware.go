package admin

import (
	"net/http"
	"strings"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// AdminRequired is HTTP middleware that ensures the request is from an admin user.
// Must be used AFTER auth middleware that sets the Authorization header.
func AdminRequired(auth contracts.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if token == "" {
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			user, err := auth.ValidateToken(r.Context(), token)
			if err != nil {
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			if user.Role != "admin" {
				http.Error(w, `{"error":"Forbidden: admin access required"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
