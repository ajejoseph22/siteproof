// Package service contains business logic services.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ajejoseph22/siteproof/internal/config"
	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/pkg/crypto"
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
	tokens, err := s.generateTokenPair(ctx, &user, "")
	if err != nil {
		return nil, nil, err
	}

	return tokens, &user, nil
}

// Register creates a new company and admin user, then issues tokens.
func (s *AuthService) Register(ctx context.Context, input *domain.RegisterInput) (*TokenPair, *domain.User, error) {
	if err := input.Validate(); err != nil {
		return nil, nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("begin registration: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	`, input.Email).Scan(&exists); err != nil {
		return nil, nil, fmt.Errorf("check email: %w", err)
	}
	if exists {
		return nil, nil, domain.NewDomainError(domain.ErrConflict, "EMAIL_EXISTS", "email already exists")
	}

	baseSlug := slugify(input.CompanyName)
	if baseSlug == "" {
		var errs domain.ValidationErrors
		errs.Add("company_name", "must include letters or numbers")
		return nil, nil, errs
	}

	var company domain.Company
	created := false
	for i := 0; i < 20; i++ {
		slug := baseSlug
		if i > 0 {
			slug = fmt.Sprintf("%s-%d", baseSlug, i+1)
		}

		err = tx.QueryRow(ctx, `
			INSERT INTO companies (name, slug)
			VALUES ($1, $2)
			RETURNING id, name, slug, settings, created_at, updated_at
		`, input.CompanyName, slug).Scan(
			&company.ID, &company.Name, &company.Slug, &company.Settings, &company.CreatedAt, &company.UpdatedAt,
		)
		if err == nil {
			created = true
			break
		}
		if isUniqueViolation(err) {
			continue
		}
		return nil, nil, fmt.Errorf("create company: %w", err)
	}
	if !created {
		return nil, nil, domain.NewDomainError(domain.ErrConflict, "COMPANY_SLUG_EXISTS", "company slug already exists")
	}

	hash, err := crypto.HashPassword(input.Password)
	if err != nil {
		return nil, nil, err
	}

	var user domain.User
	err = tx.QueryRow(ctx, `
		INSERT INTO users (company_id, email, password_hash, name, phone, role, active)
		VALUES ($1, $2, $3, $4, $5, $6, true)
		RETURNING id, company_id, email, name, phone, role, active, created_at, updated_at
	`, company.ID, input.Email, hash, input.Name, input.Phone, domain.RoleAdmin).Scan(
		&user.ID, &user.CompanyID, &user.Email, &user.Name, &user.Phone,
		&user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, nil, domain.NewDomainError(domain.ErrConflict, "EMAIL_EXISTS", "email already exists")
		}
		return nil, nil, fmt.Errorf("create user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("commit registration: %w", err)
	}

	tokens, err := s.generateTokenPair(ctx, &user, "")
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
	tokens, err := s.generateTokenPair(ctx, &user, "")
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

func (s *AuthService) generateTokenPair(ctx context.Context, user *domain.User, deviceName string) (*TokenPair, error) {
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
	_, err = s.pool.Exec(ctx, `
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

func slugify(input string) string {
	lower := strings.ToLower(input)
	var b strings.Builder
	prevHyphen := false

	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevHyphen = false
			continue
		}
		if !prevHyphen && b.Len() > 0 {
			b.WriteByte('-')
			prevHyphen = true
		}
	}

	return strings.Trim(b.String(), "-")
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
