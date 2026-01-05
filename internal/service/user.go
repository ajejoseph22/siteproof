package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/ajejoseph22/siteproof/pkg/crypto"
)

// UserService handles user operations.
type UserService struct {
	pool *pgxpool.Pool
}

// NewUserService creates a new user service.
func NewUserService(pool *pgxpool.Pool) *UserService {
	return &UserService{pool: pool}
}

// Create creates a new user.
func (s *UserService) Create(ctx context.Context, input *domain.CreateUserInput) (*domain.User, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Check if email exists
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	`, input.Email).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if exists {
		return nil, domain.NewDomainError(domain.ErrConflict, "EMAIL_EXISTS", "email already exists")
	}

	// Hash password
	hash, err := crypto.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	var user domain.User
	err = s.pool.QueryRow(ctx, `
		INSERT INTO users (company_id, email, password_hash, name, phone, role, active)
		VALUES ($1, $2, $3, $4, $5, $6, true)
		RETURNING id, company_id, email, name, phone, role, active, created_at, updated_at
	`, input.CompanyID, input.Email, hash, input.Name, input.Phone, input.Role).Scan(
		&user.ID, &user.CompanyID, &user.Email, &user.Name, &user.Phone,
		&user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

// GetByID retrieves a user by ID.
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, company_id, email, name, phone, role, active, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&user.ID, &user.CompanyID, &user.Email, &user.Name, &user.Phone,
		&user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email.
func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, company_id, email, name, phone, role, active, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`, email).Scan(
		&user.ID, &user.CompanyID, &user.Email, &user.Name, &user.Phone,
		&user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &user, nil
}

// ListByCompany retrieves users for a company.
func (s *UserService) ListByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]*domain.User, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, company_id, email, name, phone, role, active, created_at, updated_at
		FROM users
		WHERE company_id = $1 AND deleted_at IS NULL
		ORDER BY name
		LIMIT $2 OFFSET $3
	`, companyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID, &user.CompanyID, &user.Email, &user.Name, &user.Phone,
			&user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// Update updates a user.
func (s *UserService) Update(ctx context.Context, id uuid.UUID, input *domain.UpdateUserInput) (*domain.User, error) {
	var user domain.User
	err := s.pool.QueryRow(ctx, `
		UPDATE users
		SET name = COALESCE($2, name),
		    phone = COALESCE($3, phone),
		    role = COALESCE($4, role),
		    active = COALESCE($5, active),
		    updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, company_id, email, name, phone, role, active, created_at, updated_at
	`, id, input.Name, input.Phone, input.Role, input.Active).Scan(
		&user.ID, &user.CompanyID, &user.Email, &user.Name, &user.Phone,
		&user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &user, nil
}

// Delete soft deletes a user.
func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE users
		SET deleted_at = now(), active = false, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
