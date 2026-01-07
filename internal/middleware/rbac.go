package middleware

import (
	"net/http"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/internal/httputil"
)

// RequirePermission creates a middleware that enforces a permission.
func RequirePermission(permission domain.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := RoleFromContext(r.Context())
			if !ok {
				httputil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
				return
			}

			if !domain.RoleHasPermission(role, permission) {
				httputil.WriteError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
