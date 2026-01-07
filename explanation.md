# SiteProof Codebase Explanation

*A Guide for Junior Developers*

---

## 📋 Table of Contents

1. [What is SiteProof?](#what-is-siteproof)
2. [Project Structure Overview](#project-structure-overview)
3. [The Big Picture: How Everything Connects](#the-big-picture)
4. [Database Layer (Migrations)](#database-layer)
5. [Go Backend Architecture](#go-backend-architecture)
6. [Docker & Deployment](#docker--deployment)
7. [PDF Generation](#pdf-generation)
8. [Request Flow: A Complete Example](#request-flow)
9. [Key Concepts for Juniors](#key-concepts)
10. [How to Run & Develop](#how-to-run--develop)

---

## What is SiteProof?

SiteProof is a **B2B SaaS application** for the construction industry. It solves a real problem:

> **Problem**: When a subcontractor (like an electrician) does extra work beyond their original contract, they need proof. Otherwise, they might not get paid.

**Solution**: SiteProof lets foremen on construction sites:

1. 📸 Take photos of the extra work
2. 🎤 Dictate a description (voice-to-text)
3. ✍️ Get the GC (General Contractor) superintendent to sign
4. 📄 Generate an **immutable PDF** as proof
5. 📧 Email it to all parties instantly

Think of it as a **digital paper trail** that can't be tampered with.

---

## Project Structure Overview

Here's what each folder contains:

```
siteproof/
├── ADR/                    # Architecture Decision Records
│   └── architecture.md     # Main design document
│
├── cmd/                    # Application entry points
│   └── api/
│       └── main.go         # 🚀 The app starts here!
│
├── internal/               # Private application code
│   ├── config/             # Configuration loading
│   ├── domain/             # Business entities & logic
│   ├── handler/            # HTTP request handlers
│   ├── middleware/         # HTTP middleware (auth, logging)
│   ├── repository/         # Database access layer
│   └── service/            # Business logic services
│
├── migrations/             # Database schema changes
├── pkg/                    # Public utility packages
│   └── crypto/             # Hashing & password utilities
├── templates/              # Typst templates for PDFs
│
├── Dockerfile              # How to build the container
├── docker-compose.yml      # How to run all services
├── Makefile               # Helper commands
├── go.mod                 # Go dependencies
└── sqlc.yaml              # SQL code generator config
```

### Why `internal/` vs `pkg/`?

In Go:
- **`internal/`**: Code that **only this project** can import. No other project can use it.
- **`pkg/`**: Code that **any project** could import. It's public and reusable.

We put `crypto` in `pkg/` because password hashing is generic enough that other projects might want to use it.

---

## The Big Picture

Here's how all the pieces fit together:

```
┌─────────────────────────────────────────────────────────────────────┐
│                        USER'S PHONE/BROWSER                         │
│                     (React Native / React Web App)                  │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ HTTPS requests (JSON)
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          DOCKER CONTAINER                           │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                     Go API Server (port 8080)                 │  │
│  │                                                               │  │
│  │   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐      │  │
│  │   │  Handlers   │───▶│  Services   │───▶│ Repository  │      │  │
│  │   │ (HTTP glue) │    │ (business   │    │ (database   │      │  │
│  │   │             │    │  logic)     │    │  queries)   │      │  │
│  │   └─────────────┘    └─────────────┘    └─────────────┘      │  │
│  │                            │                   │              │  │
│  │                            ▼                   ▼              │  │
│  │                     ┌─────────────┐    ┌─────────────┐       │  │
│  │                     │   Domain    │    │ PostgreSQL  │       │  │
│  │                     │  (entities) │    │  Database   │       │  │
│  │                     └─────────────┘    └─────────────┘       │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │                     PostgreSQL Container                        ││
│  │                      (stores all data)                          ││
│  └─────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
```

**Data flows in layers:**

1. **Handler** → Receives HTTP request, parses JSON, calls service
2. **Service** → Contains business logic, validates rules, calls repository
3. **Repository** → Talks to the database
4. **Domain** → Pure data structures, no dependencies

This is called **Clean Architecture** or **Layered Architecture**.

---

## Database Layer

### What are Migrations?

Migrations are **version-controlled changes to your database schema**.

Think of it like Git for your database structure:
- Migration `000001` creates basic utilities
- Migration `000002` creates `companies` table
- Migration `000003` creates `users` table
- ...and so on

Each migration has **two files**:
- `.up.sql` → Applies the change (creates tables)
- `.down.sql` → Reverts the change (drops tables)

**Example** from `000002_companies.up.sql`:

```sql
CREATE TABLE companies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    slug            TEXT UNIQUE NOT NULL,
    settings        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);
```

### Key Tables Explained

```
┌─────────────────────────────────────────────────────────────────────┐
│                         DATABASE SCHEMA                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────┐         ┌─────────────┐         ┌─────────────┐   │
│  │  companies  │────────▶│    users    │         │  projects   │   │
│  │             │  1:many │             │         │             │   │
│  │  - id       │         │  - id       │         │  - id       │   │
│  │  - name     │         │  - email    │         │  - name     │   │
│  │  - settings │         │  - role     │         │  - address  │   │
│  └─────────────┘         └──────┬──────┘         └──────┬──────┘   │
│                                 │                       │           │
│                                 │  creates              │           │
│                                 ▼                       ▼           │
│                          ┌─────────────────────────────────┐        │
│                          │          tickets                │        │
│                          │                                 │        │
│                          │  - id (UUID)                    │        │
│                          │  - ticket_number (TM-2024-001)  │        │
│                          │  - title                        │        │
│                          │  - description_raw (voice)      │        │
│                          │  - description_cleaned (AI)     │        │
│                          │  - status (draft/approved/...)  │        │
│                          │  - content_hash (SHA-256)       │        │
│                          └─────────────────────────────────┘        │
│                                    │                                 │
│                    ┌───────────────┼───────────────┐                │
│                    │               │               │                │
│                    ▼               ▼               ▼                │
│              ┌─────────┐    ┌───────────┐   ┌───────────┐          │
│              │ photos  │    │ signatures│   │audit_logs │          │
│              └─────────┘    └───────────┘   └───────────┘          │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### The Audit Log: Tamper-Proof Records

One of the most important tables is `audit_logs`. It uses a **hash chain**:

```sql
-- Each entry has:
prev_hash   TEXT,      -- Hash of the PREVIOUS entry
entry_hash  TEXT,      -- Hash of THIS entry
```

**How it works:**

```
┌─────────────────────────────────────────────────────────────────────┐
│                         HASH CHAIN                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Entry 1                Entry 2                Entry 3              │
│  ┌──────────┐          ┌──────────┐          ┌──────────┐          │
│  │prev: NULL│          │prev: abc │          │prev: def │          │
│  │hash: abc │──────────│hash: def │──────────│hash: ghi │          │
│  │data: ... │          │data: ... │          │data: ... │          │
│  └──────────┘          └──────────┘          └──────────┘          │
│                                                                      │
│  If ANYONE modifies Entry 2, its hash changes to "xyz"              │
│  But Entry 3 expects prev_hash to be "def"                          │
│  THE CHAIN IS BROKEN! We know someone tampered with it.             │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

The database **trigger** automatically computes these hashes:

```sql
CREATE TRIGGER audit_hash_chain
    BEFORE INSERT ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION compute_audit_hash();
```

---

## Go Backend Architecture

### Entry Point: `main.go`

Let's walk through what happens when the server starts:

```go
func main() {
    // 1. Setup JSON logging (structured logs are easier to search)
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    // 2. Load configuration from environment (optional .env supported)
    cfg, err := config.Load()

    // 3. Connect to PostgreSQL database
    pool, err := pgxpool.New(ctx, cfg.Database.URL)

    // 4. Create services (business logic)
    authService := service.NewAuthService(pool, &cfg.JWT)
    ticketService := service.NewTicketService(pool)

    // 5. Create handlers (HTTP endpoints)
    authHandler := handler.NewAuthHandler(authService)
    ticketHandler := handler.NewTicketHandler(ticketService)

    // 6. Setup router with routes
    r := chi.NewRouter()
    r.Post("/api/v1/auth/login", authHandler.Login)
    r.Post("/api/v1/tickets", ticketHandler.Create)
    // ...more routes

    // 7. Start HTTP server
    srv := &http.Server{Addr: ":8080", Handler: r}
    srv.ListenAndServe()
}
```

### Domain Layer: Pure Business Objects

The domain layer contains **structs that model real-world concepts**.

**Example:** `internal/domain/ticket.go`

```go
// TicketStatus is a custom type (not just string)
// This gives us type safety - you can't accidentally pass "banana" as a status
type TicketStatus string

const (
    TicketStatusDraft         TicketStatus = "draft"
    TicketStatusPendingReview TicketStatus = "pending_review"
    TicketStatusApproved      TicketStatus = "approved"
    TicketStatusRejected      TicketStatus = "rejected"
    TicketStatusLocked        TicketStatus = "locked"
)

// Ticket represents a T&M or variation ticket
type Ticket struct {
    ID                 uuid.UUID    `json:"id"`
    CompanyID          uuid.UUID    `json:"company_id"`
    TicketNumber       string       `json:"ticket_number"`
    Title              string       `json:"title"`
    Status             TicketStatus `json:"status"`
    // ...more fields
}
```

**Business logic methods on the struct:**

```go
// CanTransitionTo defines valid status changes
func (s TicketStatus) CanTransitionTo(target TicketStatus) bool {
    switch s {
    case TicketStatusDraft:
        return target == TicketStatusPendingReview
    case TicketStatusPendingReview:
        return target == TicketStatusApproved || target == TicketStatusRejected
    case TicketStatusLocked:
        return false // Can NEVER change from locked!
    }
    return false
}
```

This is called a **state machine** - it ensures tickets follow valid workflows:

```
┌───────────────────────────────────────────────────────────────────────┐
│                      TICKET STATUS FLOW                                │
├───────────────────────────────────────────────────────────────────────┤
│                                                                        │
│                              ┌────────────┐                            │
│          ┌──────────────────│   DRAFT    │◀─────────────┐             │
│          │   (foreman       └─────┬──────┘  (rejected,  │             │
│          │    edits)              │          edit again)│             │
│          │                        ▼                     │             │
│          │               ┌────────────────┐             │             │
│          │               │ PENDING_REVIEW │─────────────┤             │
│          │               └───────┬────────┘             │             │
│          │                       │                      │             │
│          │           ┌───────────┴───────────┐          │             │
│          │           │                       │          │             │
│          │           ▼                       ▼          │             │
│          │   ┌────────────┐          ┌────────────┐     │             │
│          │   │  APPROVED  │          │  REJECTED  │─────┘             │
│          │   └─────┬──────┘          └────────────┘                   │
│          │         │                                                   │
│          │         ▼                                                   │
│          │   ┌────────────┐                                           │
│          │   │   LOCKED   │  ← PDF generated, CAN'T CHANGE!           │
│          │   └────────────┘                                           │
│          │                                                             │
└──────────┴─────────────────────────────────────────────────────────────┘
```

### User Roles

```go
type Role string

const (
    RoleForeman    Role = "foreman"     // Creates tickets, takes photos
    RoleGCSuper    Role = "gc_super"    // Signs tickets
    RolePM         Role = "pm"          // Project manager, reviews
    RoleAccounting Role = "accounting"  // Views for billing
    RoleAdmin      Role = "admin"       // Manages everything
)
```

Each role has different permissions:

```go
func (u *User) CanCreateTicket() bool {
    return u.Role == RoleForeman || u.Role == RoleAdmin
}

func (u *User) CanSignAsGC() bool {
    return u.Role == RoleGCSuper
}
```

### Service Layer: Business Logic

Services contain the **rules** of the application.

**Example:** `internal/service/auth.go`

```go
func (s *AuthService) Login(ctx context.Context, input *domain.LoginInput) (*TokenPair, *domain.User, error) {
    // 1. Validate input
    if err := input.Validate(); err != nil {
        return nil, nil, err
    }

    // 2. Find user by email in database
    var user domain.User
    err := s.pool.QueryRow(ctx, `
        SELECT id, company_id, email, password_hash, ...
        FROM users
        WHERE email = $1 AND deleted_at IS NULL
    `, input.Email).Scan(&user.ID, ...)

    // 3. Verify password using bcrypt
    if !crypto.CheckPassword(input.Password, user.PasswordHash) {
        return nil, nil, domain.ErrUnauthorized
    }

    // 4. Check if user is active
    if !user.Active {
        return nil, nil, domain.ErrUnauthorized
    }

    // 5. Generate JWT tokens
    tokens, err := s.generateTokenPair(&user)
    return tokens, &user, nil
}
```

### Handler Layer: HTTP Glue

Handlers **translate HTTP requests into service calls**.

**Example:** `internal/handler/ticket.go`

```go
func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. Get user from context (set by auth middleware)
    userID, ok := middleware.UserIDFromContext(r.Context())
    if !ok {
        WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
        return
    }

    // 2. Parse JSON request body
    var req CreateTicketRequest
    if err := DecodeJSON(r, &req); err != nil {
        WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
        return
    }

    // 3. Build domain input
    input := &domain.CreateTicketInput{
        CompanyID:   companyID,
        ProjectID:   projectID,
        CreatedByID: userID,
        Title:       req.Title,
    }

    // 4. Call service
    ticket, err := h.ticketService.Create(r.Context(), input)

    // 5. Return JSON response
    WriteJSON(w, http.StatusCreated, ticket)
}
```

### Middleware: Cross-Cutting Concerns

Middleware runs **before** and **after** every request.

**Authentication Middleware:**

```go
func Auth(authService *AuthService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. Get token from Authorization header
            authHeader := r.Header.Get("Authorization")  // "Bearer eyJ..."
            
            // 2. Parse "Bearer <token>"
            parts := strings.SplitN(authHeader, " ", 2)
            token := parts[1]
            
            // 3. Validate token
            claims, err := authService.ValidateAccessToken(token)
            if err != nil {
                WriteError(w, 401, "UNAUTHORIZED", "invalid token")
                return
            }
            
            // 4. Add user info to request context
            ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
            ctx = context.WithValue(ctx, companyIDKey, claims.CompanyID)
            
            // 5. Continue to actual handler
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

**How middleware chains work:**

```
Request → [RequestID] → [Logging] → [Auth] → [Handler] → Response
              │             │         │
              │             │         └── Validates JWT token
              │             └── Logs request method, path, duration
              └── Adds unique ID for tracing
```

---

## Docker & Deployment

### Dockerfile Explained

The Dockerfile uses a **multi-stage build** for smaller images:

```dockerfile
# ============================================
# STAGE 1: Build the Go binary
# ============================================
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0 = no C dependencies (pure Go)
# -ldflags="-w -s" = strip debug info (smaller binary)
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/bin/siteproof-api ./cmd/api


# ============================================
# STAGE 2: Create minimal runtime image
# ============================================
FROM alpine:3.20

# Install only what we need to run
RUN apk add --no-cache ca-certificates tzdata curl typst

# Create non-root user (security!)
RUN adduser -u 1000 -D siteproof

# Copy binary from builder stage
COPY --from=builder /app/bin/siteproof-api ./bin/

# Run as non-root user
USER siteproof

CMD ["./bin/siteproof-api"]
```

**Why multi-stage?**
- Build stage has Go compiler (~1GB)
- Final image only has the binary (~50MB)

### docker-compose.yml Explained

Docker Compose orchestrates **multiple containers**:

```yaml
version: "3.8"

services:
  # The Go API server
  api:
    build: .
    ports:
      - "8080:8080"  # Expose port 8080 on host
    environment:
      - DATABASE_URL=postgres://siteproof:password@db:5432/siteproof
      - JWT_SECRET=${JWT_SECRET}  # Read from .env file
    depends_on:
      db:
        condition: service_healthy  # Wait for DB to be ready!
    volumes:
      - uploads:/data/uploads  # Persist uploaded files

  # PostgreSQL database
  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_DB=siteproof
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data  # Persist database
    healthcheck:
      test: ["CMD", "pg_isready"]  # How to check if healthy

  # Run migrations
  migrate:
    image: migrate/migrate:v4.17.0
    command: ["-path", "/migrations", "-database", "...", "up"]
    volumes:
      - ./migrations:/migrations:ro

volumes:
  uploads:   # Named volume for files
  pgdata:    # Named volume for database
```

**Container communication:**

```
┌─────────────────────────────────────────────────────────┐
│                    Docker Network                       │
│                                                         │
│  ┌─────────────┐       ┌─────────────┐                 │
│  │     api     │──────▶│     db      │                 │
│  │  port 8080  │       │  port 5432  │                 │
│  └─────────────┘       └─────────────┘                 │
│        │                                                │
│        │ DATABASE_URL=postgres://...@db:5432           │
│        │                        ↑                      │
│        │                        │                      │
│        │               Container name = hostname!       │
│                                                         │
└─────────────────────────────────────────────────────────┘
        │
        ▼
   Host machine
   http://localhost:8080
```

---

## PDF Generation

SiteProof uses **Typst** for PDF generation. Typst is like LaTeX but modern and fast.

### Template Structure

`templates/ticket.typ`:

```typst
#let ticket(
  ticket_number: "",
  project_name: "",
  line_items: (),
  // ...more parameters
) = {
  // Page setup
  set page(paper: "us-letter", margin: 1in)
  
  // Header
  align(center)[
    #text(size: 18pt, weight: "bold")[Time & Materials Ticket]
    #text(size: 14pt)[#ticket_number]
  ]
  
  // Line items table
  table(
    columns: (auto, 1fr, auto, auto),
    [*Type*], [*Description*], [*Qty*], [*Total*],
    ..line_items.map(item => (
      item.type,
      item.description,
      str(item.quantity),
      "$" + str(item.total),
    )).flatten(),
  )
  
  // Signature boxes
  grid(
    columns: (1fr, 1fr),
    [
      *Subcontractor:*
      #image(foreman_signature)
      Signed: #foreman_signed_at
    ],
    [
      *GC Superintendent:*  
      #image(gc_signature)
      Signed: #gc_signed_at
    ],
  )
}
```

The Go server:
1. Loads ticket data from database
2. Passes it to Typst as JSON
3. Typst renders to PDF
4. PDF is saved and hash is computed

---

## Request Flow: A Complete Example

Let's trace what happens when a foreman creates a ticket:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    CREATING A TICKET: FULL FLOW                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  1. MOBILE APP                                                           │
│     ────────────                                                         │
│     User fills form, taps "Create"                                       │
│                                                                          │
│     POST /api/v1/tickets                                                 │
│     Headers:                                                             │
│       Authorization: Bearer eyJhbGciOi...                                │
│       Content-Type: application/json                                     │
│     Body:                                                                │
│       {                                                                  │
│         "project_id": "123e4567-e89b-...",                               │
│         "ticket_type": "time_materials",                                 │
│         "title": "Extra conduit run",                                    │
│         "work_date": "2024-12-31"                                        │
│       }                                                                  │
│                                                                          │
│                              │                                           │
│                              ▼                                           │
│  2. MIDDLEWARE STACK                                                     │
│     ─────────────────                                                    │
│     [RequestID] → Adds X-Request-ID: "abc-123"                           │
│     [Logging]   → Logs: POST /api/v1/tickets                             │
│     [Auth]      → Validates JWT token                                    │
│                 → Extracts: userID="...", companyID="...", role="foreman"│
│                 → Adds to context                                        │
│                                                                          │
│                              │                                           │
│                              ▼                                           │
│  3. HANDLER                                                              │
│     ───────                                                              │
│     ticketHandler.Create()                                               │
│     - Parses JSON body                                                   │
│     - Gets userID from context                                           │
│     - Builds CreateTicketInput                                           │
│     - Calls ticketService.Create()                                       │
│                                                                          │
│                              │                                           │
│                              ▼                                           │
│  4. SERVICE                                                              │
│     ───────                                                              │
│     ticketService.Create()                                               │
│     - Validates input                                                    │
│     - Generates ticket_number: "TM-2024-0001"                            │
│     - Starts database transaction                                        │
│     - Inserts ticket row                                                 │
│     - Creates audit_log entry                                            │
│     - Commits transaction                                                │
│                                                                          │
│                              │                                           │
│                              ▼                                           │
│  5. DATABASE                                                             │
│     ────────                                                             │
│     INSERT INTO tickets (...) VALUES (...)                               │
│     INSERT INTO audit_logs (...) VALUES (...)                            │
│     → Trigger computes hash chain!                                       │
│                                                                          │
│                              │                                           │
│                              ▼                                           │
│  6. RESPONSE                                                             │
│     ────────                                                             │
│     HTTP 201 Created                                                     │
│     {                                                                    │
│       "id": "789...",                                                    │
│       "ticket_number": "TM-2024-0001",                                   │
│       "status": "draft",                                                 │
│       "created_at": "2024-12-31T10:30:00Z"                               │
│     }                                                                    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Key Concepts for Juniors

### 1. UUID vs Auto-Increment IDs

```go
// We use UUIDs like: "123e4567-e89b-12d3-a456-426614174000"
// Instead of integers like: 1, 2, 3...

// Why?
// - Can generate IDs on the client (offline support)
// - No information leakage ("user 5" reveals you have few users)
// - Easier to merge databases
// - Can't guess other IDs
```

### 2. JWT (JSON Web Tokens)

```
┌─────────────────────────────────────────────────────────────────────┐
│                        JWT TOKEN                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoiMTIzIn0.abc123                │
│  ─────────────────────  ─────────────────────── ──────              │
│         Header              Payload            Signature             │
│                                                                      │
│  Header: {"alg": "HS256"}  (algorithm used)                          │
│                                                                      │
│  Payload: {                                                          │
│    "user_id": "123...",                                              │
│    "company_id": "456...",                                           │
│    "role": "foreman",                                                │
│    "exp": 1735689600        (expiration timestamp)                   │
│  }                                                                   │
│                                                                      │
│  Signature: HMAC-SHA256(header + payload, SECRET_KEY)                │
│                                                                      │
│  The server can verify the signature to trust the payload            │
│  without hitting the database on every request!                      │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 3. bcrypt for Passwords

```go
// NEVER store plain passwords!

// When user registers:
hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
// hash = "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.E..."

// When user logs in:
err := bcrypt.CompareHashAndPassword(hash, []byte("password123"))
// err == nil means password matches!

// The "12" is the cost factor - higher = slower = more secure
// Takes ~250ms to hash, making brute force attacks very slow
```

### 4. Context in Go

```go
// Context carries request-scoped values and cancellation signals

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()  // Get context from request
    
    // Get values set by middleware
    userID := ctx.Value("user_id").(uuid.UUID)
    
    // Pass context to database queries
    // If request is cancelled, query is cancelled too!
    row := h.db.QueryRow(ctx, "SELECT * FROM tickets WHERE id = $1", id)
}
```

### 5. Database Transactions

```go
// A transaction ensures multiple operations succeed or fail TOGETHER

func (s *Service) CreateTicket(ctx context.Context, input *Input) error {
    tx, _ := s.pool.Begin(ctx)  // Start transaction
    defer tx.Rollback(ctx)      // Rollback if we return early
    
    // Insert ticket
    _, err := tx.Exec(ctx, "INSERT INTO tickets ...")
    if err != nil {
        return err  // Rollback happens automatically
    }
    
    // Insert audit log
    _, err = tx.Exec(ctx, "INSERT INTO audit_logs ...")
    if err != nil {
        return err  // Rollback happens automatically
    }
    
    return tx.Commit(ctx)  // Both succeed, or neither does!
}
```

---

## How to Run & Develop

### Prerequisites

- Docker & Docker Compose
- Go 1.22+ (for local development)
- Make (optional, for helper commands)

### Quick Start

```bash
# 1. Clone and navigate
cd siteproof

# 2. Create .env file
cp .env.example .env
# Edit .env with your secrets

# 3. Start everything
docker compose up -d

# 4. Run migrations
docker compose run migrate

# 5. Check logs
docker compose logs -f api
```

### Local Development (without Docker)

```bash
# 1. Start just the database
docker compose up -d db

# 2. Run migrations
make migrate-up

# 3. Run the API locally (hot reload with air)
go run ./cmd/api

# 4. The API is now at http://localhost:8080
```

### Useful Commands

```bash
# View logs
docker compose logs -f

# Connect to database
make psql

# Create new migration
make migrate-create name=add_new_table

# Run tests
go test ./...

# Format code
go fmt ./...
```

---

## Summary

SiteProof is built with:

| Layer | Technology | Purpose |
|-------|------------|---------|
| **Frontend** | React Native / React | Mobile & web apps |
| **API** | Go + Chi router | REST API server |
| **Database** | PostgreSQL | Data storage |
| **Containerization** | Docker | Consistent deployments |
| **PDF** | Typst | Document generation |

The architecture follows **clean/layered design**:
- **Handlers** → HTTP concerns only
- **Services** → Business logic
- **Repository** → Database access
- **Domain** → Pure data structures

Key security features:
- JWT authentication
- bcrypt password hashing
- Hash-chained audit logs (tamper-evident)
- Database triggers prevent locked ticket modification

---

*Happy coding! 🚀*
