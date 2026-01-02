# SiteProof Context for Gemini

This file provides context for the SiteProof project to assist with future development tasks.

## Project Overview

**SiteProof** is a B2B SaaS application for the construction industry designed to capture, approve, and lock Time & Materials (T&M) and variation work. It acts as a digital paper trail, ensuring subcontractors get paid for extra work by generating immutable, signed PDF records.

### Key Features
*   **Ticket Management:** Create T&M tickets with voice-to-text descriptions (cleaned by AI).
*   **Evidence:** Capture photos with GPS and timestamp metadata.
*   **Signatures:** Secure digital signatures from General Contractor (GC) superintendents (link-based or in-person).
*   **Immutability:** Hash-chained audit logs and locked PDF records to prevent tampering.
*   **PDF Generation:** Fast, reproducible PDF generation using Typst.

## Architecture

The project follows a **Clean/Layered Architecture**:

*   **`cmd/api/`**: Application entry points (`main.go`).
*   **`internal/domain/`**: Pure business entities and logic (no dependencies).
*   **`internal/service/`**: Business logic orchestration.
*   **`internal/repository/`**: Database access (PostgreSQL).
*   **`internal/handler/`**: HTTP transport layer (REST API).
*   **`internal/middleware/`**: Cross-cutting concerns (Auth, Logging).
*   **`migrations/`**: Database schema changes.
*   **`templates/`**: Typst templates for PDF generation.

### Tech Stack

*   **Backend:** Go 1.22+
    *   Router: `go-chi/chi`
    *   Database: `pgx/v5`
    *   SQL Generation: `sqlc`
    *   Auth: JWT + `bcrypt`
*   **Frontend:** React 19 + Vite (TypeScript)
    *   Styling: Tailwind CSS v4
    *   State: Zustand + TanStack Query
    *   UI Components: Radix UI (likely Shadcn UI)
    *   Forms: React Hook Form + Zod
*   **Database:** PostgreSQL 16
    *   Uses `pgcrypto` for hashing.
    *   Uses `golang-migrate` for schema management.
*   **Infrastructure:** Docker & Docker Compose
    *   Reverse Proxy: Caddy
    *   PDF Engine: Typst

## Development Workflow

### Key Commands (Makefile)

*   **Start Infrastructure (Docker):**
    ```bash
    make docker-up      # Starts API, DB, and Caddy
    make docker-down    # Stops services
    make docker-logs    # Tails logs
    ```

*   **Database Management:**
    ```bash
    make db-migrate-up      # Apply migrations
    make db-migrate-down    # Rollback last migration
    make db-migrate-create NAME=my_migration # Create new migration files
    make db-reset           # Reset DB and re-apply migrations
    make db-psql            # Connect to DB shell
    ```

*   **Backend Development:**
    ```bash
    make dev            # Run API locally
    make build          # Build API binary
    make test           # Run tests
    make lint           # Run linters
    make sqlc-generate  # Regenerate Go code from SQL queries
    ```

*   **Frontend Development:**
    ```bash
    cd frontend
    npm install
    npm run dev         # Start dev server
    ```

### Conventions

*   **Database First:** Changes often start with a migration and `sqlc` queries.
*   **Type Safety:** Use `sqlc` to generate type-safe Go code for database interactions. Avoid raw SQL in Go code where possible.
*   **Immutability:** Critical data (audit logs, locked tickets) is immutable. Ensure hash chains are maintained.
*   **Error Handling:** Use custom error types in the domain layer and map them to HTTP status codes in the handler layer.
*   **Code Style:** Follow standard Go idioms (`gofmt`).

## Critical Files

*   **`explanation.md`**: Comprehensive guide for developers (READ THIS FIRST).
*   **`ADR/architecture.md`**: Detailed architectural decisions and data model.
*   **`sqlc.yaml`**: Configuration for SQL code generation.
*   **`docker-compose.yml`**: Definition of local development environment.
*   **`go.mod`**: Backend dependencies.
*   **`frontend/package.json`**: Frontend dependencies.

## Key Data Models

*   **Tickets:** The core entity (`tickets` table). State machine: `draft` -> `pending_review` -> `approved`/`rejected` -> `locked`.
*   **Audit Logs:** Append-only table (`audit_logs`) with hash chaining for tamper evidence.
*   **Signatures:** Stored in `signatures` table, linked to tickets.

## Environment Variables

See `.env.example` for required variables. Key ones include:
*   `DATABASE_URL`
*   `JWT_SECRET`
*   `POSTMARK_API_KEY` (Email)
*   `OPENAI_API_KEY` (Voice cleanup)
*   `TWILIO_ACCOUNT_SID` (SMS)
