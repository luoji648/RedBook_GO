# Redbook Go Backend Constitution

## Core Principles

### I. API Contract Parity

The Go service must keep the existing frontend working without frontend changes. Paths, HTTP methods, query/form/body parameters, response envelope, error messages, status codes, pagination fields, and authentication behavior must match the existing backend or be intentionally documented as compatibility improvements.

### II. Database Compatibility

Reuse the existing MySQL `redbook` schema and table names. GORM models must map to existing columns, indexes, and relationships without requiring destructive migrations. Any new migration must be additive and idempotent.

### III. Layered Architecture

Code is organized into handlers, services, repositories, and models. Handlers translate HTTP to service calls; services own business rules and transactions; repositories isolate GORM queries. Cross-cutting concerns belong in middleware or shared packages, not scattered through handlers.

### IV. Correctness Under Concurrency

Money, stock, likes, follows, and unread cursors must be safe under concurrent requests. Use database transactions and row-level locking where the invariant spans multiple rows. Use idempotent outbox or message queues for asynchronous side effects.

### V. Security-First

Never return phone numbers or password hashes in public payloads. Validate and normalize input. Preserve the existing password verification scheme for legacy accounts. Keep JWT secrets and external credentials in configuration or environment variables.

### VI. Observability and Testability

Every handler returns the shared response shape. Structured logs include request context, user id, and error fields. Critical auth, commerce, and concurrency paths must have unit or integration tests before being considered complete.

## Additional Constraints

- Do not modify `redbook-backend` or `redbook-frontend`.
- The Vue frontend is the compatibility oracle and is not changed by this migration.
- Use Go, Gin, and GORM as the mandatory web framework and ORM.
- Prefer standard-library or well-maintained community libraries over hand-rolled replacements.
- Module path is `redbook-go`; commands are documented in the feature quickstart.

## Development Workflow

- Follow spec-kit: specification, plan, data model, contracts, tasks, then implementation.
- Each task is independently reviewable and should compile before the next task begins.
- Commit after each logical task group; do not mix unrelated changes.

## Governance

This constitution supersedes informal implementation preferences. A deviation must be justified in `plan.md` or a task note and must preserve API compatibility.

**Version**: 1.0.0 | **Ratified**: 2026-10-17 | **Last Amended**: 2026-10-17
