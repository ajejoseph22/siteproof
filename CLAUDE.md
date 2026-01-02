# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SiteProof is a B2B SaaS for construction subcontractors to capture T&M (Time & Materials) and variation work. Foremen take photos, dictate descriptions (voice-to-text), get GC superintendent signatures, and generate immutable PDF records.

## Commands

### Backend (Go)
```bash
make dev                    # Run API locally (go run ./cmd/api)
make test                   # Run all tests (go test -v ./...)
make lint                   # Run golangci-lint
make build                  # Build binary to bin/siteproof-api
make sqlc-generate          # Regenerate Go code from SQL queries
```

### Frontend (Vite + React)
```bash
cd frontend && npm run dev      # Start dev server
cd frontend && npm run build    # Build (tsc -b && vite build)
cd frontend && npm run lint     # ESLint
```

### Database
```bash
make db-migrate-up              # Apply pending migrations
make db-migrate-down            # Rollback last migration
make db-migrate-create NAME=xyz # Create new migration files
make db-reset                   # Drop, recreate, and migrate
make db-psql                    # Open psql shell
```

### Docker
```bash
make docker-up      # Start all services
make docker-down    # Stop all services
make docker-logs    # Tail logs
```

## Architecture

Clean/layered architecture with unidirectional dependencies:

```
Handler (HTTP) → Service (business logic) → Repository (DB) → Domain (entities)
```

### Directory Structure
- `cmd/api/` - Entry point (`main.go`)
- `internal/domain/` - Business entities, validation, state machines (no dependencies)
- `internal/service/` - Business logic orchestration
- `internal/repository/` - Database access layer
- `internal/handler/` - HTTP handlers (Chi router)
- `internal/middleware/` - Auth, logging, recovery
- `internal/config/` - Configuration loading
- `pkg/crypto/` - Password hashing, SHA-256 utilities
- `migrations/` - SQL migrations (golang-migrate)
- `templates/` - Typst templates for PDF generation
- `frontend/` - React + Vite web app
- `web/` - Alternative React web app (signature canvas)

### Key Technologies
- **Backend**: Go 1.22+, Chi router, pgx/v5, sqlc, JWT
- **Frontend**: React 19, Vite, TypeScript, Tailwind v4, Zustand, TanStack Query
- **Database**: PostgreSQL 16 with pgcrypto
- **PDF**: Typst

## Data Flow Patterns

### Ticket Status State Machine
```
draft → pending_review → approved → locked
                       ↘ rejected → draft (revise)
```
Once locked, tickets cannot be modified (enforced by database triggers).

### Audit Log Hash Chain
The `audit_logs` table uses a hash chain for tamper evidence. Each entry includes `prev_hash` (hash of previous entry) and `entry_hash` (hash of current entry). The chain is computed by a PostgreSQL trigger.

### Authentication
JWT tokens with access/refresh pattern. Access tokens are short-lived, refresh tokens are stored in `refresh_tokens` table.

## Database Conventions

1. Use `sqlc` for type-safe queries - regenerate with `make sqlc-generate`
2. Migrations use 6-digit numbered prefixes (000001, 000002, ...)
3. Every `.up.sql` must have a matching `.down.sql`
4. Locked tickets and audit logs are immutable (enforced by triggers)

## Environment Variables

Key variables (see `.env.example`):
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - JWT signing key
- `POSTMARK_API_KEY` - Email delivery
- `OPENAI_API_KEY` - Voice transcription cleanup
- `TWILIO_ACCOUNT_SID` - SMS for signature verification
