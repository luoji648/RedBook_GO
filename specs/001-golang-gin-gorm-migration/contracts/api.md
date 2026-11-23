# API Contract

This contract is the compatibility oracle for the Go backend. The Vue frontend must work against these endpoints without changes.

## Response Envelope

Normal business responses use:

```json
{
  "success": true,
  "errorMsg": null,
  "data": {},
  "total": 100
}
```

- `success` is `true` on success and `false` on business errors.
- `errorMsg` is populated on business failures.
- `data` is endpoint-specific.
- `total` is used by paginated list endpoints.

Exceptions:

- `GET /like/{receiverId}` returns a raw object with `clicked` and `cnt` fields, not the envelope.
- `GET /ai/note/{noteId}/stream` returns `text/event-stream`.
- `GET /s/{code}` returns an HTTP redirect.
- Unauthenticated protected requests return HTTP `401` without the envelope.
- Validation failures may return HTTP `400` with `success:false` and a compatible message.

## Authentication

- `Auth` means a valid `Authorization: Bearer <token>` header is required.
- `Public` means no token is required.
- Public routes from the original interceptor configuration are: `/user/code`, `/user/login`, `/user/login/password`, `/note/detail/**`, `/note/recommend`, `/note/related/**`, `/note/comment/tree/**`, `/note/user/**`, `/user/public/**`, `/group/by-owner/**`, `/like/user/**`, `/collect/user/**`, `/follow/public/**`, and `/product/list`.
- `GET /product/{id}` and `POST /product/list` are public because the original only protects selected product paths.

## Account

| Method | Path | Auth | Request | Notes |
| --- | --- | --- | --- | --- |
| POST | `/user/code` | Public | form `phone` | mock SMS code in `data` |
| POST | `/user/login` | Public | JSON phone + code | returns JWT |
| POST | `/user/login/password` | Public | JSON phone + password | returns JWT |
| POST | `/user/logout` | Auth | header | blacklists JWT |
| POST | `/user/sign` | Auth | | daily sign-in |
| GET | `/user/sign/count` | Auth | | current month count |
| GET | `/user/me` | Auth | | account + profile |
| GET | `/user/public/{userId}` | Public | | profile without phone |
| PUT | `/user/profile` | Auth | JSON nickname/icon | |
| GET | `/user/info` | Auth | | extended profile |
| PUT | `/user/info` | Auth | JSON extended fields | |
| PUT | `/user/password` | Auth | JSON old/new password | invalidates old JWT |

## Notes

| Method | Path | Auth | Request | Notes |
| --- | --- | --- | --- | --- |
| POST | `/note/save` | Auth | JSON note DTO | create or update |
| POST | `/note/publish/{id}` | Auth | | publish draft |
| DELETE | `/note/{id}` | Auth | | delete own note |
| GET | `/note/detail/{id}` | Public | | detail with viewer state |
| GET | `/note/my` | Auth | `current`, `size` | own notes |
| GET | `/note/feed` | Auth | `maxTime`, `offset`, `size` | follow feed cursor |
| GET | `/note/recommend` | Public | `current`, `size` | |
| GET | `/note/related/{id}` | Public | | related notes |
| GET | `/note/user/{userId}` | Public | `current`, `size` | visibility filtered |

## Comments

| Method | Path | Auth | Request | Notes |
| --- | --- | --- | --- | --- |
| POST | `/note/comment/{noteId}` | Auth | JSON parent/reply/content | |
| GET | `/note/comment/tree/{noteId}` | Public | | no-store |
| DELETE | `/note/comment/{commentId}` | Auth | | delete own comment |

## Likes

| Method | Path | Auth | Request | Notes |
| --- | --- | --- | --- | --- |
| GET | `/like/{receiverId}` | Auth | | raw `{clicked,cnt}` |
| POST | `/like/{receiverId}` | Auth | | like |
| DELETE | `/like/{receiverId}` | Auth | | unlike |
| GET | `/like/my` | Auth | `current`, `size` | compatibility endpoint |
| GET | `/like/user/{userId}` | Public | `current`, `size` | compatibility endpoint |

## Collects and Follows

| Method | Path | Auth | Request | Notes |
| --- | --- | --- | --- | --- |
| POST | `/collect/{noteId}` | Auth | | |
| DELETE | `/collect/{noteId}` | Auth | | |
| GET | `/collect/my` | Auth | `current`, `size` | |
| GET | `/collect/user/{userId}` | Public | `current`, `size` | visibility enforced |
| POST | `/follow/{userId}` | Auth | | |
| DELETE | `/follow/{userId}` | Auth | | |
| GET | `/follow/followers` | Auth | `current`, `size` | |
| GET | `/follow/following` | Auth | `current`, `size` | |
| GET | `/follow/public/{userId}/followers` | Public | `current`, `size` | |
| GET | `/follow/public/{userId}/following` | Public | `current`, `size` | |

