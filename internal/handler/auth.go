package handler

import (
	"errors"
	"net/http"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/internal/middleware"
	"github.com/ajejoseph22/siteproof/internal/service"
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

// RegisterRequest represents a registration request.
type RegisterRequest struct {
	CompanyName string  `json:"company_name"`
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	Name        string  `json:"name"`
	Phone       *string `json:"phone,omitempty"`
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	input := &domain.RegisterInput{
		CompanyName: req.CompanyName,
		Email:       req.Email,
		Password:    req.Password,
		Name:        req.Name,
		Phone:       req.Phone,
	}

	tokens, user, err := h.authService.Register(r.Context(), input)
	if err != nil {
		WriteErrorFrom(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, LoginResponse{
		User:   user,
		Tokens: tokens,
	})
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
		if errors.Is(err, domain.ErrUnauthorized) {
			WriteError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
		} else {
			WriteErrorFrom(w, err)
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
		WriteErrorFrom(w, err)
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
		WriteErrorFrom(w, err)
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
