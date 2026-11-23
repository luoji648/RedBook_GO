# Tasks: Go Backend Migration (API Parity)

**Input**: Design documents from `specs/001-golang-gin-gorm-migration/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/api.md`, `quickstart.md`

**Tests**: Focused tests are included for authentication, commerce, and concurrency-critical paths.

## Format: `[ID] [P?] [Story] Description`

- `[P]` means the task is parallelizable.
- `[Story]` maps a task to a user story from `spec.md`.

## Phase 1: Setup

**Purpose**: Initialize the new Go module and shared project skeleton.

- [ ] T001 Initialize `redbook-go` module and create the directory tree in `cmd/server`, `internal`, `tests`, `migrations`, and `docs`
- [ ] T002 [P] Implement configuration loading with Viper in `internal/config/config.go`
- [ ] T003 [P] Implement structured Zap logger and request logger in `internal/pkg/logger/logger.go`
- [ ] T004 [P] Implement the shared response envelope in `internal/pkg/response/result.go`
- [ ] T005 [P] Implement panic recovery and business-error middleware in `internal/middleware/error.go`
- [ ] T006 [P] Implement CORS middleware matching the original rules in `internal/middleware/cors.go`
- [ ] T007 Create the Gin router bootstrap and empty domain route groups in `cmd/server/main.go`
- [ ] T008 [P] Add `.gitignore`, `Makefile`, and `config/config.example.yaml`
- [ ] T009 [P] Add the baseline schema and migration notes under `migrations/`

## Phase 2: Foundational

**Purpose**: Build the blocking prerequisites required by every user story.

- [ ] T010 Implement GORM MySQL connection and transaction helpers in `internal/repository/db.go`
- [ ] T011 [P] Create account and content models in `internal/model/user.go`, `internal/model/note.go`, and `internal/model/media.go`
- [ ] T012 [P] Create commerce models in `internal/model/product.go`, `internal/model/order.go`, and `internal/model/wallet.go`
- [ ] T013 [P] Create messaging and utility models in `internal/model/chat.go`, `internal/model/notice.go`, and `internal/model/short_link.go`
- [ ] T014 [P] Implement Redis client, cache, and Lua script support in `internal/pkg/cache/redis.go`
- [ ] T015 [P] Implement Redis distributed lock in `internal/pkg/lock/lock.go`
- [ ] T016 Implement JWT signing, parsing, and token-version validation in `internal/pkg/auth/jwt.go`
- [ ] T017 Implement legacy password encoding and verification in `internal/pkg/auth/password.go`
- [ ] T018 Implement bearer-token authentication middleware and current-user context in `internal/middleware/auth.go`
- [ ] T019 [P] Implement pagination and query binding helpers in `internal/pkg/httputil/params.go`
- [ ] T020 [P] Implement validator error translation to compatible messages in `internal/pkg/httputil/validate.go`
- [ ] T021 [P] Add Swagger annotation bootstrap and documentation mount in `cmd/server/main.go`

## Phase 3: User Story 1 - Account and Authentication (Priority: P1)

**Goal**: Preserve login, profile, sign-in, and logout behavior.

**Independent Test**: Register and log in through the unchanged frontend, then inspect profile and token behavior.

- [ ] T022 [P] [US1] Implement user repository in `internal/repository/user_repo.go`
- [ ] T023 [P] [US1] Implement user-info repository in `internal/repository/user_info_repo.go`
- [ ] T024 [US1] Implement user service including code, login, logout, sign-in, profile, and password rules in `internal/service/user_service.go`
- [ ] T025 [US1] Implement user HTTP handlers in `internal/handler/user_handler.go`
- [ ] T026 [US1] Implement token blacklist and user session cache in `internal/service/auth_service.go`
- [ ] T027 [P] [US1] Add unit tests for password and JWT flows in `tests/unit/auth_test.go`
- [ ] T028 [US1] Add integration tests for public and protected user routes in `tests/integration/user_api_test.go`

## Phase 4: User Story 2 - Notes and Social Interactions (Priority: P1)