## Products and Footprints

| Method | Path | Auth | Request | Notes |
| --- | --- | --- | --- | --- |
| GET | `/product/my` | Auth | `current`, `size` | seller products |
| GET | `/product/list` | Public | `current`, `size` | published products |
| GET | `/product/{id}` | Public | | product detail |
| POST | `/product/save` | Auth | JSON product DTO | create/update |
| PUT | `/product/{id}/stock` | Auth | JSON `stock` | published only |
| POST | `/product/admin` | Auth | JSON product DTO | legacy alias |
| POST | `/product/footprint/record` | Auth | form `productId` | upsert viewed_at |
| GET | `/product/footprint/my` | Auth | `current`, `size` | |

## Cart

| Method | Path | Auth | Request | Notes |
| --- | --- | --- | --- | --- |
| POST | `/cart/add` | Auth | form `productId`, `quantity` | |
| GET | `/cart/my` | Auth | | |
| PUT | `/cart/{cartId}` | Auth | query `quantity` | |
| DELETE | `/cart/{cartId}` | Auth | | |

## Orders

| Method | Path | Auth | Request | Notes |
| --- | --- | --- | --- | --- |
| POST | `/order/create` | Auth | JSON optional `userCouponId` | from cart |
| GET | `/order/direct-preview` | Auth | query `productId`, `quantity`, `fallbackSellerUserId` | |
| POST | `/order/create-direct` | Auth | JSON product/quantity/coupon | |
| GET | `/order/my` | Auth | `current`, `size` | |
| GET | `/order/detail/{id}` | Auth | | |
| POST | `/order/pay/{id}` | Auth | JSON `payPassword` | simulated `123456` |
| POST | `/order/refund/{id}` | Auth | | |
| DELETE | `/order/{id}` | Auth | | delete non-pending record |
| POST | `/order/{id}/close` | Auth | | alias `/order/close/{id}` |

## Wallet and Coupons

| Method | Path | Auth | Request | Notes |
| --- | --- | --- | --- | --- |
| GET | `/wallet/balance` | Auth | | `{balanceCent}` |
| POST | `/wallet/recharge/alipay` | Auth | JSON `amountCent` | simulated |
| GET | `/coupon/my` | Auth | `current`, `size` | available coupons |
| POST | `/coupon/claim-follow` | Auth | JSON `productId` | follow-seller coupon |

## Messaging

| Method | Path | Auth | Request | Notes |
| --- | --- | --- | --- | --- |
| GET | `/chat/unread-total` | Auth | | private + group + notice |
| GET | `/chat/threads` | Auth | `current`, `size` | |
| POST | `/chat/{threadId}/read` | Auth | | |
| GET | `/chat/peer-thread` | Auth | query `peerUserId` | ensure thread |
| GET | `/chat/{threadId}/messages` | Auth | `current`, `size` | |
| POST | `/chat/send` | Auth | form `toUserId`, `content` | |
| POST | `/group/create` | Auth | JSON name/joinMode/avatar | |
| GET | `/group/by-owner/{ownerUserId}` | Public | | |
| GET | `/group/my` | Auth | | |
| POST | `/group/{groupId}/join` | Auth | | |
| POST | `/group/join-request/{requestId}/handle` | Auth | form `approve` | |
| GET | `/group/{groupId}/meta` | Auth | | |
| GET | `/group/{groupId}/messages` | Auth | `current`, `size` | |
| POST | `/group/{groupId}/send` | Auth | form `content` | |
| POST | `/group/{groupId}/read` | Auth | | |
| GET | `/group/{groupId}/detail` | Auth | | |
| POST | `/group/{groupId}/kick` | Auth | form `targetUserId` | |
| PUT | `/group/{groupId}/profile` | Auth | JSON name/avatar | |

## Notifications, Sharing, OSS, and AI

| Method | Path | Auth | Request | Notes |
| --- | --- | --- | --- | --- |
| GET | `/notice/interactions` | Auth | `current`, `size`, `category` | |
| POST | `/notice/read` | Auth | form `category` | |
| POST | `/share/note/{noteId}` | Auth | | generate short link |
| POST | `/short-link/generate` | Auth | form `noteId` | |
| GET | `/s/{code}` | Public | | redirect |
| POST | `/oss/presign` | Auth | form `ext`, `contentType` | |
| POST | `/oss/upload` | Auth | multipart `file` | |
| GET | `/ai/note/{noteId}/stream` | Auth | query `mode`, `question` | SSE |

