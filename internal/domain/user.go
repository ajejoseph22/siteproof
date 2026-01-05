package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a user's role in the system.
type Role string

const (
	RoleForeman    Role = "foreman"
	RoleGCSuper    Role = "gc_super"
	RolePM         Role = "pm"
	RoleAccounting Role = "accounting"
	RoleAdmin      Role = "admin"
)

// IsValid returns true if the role is valid.
func (r Role) IsValid() bool {
	switch r {
	case RoleForeman, RoleGCSuper, RolePM, RoleAccounting, RoleAdmin:
		return true
	}
	return false
}

// User represents an application user.
type User struct {
	ID           uuid.UUID  `json:"id"`
	CompanyID    uuid.UUID  `json:"company_id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // Never serialize
	Name         string     `json:"name"`
	Phone        *string    `json:"phone,omitempty"`
	Role         Role       `json:"role"`
	Active       bool       `json:"active"`
	DeletedAt    *time.Time `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// IsDeleted returns true if the user has been soft deleted.
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// CanCreateTicket returns true if the user can create tickets.
func (u *User) CanCreateTicket() bool {
	return u.Role == RoleForeman || u.Role == RoleAdmin
}

// CanSignAsForeman returns true if the user can sign as foreman.
func (u *User) CanSignAsForeman() bool {
	return u.Role == RoleForeman
}

// CanSignAsGC returns true if the user can sign as GC superintendent.
func (u *User) CanSignAsGC() bool {
	return u.Role == RoleGCSuper
}

// CanViewAuditLogs returns true if the user can view audit logs.
func (u *User) CanViewAuditLogs() bool {
	return u.Role == RolePM || u.Role == RoleAdmin
}

// CanManageUsers returns true if the user can manage other users.
func (u *User) CanManageUsers() bool {
	return u.Role == RoleAdmin
}

// CreateUserInput represents input for creating a new user.
type CreateUserInput struct {
	CompanyID uuid.UUID
	Email     string
	Password  string
	Name      string
	Phone     *string
	Role      Role
}

// Validate validates the create user input.
func (i *CreateUserInput) Validate() error {
	var errs ValidationErrors

	if i.Email == "" {
		errs.Add("email", "is required")
	}
	if i.Password == "" {
		errs.Add("password", "is required")
	} else if len(i.Password) < 8 {
		errs.Add("password", "must be at least 8 characters")
	}
	if i.Name == "" {
		errs.Add("name", "is required")
	}
	if !i.Role.IsValid() {
		errs.Add("role", "is invalid")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// UpdateUserInput represents input for updating a user.
type UpdateUserInput struct {
	Name   *string
	Phone  *string
	Role   *Role
	Active *bool
}

// LoginInput represents input for user login.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate validates the login input.
func (i *LoginInput) Validate() error {
	var errs ValidationErrors

	if i.Email == "" {
		errs.Add("email", "is required")
	}
	if i.Password == "" {
		errs.Add("password", "is required")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}
