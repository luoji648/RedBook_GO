# Implementation Plan: Go Backend Migration (API Parity)

**Branch**: `001-golang-gin-gorm-migration` | **Date**: 2026-10-17 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/001-golang-gin-gorm-migration/spec.md`

## Summary

Build a new Go web service that is a behavior-compatible replacement for the existing Spring Boot 3 backend. The Vue frontend is left unchanged, so the Go service must reproduce the existing REST paths, request parameter styles, unified response envelope, JWT authentication behavior, MySQL schema mapping, and integrations for Redis, RocketMQ, OSS, and Dify. The implementation uses Gin for routing and GORM for persistence.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: Gin, GORM with MySQL driver, go-redis, golang-jwt, Viper, Zap, go-playground/validator, swaggo/swag, aliyun-oss-go-sdk, apache/rocketmq-client-go/v2, gin-contrib/cors

**Storage**: MySQL 8 using the existing `redbook` schema; Redis for tokens, blacklists, like caches, distributed locks, and unread/count state

**Testing**: standard `go test`, `testify`, `httptest`, `sqlmock`, and `miniredis`

**Target Platform**: Linux server; Windows is supported for local development

**Project Type**: web-service backend

**Performance Goals**: Preserve or improve existing behavior; public read endpoints target under 300 ms p95 locally, and core writes remain correct under concurrent requests

**Constraints**: No changes to `redbook-backend` or `redbook-frontend`; no destructive database migrations; existing passwords must remain valid

**Scale/Scope**: One deployable service with about 90 routes, roughly 30 mapped tables, and six business modules

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **API Contract Parity**: Pass. The plan derives routes and response shapes from the original controllers and frontend API client.
- **Database Compatibility**: Pass. GORM models map to existing tables; migrations remain additive and idempotent.
- **Layered Architecture**: Pass. Handlers, services, repositories, and models are separated.
- **Correctness Under Concurrency**: Pass. Commerce, like, follow, and unread paths are assigned transactional or atomic designs.
- **Security-First**: Pass. Legacy password verification is preserved and public payloads exclude phone and password fields.
- **Observability and Testability**: Pass. Shared logging, error handling, and focused tests are included.

No violations require justification.

## Project Structure

### Documentation (this feature)

```text
specs/001-golang-gin-gorm-migration/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── api.md
└── tasks.md
```

### Source Code (repository root)

```text
redbook-go/
├── cmd/server/main.go
├── internal/
│   ├── config/
│   ├── model/
│   ├── repository/
│   ├── service/
│   ├── handler/
│   ├── middleware/
│   └── pkg/
│       ├── response/
│       ├── auth/
│       ├── cache/
│       ├── lock/
│       ├── idgen/
│       ├── oss/
│       ├── dify/
│       └── mq/
├── migrations/
├── tests/
│   ├── unit/
│   └── integration/
├── docs/
├── go.mod
└── go.sum
```

**Structure Decision**: Use a single backend service layout. The existing frontend remains outside this repository and is not duplicated. Models map directly to the existing MySQL tables; handlers are grouped by domain prefix; services contain transactional business logic; repositories contain GORM queries.

## Complexity Tracking

No constitution violations.

