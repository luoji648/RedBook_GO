# Research: Go Backend Migration Decisions

## Decision: Gin as the HTTP framework

- **Decision**: Use `github.com/gin-gonic/gin` with standard route groups.
- **Rationale**: It is the requested framework, has mature middleware, JSON binding, multipart handling, and streaming support.
- **Alternatives considered**: `net/http`, Echo, Chi. Gin was selected for ecosystem familiarity and the ability to map Spring-style handlers naturally.

## Decision: GORM as the ORM

- **Decision**: Use `gorm.io/gorm` and `gorm.io/driver/mysql`.
- **Rationale**: It is the requested ORM and can map the existing snake_case tables through struct tags without changing the schema.
- **Alternatives considered**: `sqlx`, `ent`, raw `database/sql`. Raw SQL was rejected for maintainability across 30 tables; GORM was chosen for productivity while keeping explicit transactions for critical paths.

## Decision: Configuration

- **Decision**: Use `github.com/spf13/viper` to load `config.yaml` and environment variables.
- **Rationale**: It mirrors the original `application.yaml` plus environment override pattern and is standard for Go services.
- **Alternatives considered**: `envconfig`, `koanf`. Viper provides file and environment merging.

## Decision: Authentication and password compatibility

- **Decision**: Use `github.com/golang-jwt/jwt/v5`, Redis-backed blacklist and user cache, and reproduce the original `salt@md5(password + salt)` verification for legacy accounts.
- **Rationale**: Existing rows must remain valid without a reset. New password writes may keep the same format or use a versioned `bcrypt` marker only if backward-compatible; the safest initial implementation preserves the legacy scheme exactly.
- **Alternatives considered**: Blind bcrypt migration would break existing users. The migration keeps legacy verification and leaves hash migration as a separate optional task.

## Decision: Redis

- **Decision**: Use `github.com/redis/go-redis/v9` for JWT blacklists, user token-version state, like caches, distributed locks, and short-lived sets. Implement Lua scripts for atomic like-set and lock operations.
- **Rationale**: The original depends heavily on Redis; go-redis is the standard client and supports Lua and pipelines.
- **Alternatives considered**: `rueidis`, Redigo. go-redis was selected for breadth and active maintenance.

## Decision: In-process cache

- **Decision**: Use `github.com/allegro/bigcache` or a simple TTL cache for click-summary caching.
- **Rationale**: This replaces Caffeine without a JVM dependency.
- **Alternatives considered**: Ristretto, hand-written `sync.Map`. Either is acceptable; the plan favors a small explicit cache behind a `cache` interface.

## Decision: Message queue

- **Decision**: Use `github.com/apache/rocketmq-client-go/v2` for like-event producers and consumers. Keep the same topics (`click-topic`, `click-record-topic`, `click-summary-topic`) when MQ is enabled.
- **Rationale**: The original uses RocketMQ for like aggregation. Reusing the same topics preserves operational compatibility.
- **Alternatives considered**: Replace with an outbox and polling worker for development. The implementation will wrap the producer behind an interface so an outbox fallback can run without an MQ broker.

## Decision: Distributed ID generation

- **Decision**: Use auto-increment MySQL ids for most entities and a snowflake-style generator for `click_records` and `click_summary` ids.
- **Rationale**: Those tables use assigned ids in the original. Snowflake avoids collisions under concurrent like events.
- **Alternatives considered**: UUID. Snowflake preserves numeric ordering and is more consistent with the existing schema.

## Decision: OSS

- **Decision**: Use `github.com/aliyun/aliyun-oss-go-sdk/oss` for presigned PUT and server-side upload.
- **Rationale**: It is the official Go SDK and covers the same operations as the Java SDK.
- **Alternatives considered**: S3-compatible SDKs. Aliyun OSS SDK was chosen for direct parity.

## Decision: Dify AI SSE

- **Decision**: Proxy Dify with `http.Client` streaming and Gin's `c.Stream` or `http.Flusher`, forwarding `text/event-stream`.
- **Rationale**: Native Go HTTP supports streaming without a reactive library.
- **Alternatives considered**: `r3labs/sse`. The backend only forwards events, so a custom proxy is sufficient and avoids an extra dependency.

## Decision: OpenAPI documentation

- **Decision**: Use `github.com/swaggo/swag` plus `gin-swagger` annotations and expose a documentation endpoint equivalent to Knife4j.
- **Rationale**: The original exposes SpringDoc/Knife4j; swag provides the closest Go equivalent.
- **Alternatives considered**: Hand-written OpenAPI files. Annotations keep route and DTO definitions near the code.

## Decision: Error and response compatibility

- **Decision**: Centralize the original envelope in a `response.Result` type and convert domain errors to `success:false` with compatible messages. Return raw `401` for unauthenticated protected requests.
- **Rationale**: The frontend relies on these exact shapes and status behavior.
- **Alternatives considered**: Return `400` for all failures. That would break the frontend's auth-redirect logic and is rejected.

## Decision: CORS and trailing slash behavior

- **Decision**: Use `github.com/gin-contrib/cors` with the original allowed origin/method/header rules and add trailing-slash route compatibility.
- **Rationale**: The original permits all origins and methods in development, and the frontend proxy depends on those routes.
- **Alternatives considered**: Manual middleware. The maintained library reduces risk.

