# SiteProof Architecture Document

**Product**: Mobile-first T&M / Variation capture for construction subcontractors  
**Version**: 1.0  
**Date**: December 2024  
**Status**: Initial Architecture  

---

## Problem & Constraints

### Problem Statement
Subcontractors (electricians, plumbers, HVAC) need to capture, approve, and lock Time & Materials (T&M) and variation work on-site with photo evidence, voice-to-text descriptions, and GC superintendent signatures. The system must produce immutable PDF records emailed instantly to all parties.

### Hard Constraints
- **Team**: 1–2 engineers
- **Infrastructure**: Single Hetzner VPS initially (no K8s, no ECS)
- **Target**: MVP-first, but architected for 1,000+ companies
- **Philosophy**: Boring tech over clever tech

### Inferred Constraints
- **Users**: Construction workers on job sites, often with poor connectivity
- **Device mix**: iOS, Android, and web (foremen tablets, office PCs)
- **Legal weight**: PDFs may be used in disputes; immutability is non-negotiable
- **Offline tolerance**: Partial offline support is valuable but not MVP-critical

### Priority Order
1. Correctness (signatures, immutability, audit trail)
2. Security (access control, tamper-proofing)
3. Reliability (email delivery, data durability)
4. Maintainability (small team must move fast)
5. Performance (acceptable latency on mobile)
6. Cost (bootstrap budget)

---

## 1. Architecture Overview

