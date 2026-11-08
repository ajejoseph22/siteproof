# Repository Guidelines

## Project Structure & Module Organization
- `cmd/api/` holds the API entry point.
- `internal/` contains domain, service, repository, handler, middleware, and config layers.
- `pkg/` provides shared utilities (e.g., crypto helpers).
- `migrations/` stores database migrations using numbered `000001_*.up.sql` / `.down.sql` pairs.
- `templates/` contains Typst templates used for PDF generation.
- `client/` is the React + Vite web app.
- `openapi.yaml` documents the HTTP API surface.

## Build, Test, and Development Commands
- `make dev` runs the Go API locally (`go run ./cmd/api`).
- `make build` builds the API binary to `bin/siteproof-api`.
- `make test` runs all Go tests (`go test -v ./...`).
- `make lint` runs `golangci-lint`.
- `make sqlc-generate` regenerates Go code from SQL queries.
- `make db-migrate-up` / `make db-migrate-down` applies or rolls back migrations.
- `make docker-up` / `make docker-down` starts or stops the Docker stack.
- `cd client && npm run dev|build|lint` runs the Vite dev server, builds, or lints.

## Coding Style & Naming Conventions
- Go: run `gofmt` (or `go fmt`) before committing; keep exported identifiers in `CamelCase`.
- Frontend: TypeScript + ESLint; prefer `PascalCase` for components and `camelCase` for variables.
- SQL migrations must use 6-digit numeric prefixes and include both `.up.sql` and `.down.sql` files.

## Testing Guidelines
- Go tests live alongside code in `*_test.go` files and use the standard `testing` package.
- There is no formal coverage target; add tests for new business logic and bug fixes.
- Run `make test` for backend changes; run `cd client && npm run lint` for UI changes.

## Commit & Pull Request Guidelines
- Commit messages follow Conventional Commits (e.g., `feat: ...`, `docs(api): ...`).
- PRs should include a concise summary, test results, and linked issues when applicable.
- UI changes should include screenshots or short clips.

## Configuration & Documentation Notes
- Frontend env setup uses `client/.env.example`.
- Architecture and workflows are documented in `CLAUDE.md` and `ADR/`.