**Goal**: Preserve note CRUD, feeds, comments, likes, collects, and follows.

**Independent Test**: Publish and browse a note, then exercise comments, likes, collects, and follows through the frontend.

- [ ] T029 [P] [US2] Implement note, media, and note-product repositories in `internal/repository/note_repo.go`
- [ ] T030 [P] [US2] Implement comment, collect, follow, and like repositories in `internal/repository/interaction_repo.go`
- [ ] T031 [US2] Implement note service including visibility, feeds, recommendations, related notes, and count enrichment in `internal/service/note_service.go`
- [ ] T032 [US2] Implement note HTTP handlers in `internal/handler/note_handler.go`
- [ ] T033 [US2] Implement comment service and handlers in `internal/service/comment_service.go` and `internal/handler/comment_handler.go`
- [ ] T034 [US2] Implement collect service and handlers in `internal/service/collect_service.go` and `internal/handler/collect_handler.go`
- [ ] T035 [US2] Implement follow service and handlers in `internal/service/follow_service.go` and `internal/handler/follow_handler.go`
- [ ] T036 [US2] Implement like read/write service, aggregate cache, MQ producer wrapper, and handlers in `internal/service/like_service.go` and `internal/handler/like_handler.go`
- [ ] T037 [P] [US2] Add interaction and feed integration tests in `tests/integration/note_social_api_test.go`

## Phase 5: User Story 3 - Commerce (Priority: P1)

**Goal**: Preserve products, cart, orders, coupons, wallet, and product footprints.

**Independent Test**: Publish a product, complete cart and direct purchases, pay, refund, claim a coupon, and inspect balances.

- [ ] T038 [P] [US3] Implement product, cart, and footprint repositories in `internal/repository/product_repo.go`
- [ ] T039 [P] [US3] Implement order, order-item, coupon, and wallet repositories in `internal/repository/commerce_repo.go`
- [ ] T040 [US3] Implement product service and handlers in `internal/service/product_service.go` and `internal/handler/product_handler.go`
- [ ] T041 [US3] Implement cart service and handlers in `internal/service/cart_service.go` and `internal/handler/cart_handler.go`
- [ ] T042 [US3] Implement footprint service and handlers in `internal/service/footprint_service.go` and `internal/handler/footprint_handler.go`
- [ ] T043 [US3] Implement order service with create, preview, pay, refund, close, and delete rules in `internal/service/order_service.go`
- [ ] T044 [US3] Implement order HTTP handlers in `internal/handler/order_handler.go`
- [ ] T045 [US3] Implement wallet service and handlers with recharge ledger in `internal/service/wallet_service.go` and `internal/handler/wallet_handler.go`
- [ ] T046 [US3] Implement coupon service and handlers in `internal/service/coupon_service.go` and `internal/handler/coupon_handler.go`
- [ ] T047 [US3] Implement order timeout and seller settlement background jobs in `internal/service/order_scheduler.go`
- [ ] T048 [P] [US3] Add commerce unit tests for totals, coupon, and wallet transitions in `tests/unit/commerce_test.go`
- [ ] T049 [US3] Add commerce integration tests in `tests/integration/commerce_api_test.go`

## Phase 6: User Story 4 - Messaging and Notifications (Priority: P2)

**Goal**: Preserve private chat, group chat, unread counts, and interaction notifications.

**Independent Test**: Exchange private and group messages between sessions and verify unread totals and read cursors.

- [ ] T050 [P] [US4] Implement private-chat repositories in `internal/repository/chat_repo.go`
- [ ] T051 [P] [US4] Implement group and notice repositories in `internal/repository/group_repo.go`
- [ ] T052 [US4] Implement private chat service and handlers in `internal/service/chat_service.go` and `internal/handler/chat_handler.go`
- [ ] T053 [US4] Implement group service and handlers in `internal/service/group_service.go` and `internal/handler/group_handler.go`
- [ ] T054 [US4] Implement notice service and handlers with category read cursors in `internal/service/notice_service.go` and `internal/handler/notice_handler.go`
- [ ] T055 [US4] Implement unread aggregate logic across private, group, and notice state in `internal/service/unread_service.go`
- [ ] T056 [P] [US4] Add messaging and notification integration tests in `tests/integration/messaging_api_test.go`

