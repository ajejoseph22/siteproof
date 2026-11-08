# Roadmap

SiteProof is partly built. 22 of the 47 planned tasks are done. This page lists what's left, grouped by milestone. Each item matches a task in [`plans/prd.json`](plans/prd.json), which holds the step-by-step acceptance criteria. When a task is finished, set `"passes": true` there and tick it off here.

## Done

The foundation is in place: config loading, the multi-tenant schema and migrations, registration and JWT auth (access + refresh), RBAC and tenant scoping, idempotency keys, projects, tickets with cursor pagination, line items, photo upload and secure downloads, the signature request and capture flow, the audit-log hash chain and its verification endpoints, structured logging with request IDs, health checks, and the web auth screens.

## Milestone 1: Lock and prove a ticket

This is the core promise of the product: a signed ticket becomes a record that can be shown to be untampered.

- [ ] **Content hash at lock time.** Canonically serialize the ticket content, including line items, photo hashes and signature hashes, and store `content_hash` when the ticket moves to `locked`. Add tests showing the hash changes whenever the content does.
- [ ] **PDF generation.** Map ticket data into `templates/ticket.typ`, compile it with Typst, store the file under `/data/uploads/{company_id}/pdfs/{ticket_id}`, and save its SHA-256 hash.
- [ ] **Integrity verify endpoint.** Recompute the content, PDF and photo hashes plus the audit chain, compare them with the stored values, and return a structured result. Cover both valid and tampered data in tests.

## Milestone 2: Delivery integrations

- [ ] **Email via Postmark.** Send the PDF to the signer and to internal recipients, record each send in `email_events`, and retry or track failures.
- [ ] **Postmark webhooks.** Verify the webhook, update delivery status and timestamps in `email_events`, and log unknown event types without acting on them.
- [ ] **SMS via Twilio.** Send signing links and verification codes, record every attempt, and return clear errors when delivery fails.
- [ ] **Voice-note cleanup via OpenAI.** Clean up raw transcriptions with timeouts, retries and a circuit breaker, store both the raw and the cleaned text, and fall back to the raw text when the call fails.

## Milestone 3: Web app workflows

The React client in `client/` has its auth screens and layout. The ticket workflows still need wiring to the API.

- [ ] **Ticket list.** Cursor pagination, status filters, empty and loading states, and a mobile layout.
- [ ] **Ticket detail.** Metadata, line items and totals, the photo gallery, signature status, the audit timeline, and actions that depend on role and status.
- [ ] **Create and edit form.** React Hook Form with Zod, adding, editing and removing line items, draft persistence, server-side validation errors, and status-aware controls.
- [ ] **Photo capture and upload.** Choose or capture photos, add a caption and `taken_at`, show upload progress with retry, and refresh the detail view afterwards.
- [ ] **Signature flows.** A signature canvas for the foreman, a signing-link page for the GC, a confirmation step, and no editing once signed.
- [ ] **PDF and email status.** PDF preview and download, email delivery status, and clear errors when PDF generation fails.
- [ ] **PWA.** `vite-plugin-pwa` with auto-update, cached static assets and core routes, an offline banner, and installability checks.

## Milestone 4: Test coverage

Current tests cover only `internal/domain` and `internal/handler`.

- [ ] **Locking and RBAC.** Locked tickets can't be edited, role restrictions hold, and cross-company access is rejected.
- [ ] **Photo upload integration.** Use testcontainers-go with Postgres. Upload a photo, check its stored SHA-256 hash and the file on disk, and confirm unauthorized users can't fetch it.
- [ ] **Audit chain and verify integration.** Build up a chain of audit entries, verify it through the endpoint, then tamper with the data and confirm verification fails.

## Milestone 5: Ops, docs and auth

- [ ] **Prometheus metrics.** Track HTTP request durations plus PDF and email timings, serve them at `/metrics`, and document them.
- [ ] **Caddy.** A Caddyfile for TLS, the reverse proxy, rate limiting, compression and static frontend assets, plus deployment notes.
- [ ] **Local Docker Compose.** `docker-compose.yml` already runs Postgres and the API. It still needs a Caddy service and startup and teardown docs.
- [ ] **OpenAPI spec.** Bring `openapi.yaml` in line with the implemented endpoints: security schemes, schemas, pagination, idempotency and error codes, checked against example requests.
- [ ] **Microsoft SSO.** OAuth login and callback with MSAL, linking users by email domain, and issuing JWTs after sign-in.

## Later: Scale

- [ ] **S3-compatible storage.** A storage interface with local and S3 backends, a config switch between them, and migration notes for existing files.
- [ ] **Background PDF worker.** A job table, a worker mode in the Go binary, enqueueing jobs on lock, retries and a dead-letter queue.
- [ ] **Read replica.** A replica connection for reporting queries, setup docs, and replication-lag metrics.
