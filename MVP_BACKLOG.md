# MVP Backlog (Architecture vs Code)

## Scope
This list compares the Phase 1 MVP requirements in `ADR/architecture.md` with the current Go API and web UI code. Items below are missing or only partially implemented and are required to complete the MVP.

## Backend & API Gaps
- **Signup + company bootstrap**: Add registration that creates a company and initial admin user (current API only supports login/refresh/logout).
- **Company CRUD**: Implement service and handlers for companies (create/get/update/list) with tenant scoping.
- **Ticket list pagination**: Implement paginated list endpoints with metadata (cursor-based per architecture), not just list-by-project.
- **Ticket line items**: CRUD for `ticket_line_items` and include line items in ticket responses.
- **Photo upload + storage**: Multipart upload, SHA-256 hashing/dedupe, local file storage, and list/delete endpoints.
- **Voice cleanup**: `/api/v1/tickets/{id}/description/cleanup` with OpenAI integration and `transcription_versions` persistence.
- **Signature flow**: Create signing links, send SMS/email, verify codes, accept signatures, and enforce signing method rules.
- **Approve + lock**: Approve endpoint, compute `content_hash`, generate PDF via Typst, store `pdf_hash`, and lock ticket.
- **Email delivery**: Postmark send + webhook ingestion into `email_events`.
- **Audit logging**: Write audit logs for ticket, photo, signature, email, and auth events so the hash chain is exercised.
- **Idempotency**: Honor `Idempotency-Key` for POSTs using `idempotency_keys` (return cached response on replay).
- **Integrity verification**: `/api/v1/tickets/{id}/verify` to check content hash, PDF hash, photo hashes, and audit chain.
- **Role enforcement**: Apply `RequireRole` to endpoints (foreman vs GC vs PM vs admin).

## Frontend/UI Gaps (`client/`)
- **Line items UI**: Add/edit/remove line items during ticket creation and detail views.
- **Photo gallery + upload**: Wire photo upload and display to the API.
- **Voice notes**: Capture raw text, request cleanup, and show raw vs cleaned.
- **Signature UX**: Request signature, handle signing link flow, and display signatures on ticket detail.
- **Workflow actions**: Submit, approve/reject, and download PDF wired to backend endpoints.

## Quality & Delivery Gaps
- **CI pipeline**: Automate `go test`, `golangci-lint`, `sqlc compile`, and `typst compile`.
- **Critical tests**: Add tests for locking immutability, audit chain validity, and access control.