## Phase 7: User Story 5 - Integrations and Utilities (Priority: P2)

**Goal**: Preserve sharing, short links, OSS upload, and Dify AI streaming.

**Independent Test**: Generate a short link, upload an image, and stream AI output for all supported modes.

- [ ] T057 [US5] Implement short-link repository, service, generation, redirect, and handlers in `internal/service/short_link_service.go` and `internal/handler/short_link_handler.go`
- [ ] T058 [US5] Implement OSS presigned and server-upload service in `internal/pkg/oss/oss.go`
- [ ] T059 [US5] Implement OSS HTTP handlers in `internal/handler/oss_handler.go`
- [ ] T060 [US5] Implement Dify streaming proxy in `internal/pkg/dify/client.go`
- [ ] T061 [US5] Implement AI stream handler and share handler in `internal/handler/ai_handler.go` and `internal/handler/share_handler.go`
- [ ] T062 [P] [US5] Add integration tests for short links and OSS with mocked clients in `tests/integration/integrations_api_test.go`

## Phase 8: User Story 6 - Operational Compatibility and Polish (Priority: P1)

**Goal**: Prove end-to-end compatibility and harden the service.

**Independent Test**: Run the unchanged frontend against the new backend and compare the documented contract.

- [ ] T063 [P] [US6] Complete OpenAPI annotations and verify generated documentation in `docs/`
- [ ] T064 [P] [US6] Add a contract smoke suite that asserts method, path, auth, and envelope behavior in `tests/integration/contract_test.go`
- [ ] T065 [P] [US6] Add request-id and user-id fields to structured logs in `internal/middleware/logger.go`
- [ ] T066 [P] [US6] Validate required configuration and fail fast with clear messages in `internal/config/config.go`
- [ ] T067 [US6] Run the unchanged Vue frontend against the Go service and fix compatibility gaps found during smoke testing
- [ ] T068 [P] [US6] Add concurrency tests for wallet, stock, likes, follows, and coupon claims in `tests/integration/concurrency_test.go`
- [ ] T069 [P] [US6] Review secret handling, public payload filtering, and CORS restrictions in `internal/pkg/auth` and `internal/handler`
- [ ] T070 [US6] Execute `quickstart.md` validation and record the result in the feature checklist

## Dependencies and Execution Order

- **Phase 1**: No dependencies.
- **Phase 2**: Depends on Phase 1; blocks all user stories.
- **User Story 1**: Depends on Phase 2.
- **User Story 2**: Depends on Phase 2 and logically benefits from User Story 1 auth but is independently testable with a seeded token.
- **User Story 3**: Depends on Phase 2 and User Story 1 for authenticated sellers/buyers.
- **User Story 4**: Depends on Phase 2 and User Story 1.
- **User Story 5**: Depends on Phase 2 and User Story 1.
- **User Story 6**: Depends on all desired user stories being implemented.

## Parallel Opportunities

- All `[P]` setup and foundational tasks can run in parallel where files are independent.
- User Story repositories marked `[P]` can be implemented in parallel within each story.
- After Phase 2, User Stories 1 through 5 can be implemented in parallel by separate developers.
- Test tasks marked `[P]` can run in parallel.

## Implementation Strategy

### MVP First

1. Complete Phase 1 and Phase 2.
2. Complete User Story 1 and verify login/profile flows.
3. Complete User Story 2 to prove the core community experience.
4. Complete User Story 3 for commerce.
5. Add messaging, integrations, and final contract hardening.

### Validation Gates

- `go build ./...` must pass after each phase.
- `go test ./...` must pass before advancing from foundational and commerce phases.
- The unchanged Vue frontend must remain runnable without edits at every user-story checkpoint.

