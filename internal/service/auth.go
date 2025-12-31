// Package service contains business logic services.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourusername/siteproof/internal/config"
	"github.com/yourusername/siteproof/internal/domain"
	"github.com/yourusername/siteproof/pkg/crypto"
)

// AuthService handles authentication operations.
type AuthService struct {
	pool   *pgxpool.Pool
	config *config.JWTConfig
}

// NewAuthService creates a new auth service.
func NewAuthService(pool *pgxpool.Pool, cfg *config.JWTConfig) *AuthService {
	return &AuthService{
		pool:   pool,
		config: cfg,
	}
}

// TokenPair represents an access and refresh token pair.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Claims represents JWT claims.
type Claims struct {
	UserID    uuid.UUID   `json:"user_id"`
	CompanyID uuid.UUID   `json:"company_id"`
	Role      domain.Role `json:"role"`
	jwt.RegisteredClaims
}

// Login authenticates a user and returns tokens.
func (s *AuthService) Login(ctx context.Context, input *domain.LoginInput) (*TokenPair, *domain.User, error) {
	if err := input.Validate(); err != nil {
		return nil, nil, err
	}

	// Get user by email
	var user domain.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, company_id, email, password_hash, name, phone, role, active, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`, input.Email).Scan(
		&user.ID, &user.CompanyID, &user.Email, &user.PasswordHash,
		&user.Name, &user.Phone, &user.Role, &user.Active,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, nil, domain.ErrUnauthorized
	}

	// Check password
	if !crypto.CheckPassword(input.Password, user.PasswordHash) {
		return nil, nil, domain.ErrUnauthorized
	}

	// Check if user is active
	if !user.Active {
		return nil, nil, domain.ErrUnauthorized
	}

	// Generate tokens
	tokens, err := s.generateTokenPair(&user, "")
	if err != nil {
		return nil, nil, err
	}

	return tokens, &user, nil
}

// RefreshTokens generates new tokens from a refresh token.
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	tokenHash := crypto.HashSHA256String(refreshToken)

	// Get refresh token from database
	var tokenID uuid.UUID
	var userID uuid.UUID
	var expiresAt time.Time
	var revokedAt *time.Time

	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, tokenHash).Scan(&tokenID, &userID, &expiresAt, &revokedAt)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	// Check if token is revoked or expired
	if revokedAt != nil || time.Now().After(expiresAt) {
		return nil, domain.ErrExpiredToken
	}

	// Get user
	var user domain.User
	err = s.pool.QueryRow(ctx, `
		SELECT id, company_id, email, name, phone, role, active
		FROM users
		WHERE id = $1 AND deleted_at IS NULL AND active = true
	`, userID).Scan(
		&user.ID, &user.CompanyID, &user.Email,
		&user.Name, &user.Phone, &user.Role, &user.Active,
	)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	// Revoke old refresh token
	_, err = s.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now() WHERE id = $1
	`, tokenID)
	if err != nil {
		return nil, fmt.Errorf("revoke old token: %w", err)
	}

	// Generate new tokens
	tokens, err := s.generateTokenPair(&user, "")
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// Logout revokes all refresh tokens for a user.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	return err
}

// ValidateAccessToken validates an access token and returns the claims.
func (s *AuthService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrExpiredToken
		}
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}

func (s *AuthService) generateTokenPair(user *domain.User, deviceName string) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(s.config.AccessTokenTTL)
	refreshExpiry := now.Add(s.config.RefreshTokenTTL)

	// Generate access token
	claims := &Claims{
		UserID:    user.ID,
		CompanyID: user.CompanyID,
		Role:      user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	// Generate refresh token
	refreshTokenRaw, err := crypto.GenerateToken()
	if err != nil {
		return nil, err
	}
	refreshTokenHash := crypto.HashSHA256String(refreshTokenRaw)

	// Store refresh token
	_, err = s.pool.Exec(context.Background(), `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, device_name)
		VALUES ($1, $2, $3, $4)
	`, user.ID, refreshTokenHash, refreshExpiry, deviceName)
	if err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenRaw,
		ExpiresAt:    accessExpiry,
	}, nil
}

