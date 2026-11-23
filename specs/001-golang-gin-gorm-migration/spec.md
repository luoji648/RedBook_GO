# Feature Specification: Go Backend Migration (API Parity)

**Feature Branch**: `001-golang-gin-gorm-migration`

**Created**: 2026-10-17

**Status**: Draft

**Input**: User description: "把项目用 golang 重写，用 gin+gorm，创建一个新项目，保持前端不动。"

## User Scenarios & Testing

### User Story 1 - Account and Authentication (Priority: P1)

An existing user can sign up or sign in with a phone number and verification code, sign in with a password, retrieve their own and public profiles, update profile fields, sign in daily, and log out. The new backend accepts the same token format and enforces the same access rules.

**Why this priority**: Authentication gates every other feature and must be compatible before content, commerce, or messaging can be exercised.

**Independent Test**: Start the new backend against an existing database, register and log in through the unchanged frontend, then verify profile, sign-in, and logout flows without modifying any frontend code.

**Acceptance Scenarios**:

1. **Given** a valid phone number, **When** the frontend requests a code, **Then** the service returns a six-digit code in the standard response envelope.
2. **Given** a valid code, **When** the user logs in, **Then** the service returns a bearer token accepted by subsequent authenticated requests.
3. **Given** a valid phone and password, **When** the user logs in by password, **Then** the service returns a valid token and does not expose the password hash.
4. **Given** an authenticated user, **When** they update nickname or avatar, **Then** the change is persisted and returned consistently.
5. **Given** an invalid or expired token, **When** the frontend calls a protected endpoint, **Then** the service returns `401` and the frontend redirects to login.

---

### User Story 2 - Notes and Social Interactions (Priority: P1)

Users can create, publish, edit, delete, and browse notes with media and linked products; view feeds, recommendations, related notes, and user profiles; and interact through comments, likes, collections, and follows.

**Why this priority**: Notes and interactions are the core community experience and depend only on account identity.

**Independent Test**: Using the unchanged frontend, create a note with images, publish it, view it in public and personal lists, then like, collect, comment, and follow another user; each action returns the same result shape as the original backend.

**Acceptance Scenarios**:

1. **Given** an authenticated user, **When** they save a note with media and product links, **Then** the note and its media/product rows are persisted.
2. **Given** a published note, **When** another user views its detail, **Then** the service returns title, content, media, counts, author, and viewer interaction state without leaking private fields.
3. **Given** a note, **When** a user comments, **Then** the comment tree includes root comments and replies with author and reply-target information.
4. **Given** a note, **When** a user likes or collects it, **Then** the service persists the interaction and updates counts atomically.
5. **Given** a user follows another user, **When** either party requests their follower or following list, **Then** the list is paginated and respects privacy visibility.

---

### User Story 3 - Commerce (Priority: P1)

Sellers can manage products, and buyers can browse products, use a cart, create and pay for orders, receive refunds, claim coupons, record product footprints, and recharge a wallet.

**Why this priority**: Commerce is a substantial, independent value slice and exercises transactions, inventory, money, and idempotency rules.

**Independent Test**: From the unchanged frontend, publish a product, add it to a cart, create an order, pay with the simulated wallet password, verify inventory and wallet balance changes, then test refund and coupon flows.

**Acceptance Scenarios**:

1. **Given** a seller, **When** they create or update a product, **Then** the product is stored with seller, price, stock, cover, and status fields.
2. **Given** a buyer with cart items, **When** they create an order, **Then** the order totals and line items are calculated in cents.
3. **Given** a pending order and sufficient wallet balance, **When** the buyer pays, **Then** the order, inventory, wallet, and coupon state change in one consistent transaction.
4. **Given** a paid order within the refund window, **When** the buyer refunds, **Then** the wallet is credited and order status changes accordingly.
5. **Given** a user who follows a seller, **When** they claim a follow coupon, **Then** the coupon is issued once per product and appears in their coupon list.

---

### User Story 4 - Messaging and Notifications (Priority: P2)