### High-Level System Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              MOBILE CLIENTS                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                         │
│  │   iOS App   │  │ Android App │  │   Web App   │                         │
│  │  (React     │  │  (React     │  │  (React +   │                         │
│  │   Native)   │  │   Native)   │  │   Vite)     │                         │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘                         │
│         │                │                │                                 │
│         └────────────────┼────────────────┘                                 │
│                          │                                                  │
│                    HTTPS (REST API)                                         │
└──────────────────────────┼──────────────────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                         HETZNER VPS (Single Server)                          │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │                         Caddy (Reverse Proxy)                          │  │
│  │                    - TLS termination (auto Let's Encrypt)              │  │
│  │                    - Rate limiting                                     │  │
│  │                    - Static file serving                               │  │
│  └────────────────────────────────┬───────────────────────────────────────┘  │
│                                   │                                          │
│                                   ▼                                          │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │                   Go API Server (single binary)                        │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                  │  │
│  │  │   Handlers   │  │   Domain     │  │    Infra     │                  │  │
│  │  │  (HTTP/JSON) │  │   (Business  │  │  (DB, Email, │                  │  │
│  │  │              │→ │    Logic)    │→ │   Storage)   │                  │  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘                  │  │
│  └──────────────────────────────┬─────────────────────────────────────────┘  │
│                                 │                                            │
│           ┌─────────────────────┼─────────────────────┐                      │
│           │                     │                     │                      │
│           ▼                     ▼                     ▼                      │
│  ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐            │
│  │   PostgreSQL    │   │   Local File    │   │  Typst (PDF     │            │
│  │   (Database)    │   │   Storage       │   │   Generation)   │            │
│  │   - pgcrypto    │   │   /data/uploads │   │   CLI binary    │            │
│  └─────────────────┘   └─────────────────┘   └─────────────────┘            │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
                           │
       ┌───────────────────┼───────────────────┐
       │                   │                   │
       ▼                   ▼                   ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   Postmark   │  │   OpenAI     │  │   Twilio     │  │   Native     │
│   (Email)    │  │   (Voice     │  │   (SMS for   │  │   Device     │
│              │  │   Cleanup)   │  │   signing)   │  │   STT APIs   │
└──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘
```

### Key Component Responsibilities

| Component | Responsibility |
|-----------|---------------|
| **Caddy** | TLS termination, rate limiting, static assets, reverse proxy |
| **Go API Server** | All business logic, authentication, API endpoints, background jobs |
| **PostgreSQL** | Primary data store, audit logs, file metadata, email records |
| **Local File Storage** | Photos, signed PDFs (migrates to S3-compatible later) |
| **Typst** | PDF generation from templates (fast, reproducible) |
| **Postmark** | Transactional email delivery with delivery tracking |
| **OpenAI API** | Voice transcription cleanup (grammar, professionalism) |
| **Twilio** | SMS delivery for signature verification codes |
| **Device STT** | iOS/Android native speech-to-text (raw capture) |

### Why This Architecture Fits Construction Workflows

1. **Single binary deployment**: No container orchestration means faster deploys and simpler debugging. A foreman calling about a bug doesn't wait for pods to reschedule.

2. **Local file storage first**: Hetzner's included SSD storage is fast and cheap. Photos upload quickly. Migration to object storage is a configuration change, not a rewrite.

3. **Synchronous PDF generation**: Foremen need the PDF immediately to get signatures. Background queues add latency and complexity for the core happy path.

4. **Immutability at the database level**: PostgreSQL triggers + hash chains prevent tampering. This is simpler than blockchain but equally auditable.

5. **Voice cleanup via LLM**: Construction speech is informal. "Uh, we ran like 50 feet of extra conduit because the architect screwed up" becomes professional language suitable for invoices and disputes.

---

## 2. Tech Stack (with Justification)

### Frontend (Web): React + Vite

**Decision**: React 18 + Vite + React Router v6  
**Options considered**: Next.js, Vite + React, Remix  

**Why Vite over Next.js**:
- **Simpler**: No server/client component confusion—it's just a SPA
- **Faster**: Vite HMR is ~50ms vs Next.js ~200ms
- **No SSR needed**: This is a logged-in app, no SEO requirements
- **Better PWA support**: `vite-plugin-pwa` is more mature than `next-pwa`
- **Static output**: Just files served by Caddy—no Node.js runtime needed
- **No Vercel dependency**: Vercel Pro costs $20+/mo for commercial use

**Why NOT Next.js**:
- SSR/SSG solves problems this app doesn't have
- App Router adds complexity for a SPA
- Static export removes most Next.js benefits anyway
- Overkill for logged-in CRUD app with separate backend

**PWA support** (important for construction sites with poor connectivity):
```ts
// vite.config.ts
import { VitePWA } from 'vite-plugin-pwa'
export default {
  plugins: [VitePWA({ registerType: 'autoUpdate' })]
}
```

**Key libraries**:
| Purpose | Library |
|---------|---------|
| Routing | React Router v6 |
| Styling | Tailwind CSS |
| Forms | React Hook Form + Zod |
| Server state | TanStack Query (React Query) |
| Local state | Zustand |
| HTTP | Native fetch |
| Signature | react-signature-canvas |

**Build & deployment**: `vite build` → static files → Caddy serves from `/srv/web`

### Frontend (Mobile): React Native + Expo

**Decision**: React Native with Expo managed workflow  
**Options considered**: React Native, Flutter, native iOS/Android, PWA-only  

**Why React Native + Expo**:
- **Code sharing**: TypeScript across web and mobile; share types, validation, API client
- **Expo simplifies**: Managed builds, OTA updates, camera/permissions handled
- **Two apps, one codebase**: iOS + Android from single codebase
- **Offline-first**: AsyncStorage + React Query persistence for poor connectivity

**Why NOT Flutter**: Different language (Dart), no code sharing with web  
**Why NOT PWA-only**: Limited camera access, no app store presence (matters for B2B credibility)

**Key libraries**:
| Purpose | Library |
|---------|---------|
| Navigation | React Navigation |
| Camera | expo-camera, expo-image-picker |
| Voice-to-text | react-native-voice (native STT) |
| Signature | react-native-signature-canvas |
| Offline storage | AsyncStorage + React Query persistence |

**Shared code structure**:
```
/packages
  /shared        # Types, API client, Zod schemas, utils
  /web           # Vite + React app
  /mobile        # Expo + React Native app
```

### Backend: Go 1.22+

**Decision**: Go  
**Options considered**: Node.js/TypeScript, Python/FastAPI, Go, Rust  
**Why chosen**:
- Single static binary simplifies deployment
- Excellent standard library (HTTP, JSON, crypto)
- Low memory footprint for VPS
- Strong typing catches bugs at compile time
- goroutines handle concurrent requests without callback hell
- 1–2 engineers can reason about the entire codebase

**Consequences**: 
- No ORM; use `sqlc` for type-safe SQL
- Smaller ecosystem than Node/Python
- Team must know Go (straightforward to learn)

**How to revisit**: If hiring becomes difficult, consider Node.js/TypeScript migration. Domain logic in Go is framework-free and conceptually portable.

### API Style: REST + JSON

**Decision**: REST with JSON  
**Options considered**: GraphQL, gRPC, REST  
**Why chosen**:
- Simpler tooling (curl, Postman, browser devtools)
- Mobile clients don't need GraphQL's flexibility
- Caching is straightforward (HTTP semantics)
- Easier to debug on construction sites ("what did the request look like?")

**Consequences**:
- May need multiple endpoints for complex screens (acceptable)
- No automatic type generation (mitigated by OpenAPI spec)

**API conventions**:
- Versioned: `/api/v1/...`
- Resource-oriented: `/api/v1/tickets`, `/api/v1/tickets/{id}/photos`
- Consistent error format: `{"error": {"code": "TICKET_LOCKED", "message": "..."}}`

**Pagination** (cursor-based for lists):
```
GET /api/v1/tickets?limit=20&cursor=abc123&status=approved&sort=-created_at

Response:
{
  "data": [...],
  "pagination": {
    "limit": 20,
    "next_cursor": "def456",   // null if no more pages
    "has_more": true
  }
}
```

**Idempotency** (for unsafe operations):
```
POST /api/v1/tickets
Headers:
  Idempotency-Key: client-generated-uuid

- Server stores key → response mapping for 24 hours
- Duplicate requests return cached response
- Prevents double-submission on flaky mobile connections
```

**API contract examples**:

```
# Create Ticket
POST /api/v1/tickets
Request:
{
  "project_id": "uuid",
  "title": "Extra conduit run - Building B",
  "work_date": "2024-12-29",
  "ticket_type": "time_materials"
}
Response (201):
{
  "id": "uuid",
  "ticket_number": "TM-2024-0042",
  "status": "draft",
  "created_at": "2024-12-29T10:30:00Z"
}

# Upload Photo
POST /api/v1/tickets/{id}/photos
Content-Type: multipart/form-data
Fields: file, caption, taken_at, latitude, longitude, client_hash
Response (201):
{
  "id": "uuid",
  "file_path": "/files/photos/...",
  "thumbnail_url": "/files/thumbs/..."
}

# Request Signature
POST /api/v1/tickets/{id}/signature-request
Request:
{
  "signer_email": "gcsuper@contractor.com",
  "signer_phone": "+1234567890"
}
Response (201):
{
  "signing_link_id": "uuid",
  "expires_at": "2024-12-30T10:30:00Z",
  "sent_via": ["sms", "email"]
}

# Error Responses
{
  "error": {
    "code": "TICKET_LOCKED",
    "message": "Cannot modify a locked ticket",
    "details": {
      "ticket_id": "uuid",
      "locked_at": "2024-12-29T10:30:00Z"
    }
  }
}

Error codes:
- VALIDATION_ERROR (400) - Invalid input
- UNAUTHORIZED (401) - Missing or invalid token
- FORBIDDEN (403) - No permission for this resource
- NOT_FOUND (404) - Resource doesn't exist
- CONFLICT (409) - State conflict (e.g., already signed)
- TICKET_LOCKED (422) - Cannot modify locked ticket
- RATE_LIMITED (429) - Too many requests
```

### Database: PostgreSQL 16

**Decision**: PostgreSQL  
**Options considered**: PostgreSQL, SQLite, MySQL  
**Why chosen**:
- JSONB for flexible metadata without schema changes
- `pgcrypto` for hashing/signatures
- Row-level security for multi-tenancy
- Mature, boring, battle-tested
- Excellent tooling (pg_dump, logical replication for future scaling)

**Schema style**: Normalized with strategic denormalization
- Core entities normalized (users, tickets, projects)
- Audit logs append-only with hash chains
- File metadata in DB; files on disk (paths stored in DB)

**Migrations**: `golang-migrate/migrate`
- SQL-based migrations (not ORM-generated)
- Forward and backward scripts required
- Applied via CI/CD with verification

### File Storage: Local Filesystem → S3-Compatible

**Decision**: Local filesystem initially  
**Options considered**: S3 from day one, local only, database BLOBs  
**Why chosen**:
- Hetzner includes 80GB+ NVMe; sufficient for MVP
- Simpler debugging (files are just files)
- Clear migration path to S3-compatible (Hetzner Object Storage, Backblaze B2)

**Directory structure**:
```
/data/uploads/
  /{company_id}/
    /photos/{ticket_id}/{uuid}.jpg
    /pdfs/{ticket_id}/{uuid}.pdf
    /signatures/{ticket_id}/{uuid}.png
```

**Migration trigger**: When storage exceeds 50GB or when multi-region is needed.

### PDF Generation: Typst

**Decision**: Typst  
**Options considered**: wkhtmltopdf, Puppeteer, WeasyPrint, LaTeX, Typst  
**Why chosen**:
- Extremely fast (100ms for typical document)
- Reproducible (same input = same output, byte-for-byte)
- Single binary, no browser dependency
- Clean template syntax
- Active development, modern tooling

**Consequences**:
- Less CSS-like than HTML-based tools
- Team learns Typst syntax (small learning curve)

**Template structure**:
```
/templates/
  /ticket.typ         # Main T&M ticket template
  /partials/
    /header.typ       # Company logo, project info
    /line_items.typ   # Labor/materials table
    /signatures.typ   # Signature blocks
    /footer.typ       # Legal text, timestamps
```

### Email Delivery: Postmark

**Decision**: Postmark  
**Options considered**: SendGrid, Amazon SES, Postmark, self-hosted  
**Why chosen**:
- Best-in-class deliverability (construction companies have strict spam filters)
- Excellent delivery tracking API
- Webhook support for bounce/delivery events
- Simple pricing, no complex tiers
- Fast integration (hours, not days)

**Consequences**:
- Vendor dependency (acceptable; email providers are commoditized)
- Cost scales with volume (~$1.25 per 1000 emails)

**Webhook integration**: Postmark POSTs delivery status to `/api/v1/webhooks/postmark`, updating `email_events` table.

### Authentication: Manual JWT + Microsoft OAuth

**Decision**: Manual email/password auth with JWT tokens + optional Microsoft OAuth  
**Options considered**: Firebase Auth, Auth0, Clerk, WorkOS, manual implementation  

**Why manual over managed auth**:
- **B2B construction reality**: Social login (Google, Apple) rarely used in this industry
- **Microsoft dominates**: Construction companies use Microsoft 365—"Sign in with Microsoft" is the only social login that makes sense
- **Full control**: No external dependency during auth flow
- **Simple to build**: Email/password + JWT is 2-3 days of work
- **Cost**: $0 forever

**Why NOT Firebase/Clerk**:
- Social login (Google, Apple) won't be used—construction workers use company email
- Phone auth can be done with Twilio directly
- Adds external dependency for minimal benefit

**Auth methods supported**:

| Method | Implementation | When used |
|--------|---------------|-----------|
| Email/password | Manual (bcrypt + JWT) | Default for all users |
| Sign in with Microsoft | MSAL library | Optional, for O365 users |
| Enterprise SSO (Azure AD/Okta) | WorkOS (Phase 2) | When enterprise customers require |

**Password handling**:
```go
// Registration: hash with bcrypt
hash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

// Login: verify
err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
```

**Token structure**:
- Access token: 15-minute expiry, contains `user_id`, `company_id`, `role`
- Refresh token: 30-day expiry, stored in DB (revocable), rotated on use

**Microsoft OAuth flow** (optional):
1. User clicks "Sign in with Microsoft"
2. Redirect to Microsoft login (MSAL library)
3. Microsoft returns authorization code
4. Backend exchanges code for user info
5. Create/match user in PostgreSQL by email
6. Issue JWT tokens (same as email/password flow)

**SMS for signature verification**: Use Twilio directly (see signing flow). Keeps auth simple.

**Enterprise SSO (Phase 2)**: When large GC requires Azure AD SSO, add WorkOS alongside manual auth. WorkOS handles SSO; manual auth continues for regular users.

### Voice Transcription Cleanup: OpenAI GPT-4o-mini

**Decision**: OpenAI API for transcription cleanup  
**Options considered**: Claude, GPT-4, GPT-4o-mini, local LLM (Llama 3, Mistral, Phi-3), rule-based  

**Why NOT local LLM**:
- **Hardware constraint**: Hetzner VPS has no GPU and only 8GB RAM. CPU inference on Llama 3 8B takes 30-60 seconds—unacceptable for on-site foremen.
- **Cost at MVP scale**: At 2,000 cleanups/day, OpenAI costs ~$0.16/day. A GPU server (Hetzner GEX44) costs ~€150/month. Break-even requires massive scale.
- **Operational overhead**: Running Ollama/vLLM adds another service for a 1-2 person team to manage.
- **Quality gap**: GPT-4o-mini excels at grammar correction. Smaller local models require fine-tuning to match.

**Why NOT rule-based**:
- Construction speech is highly variable ("we ran like 50 feet of extra conduit")
- Rules can't handle context-dependent grammar correction
- Maintenance burden grows with edge cases

**Why chosen (GPT-4o-mini)**:
- Best cost/quality ratio for grammar and professional rewriting
- Fast enough for synchronous use (typically <2s)
- No infrastructure to manage
- Excellent at preserving factual content while cleaning language

**Consequences**:
- External dependency (mitigated by graceful degradation to raw text)
- Cost per transcription (~$0.00008 at current rates)
- Must preserve raw transcript for audit
- Vendor lock-in (acceptable; easy to swap LLM providers)

**Revisit trigger**: When volume exceeds 50,000 cleanups/month OR enterprise customers require on-prem processing, evaluate:
- Dedicated GPU server with Llama 3 70B or successor
- Fine-tuned smaller model (Phi-4, Mistral) for this specific task
- Hybrid: local for bulk, OpenAI for edge cases

---

## 3. Infrastructure & Deployment

### Hetzner VPS Layout

**Server spec (initial)**:
- CPX31 (4 vCPU AMD, 8GB RAM, 160GB NVMe)
- ~€15/month
- Location: Ashburn, VA (US East) for US customers, or Falkenstein for EU

**Why this size**:
- 4 vCPU handles concurrent PDF generation
- 8GB RAM comfortable for Go + PostgreSQL
- 160GB storage handles 50,000+ photos before needing object storage

**Upgrade path**: CPX41 (8 vCPU, 16GB) or CPX51 (16 vCPU, 32GB) are in-place upgrades.

### Operating System

**Decision**: Ubuntu 24.04 LTS  
**Why**: 
- 5-year support window
- Excellent PostgreSQL packaging
- Team familiarity
- Large troubleshooting knowledge base

### Directory Layout

```
/opt/siteproof/
  /bin/
    siteproof-api          # Main binary
    typst                  # PDF generator
  /config/
    config.yaml            # Non-secret configuration
  /templates/
    *.typ                  # Typst templates
  /migrations/
    *.up.sql, *.down.sql   # Database migrations

/data/
  /uploads/                # User-uploaded files
  /backups/                # Local backup staging

/var/log/siteproof/
  api.log                  # Application logs (JSON)
```

### Reverse Proxy: Caddy 2

**Decision**: Caddy  
**Options considered**: Nginx, Traefik, Caddy  
**Why chosen**:
- Automatic TLS (Let's Encrypt) with zero configuration
- Simple config syntax
- Built-in rate limiting
- Excellent defaults (HTTP/2, OCSP stapling)

**Caddyfile**:
```caddyfile
siteproof.app {
    # Rate limiting
    rate_limit {
        zone api {
            key {remote_host}
            events 100
            window 1m
        }
    }

    # API proxy
    handle /api/* {
        reverse_proxy localhost:8080
    }

    # Static files (web app)
    handle {
        root * /opt/siteproof/web
        try_files {path} /index.html
        file_server
    }

    # Uploaded files (serve directly)
    handle /files/* {
        root * /data/uploads
        file_server
    }

    log {
        output file /var/log/caddy/access.log
        format json
    }
}
```

### Containerization: Docker + docker-compose

**Decision**: Docker containers managed by docker-compose  
**Options considered**: Bare metal + systemd, Docker, Kubernetes  

**Why Docker over bare systemd**:
- **Easy replication**: `docker compose up -d` on any VPS—identical environment
- **Simpler scaling**: Adding VPS #2 or #3 is trivial
- **Consistent environments**: Same containers in dev, CI, and prod
- **Dependency bundling**: Go, Typst, Caddy all containerized
- **Easy rollback**: Pull previous image tag

**Why NOT Kubernetes**:
- Overkill for 1-3 VPSes
- Adds significant operational complexity
- Docker-compose is sufficient until 5+ services or auto-scaling needed

**docker-compose.yml**:
```yaml
version: "3.8"

services:
  api:
    image: ghcr.io/yourcompany/siteproof-api:${VERSION:-latest}
    restart: unless-stopped
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - POSTMARK_API_KEY=${POSTMARK_API_KEY}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - TWILIO_ACCOUNT_SID=${TWILIO_ACCOUNT_SID}
      - TWILIO_AUTH_TOKEN=${TWILIO_AUTH_TOKEN}
      - JWT_SECRET=${JWT_SECRET}
    volumes:
      - uploads:/data/uploads
      - ./templates:/opt/siteproof/templates:ro
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3

  caddy:
    image: caddy:2-alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - ./web:/srv/web:ro           # Static frontend (vite build output)
      - uploads:/srv/uploads:ro      # Uploaded files
      - caddy_data:/data
      - caddy_config:/config

  # PostgreSQL - only on primary VPS, or use managed DB
  db:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      - POSTGRES_DB=siteproof
      - POSTGRES_USER=siteproof
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
    # Not exposed externally - only accessible to api container

volumes:
  uploads:
  caddy_data:
  caddy_config:
  pgdata:
```

**Dockerfile** (Go API):
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o siteproof-api ./cmd/api

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata curl
# Install Typst for PDF generation
RUN apk add --no-cache --repository=https://dl-cdn.alpinelinux.org/alpine/edge/testing typst
WORKDIR /opt/siteproof
COPY --from=builder /app/siteproof-api ./bin/
COPY templates ./templates/
EXPOSE 8080
CMD ["./bin/siteproof-api"]
```

**Deployment to VPS**:
```bash
# On any VPS (initial or scaled)
git clone <repo> /opt/siteproof  # Or just docker-compose.yml + .env
cd /opt/siteproof
cp .env.example .env             # Fill in secrets
docker compose pull
docker compose up -d
```

### Secrets Management

**Decision**: `.env` file for Docker, never committed to git  
**Options considered**: HashiCorp Vault, AWS Secrets Manager, env files  
**Why chosen**:
- Simplest approach for Docker deployments
- No external dependency
- `.env` stays on server, never in repo

**`.env` file** (on each VPS):
```bash
# Database
DATABASE_URL=postgres://siteproof:xxx@db:5432/siteproof?sslmode=disable
DB_PASSWORD=xxx

# External services
POSTMARK_API_KEY=xxx
OPENAI_API_KEY=xxx
TWILIO_ACCOUNT_SID=xxx
TWILIO_AUTH_TOKEN=xxx

# Auth
JWT_SECRET=xxx
MICROSOFT_CLIENT_ID=xxx
MICROSOFT_CLIENT_SECRET=xxx

# Versioning
VERSION=1.2.3
```

**Security**:
- `.env` file: `chmod 600`, owned by root
- `.env` in `.gitignore` (never committed)
- Use `.env.example` with placeholder values for documentation

**Phase 2**: Migrate to Infisical, Doppler, or 1Password Secrets Automation when team grows.

### Backups

**Database backups**:
```bash
# Daily at 3 AM UTC
0 3 * * * pg_dump siteproof | gzip > /data/backups/db/siteproof-$(date +\%Y\%m\%d).sql.gz

# Retain 30 days locally
find /data/backups/db -mtime +30 -delete

# Sync to Hetzner Storage Box (off-server)
rsync -az /data/backups/ u123456@u123456.your-storagebox.de:backups/
```

**File backups**:
```bash
# Incremental rsync to Storage Box
0 4 * * * rsync -az /data/uploads/ u123456@u123456.your-storagebox.de:uploads/
```

**Hetzner Storage Box**: ~€3/month for 100GB, accessible via SFTP/rsync.

**Recovery testing**: Monthly restore to a test VPS (documented in runbook).

### CI/CD

**Decision**: GitHub Actions → Docker build → Push to registry → Pull on VPS  
**Why**: Simple, Docker-native, easy to scale to multiple VPSes.

**Pipeline**:
```yaml
name: Build and Deploy
on:
  push:
    branches: [main]
    tags: ['v*']

env:
  REGISTRY: ghcr.io
  API_IMAGE: ghcr.io/${{ github.repository }}/api
  WEB_IMAGE: ghcr.io/${{ github.repository }}/web

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test ./...

  build-api:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: |
            ${{ env.API_IMAGE }}:latest
            ${{ env.API_IMAGE }}:${{ github.sha }}

  build-web:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: cd packages/web && npm ci && npm run build
      - uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          context: ./packages/web
          push: true
          tags: ${{ env.WEB_IMAGE }}:latest

  deploy:
    needs: [build-api, build-web]
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to VPS(es)
        env:
          SSH_KEY: ${{ secrets.SSH_KEY }}
          VPS_HOSTS: ${{ secrets.VPS_HOSTS }}  # "vps1.example.com,vps2.example.com"
        run: |
          echo "$SSH_KEY" > key && chmod 600 key
          for host in ${VPS_HOSTS//,/ }; do
            ssh -i key deploy@$host "cd /opt/siteproof && docker compose pull && docker compose up -d"
          done
```

**Scaling deployment**: Add VPS hostnames to `VPS_HOSTS` secret. Same deploy command runs on all.

**Deployment strategy**: Rolling update via docker-compose (containers restart one by one). Acceptable for MVP given:
- Go binary starts in <1 second
- Stateless server (requests in flight may fail, clients retry)

**Phase 2**: Add health check endpoint and rolling restart script.

---

## 4. Data Model (Conceptual)

### Entity Relationship Overview

```
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│  companies  │──────<│    users    │       │   projects  │
└─────────────┘       └─────────────┘       └──────┬──────┘
                             │                      │
                             │                      │
                      ┌──────┴──────┐               │
                      │             │               │
                      ▼             ▼               ▼
               ┌─────────────┐  ┌─────────────────────────┐
               │   tickets   │──│   ticket_line_items     │
               └──────┬──────┘  └─────────────────────────┘
                      │
        ┌─────────────┼─────────────┬─────────────┐
        │             │             │             │
        ▼             ▼             ▼             ▼
┌─────────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐
│   photos    │ │ signatures│ │audit_logs │ │email_events│
└─────────────┘ └───────────┘ └───────────┘ └───────────┘
```

### Core Tables

#### `companies`
```sql
CREATE TABLE companies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    slug            TEXT UNIQUE NOT NULL,  -- URL-friendly identifier
    settings        JSONB DEFAULT '{}',    -- Company-specific config
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);
```

#### `users`
```sql
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      UUID NOT NULL REFERENCES companies(id),
    email           TEXT UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    name            TEXT NOT NULL,
    role            TEXT NOT NULL CHECK (role IN (
                        'foreman',      -- Creates tickets, captures photos
                        'gc_super',     -- Signs tickets on behalf of GC
                        'pm',           -- Project manager, reviews tickets
                        'accounting',   -- Views approved tickets for billing
                        'admin'         -- Company admin, manages users
                    )),
    phone           TEXT,                  -- For SMS notifications (future)
    active          BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_users_company ON users(company_id);
CREATE INDEX idx_users_email ON users(email);
```

#### `projects`

**Who creates projects**: Admin or PM creates projects **before field work begins** via web interface. This ensures GC contact info, site address, and contract details are configured before foremen create tickets. Foremen select from a dropdown of active projects when creating tickets.

**Edge case**: If foreman is on a new job not yet in system, they call the office to add it (ensures data quality). Phase 2 may add "quick project" creation for foremen.

```sql
CREATE TABLE projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      UUID NOT NULL REFERENCES companies(id),
    name            TEXT NOT NULL,
    address         TEXT,
    gc_company_name TEXT,                  -- General contractor name
    gc_contact_email TEXT,                 -- GC superintendent email (required for ticket approval emails)
    status          TEXT DEFAULT 'active' CHECK (status IN ('active', 'completed', 'archived')),
    timezone        TEXT DEFAULT 'America/New_York', -- For PDF timestamp display
    settings        JSONB DEFAULT '{}',    -- Project-specific config
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_projects_company ON projects(company_id);
```

#### `tickets` (T&M / Variation records)
```sql
CREATE TABLE tickets (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id          UUID NOT NULL REFERENCES companies(id),
    project_id          UUID NOT NULL REFERENCES projects(id),
    created_by_id       UUID NOT NULL REFERENCES users(id),
    
    -- Ticket identification
    ticket_number       TEXT NOT NULL,         -- Human-readable: "TM-2024-0042"
    ticket_type         TEXT NOT NULL CHECK (ticket_type IN ('time_materials', 'variation')),
    
    -- Content
    title               TEXT NOT NULL,
    description_raw     TEXT,                  -- Original voice transcription (preserved)
    description_cleaned TEXT,                  -- LLM-cleaned version (displayed)
    description_final   TEXT,                  -- User-edited final version
    
    -- Work details
    work_date           DATE NOT NULL,
    
    -- Status workflow
    status              TEXT DEFAULT 'draft' CHECK (status IN (
                            'draft',           -- Being created
                            'pending_review',  -- Waiting for GC signature
                            'approved',        -- GC signed
                            'rejected',        -- GC rejected
                            'locked'           -- PDF generated, immutable
                        )),
    
    -- Signature tracking
    gc_signed_at        TIMESTAMPTZ,
    gc_signed_by_id     UUID REFERENCES users(id),
    
    -- Rejection tracking (if GC refuses to sign)
    rejected_at         TIMESTAMPTZ,
    rejected_by_id      UUID REFERENCES users(id),
    rejection_reason    TEXT,                  -- Required when rejecting
    
    -- PDF record
    pdf_generated_at    TIMESTAMPTZ,
    pdf_path            TEXT,                  -- Path to locked PDF
    pdf_hash            TEXT,                  -- SHA-256 of PDF file
    
    -- Immutability
    locked_at           TIMESTAMPTZ,
    content_hash        TEXT,                  -- Hash of all content at lock time
    
    -- Metadata
    created_at          TIMESTAMPTZ DEFAULT now(),
    updated_at          TIMESTAMPTZ DEFAULT now(),
    
    UNIQUE(company_id, ticket_number)
);

CREATE INDEX idx_tickets_company ON tickets(company_id);
CREATE INDEX idx_tickets_project ON tickets(project_id);
CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_work_date ON tickets(work_date);
```

#### `ticket_line_items`

**Important**: Line items capture **only the extra/additional work** outside the original contract scope. This is the core purpose of T&M tickets. The original contract work is tracked in the subcontractor's accounting/job costing system, not here.

**Example**: Original scope was 200 ft of conduit. Due to a design change, 250 ft was installed. The ticket captures only the **50 ft extra**:
| Type | Description | Qty | Unit |
|------|-------------|-----|------|
| Labor | Electrician - overtime | 4 | hours |
| Material | 3/4" EMT conduit | 50 | linear ft |
| Equipment | Scissor lift rental | 0.5 | day |

```sql
CREATE TABLE ticket_line_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id       UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    
    item_type       TEXT NOT NULL CHECK (item_type IN ('labor', 'material', 'equipment')),
    description     TEXT NOT NULL,
    quantity        DECIMAL(10, 2) NOT NULL,
    unit            TEXT NOT NULL,             -- 'hours', 'each', 'linear_ft', etc.
    unit_rate       DECIMAL(10, 2),            -- Optional: price per unit
    total           DECIMAL(10, 2),            -- quantity * unit_rate (computed or manual)
    
    sort_order      INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_line_items_ticket ON ticket_line_items(ticket_id);
```

#### `photos`
```sql
CREATE TABLE photos (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id       UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    uploaded_by_id  UUID NOT NULL REFERENCES users(id),
    
    file_path       TEXT NOT NULL,             -- Relative path in storage
    file_size       INTEGER NOT NULL,          -- Bytes
    file_hash       TEXT NOT NULL,             -- SHA-256 for integrity
    mime_type       TEXT NOT NULL,
    
    caption         TEXT,
    taken_at        TIMESTAMPTZ,               -- EXIF timestamp if available
    latitude        DECIMAL(9, 6),             -- GPS from EXIF
    longitude       DECIMAL(9, 6),
    
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_photos_ticket ON photos(ticket_id);
```

#### `signing_links` (Secure signature request tokens)

Used for link-based signing where GC signs on their own device, preventing forgery.

```sql
CREATE TABLE signing_links (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id       UUID NOT NULL REFERENCES tickets(id),
    
    -- Token for URL: /sign/{token}
    token           TEXT UNIQUE NOT NULL,      -- Cryptographically random, 32+ chars
    
    -- Who should sign
    signer_type     TEXT NOT NULL CHECK (signer_type IN ('gc_super')),
    signer_email    TEXT NOT NULL,             -- Where link was sent
    signer_phone    TEXT,                      -- For SMS delivery
    
    -- Delivery tracking
    sent_via        TEXT[] DEFAULT '{}',       -- ['email', 'sms']
    sent_at         TIMESTAMPTZ DEFAULT now(),
    
    -- Expiration and usage
    expires_at      TIMESTAMPTZ NOT NULL,      -- Default: 24 hours from creation
    used_at         TIMESTAMPTZ,               -- When signature was captured
    revoked_at      TIMESTAMPTZ,               -- If foreman cancels request
    
    -- SMS verification (for same-device fallback)
    verification_code TEXT,                    -- 6-digit code for same-device signing
    verification_attempts INTEGER DEFAULT 0,   -- Max 3 attempts
    
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_signing_links_token ON signing_links(token);
CREATE INDEX idx_signing_links_ticket ON signing_links(ticket_id);
```

#### `signatures`
```sql
CREATE TABLE signatures (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id       UUID NOT NULL REFERENCES tickets(id),
    signing_link_id UUID REFERENCES signing_links(id), -- Which link was used
    signer_id       UUID REFERENCES users(id), -- NULL if external signer
    
    signer_type     TEXT NOT NULL CHECK (signer_type IN ('foreman', 'gc_super')),
    signer_name     TEXT NOT NULL,             -- Name as signed
    signer_email    TEXT,                      -- Email at time of signing
    
    signature_image TEXT NOT NULL,             -- Path to signature PNG
    signature_hash  TEXT NOT NULL,             -- SHA-256 of image
    
    -- Signing method for audit/dispute resolution
    signing_method  TEXT NOT NULL CHECK (signing_method IN (
                        'link_own_device',     -- GC opened link on their device (strongest)
                        'link_same_device',    -- GC used link on foreman's device + SMS code
                        'same_device_verified' -- Same device with SMS verification (fallback)
                    )),
    
    -- Device fingerprinting for fraud detection
    ip_address      INET,
    user_agent      TEXT,
    device_id       TEXT,                      -- Client-generated device fingerprint
    
    signed_at       TIMESTAMPTZ DEFAULT now(),
    
    UNIQUE(ticket_id, signer_type)
);

CREATE INDEX idx_signatures_ticket ON signatures(ticket_id);
```

#### `audit_logs` (Append-only, hash-chained)
```sql
CREATE TABLE audit_logs (
    id              BIGSERIAL PRIMARY KEY,
    
    -- What changed
    entity_type     TEXT NOT NULL,             -- 'ticket', 'photo', 'signature', etc.
    entity_id       UUID NOT NULL,
    action          TEXT NOT NULL,             -- 'created', 'updated', 'signed', 'locked'
    
    -- Who and when
    actor_id        UUID REFERENCES users(id),
    actor_ip        INET,
    occurred_at     TIMESTAMPTZ DEFAULT now(),
    
    -- Change details
    changes         JSONB,                     -- {field: {old: x, new: y}}
    
    -- Hash chain for tamper evidence
    prev_hash       TEXT,                      -- Hash of previous log entry
    entry_hash      TEXT NOT NULL,             -- Hash of this entry
    
    -- Immutability enforced
    CONSTRAINT no_updates CHECK (true)         -- Symbolic; enforced by trigger
);

CREATE INDEX idx_audit_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_actor ON audit_logs(actor_id);
CREATE INDEX idx_audit_time ON audit_logs(occurred_at);

-- Trigger to compute hash chain
CREATE OR REPLACE FUNCTION compute_audit_hash() RETURNS TRIGGER AS $$
BEGIN
    SELECT entry_hash INTO NEW.prev_hash 
    FROM audit_logs 
    ORDER BY id DESC LIMIT 1;
    
    NEW.entry_hash := encode(
        sha256(
            (NEW.prev_hash || NEW.entity_type || NEW.entity_id || 
             NEW.action || NEW.occurred_at || NEW.changes::text)::bytea
        ), 
        'hex'
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_hash_chain
BEFORE INSERT ON audit_logs
FOR EACH ROW EXECUTE FUNCTION compute_audit_hash();

-- Prevent updates/deletes
CREATE OR REPLACE FUNCTION prevent_audit_modification() RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Audit logs cannot be modified or deleted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_immutable
BEFORE UPDATE OR DELETE ON audit_logs
FOR EACH ROW EXECUTE FUNCTION prevent_audit_modification();
```

#### `email_events`
```sql
CREATE TABLE email_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id       UUID REFERENCES tickets(id),
    
    -- Postmark tracking
    message_id      TEXT UNIQUE NOT NULL,      -- Postmark MessageID
    
    -- Recipients
    to_email        TEXT NOT NULL,
    to_name         TEXT,
    
    -- Content reference
    subject         TEXT NOT NULL,
    template        TEXT,                      -- 'ticket_approved', 'ticket_rejected'
    
    -- Status tracking
    status          TEXT DEFAULT 'sent' CHECK (status IN (
                        'sent', 'delivered', 'opened', 'bounced', 'spam_complaint'
                    )),
    
    -- Timestamps
    sent_at         TIMESTAMPTZ DEFAULT now(),
    delivered_at    TIMESTAMPTZ,
    opened_at       TIMESTAMPTZ,
    bounced_at      TIMESTAMPTZ,
    
    -- Error details
    error_code      TEXT,
    error_message   TEXT
);

CREATE INDEX idx_email_ticket ON email_events(ticket_id);
CREATE INDEX idx_email_message ON email_events(message_id);
```

#### `transcription_versions` (Voice cleanup audit trail)
```sql
CREATE TABLE transcription_versions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id       UUID NOT NULL REFERENCES tickets(id),
    
    version         INTEGER NOT NULL,          -- 1 = raw, 2 = cleaned, 3+ = edits
    content         TEXT NOT NULL,
    source          TEXT NOT NULL CHECK (source IN (
                        'voice_raw',           -- Original STT output
                        'llm_cleaned',         -- After GPT cleanup
                        'user_edit'            -- Manual edit by user
                    )),
    
    -- LLM metadata (if applicable)
    llm_model       TEXT,                      -- 'gpt-4o-mini'
    llm_prompt_hash TEXT,                      -- Hash of system prompt used
    
    created_by_id   UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT now(),
    
    UNIQUE(ticket_id, version)
);

CREATE INDEX idx_transcription_ticket ON transcription_versions(ticket_id);
```

#### `refresh_tokens`
```sql
CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    token_hash      TEXT NOT NULL,             -- SHA-256 of token
    expires_at      TIMESTAMPTZ NOT NULL,
    
    -- Device info for session management
    device_name     TEXT,
    last_used_at    TIMESTAMPTZ DEFAULT now(),
    
    created_at      TIMESTAMPTZ DEFAULT now(),
    revoked_at      TIMESTAMPTZ
);

CREATE INDEX idx_refresh_user ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_token ON refresh_tokens(token_hash);
```

---

## 5. Critical Flows (Step-by-Step)

### Flow 1: Create T&M Ticket

```
┌─────────────────────────────────────────────────────────────────┐
│                    Foreman on Job Site                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 1. Open app → Select project → Tap "New T&M Ticket"             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. App shows form:                                              │
│    - Work date (default: today)                                 │
│    - Title (brief description)                                  │
│    - Description (voice or text)                                │
│    - Line items (labor, materials, equipment)                   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. API: POST /api/v1/tickets                                    │
│    Request: { project_id, title, work_date }                    │
│    Response: { id, ticket_number: "TM-2024-0042", status: draft}│
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. Server actions:                                              │
│    - Generate sequential ticket_number per company              │
│    - Create ticket record (status: draft)                       │
│    - Create audit_log entry (action: created)                   │
└─────────────────────────────────────────────────────────────────┘
```

### Flow 2: Capture Photo

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Foreman taps camera icon on ticket screen                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. Native camera opens → Takes photo → Returns to app           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. Client-side processing:                                      │
│    - Extract EXIF (timestamp, GPS)                              │
│    - Compress to max 2048px, JPEG 80%                           │
│    - Compute SHA-256 hash                                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. API: POST /api/v1/tickets/{id}/photos                        │
│    Content-Type: multipart/form-data                            │
│    Fields: file, caption, taken_at, latitude, longitude, hash   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. Server actions:                                              │
│    - Verify ticket exists, status=draft, user has access        │
│    - Verify hash matches uploaded file                          │
│    - Save to /data/uploads/{company}/{ticket_id}/{uuid}.jpg     │
│    - Create photo record with file_path, file_hash              │
│    - Create audit_log entry (action: photo_added)               │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 6. Response: { id, file_path, thumbnail_url }                   │
└─────────────────────────────────────────────────────────────────┘
```

### Flow 3: Voice-to-Text Description

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Foreman taps microphone icon                                 │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. Native STT activates (iOS: SFSpeechRecognizer,               │
│    Android: SpeechRecognizer)                                   │
│    - Runs on-device, no network required                        │
│    - Returns raw text to app                                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. Foreman reviews raw text, taps "Clean Up"                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. API: POST /api/v1/tickets/{id}/description/cleanup           │
│    Request: { raw_text: "uh we ran like 50 feet of extra..." }  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. Server processing (SYNCHRONOUS):                             │
│                                                                 │
│    a) Save raw transcript immediately:                          │
│       - Update ticket.description_raw = raw_text                │
│       - Create transcription_version (version=1, source=raw)    │
│                                                                 │
│    b) Call OpenAI GPT-4o-mini:                                  │
│       System prompt:                                            │
│       "You are a construction documentation assistant.          │
│        Clean up this voice transcription:                       │
│        - Remove filler words (uh, um, like, you know)           │
│        - Fix grammar and sentence structure                     │
│        - Make it professional and clear                         │
│        - Preserve all factual details exactly                   │
│        - Do not add information or legal language               │
│        - Output should be suitable for invoices"                │
│                                                                 │
│    c) Save cleaned version:                                     │
│       - Update ticket.description_cleaned = llm_output          │
│       - Create transcription_version (version=2, source=cleaned)│
│       - Create audit_log (action: description_cleaned)          │
│                                                                 │
│    d) Return cleaned text to client                             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 6. Response: { cleaned_text: "Ran 50 linear feet of..." }       │
│    - Latency target: <3 seconds                                 │
│    - If OpenAI fails: return raw text, log error, allow retry   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 7. Foreman can edit cleaned text before finalizing              │
│    - Edits are stored LOCALLY in app state (not auto-saved)     │
│    - Visual indicator: "Unsaved changes" badge                  │
│    - Tapping "Save Description" → API call → new version saved  │
│    - description_final = user's accepted version                │
└─────────────────────────────────────────────────────────────────┘
```

**Edit Save Behavior**: Explicit save, not debounced auto-save.

| Approach | Decision |
|----------|----------|
| Debounced auto-save | ❌ Rejected: spotty connectivity causes errors; clutters audit trail |
| **Explicit save** | ✅ Chosen: clear mental model; works offline; clean audit trail |

**Flow**:
1. User edits text field → changes stored **locally in app state**
2. "Unsaved changes" indicator appears
3. User taps **"Save Description"** → `POST /api/v1/tickets/{id}/description`
4. Server creates `transcription_version` record (source: `user_edit`)
5. If app closes with unsaved changes: prompt to save or discard

Each `transcription_version` represents an **intentional checkpoint**, not every keystroke.

**Voice Cleanup Prompt (Full)**:
```
You are a construction documentation assistant helping clean up voice transcriptions from job site foremen.

Your task:
1. Remove filler words: "uh", "um", "like", "you know", "I mean", repeated words
2. Fix grammar and punctuation
3. Normalize tense (prefer past tense for completed work)
4. Make the language professional and clear
5. Preserve ALL factual information exactly as stated
6. Convert measurements to standard format (e.g., "fifty feet" → "50 feet")

Do NOT:
- Add information not in the original
- Remove factual details
- Add legal or contractual language
- Summarize or shorten significantly
- Change quantities, times, or measurements

The output should be suitable for:
- Client invoices
- Project manager review
- Potential dispute resolution

Input transcription:
{raw_text}

Output only the cleaned text, nothing else.
```

### Flow 4: GC Signature (Secure Link-Based)

**Security requirement**: GC must sign on their own device OR provide SMS verification to prevent foreman forgery.

#### Option A: Link-Based Signing (Default, Recommended)

GC receives signing link via SMS/email, opens on their own device.

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Foreman completes ticket → Taps "Request GC Signature"       │
│    - Status changes: draft → pending_review                     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. API: POST /api/v1/tickets/{id}/signature-request             │
│    Server actions:                                              │
│    - Generate cryptographic token (32+ chars)                   │
│    - Create signing_link record (expires in 24h)                │
│    - Generate 6-digit SMS verification code (for fallback)      │
│    - Send SMS to GC's phone: "Sign T&M ticket: [link]"          │
│    - Send email to GC with link + ticket summary                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. GC receives SMS/email, opens link on THEIR device            │
│    URL: https://siteproof.app/sign/{token}                      │
│    - No app install required (mobile web)                       │
│    - Token validates request is legitimate                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. GC reviews ticket on signing page:                           │
│    - Description, line items, photos                            │
│    - Three options: Approve, Reject, or Request Changes         │
└─────────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              │               │               │
              ▼               ▼               ▼
      ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
      │   APPROVE   │ │   REJECT    │ │  REQUEST    │
      │  (sign it)  │ │(with reason)│ │  CHANGES    │
      └──────┬──────┘ └──────┬──────┘ └──────┬──────┘
             │               │               │
             │               ▼               ▼
             │        ┌─────────────────────────────┐
             │        │ Ticket returns to foreman   │
             │        │ (see Flow 4b for details)   │
             │        └─────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. GC draws signature on their device                           │
│    - Canvas for finger/stylus signature                         │
│    - Name confirmation field                                    │
│    - Legal acknowledgment checkbox                              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 6. API: POST /api/v1/signing-links/{token}/sign                 │
│    Request: {                                                   │
│      signer_name: "John Smith",                                 │
│      signature_image: "data:image/png;base64,...",              │
│      acknowledgment: true                                       │
│    }                                                            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 7. Server actions:                                              │
│    - Verify token is valid and not expired                      │
│    - Verify ticket status = pending_review                      │
│    - Mark signing_link as used (used_at = now())                │
│    - Decode base64, save signature PNG                          │
│    - Compute SHA-256 of signature image                         │
│    - Create signature record:                                   │
│        signing_method = 'link_own_device'                       │
│        ip_address = GC's IP (different from foreman's!)         │
│        device_id = GC's device fingerprint                      │
│    - Update ticket: status=approved, gc_signed_at=now()         │
│    - Create audit_log (action: signed, method: link_own_device) │
│    - TRIGGER: PDF generation + email dispatch                   │
└─────────────────────────────────────────────────────────────────┘
```

**Why this is secure**: 
- Link sent to GC's phone/email → foreman can't intercept
- Signature captured from GC's IP address, not foreman's device
- Device fingerprint differs from foreman's device
- Audit trail clearly shows signing occurred on separate device

#### Option B: Same-Device with SMS Verification (Fallback)

For situations where GC **has their phone** but prefers to sign on foreman's device:
- GC's phone has poor connectivity or no data plan
- GC's phone is too small / broken screen
- GC's phone battery is low
- Foreman's tablet has bigger screen, easier to sign
- Time pressure (faster than waiting for GC to open link)

**Requirement**: GC must have their phone to receive the SMS verification code.

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Foreman shows ticket to GC on their device                   │
│    GC taps "Sign on this device"                                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. System sends SMS to GC's registered phone:                   │
│    "Your SiteProof verification code is: 847293"                │
│    (6-digit code, expires in 10 minutes, max 3 attempts)        │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. GC checks their phone, reads code aloud                      │
│    Foreman enters code on their device                          │
│    (Proves GC is physically present and consenting)             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. API: POST /api/v1/signing-links/{token}/verify               │
│    Request: { code: "847293" }                                  │
│    - Server verifies code matches                               │
│    - If wrong, increment attempts (max 3)                       │
│    - If correct, allow signature capture                        │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. GC draws signature on foreman's device                       │
│    - Same signature capture UI                                  │
│    - signing_method = 'same_device_verified'                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 6. Server creates signature with:                               │
│    - signing_method = 'same_device_verified'                    │
│    - verification_code_used = true                              │
│    - IP/device matches foreman's (expected)                     │
│    - Audit log notes: "Signed on foreman device with SMS verify"│
└─────────────────────────────────────────────────────────────────┘
```

**Forgery mitigation**: GC must have their phone to receive SMS code. Foreman cannot forge without physical access to GC's phone.

#### What If GC Doesn't Have Their Phone?

**They cannot sign on-site.** Options:

| Situation | Solution |
|-----------|----------|
| GC left phone in truck | Go get it |
| GC forgot phone at home | Use Option A: send link to email, GC signs later from office computer |
| GC has no phone at all | GC must sign from office computer via emailed link |
| Urgent, can't wait | Foreman documents the work, gets signature next site visit |

**Design decision**: We do not allow unverified same-device signing. The friction of "go get your phone" is preferable to the fraud risk of allowing forged signatures.

#### Signing Method Comparison

| Method | Forgery Risk | UX Friction | When to Use |
|--------|--------------|-------------|-------------|
| `link_own_device` | Very Low | Medium | Default for all signatures |
| `same_device_verified` | Low | Low | GC doesn't have phone handy, time pressure |
| ~~`same_device_unverified`~~ | **High** | None | **NOT ALLOWED** - removed from system |

**Policy**: Same-device signing without SMS verification is **not permitted**. The system enforces this at the API level.

### Flow 4b: GC Rejection (When Superintendent Refuses to Sign)

**Common reasons GC may refuse**:
- "This wasn't authorized extra work"
- "Quantities are incorrect"
- "I wasn't on site when this happened"
- "Description doesn't match what I observed"

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. GC Super taps "Reject" on review screen                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. Modal requires rejection reason (free text, required)        │
│    Example: "Quantities disputed - I counted 30 ft, not 50 ft"  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. API: POST /api/v1/tickets/{id}/reject                        │
│    Request: { reason: "Quantities disputed..." }                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. Server actions:                                              │
│    - Verify ticket status = pending_review                      │
│    - Update ticket:                                             │
│        status = 'rejected'                                      │
│        rejected_at = now()                                      │
│        rejected_by_id = gc_user_id                              │
│        rejection_reason = reason                                │
│    - Create audit_log (action: rejected, reason, rejector)      │
│    - Send notification email to foreman                         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. Foreman receives notification, can:                          │
│    - Edit ticket (clears rejection, status → draft)             │
│    - Add clarifying photos or description                       │
│    - Resubmit for signature (status → pending_review)           │
│    - Escalate to PM (Phase 2 feature)                           │
└─────────────────────────────────────────────────────────────────┘
```

**Rejection API**:
```
POST /api/v1/tickets/{id}/reject
Request:  { "reason": "string (required)" }
Response: { "id": "uuid", "status": "rejected", "rejected_at": "..." }
```

**Business note**: Most disputes are resolved verbally on-site. This flow handles cases where GC formally refuses. Escalation workflows (involving PMs from both parties) are deferred to Phase 2.

### Flow 5: PDF Locking

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Triggered immediately after GC signature                     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. Server compiles PDF data:                                    │
│    - Ticket details (title, description_final, work_date)       │
│    - Line items                                                 │
│    - All photos (embedded as images)                            │
│    - Both signatures (foreman + GC super)                       │
│    - Timestamps for all actions                                 │
│    - Audit hash chain reference                                 │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. Generate PDF via Typst:                                      │
│    - Render /templates/ticket.typ with JSON data                │
│    - Output: /data/uploads/{company}/pdfs/{ticket_id}/{uuid}.pdf│
│    - Compute SHA-256 of PDF file                                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. Lock ticket:                                                 │
│    - Compute content_hash (SHA-256 of all ticket data)          │
│    - Update ticket:                                             │
│        status = 'locked'                                        │
│        locked_at = now()                                        │
│        pdf_path = path                                          │
│        pdf_hash = hash                                          │
│        content_hash = hash                                      │
│    - Create audit_log (action: locked, hashes)                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. Immutability enforcement:                                    │
│    - Database trigger prevents UPDATE on locked tickets         │
│    - Only allowed update: adding email_events                   │
└─────────────────────────────────────────────────────────────────┘
```

**Ticket Lock Trigger**:
```sql
CREATE OR REPLACE FUNCTION prevent_locked_ticket_changes() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status = 'locked' THEN
        -- Only allow email event tracking updates
        IF NEW.status = OLD.status AND 
           NEW.content_hash = OLD.content_hash AND
           NEW.pdf_hash = OLD.pdf_hash THEN
            RETURN NEW;
        END IF;
        RAISE EXCEPTION 'Locked tickets cannot be modified';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ticket_lock_enforcement
BEFORE UPDATE ON tickets
FOR EACH ROW EXECUTE FUNCTION prevent_locked_ticket_changes();
```

### Flow 6: Email Dispatch

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Triggered immediately after PDF locking                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. Determine recipients:                                        │
│    - Foreman who created ticket                                 │
│    - GC superintendent who signed                               │
│    - Project's gc_contact_email                                 │
│    - Company's PM and accounting (if configured)                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. For each recipient, call Postmark API:                       │
│    POST https://api.postmarkapp.com/email                       │
│    {                                                            │
│      "From": "tickets@siteproof.app",                           │
│      "To": "gcsuper@contractor.com",                            │
│      "Subject": "T&M Ticket TM-2024-0042 Approved - Project X", │
│      "HtmlBody": "<rendered email template>",                   │
│      "Attachments": [{                                          │
│        "Name": "TM-2024-0042.pdf",                              │
│        "Content": "<base64 PDF>",                               │
│        "ContentType": "application/pdf"                         │
│      }],                                                        │
│      "MessageStream": "outbound"                                │
│    }                                                            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. For each Postmark response:                                  │
│    - Create email_events record with message_id                 │
│    - Status = 'sent'                                            │
│    - Create audit_log (action: email_sent, recipient)           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. Postmark webhooks (async, later):                            │
│    POST /api/v1/webhooks/postmark                               │
│    - Update email_events with delivered_at, opened_at, etc.     │
└─────────────────────────────────────────────────────────────────┘
```

---

## 6. Security & Correctness

### Data Classification

| Classification | Fields | Handling |
|----------------|--------|----------|
| **PII** | `users.email`, `users.phone`, `users.name`, `signatures.signer_name` | Redacted in logs, deletable on request |
| **Sensitive** | `users.password_hash`, JWT secrets, API keys, signature images | Never logged, encrypted at rest (PostgreSQL TDE if needed) |
| **Business Confidential** | Ticket details, photos, line items, PDFs | Access controlled by `company_id`, audit logged |
| **Audit Data** | `audit_logs.*`, `transcription_versions.*` | Immutable, retained for legal compliance |
| **Public** | None | N/A - all data requires authentication |

### Data Retention Policy

| Data Type | Retention Period | Deletion Trigger | Notes |
|-----------|------------------|------------------|-------|
| **Locked tickets + PDFs** | Indefinite | None | Legal records, immutable by design |
| **Draft tickets** | 90 days if abandoned | Auto-cleanup job | Notify user before deletion |
| **Rejected tickets** | 1 year | Auto-cleanup job | May be needed for dispute reference |
| **Photos** | Tied to ticket lifecycle | Cascade delete with ticket | |
| **Audit logs** | 7 years minimum | Legal requirement | May be extended per contract |
| **User accounts** | Until deletion request | GDPR/CCPA request | Soft delete, anonymize after 30 days |
| **Transcription versions** | Tied to ticket lifecycle | Cascade delete with ticket | |
| **Email events** | 2 years | Auto-cleanup job | Delivery tracking |
| **Signing links** | 30 days after use/expiry | Auto-cleanup job | Short-lived tokens |

**User deletion flow**:
1. User requests account deletion
2. Soft delete: `users.deleted_at = now()`, `users.active = false`
3. 30-day grace period (can restore)
4. Hard delete: Anonymize PII, retain audit logs with anonymized actor

### Threat Model Summary

| Asset | Threat | Mitigation |
|-------|--------|------------|
| User credentials | Brute force, credential stuffing | bcrypt hashing, rate limiting (5 attempts/min), account lockout |
| JWT tokens | Token theft | Short expiry (15min), refresh rotation, HTTPS only |
| Ticket data | Unauthorized access | Row-level auth checks, company_id scoping |
| Signed tickets | Tampering post-signature | Hash chains, database triggers, PDF hash verification |
| **GC signatures** | **Foreman forgery** | **Link-based signing on GC's device; SMS verification fallback** |
| Photos | Modification | SHA-256 hash at upload, stored hash comparison |
| Audit logs | Deletion/modification | Append-only table, hash chain, database triggers |
| API | Abuse, DoS | Caddy rate limiting, input validation, request size limits |
| Secrets | Exposure | Env files with restricted permissions, not in git |

### Preventing Signature Forgery

**Threat**: Foreman creates inflated T&M ticket, draws fake GC signature on their own device, submits for payment.

**Attack scenario**:
1. Foreman inflates hours/materials on ticket
2. Foreman draws squiggle claiming it's GC's signature
3. Ticket locks, PDF emails to GC
4. GC receives email with signature they never made

**Mitigations (defense in depth)**:

| Layer | Control | How it helps |
|-------|---------|--------------|
| **Primary** | Link-based signing | GC signs on their own device; IP/device fingerprint differs from foreman |
| **Secondary** | SMS verification | Even on same device, GC must have their phone to receive code |
| **Detective** | IP address logging | Audit shows if signature came from same IP as foreman |
| **Detective** | Device fingerprinting | Client-side fingerprint detects same-device signing |
| **Detective** | Immediate email to GC | GC sees PDF within seconds; can dispute immediately |
| **Deterrent** | Audit trail | Clear evidence chain for fraud investigation |

**Enforcement**:
- API **rejects** signature requests without valid signing_link token
- Same-device signing **requires** SMS verification code
- `signing_method` field in signatures table records how signature was obtained
- PDF includes signing method: "Signed via secure link on signer's device" vs "Signed on requesting party's device with SMS verification"

**Residual risk accepted**: If foreman has physical access to GC's phone, they could intercept the SMS verification code and forge the signature. This is a physical security problem that software cannot fully solve. 

We accept this risk because:
1. The attack requires physical access to GC's phone (uncommon)
2. Foreman must be willing to commit fraud (legal consequences: contract fraud, potential criminal charges)
3. GC will eventually see the emailed PDF (when they check email on another device or return to office)
4. Audit trail creates evidence for disputes

**Not a mitigation**: Sending additional SMS or email notifications does not prevent this attack—if foreman has GC's phone, they can intercept/delete both channels.

**Future consideration**: For high-value tickets (>$X threshold), require GC to click confirmation link in email from a different device, or require in-person identity verification. This adds friction and is deferred unless fraud becomes a problem.

### Preventing Tampering After Signature

**Multi-layer protection**:

1. **Database level**: 
   - Trigger prevents UPDATE on tickets where `status = 'locked'`
   - Audit logs are append-only (trigger prevents UPDATE/DELETE)

2. **Content hashing**:
   - `content_hash` = SHA-256 of (title + description + line items + photo hashes + signature hashes)
   - Stored at lock time, verifiable anytime

3. **PDF hashing**:
   - `pdf_hash` = SHA-256 of generated PDF file
   - PDF can be re-verified against stored hash

4. **Audit hash chain**:
   - Each audit log entry includes hash of previous entry
   - Broken chain = tampering detected
   - Verification query available in admin tools

**Verification endpoint**:
```
GET /api/v1/tickets/{id}/verify
Response: {
  "valid": true,
  "content_hash_match": true,
  "pdf_hash_match": true,
  "audit_chain_valid": true,
  "photo_hashes_valid": true
}
```

### Audit Trail Design

**What is logged**:
- Every ticket state change (draft → pending → approved → locked)
- Every photo upload
- Every signature capture
- Every email sent
- Every login
- Every description cleanup (with before/after)

**Log entry structure**:
```json
{
  "id": 12345,
  "entity_type": "ticket",
  "entity_id": "uuid",
  "action": "signed",
  "actor_id": "user-uuid",
  "actor_ip": "192.168.1.1",
  "occurred_at": "2024-12-29T10:30:00Z",
  "changes": {
    "status": {"old": "pending_review", "new": "approved"},
    "gc_signed_by_id": {"old": null, "new": "signer-uuid"}
  },
  "prev_hash": "abc123...",
  "entry_hash": "def456..."
}
```

### Access Control

**Role-based permissions**:

| Action | foreman | gc_super | pm | accounting | admin |
|--------|---------|----------|-----|------------|-------|
| Create ticket | ✓ | - | - | - | ✓ |
| Edit draft ticket | ✓ (own) | - | ✓ | - | ✓ |
| Upload photos | ✓ | - | - | - | ✓ |
| Sign as foreman | ✓ (own) | - | - | - | - |
| Sign as GC | - | ✓ | - | - | - |
| View tickets | ✓ (own) | ✓ (signed) | ✓ | ✓ | ✓ |
| View audit logs | - | - | ✓ | - | ✓ |
| Manage users | - | - | - | - | ✓ |

**Implementation**: Middleware checks `user.role` and `user.company_id` on every request. Queries always include `WHERE company_id = $user_company_id`.

### Data Integrity Guarantees

1. **Referential integrity**: Foreign keys enforced at database level
2. **Unique constraints**: ticket_number per company, one signature per type per ticket
3. **Check constraints**: Valid status transitions, valid roles
4. **Optimistic locking**: `updated_at` column checked on updates (prevents lost updates)

### Consistency Model

| Operation | Consistency | Rationale |
|-----------|-------------|-----------|
| Ticket CRUD | **Strong** (PostgreSQL ACID) | Cannot lose user data |
| Signature + status change + PDF generation | **Strong** (single DB transaction) | Atomic state transition |
| Photo upload | **Strong** | File saved before DB commit |
| Audit log writes | **Strong** | Must never lose audit trail |
| Email delivery status | **Eventual** (webhook callback) | Acceptable delay; Postmark is source of truth |
| Voice cleanup | **Synchronous** | User waits for result |

**Transaction boundaries**:
- Signature capture is atomic: `BEGIN` → save signature → update ticket status → create audit log → `COMMIT`
- If any step fails, entire operation rolls back
- PDF generation happens after commit (if it fails, ticket is approved but PDF is regenerated on retry)

### External Service Resilience

| Service | Timeout | Retry Strategy | Circuit Breaker | Fallback |
|---------|---------|----------------|-----------------|----------|
| **OpenAI** (voice cleanup) | 10s | 2× with 1s backoff | Open after 5 consecutive failures, reset after 30s | Return raw transcription text |
| **Postmark** (email) | 5s | 3× with exponential backoff | Open after 10 failures | Queue to `failed_emails` table, retry via background job |
| **Twilio** (SMS) | 5s | 2× with 500ms backoff | Open after 5 failures | Show "SMS failed, try link-based signing" |
| **PostgreSQL** | 3s | No retry (fail fast) | N/A | Return 503 to client |

**Implementation** (Go):
```go
// Circuit breaker for OpenAI
openaiBreaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "openai",
    MaxRequests: 3,              // Requests allowed in half-open state
    Interval:    30 * time.Second,
    Timeout:     30 * time.Second, // Time to stay open before half-open
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        return counts.ConsecutiveFailures >= 5
    },
})
```

**Idempotency implementation**:
- `POST` operations accept `Idempotency-Key` header
- Server stores `hash(key + endpoint) → response` in Redis or PostgreSQL for 24 hours
- Duplicate requests within window return cached response with `X-Idempotent-Replayed: true` header
- Natural idempotency: photo uploads use SHA-256 hash as dedup key

### Time & Timezone Correctness

**Rules**:
- All timestamps stored as `TIMESTAMPTZ` (UTC in database)
- API accepts/returns ISO 8601 with timezone: `2024-12-29T10:30:00-05:00`
- `work_date` is `DATE` type (no time component, represents calendar day)
- Server always operates in UTC
- Client renders in user's local timezone
- PDF displays timestamps in project's configured timezone

**Timezone handling in PDF**:
```
Signed: December 29, 2024 at 10:30 AM EST
```

Project settings include `timezone` field (e.g., "America/New_York").

---

## 7. Performance & Scalability

### Expected Load Assumptions (MVP to Scale)

| Metric | MVP (10 companies) | Growth (100 companies) | Scale (1000 companies) |
|--------|-------------------|----------------------|----------------------|
| Active users | 50 | 500 | 5,000 |
| Tickets/day | 20 | 200 | 2,000 |
| Photos/day | 100 | 1,000 | 10,000 |
| API requests/day | 5,000 | 50,000 | 500,000 |
| Peak QPS | 1 | 10 | 100 |
| Storage/month | 2 GB | 20 GB | 200 GB |

### Latency Targets

| Operation | Target p95 |
|-----------|-----------|
| API read (list tickets) | < 100ms |
| API write (create ticket) | < 200ms |
| Photo upload (2MB) | < 2s |
| Voice cleanup (OpenAI) | < 3s |
| PDF generation | < 2s |
| Email dispatch | < 1s (async ok) |

### Bottlenecks & Mitigations

| Bottleneck | When it matters | Mitigation |
|------------|-----------------|------------|
| Database connections | > 50 concurrent requests | pgBouncer connection pooling |
| File storage IOPS | > 1000 photos/hour | Migrate to S3-compatible |
| PDF generation CPU | > 100 PDFs/hour | Dedicated PDF worker process |
| OpenAI rate limits | > 500 requests/min | Queue with backoff, fallback to raw text |
| Email sending | > 1000 emails/hour | Postmark handles; our code is non-blocking |
| Single server CPU | > 200 QPS sustained | Vertical scale or add second server |

### Scaling Path: 10 → 1,000 Customers

**Phase 1 (10-50 companies): Single VPS**
- Upgrade Hetzner VPS as needed (CPX31 → CPX41 → CPX51)
- Add pgBouncer when connections exceed 50
- Monitor with simple metrics (Prometheus + node_exporter)

**Phase 2 (50-200 companies): Optimize**
- Move file storage to Hetzner Object Storage (S3-compatible)
- Add read replica for PostgreSQL (reporting queries)
- Implement CDN for static assets and photos
- Add Redis for session caching (optional)

**Phase 3 (200-1000 companies): Distribute**
- Separate API server from database server
- Multiple API servers behind load balancer
- Background worker process for PDF generation
- Consider multi-region for latency (US + EU)

### When to Consider ECS/Kubernetes (and When NOT to)

**VPS replication (2-3 servers + load balancer) works well when**:
- Single service (monolith) deployed to multiple instances
- Simple topology: app servers + database
- Deploys are weekly or less frequent
- Team can SSH and debug directly
- Brief downtime during deploy is acceptable
- Load is predictable (no auto-scaling needed)

**Example VPS topology at scale**:
```
[Cloudflare/HAProxy Load Balancer]
              ↓
[VPS-1: App]  [VPS-2: App]  [VPS-3: App]
              ↓
[VPS-4: PostgreSQL Primary]  [VPS-5: PostgreSQL Replica]
```

Deploy script: parallel SSH to all app servers, rolling restart. This handles **1,000+ companies, 100+ QPS** comfortably.

**ECS/Kubernetes becomes worthwhile when**:

| Trigger | Why it forces the change |
|---------|-------------------------|
| **5+ independently deployable services** | Coordinating deploys manually becomes error-prone |
| **Auto-scaling requirements** | Unpredictable burst traffic (10× spikes) needs automatic instance spin-up |
| **Zero-downtime deploys mandatory** | Rolling deploys with health checks, automatic rollback |
| **Team grows to 5+ engineers** | Need standardized deployment process, not tribal knowledge |
| **Multi-region with failover** | Complex routing, health checks, DNS failover |
| **Compliance requirements** | Audit trail of deployments, immutable infrastructure |

**For SiteProof specifically**:
- Stay on VPS through ~1,000 companies
- K8s/ECS adds value when you split out separate services (PDF worker, notification service, analytics)
- Consider **managed K8s** (DigitalOcean DOKS, Hetzner Cloud) or **Railway/Render** before full ECS/EKS
- The overhead of K8s isn't cost—it's learning curve and operational burden for a small team

**Cost comparison at scale**:
| Approach | Monthly Cost | Operational Overhead |
|----------|-------------|---------------------|
| 3× VPS (CPX31) | ~€45 | Low (SSH, systemd) |
| Managed K8s (DOKS) | ~€60 + nodes | Medium (kubectl, helm) |
| AWS ECS/Fargate | ~$150-300 | Medium-High (AWS complexity) |

**Recommendation**: Don't adopt K8s/ECS until you have multiple services OR unpredictable scaling needs. For a monolithic Go app, VPS scales further than most people think.

### What Changes When Scaling

| Aspect | MVP | At Scale |
|--------|-----|----------|
| File storage | Local disk | S3-compatible |
| Database | Single PostgreSQL | Primary + read replicas |
| App servers | Single Go binary | 2-3 instances + load balancer |
| PDF generation | Inline | Dedicated worker |
| Email | Synchronous | Queue-based |
| Monitoring | Basic logs | Full observability stack |

**What doesn't change**:
- Core data model
- API contracts
- Business logic
- Authentication flow
- PDF templates

---

## 8. Explicit Tradeoffs

### What Is Intentionally NOT Built (MVP)

| Feature | Why deferred |
|---------|--------------|
| Offline mode | Adds significant complexity (sync conflicts); users have cellular |
| Native apps | React Native web wrapper sufficient initially; native later for UX |
| Push notifications | Email is sufficient; add when user research demands it |
| Multi-language | English-only; construction crews in target market speak English |
| SSO/SAML | Small companies don't need it; add for enterprise deals |
| Reporting/analytics | Basic list views sufficient; add when PMs request dashboards |
| Time tracking integration | MVP is standalone; integrations are Phase 2 |
| Payment processing | Tickets are documentation, not invoices; billing is separate |
| Advanced workflows | Approval chains, multi-level sign-off; start simple |

### Risks Accepted Early

| Risk | Likelihood | Impact | Acceptance Rationale |
|------|-----------|--------|---------------------|
| Single point of failure (one VPS) | Medium | High | Acceptable for MVP; fast recovery from backups |
| OpenAI dependency for voice cleanup | Low | Medium | Fallback to raw text; not a showstopper |
| No real-time sync | Low | Low | Polling is acceptable; WebSockets later |
| Manual deployment | Medium | Low | Acceptable for 1-2 engineers; automate later |
| Limited monitoring | Medium | Medium | Start with logs; add metrics when needed |

### Deferred Decisions

| Decision | Defer until |
|----------|-------------|
| Database sharding strategy | > 1000 companies |
| Multi-region architecture | International expansion |
| Native mobile apps | User feedback indicates need |
| Queue system selection | Background processing becomes bottleneck |
| Feature flag system | Multiple concurrent experiments needed |

---

## 9. Phase 2 Evolution

### Object Storage Migration (Trigger: 50GB files OR multi-server)

**When**: Storage exceeds 50GB or when adding second server  
**Why**: Local disk doesn't scale; shared storage needed for multiple servers  
**How**:
1. Deploy Hetzner Object Storage (or Backblaze B2)
2. Add S3 client to Go app (configurable backend)
3. Migrate existing files with background job
4. Update file paths in database
5. Keep local storage as write-through cache

**Effort**: 1-2 days  
**Risk**: Low (abstraction already in place)

### Background Workers (Trigger: PDF generation > 2s p95)

**When**: PDF generation becomes slow or blocks requests  
**Why**: Synchronous PDF generation holds HTTP connections  
**How**:
1. Add simple job queue (PostgreSQL-based: `pgq` or `river`)
2. Create worker process (same Go binary, different mode)
3. PDF generation becomes: create job → return 202 → worker processes → webhook on completion
4. Client polls or receives push notification

**Effort**: 2-3 days  
**Risk**: Low (well-understood pattern)

### Read Replicas (Trigger: Reporting queries slow main DB)

**When**: Analytics/reporting queries impact ticket creation latency  
**Why**: Read-heavy reporting shouldn't compete with writes  
**How**:
1. Enable PostgreSQL logical replication
2. Create read replica on separate VPS
3. Route reporting queries to replica
4. Keep ticket creation/updates on primary

**Effort**: 1 day setup, ongoing monitoring  
**Risk**: Low (PostgreSQL replication is mature)

### Multi-Region Support (Trigger: International expansion)

**When**: Customers in EU/APAC need low latency  
**Why**: Cross-Atlantic latency impacts mobile UX  
**How**:
1. Deploy additional VPS in target region
2. Use global load balancer (Cloudflare)
3. Replicate database (active-passive or CockroachDB)
4. Object storage with CDN edge caching

**Effort**: Significant (weeks)  
**Risk**: Medium (distributed systems complexity)

### Enterprise SSO via WorkOS (Trigger: Enterprise customer request)

**When**: Large GC or national subcontractor requires Okta/Azure AD SSO for their users  
**Why**: Enterprise IT departments mandate SSO for security compliance; deal-blocking for large contracts

**How**:
1. Sign up for WorkOS (free to start, ~$125/mo per SSO connection when active)
2. Integrate WorkOS SDK alongside manual auth
3. Routing logic: If user's email domain has SSO configured → WorkOS; otherwise → manual auth
4. WorkOS returns user profile → create/match user in PostgreSQL → issue JWT
5. Manual email/password auth continues for non-SSO users

**Effort**: 1-2 days integration  
**Risk**: Low (WorkOS designed to complement existing auth)

**Also enables**:
- SCIM directory sync (auto-provision/deprovision users from enterprise IdP)
- Audit logs for compliance
- "Sign in with Microsoft" via Azure AD SSO (replaces basic OAuth)

### Local LLM for Voice Cleanup (Trigger: 50K+ cleanups/month OR enterprise privacy requirements)

**When**: 
- Volume exceeds ~50,000 transcription cleanups/month (cost savings justify GPU)
- Enterprise customers require on-premises data processing
- Smaller models achieve GPT-4o-mini quality for grammar correction

**Why**: Eliminates external dependency, reduces per-request cost to near-zero, keeps data on-prem  

**How**:
1. Add dedicated GPU server (Hetzner GEX44 with RTX 4000, ~€150/month)
2. Deploy Ollama or vLLM with Llama 3 70B (or fine-tuned smaller model)
3. Update API client to route to local endpoint
4. Keep OpenAI as fallback for edge cases or overflow

**Effort**: 2-3 days setup + ongoing model management  
**Risk**: Low (abstraction already in place; swap API endpoint)

---

## 10. Observability & Operability

### Logging

**Structured JSON logs** via Go's `slog`:
```json
{
  "time": "2024-12-29T10:30:00Z",
  "level": "INFO",
  "msg": "ticket_created",
  "request_id": "abc-123",
  "user_id": "user-uuid",
  "company_id": "company-uuid",
  "ticket_id": "ticket-uuid",
  "duration_ms": 45
}
```

**Log aggregation**: Initially `journalctl` + grep. Add Loki or similar when needed.

### Metrics (Phase 1.5)

When load justifies it, add Prometheus metrics:
- `http_requests_total{method, path, status}`
- `http_request_duration_seconds{method, path}`
- `pdf_generation_duration_seconds`
- `openai_request_duration_seconds`
- `email_sent_total{status}`

### Health Checks

```
GET /health
{
  "status": "ok",
  "database": "ok",
  "storage": "ok",
  "version": "1.2.3",
  "uptime_seconds": 86400
}
```

### Runbook Stubs

**Service won't start**:
1. Check logs: `journalctl -u siteproof-api -n 100`
2. Verify config: `cat /opt/siteproof/config/config.yaml`
3. Check database: `psql -c "SELECT 1"`
4. Check disk space: `df -h`

**High latency**:
1. Check database: `SELECT * FROM pg_stat_activity`
2. Check slow queries: `SELECT * FROM pg_stat_statements ORDER BY mean_time DESC`
3. Check memory: `free -m`
4. Check CPU: `top`

**Email not delivering**:
1. Check Postmark dashboard for bounces
2. Query `email_events` for error codes
3. Verify recipient email validity
4. Check spam folder

---

## 11. Testing Strategy

### Test Pyramid

| Level | Coverage | Tools |
|-------|----------|-------|
| Unit | Domain logic, hash functions, validation | Go `testing` |
| Integration | Database operations, file storage | `testcontainers-go` |
| API | HTTP handlers, auth flow | Go `httptest` |
| E2E | Critical flows (create → sign → email) | Playwright (future) |

### Quality Gates (CI)

1. `go vet` - static analysis
2. `golangci-lint` - linting
3. `go test ./...` - all tests
4. `sqlc compile` - SQL validation
5. `typst compile templates/ticket.typ` - template validation

### Critical Test Cases

- [ ] Ticket cannot be modified after locking (trigger enforcement)
- [ ] Audit log hash chain is valid
- [ ] PDF hash matches generated file
- [ ] Photo hash matches uploaded file
- [ ] Expired JWT returns 401
- [ ] User cannot access other company's tickets
- [ ] Voice cleanup preserves all facts

---

## 12. Delivery Plan (Phased)

### Phase 1: Core MVP (4-6 weeks)

**Week 1-2**: Foundation
- [ ] Project setup, CI/CD, database migrations
- [ ] User auth (register, login, JWT)
- [ ] Company and project CRUD

**Week 3-4**: Ticket flow
- [ ] Ticket CRUD with line items
- [ ] Photo upload with hashing
- [ ] Voice-to-text cleanup integration
- [ ] Signature capture

**Week 5-6**: Locking & delivery
- [ ] PDF generation with Typst
- [ ] Ticket locking with hash chain
- [ ] Email dispatch via Postmark
- [ ] Basic web UI

### Phase 1.5: Polish (2 weeks)

- [ ] Mobile-responsive web UI
- [ ] Error handling improvements
- [ ] Monitoring basics
- [ ] Documentation

### Phase 2: Scale Prep (when needed)

- [ ] Object storage migration
- [ ] Background workers
- [ ] Native mobile apps
- [ ] Advanced reporting

---

## Appendix: ADR Index

| ID | Decision | Status |
|----|----------|--------|
| ADR-001 | Use Go for backend | Accepted |
| ADR-002 | PostgreSQL as primary database | Accepted |
| ADR-003 | Local file storage initially | Accepted |
| ADR-004 | Typst for PDF generation | Accepted |
| ADR-005 | Postmark for email | Accepted |
| ADR-006 | Hash chain for audit logs | Accepted |
| ADR-007 | Synchronous voice cleanup | Accepted |
| ADR-008 | Manual JWT auth + Microsoft OAuth | Accepted |
| ADR-009 | Docker + docker-compose for deployment | Accepted |
| ADR-010 | React + Vite for web frontend (PWA) | Accepted |
| ADR-011 | React Native + Expo for mobile | Accepted |
| ADR-012 | Twilio for SMS (signing verification) | Accepted |
| ADR-013 | WorkOS for future enterprise SSO | Proposed |

---

*Document generated following software-architecture skill methodology. Priority order: correctness → security → reliability → maintainability → performance → cost.*

