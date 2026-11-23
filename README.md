# Redbook Go Backend

This repository is the planned Go replacement for `redbook-backend`. The existing Vue frontend remains unchanged and is the compatibility oracle.

## Status

Planning artifacts are complete; implementation has not started.

- Specification: `specs/001-golang-gin-gorm-migration/spec.md`
- Implementation plan: `specs/001-golang-gin-gorm-migration/plan.md`
- Data model: `specs/001-golang-gin-gorm-migration/data-model.md`
- API contract: `specs/001-golang-gin-gorm-migration/contracts/api.md`
- Task breakdown: `specs/001-golang-gin-gorm-migration/tasks.md`
- Validation guide: `specs/001-golang-gin-gorm-migration/quickstart.md`

## Tech Direction

- Go 1.23+
- Gin for HTTP routing
- GORM with the MySQL driver
- Redis, JWT, OSS, RocketMQ, and Dify integrations as described in `plan.md` and `research.md`

See `specs/001-golang-gin-gorm-migration/quickstart.md` for the future run and validation workflow.
