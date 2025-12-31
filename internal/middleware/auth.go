// Package middleware provides HTTP middleware functions.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/yourusername/siteproof/internal/domain"
	"github.com/yourusername/siteproof/internal/handler"
	"github.com/yourusername/siteproof/internal/service"
)

// contextKey is a custom type for context keys.
type contextKey string

const (
	userIDKey    contextKey = "user_id"
	companyIDKey contextKey = "company_id"
	roleKey      contextKey = "role"
)

// Auth creates an authentication middleware.
func Auth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				handler.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
				return
			}

			// Parse Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				handler.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header format")
				return
			}

			// Validate token
			claims, err := authService.ValidateAccessToken(parts[1])
			if err != nil {
				if err == domain.ErrExpiredToken {
					handler.WriteError(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "access token has expired")
					return
				}
				handler.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid access token")
				return
			}

			// Add claims to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, companyIDKey, claims.CompanyID)
			ctx = context.WithValue(ctx, roleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole creates a middleware that requires a specific role.
func RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	roleSet := make(map[domain.Role]bool)
	for _, r := range roles {
		roleSet[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := RoleFromContext(r.Context())
			if !ok {
				handler.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
				return
			}

			if !roleSet[role] {
				handler.WriteError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// UserIDFromContext extracts the user ID from context.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

// CompanyIDFromContext extracts the company ID from context.
func CompanyIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(companyIDKey).(uuid.UUID)
	return id, ok
}

// RoleFromContext extracts the role from context.
func RoleFromContext(ctx context.Context) (domain.Role, bool) {
	role, ok := ctx.Value(roleKey).(domain.Role)
	return role, ok
}

