# Quickstart Validation Guide

This guide validates the migrated Go backend without changing the Vue frontend.

## Prerequisites

- Go 1.23+
- MySQL 8 running with the existing `redbook` database initialized by `redbook-backend/src/main/resources/db/schema.sql` and applicable migrations
- Redis running on the configured host/port
- Optional: RocketMQ for like-event consumers
- Optional: Aliyun OSS credentials and Dify endpoint for integration endpoints

## Setup

```bash
cd redbook-go
go mod tidy
go build ./...
```

Copy and edit configuration, for example `config/config.example.yaml` to `config/config.yaml`, then set database, Redis, JWT, MQ, OSS, and Dify values.

## Run

```bash
go run ./cmd/server
```

The service listens on port `8081` by default to match the frontend proxy.

Start the frontend separately from its existing directory:

```bash
cd ../redbook-frontend
npm install
npm run dev
```

## Smoke Scenarios

1. Open `http://localhost:5173` and register or log in through the unchanged UI.
2. Publish a note with an uploaded image, then reload the public and personal note pages.
3. Like, collect, comment, and follow another user; verify counters and privacy.
4. Create a product, add it to the cart, create and pay an order with password `123456`.
5. Check wallet balance, claim a follow coupon, and test refund.
6. Open private chat and group chat with two accounts; verify unread counts and read cursors.
7. Generate a short link and open its `/s/{code}` URL.
8. Call the AI stream endpoint with `mode=SUMMARY_POST`, `SUMMARY_COMMENTS`, and `QA` using an SSE client.

## Automated Checks

```bash
go test ./...
```

Run the contract smoke suite once handlers exist:

```bash
go test ./tests/integration -run TestApiParity -v
```

## Expected Outcomes

- All frontend API calls return the documented success envelope and no frontend code changes are required.
- Public routes work without a token; protected routes return `401` without a token.
- Legacy accounts log in with the existing password scheme.
- Order payment, wallet, coupon, inventory, and refund transitions remain consistent under concurrent requests.

