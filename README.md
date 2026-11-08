# SiteProof

SiteProof lets construction subcontractors (electricians, plumbers, HVAC crews) record Time & Materials and variation work on site. A foreman logs the work with photos and a dictated description, the general contractor's superintendent signs it, and the ticket is locked into a tamper-evident record that can be used if the work is ever disputed.

> **Status:** work in progress. The core API (auth, tenants, RBAC, projects, tickets, line items, photos, signatures, audit chain) is in place. PDF generation, email/SMS delivery, voice cleanup and most of the web UI are still open. See [Roadmap](#roadmap).

## How it works

```
draft → pending_review → approved → locked
                       ↘ rejected → draft
```

- **Tickets** move through the state machine above. Once a ticket is locked, database triggers block any further change.
- **Photos** are hashed (SHA-256) on upload and served through authenticated download endpoints.
- **Signatures** are captured on the foreman's device or through a signing link sent to the superintendent.
- **Audit log** entries form a hash chain (`prev_hash` → `entry_hash`) computed in Postgres, and the API can verify the chain.
- **Multi-tenant**: every row is scoped to a company, and role checks (foreman, GC super, PM, accounting, admin) run in middleware.

## Tech stack

| Layer    | Tools |
|----------|-------|
| API      | Go 1.22, chi, pgx/v5, sqlc, JWT (access + refresh) |
| Database | PostgreSQL 16, golang-migrate |
| Web      | React 19, Vite, TypeScript, Tailwind v4, Zustand, TanStack Query |
| PDF      | Typst (`templates/ticket.typ`) |
| Ops      | Docker Compose, GitHub Actions |

## Repository layout

```
cmd/api/              API entry point
internal/domain/      Entities, validation, state machine (no dependencies)
internal/service/     Business logic
internal/repository/  sqlc queries (queries/) and generated code (db/)
internal/handler/     HTTP handlers
internal/middleware/  Auth, RBAC, idempotency, logging, recovery
internal/config/      Config loading
pkg/                  Crypto and logging helpers
migrations/           SQL migrations (000001_*.up.sql / .down.sql)
templates/            Typst PDF templates
client/               React web app
openapi.yaml          API specification
ADR/architecture.md   Architecture and design decisions
```

## Getting started

Prerequisites: Go 1.22+, Node 20+, Docker, [sqlc](https://sqlc.dev) and [golang-migrate](https://github.com/golang-migrate/migrate).

```bash
# 1. Configure
cp .env.example .env          # set JWT_SECRET (32+ chars) and DATABASE_URL

# 2. Start Postgres and apply migrations
make docker-up                # or point DATABASE_URL at your own Postgres
make db-migrate-up
make db-seed                  # optional demo data

# 3. Run the API (http://localhost:8080)
make dev

# 4. Run the web app (http://localhost:5173)
cd client
cp .env.example .env          # VITE_DEMO_MODE=true uses mock data, false calls the API
npm install
npm run dev
```

Run `make help` to list every target.

## Development

```bash
make test             # go test ./...
make lint             # golangci-lint
make sqlc-generate    # regenerate internal/repository/db after editing queries
cd client && npm run lint && npm run build
```

- Generated sqlc code is committed, and CI fails if it drifts from the queries.
- Every migration needs a matching `.down.sql`.
- Commits follow [Conventional Commits](https://www.conventionalcommits.org).

CI runs vet, tests, lint, sqlc verification, a Typst compile and the client build on every push and pull request. Building the image and deploying run only when triggered by hand.

## Roadmap

The full task list, with acceptance steps and status, lives in [`plans/prd.json`](plans/prd.json). [`MVP_BACKLOG.md`](MVP_BACKLOG.md) summarises the gaps. Main open items:

- Content hashing at lock time, PDF generation with Typst, and a ticket integrity verify endpoint
- Email via Postmark (with delivery webhooks), SMS signing links via Twilio, and voice-note cleanup via OpenAI
- Web UI for ticket list and detail, the create/edit form with line items, photo upload, signature flows and PDF preview
- Tests for ticket locking, RBAC, photo hashing and the audit chain
- Prometheus metrics, a Caddy config, and S3-compatible storage
