# Database Migrations

This directory contains SQL migrations for the SiteProof database using [golang-migrate](https://github.com/golang-migrate/migrate).

## Migration Structure

```
migrations/
├── 000001_extensions_and_utilities.up.sql    # pgcrypto, utility functions
├── 000001_extensions_and_utilities.down.sql
├── 000002_companies.up.sql                   # Multi-tenant root entity
├── 000002_companies.down.sql
├── 000003_users.up.sql                       # User accounts with RBAC
├── 000003_users.down.sql
├── 000004_projects.up.sql                    # Construction job sites
├── 000004_projects.down.sql
├── 000005_tickets.up.sql                     # T&M and variation records
├── 000005_tickets.down.sql
├── 000006_ticket_line_items.up.sql           # Labor, materials, equipment
├── 000006_ticket_line_items.down.sql
├── 000007_photos.up.sql                      # Photo evidence
├── 000007_photos.down.sql
├── 000008_signing_links.up.sql               # Secure signature requests
├── 000008_signing_links.down.sql
├── 000009_signatures.up.sql                  # Digital signatures
├── 000009_signatures.down.sql
├── 000010_audit_logs.up.sql                  # Hash-chained audit trail
├── 000010_audit_logs.down.sql
├── 000011_email_events.up.sql                # Email delivery tracking
├── 000011_email_events.down.sql
├── 000012_transcription_versions.up.sql      # Voice-to-text history
├── 000012_transcription_versions.down.sql
├── 000013_refresh_tokens.up.sql              # JWT refresh tokens
├── 000013_refresh_tokens.down.sql
├── 000014_ticket_lock_enforcement.up.sql     # Immutability triggers
├── 000014_ticket_lock_enforcement.down.sql
├── 000015_idempotency_keys.up.sql            # Request idempotency
├── 000015_idempotency_keys.down.sql
├── 000016_seed_data.up.sql                   # Development seed data
└── 000016_seed_data.down.sql
```

## Usage

### Prerequisites

Install golang-migrate:

```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Docker
docker pull migrate/migrate
```

### Running Migrations

```bash
# Set database URL
export DATABASE_URL="postgres://user:pass@localhost:5432/siteproof?sslmode=disable"

# Apply all migrations
make db-migrate-up

# Rollback last migration
make db-migrate-down

# Force to specific version (use with caution)
make db-migrate-force VERSION=5

# Check current version
make db-migrate-version

# Create new migration
make db-migrate-create NAME=add_new_feature
```

### Docker Compose

If using the Docker setup, run migrations inside the container:

```bash
docker compose exec api migrate -path=/opt/siteproof/migrations -database="$DATABASE_URL" up
```

## Key Database Features

### Hash-Chained Audit Logs
- `audit_logs` table uses SHA-256 hash chain for tamper evidence
- Each entry's `entry_hash` includes hash of previous entry
- Trigger prevents UPDATE/DELETE on audit logs

### Ticket Immutability
- `ticket_lock_enforcement` trigger prevents modification of locked tickets
- `ticket_status_transition` trigger enforces valid status workflow
- `content_hash` and `pdf_hash` stored at lock time for verification

### Status Workflow

```
draft → pending_review → approved → locked
                      ↘ rejected → draft (revise and resubmit)
```

## Conventions

1. **Naming**: `{number}_{description}.{up|down}.sql`
2. **Numbers**: 6-digit, zero-padded (000001, 000002, ...)
3. **Reversibility**: Every `.up.sql` has matching `.down.sql`
4. **Idempotency**: Use `IF EXISTS` / `IF NOT EXISTS` where appropriate
5. **Comments**: Document tables, columns, and complex logic
6. **Constraints**: Named constraints for clear error messages

## Testing Migrations

```bash
# Fresh database
make db-reset

# Verify up/down cycle
make db-migrate-up
make db-migrate-down
make db-migrate-up

# Check for drift
make db-migrate-version
```

## Production Checklist

- [ ] Backup database before migration
- [ ] Test migration on staging first
- [ ] Review `.down.sql` for data loss implications
- [ ] Monitor database locks during migration
- [ ] Have rollback plan ready

