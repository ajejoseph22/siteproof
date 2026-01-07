# SiteProof Makefile
# Database and development commands

# =============================================================================
# CONFIGURATION
# =============================================================================

# Database connection (override with environment variables)
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= siteproof
DB_PASSWORD ?= siteproof
DB_NAME ?= siteproof
DB_SSLMODE ?= disable

DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
MIGRATIONS_PATH ?= migrations
GOCACHE_DIR ?= $(CURDIR)/.cache/go-build
GOMODCACHE_DIR ?= $(CURDIR)/.cache/go-mod
GOPATH_DIR ?= $(CURDIR)/.cache/go
GOLANGCI_LINT_CACHE_DIR ?= $(CURDIR)/.cache/golangci-lint

# =============================================================================
# DATABASE COMMANDS
# =============================================================================

.PHONY: db-migrate-up db-migrate-down db-migrate-version db-migrate-force db-migrate-create db-reset db-psqlainer

## db-migrate-up: Apply all pending migrations
db-migrate-up:
	@echo "Applying migrations..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up
	@echo "Migrations applied successfully"

## db-migrate-down: Rollback the last migration
db-migrate-down:
	@echo "Rolling back last migration..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down 1
	@echo "Rollback complete"

## db-migrate-down-all: Rollback all migrations (DANGEROUS)
db-migrate-down-all:
	@echo "WARNING: This will drop all tables!"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ]
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down -all

## db-migrate-version: Show current migration version
db-migrate-version:
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" version

## db-migrate-force: Force migration version (use with caution)
db-migrate-force:
ifndef VERSION
	$(error VERSION is required. Usage: make db-migrate-force VERSION=5)
endif
	@echo "Forcing migration version to $(VERSION)..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" force $(VERSION)

## db-migrate-create: Create a new migration file
db-migrate-create:
ifndef NAME
	$(error NAME is required. Usage: make db-migrate-create NAME=add_feature)
endif
	@echo "Creating migration: $(NAME)"
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq -digits 6 $(NAME)

## db-reset: Drop and recreate database, apply all migrations
db-reset:
	@echo "Resetting database..."
	psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -c "DROP DATABASE IF EXISTS $(DB_NAME);" postgres
	psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -c "CREATE DATABASE $(DB_NAME);" postgres
	$(MAKE) db-migrate-up
	@echo "Database reset complete"

## db-psql: Open psql shell
db-psql:
	psql "$(DATABASE_URL)"

## db-dump: Dump database to file
db-dump:
	@mkdir -p backups
	pg_dump "$(DATABASE_URL)" > backups/siteproof-$(shell date +%Y%m%d-%H%M%S).sql
	@echo "Database dumped to backups/"

## db-seed: Seed database with dev/test data (WARNING: clears existing data)
db-seed:
	@echo "WARNING: This will clear existing data!"
	@read -p "Continue? [y/N] " confirm && [ "$$confirm" = "y" ]
	psql "$(DATABASE_URL)" -f scripts/seed.sql
	@echo "Database seeded successfully"

## db-seed-docker: Seed database in Docker container
db-seed-docker:
	@echo "WARNING: This will clear existing data!"
	@read -p "Continue? [y/N] " confirm && [ "$$confirm" = "y" ]
	docker compose exec -T db psql -U siteproof -d siteproof < scripts/seed.sql
	@echo "Database seeded successfully"

# =============================================================================
# DOCKER COMMANDS
# =============================================================================

.PHONY: docker-up docker-down docker-logs docker-migrate

## docker-up: Start all services
docker-up:
	docker compose up -d

## docker-down: Stop all services
docker-down:
	docker compose down

## docker-logs: Tail logs for all services
docker-logs:
	docker compose logs -f

## docker-migrate: Run migrations in docker
docker-migrate:
	docker compose exec api migrate -path=/opt/siteproof/migrations -database="$(DATABASE_URL)" up

## docker-psql: Open psql shell in docker
docker-psql:
	docker compose exec db psql -U $(DB_USER) -d $(DB_NAME)

# =============================================================================
# DEVELOPMENT
# =============================================================================

.PHONY: dev test lint fmt

## dev: Start development server
dev:
	go run ./cmd/api

## test: Run all tests
test:
	@mkdir -p $(GOCACHE_DIR)
	GOCACHE=$(GOCACHE_DIR) go test -v ./...

## lint: Run linters
lint:
	@mkdir -p $(GOCACHE_DIR) $(GOMODCACHE_DIR) $(GOPATH_DIR) $(GOLANGCI_LINT_CACHE_DIR)
	GOCACHE=$(GOCACHE_DIR) GOMODCACHE=$(GOMODCACHE_DIR) GOPATH=$(GOPATH_DIR) GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE_DIR) golangci-lint run

## fmt: Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# =============================================================================
# BUILD
# =============================================================================

.PHONY: build build-docker

## build: Build the API binary
build:
	CGO_ENABLED=0 go build -o bin/siteproof-api ./cmd/api

## build-docker: Build Docker image
build-docker:
	docker build -t siteproof-api:latest .

# =============================================================================
# SQLC (Type-safe SQL)
# =============================================================================

.PHONY: sqlc-generate sqlc-verify

## sqlc-generate: Generate Go code from SQL
sqlc-generate:
	sqlc generate

## sqlc-verify: Verify SQL queries compile
sqlc-verify:
	sqlc compile

# =============================================================================
# HELP
# =============================================================================

.PHONY: help
help:
	@echo "SiteProof Development Commands"
	@echo ""
	@echo "Database:"
	@echo "  make db-migrate-up       Apply all pending migrations"
	@echo "  make db-migrate-down     Rollback the last migration"
	@echo "  make db-migrate-version  Show current migration version"
	@echo "  make db-migrate-create NAME=xyz  Create new migration"
	@echo "  make db-reset            Drop and recreate database"
	@echo "  make db-seed             Seed with dev data (dev/test)"
	@echo "  make db-seed-docker      Seed with dev data (Docker)"
	@echo "  make db-psql             Open psql shell"
	@echo "  make db-dump             Dump database to file"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up           Start all services"
	@echo "  make docker-down         Stop all services"
	@echo "  make docker-logs         Tail logs"
	@echo "  make docker-migrate      Run migrations in docker"
	@echo ""
	@echo "Development:"
	@echo "  make dev                 Start development server"
	@echo "  make test                Run tests"
	@echo "  make lint                Run linters"
	@echo "  make fmt                 Format code"
	@echo ""
	@echo "Build:"
	@echo "  make build               Build API binary"
	@echo "  make build-docker        Build Docker image"

.DEFAULT_GOAL := help
