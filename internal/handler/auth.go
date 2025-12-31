package handler

import (
	"net/http"

	"github.com/yourusername/siteproof/internal/domain"
	"github.com/yourusername/siteproof/internal/middleware"
	"github.com/yourusername/siteproof/internal/service"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a login response.
type LoginResponse struct {
	User   *domain.User       `json:"user"`
	Tokens *service.TokenPair `json:"tokens"`
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	input := &domain.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	tokens, user, err := h.authService.Login(r.Context(), input)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			WriteError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
		default:
			if _, ok := err.(domain.ValidationErrors); ok {
				WriteErrorWithDetails(w, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", err)
			} else {
				WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "login failed")
			}
		}
		return
	}

	WriteJSON(w, http.StatusOK, LoginResponse{
		User:   user,
		Tokens: tokens,
	})
}

// RefreshRequest represents a token refresh request.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "refresh_token is required")
		return
	}

	tokens, err := h.authService.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case domain.ErrInvalidToken:
			WriteError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid refresh token")
		case domain.ErrExpiredToken:
			WriteError(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "refresh token has expired")
		default:
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "token refresh failed")
		}
		return
	}

	WriteJSON(w, http.StatusOK, tokens)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	if err := h.authService.Logout(r.Context(), userID); err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "logout failed")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

// Me handles GET /api/v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	companyID, _ := middleware.CompanyIDFromContext(r.Context())
	role, _ := middleware.RoleFromContext(r.Context())

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":    userID,
		"company_id": companyID,
		"role":       role,
	})
}