Users can exchange private messages, create and join groups, manage group membership, mark conversations read, and see unread counts for chat and interaction notifications.

**Why this priority**: Messaging builds on identity but can be delivered after the core community and commerce paths are stable.

**Independent Test**: Use two frontend sessions to exchange private messages, then create a group, send group messages, join with approval, mark read, and verify unread totals.

**Acceptance Scenarios**:

1. **Given** two users, **When** one sends a private message, **Then** a stable thread is created and the recipient sees an updated unread count.
2. **Given** a group creator, **When** they create a group, **Then** the group and owner membership are persisted.
3. **Given** a join request, **When** the owner approves or rejects it, **Then** membership or request status is updated and the applicant is notified.
4. **Given** a user who views interactions, **When** they mark a category read, **Then** subsequent unread counts reflect the read cursor.

---

### User Story 5 - Integrations and Utilities (Priority: P2)

Users can share notes via short links, upload images through OSS presigned or server-upload endpoints, and use the Dify-powered note AI assistant with streaming output.

**Why this priority**: These are optional integrations that must preserve their external contracts but can be validated after core modules.

**Independent Test**: Generate a short link and open the redirect path, upload an image through the unchanged frontend, and call the AI stream endpoint with each supported mode.

**Acceptance Scenarios**:

1. **Given** a note, **When** a user generates a short link, **Then** the service returns a short URI that redirects to the note URL.
2. **Given** an image file, **When** the frontend uploads it, **Then** the service returns object key and public URL.
3. **Given** a note id and AI mode, **When** the frontend opens the SSE stream, **Then** events are streamed in the expected text/event-stream format.

---

### User Story 6 - Operational Compatibility (Priority: P1)

The service returns the same response envelope and error behavior, applies the same authentication and CORS rules, supports the same pagination conventions, exposes OpenAPI documentation, and can start with the same MySQL and Redis infrastructure.

**Why this priority**: Without these shared contracts, individual module parity is not verifiable.

**Independent Test**: Run the new service, load the unchanged frontend, exercise authenticated and public endpoints, and compare success/error shapes with the original backend.

**Acceptance Scenarios**:

1. **Given** any business endpoint, **When** it succeeds, **Then** the response contains `success`, `errorMsg`, `data`, and `total` with the same meaning as the original.
2. **Given** a business error, **When** it occurs, **Then** the service returns `success: false` and a compatible `errorMsg`.
3. **Given** an unauthenticated protected request, **When** it is made, **Then** the service returns `401` without a business envelope.
4. **Given** a public endpoint, **When** it is called, **Then** it does not require a token.

---

### Edge Cases

- Phone numbers that are malformed, duplicated, or have exceeded the daily SMS limit.
- Note visibility values that are invalid or changed while the note is referenced by another user.
- Concurrent likes, follows, stock updates, coupon claims, and wallet payments.
- Orders that expire, are paid twice, are closed while paying, or use an expired coupon.
- Group membership after a user is kicked, leaves, or the owner is deleted.
- Short-link collisions, expired links, and redirects to deleted notes.
- OSS credentials missing, oversized files, and unsupported media types.
- Dify unavailable, returning non-SSE errors, or producing partial streams.

## Requirements

### Functional Requirements

- **FR-001**: The backend MUST be a new, independently buildable service and MUST NOT modify `redbook-backend` or `redbook-frontend`.
- **FR-002**: The backend MUST preserve every endpoint, method, parameter source, path variable, and pagination parameter used by the existing frontend.
- **FR-003**: The backend MUST return the shared envelope `{success, errorMsg, data, total}` for normal business responses.
- **FR-004**: The backend MUST return `401` for missing, invalid, expired, blacklisted, or token-version-mismatched credentials on protected routes.
- **FR-005**: The backend MUST support public routes without authentication exactly as listed in the contract.
- **FR-006**: The backend MUST support verification-code login, password login, logout, profile retrieval/update, daily sign-in, and password change.
- **FR-007**: The backend MUST support note save/publish/delete/detail/my/feed/recommend/related/user list operations with media and linked products.
- **FR-008**: The backend MUST support comment add/tree/delete, like/unlike/view, collect/uncollect/my/user, and follow/unfollow/list operations.
- **FR-009**: The backend MUST support product list/my/get/save/stock plus product-footprint record/list.
- **FR-010**: The backend MUST support cart add/my/update/remove.
- **FR-011**: The backend MUST support order create, direct preview, direct create, list, detail, pay, refund, close, and delete record.
- **FR-012**: The backend MUST support wallet balance and simulated Alipay recharge.
- **FR-013**: The backend MUST support coupon list and follow-seller coupon claim.
- **FR-014**: The backend MUST support private chat unread totals, thread list, peer-thread resolution, message list, send, and mark-read.
- **FR-015**: The backend MUST support group create, owner listing, my groups, join, join-request handling, metadata, messages, send, mark-read, detail, kick, and profile update.
- **FR-016**: The backend MUST support interaction notice paging and category read cursors.
- **FR-017**: The backend MUST support note share and short-link generation plus `/s/{code}` redirect.
- **FR-018**: The backend MUST support OSS presigned upload and server-side multipart upload.
- **FR-019**: The backend MUST support Dify AI streaming over `text/event-stream`.
- **FR-020**: The backend MUST reuse the existing MySQL schema and preserve all relationships and uniqueness constraints.
- **FR-021**: The backend MUST preserve the existing password hash format so legacy accounts remain valid.
- **FR-022**: The backend MUST apply consistent validation and return compatible user-facing error messages.
- **FR-023**: The backend MUST expose configuration for database, Redis, JWT, MQ, OSS, and Dify through environment variables or a config file.
- **FR-024**: The backend MUST emit structured logs and an OpenAPI-compatible documentation surface.

### Key Entities

- **User**: account identity, phone, password hash, nickname, avatar, token version.
- **UserInfo**: public and private profile fields such as city, introduction, counts, gender, birthday, credits, level, and privacy flags.
- **Note**: content, type, status, visibility, and counters.
- **NoteMedia, NoteProduct**: ordered media and linked products for a note.
- **Like/Collect/Follow/Comment**: interaction records and related count semantics.
- **Product, Cart, Order, OrderItem, UserCoupon, ProductFootprint**: commerce data.
- **UserWallet, WalletRecharge**: money balance and recharge ledger.
- **ChatThread, ChatMessage, ChatGroup, ChatGroupMember, ChatGroupJoinRequest, GroupMessage**: messaging data.
- **NoticeRead, ShortLink, ClickRecords, ClickSummary**: notification cursors and utility/aggregate records.

## Success Criteria

### Measurable Outcomes

- **SC-001**: All frontend API calls used by the unchanged Vue app succeed against the new service in an integration smoke test.
- **SC-002**: 100% of the documented endpoints preserve their method, path, auth requirement, and top-level response shape.
- **SC-003**: Existing users can log in without re-registering and without resetting passwords.
- **SC-004**: Core flows (login, publish note, comment, like, follow, product browse, cart, order, pay, chat, notice) each complete end to end.
- **SC-005**: P95 latency for public read endpoints is under 300 ms under the project's standard local load.
- **SC-006**: Critical auth and commerce paths have at least 70% line coverage.
- **SC-007**: No destructive database migration is required; the original MySQL data remains usable by both backends during comparison.

## Assumptions

- The Vue frontend remains the compatibility oracle and is not changed.
- The compatibility surface is the union of the original backend controllers and frontend-declared API functions; any frontend function that the original backend did not expose is treated as a compatibility gap to implement in Go.
- The existing MySQL `redbook` database, Redis instance, and MQ topic names are reused.
- Verification-code sending remains a mock returning the code in the response, matching the original.
- Alipay recharge remains a simulation and does not call a real payment provider.
- Dify, OSS, and RocketMQ are optional at startup where the original used placeholders or could degrade.
- Java-specific infrastructure such as Caffeine and Canal is replaced by equivalent or simpler in-process cache and explicit invalidation unless parity requires otherwise.

