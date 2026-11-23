# Data Model

The Go service reuses the existing MySQL `redbook` schema. GORM structs map to these tables with `gorm:"column:<snake_case>"` tags. Money is stored in integer cents. Timestamps map to `time.Time` or `*time.Time` where nullable.

## Account

### tb_user

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | auto increment primary key |
| phone | string | unique |
| password | string | legacy `salt@md5(...)` |
| nick_name | string | |
| icon | string | avatar URL |
| token_version | int32 | increments to invalidate old JWTs |
| create_time | time | |
| update_time | time | |

### tb_user_info

| Field | Type | Notes |
| --- | --- | --- |
| user_id | int64 | primary key, references `tb_user.id` |
| city | string | |
| introduce | string | |
| fans | int32 | |
| followee | int32 | |
| gender | bool | stored as tinyint |
| birthday | *time.Time | |
| credits | int32 | |
| level | int32 | |
| collect_public | bool | |
| like_public | bool | |
| create_time | time | |
| update_time | time | |

## Content

### tb_note

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| user_id | int64 | author |
| title | string | |
| content | string | long text |
| type | int32 | 0 image, 1 video |
| status | int32 | 0 draft, 1 published, 2 offline |
| visibility | int32 | 0 public, 1 followers, 2 private |
| like_count | int32 | legacy display field |
| collect_count | int32 | |
| comment_count | int32 | |
| create_time | time | |
| update_time | time | |

### tb_note_media

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| note_id | int64 | references note |
| url | string | |
| sort_order | int32 | |
| media_type | int32 | 0 image, 1 video |

### tb_note_product

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| note_id | int64 | |
| product_id | int64 | |
| sort_order | int32 | unique `(note_id, product_id)` |

## Social

### tb_note_comment

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| note_id | int64 | |
| user_id | int64 | author |
| parent_id | *int64 | root comment id for replies |
| reply_to_user_id | *int64 | |
| content | string | |
| status | int32 | 0 ok, 1 deleted |
| create_time | time | |

### tb_note_collect

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| user_id | int64 | unique with note |
| note_id | int64 | |
| create_time | time | |

### tb_user_follow

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| follower_id | int64 | unique with followee |
| followee_id | int64 | |
| create_time | time | |

### click_records

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | snowflake id |
| receiver_id | int64 | note id |
| clicker_id | int64 | user id |
| click_type | int32 | like or unlike |
| created | time | |

### click_summary

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | snowflake id |
| receiver_id | int64 | unique |
| like_count | int64 | |
| created | time | |
| modified | time | |

## Commerce

### tb_product

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| title | string | |
| cover | string | |
| price_cent | int64 | |
| stock | int32 | |
| status | int32 | 1 published |
| seller_id | *int64 | |
| create_time | time | |
| update_time | time | |

### tb_cart

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| user_id | int64 | unique with product |
| product_id | int64 | |
| quantity | int32 | |
| create_time | time | |
| update_time | time | |

### tb_order

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| user_id | int64 | |
| total_cent | int64 | |
| pay_cent | *int64 | actual paid amount |
| discount_cent | int64 | coupon deduction |
| user_coupon_id | *int64 | |
| status | int32 | 0 created, 1 paid, 2 cancelled, 3 refunded |
| pay_time | *time.Time | |
| seller_settled | int32 | |
| from_cart | int32 | |
| create_time | time | |
| update_time | time | |

### tb_order_item

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| order_id | int64 | |
| product_id | int64 | |
| quantity | int32 | |
| price_cent | int64 | snapshot of product price |

### tb_user_coupon

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| user_id | int64 | |
| product_id | *int64 | optional target product |
| title | string | |
| discount_cent | int64 | |
| min_order_cent | int64 | |
| status | int32 | 0 unused, 1 used, 2 expired |
| expire_time | time | |
| create_time | time | |

### tb_product_footprint

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| user_id | int64 | unique with product |
| product_id | int64 | |
| viewed_at | time | |

### tb_user_wallet

| Field | Type | Notes |
| --- | --- | --- |
| user_id | int64 | primary key |
| balance_cent | int64 | |
| create_time | time | |
| update_time | time | |

### tb_wallet_recharge

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| user_id | int64 | |
| out_trade_no | string | unique |
| amount_cent | int64 | |
| alipay_trade_no | *string | |
| status | int32 | 0 pending, 1 success, 2 closed |
| create_time | time | |
| pay_time | *time.Time | |

## Messaging

### tb_chat_thread

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| user_low | int64 | lower user id |
| user_high | int64 | higher user id |
| last_msg_time | *time.Time | |
| user_low_read_msg_id | *int64 | |
| user_high_read_msg_id | *int64 | |

### tb_chat_message

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| thread_id | int64 | |
| sender_id | int64 | |
| content | string | |
| read_flag | int32 | |
| create_time | time | |

### tb_chat_group

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| owner_id | int64 | |
| name | string | |
| avatar | *string | |
| join_mode | int32 | 0 open, 1 approval |
| create_time | time | |
| last_msg_time | *time.Time | |

### tb_chat_group_member

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| group_id | int64 | unique with user |
| user_id | int64 | |
| role | int32 | 0 member, 1 owner |
| member_status | int32 | 0 active, 1 removed |
| create_time | time | |
| last_read_msg_id | *int64 | |

### tb_chat_group_join_request

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| group_id | int64 | unique with applicant |
| applicant_id | int64 | |
| status | int32 | 0 pending, 1 approved, 2 rejected |
| create_time | time | |
| update_time | time | |

### tb_group_message

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| group_id | int64 | |
| sender_id | int64 | |
| content | string | |
| create_time | time | |

## Utility

### tb_notice_read

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| user_id | int64 | unique with category |
| category | string | `like_collect`, `follow`, `comment` |
| last_read_time | time | |

### tb_short_link

| Field | Type | Notes |
| --- | --- | --- |
| id | int64 | |
| short_uri | string | unique |
| long_url | *string | null when pooled |
| valid_date | *time.Time | |
| create_time | time | |
| del_flag | int32 | |

## Relationship Summary

- `User` 1:1 `UserInfo`, 1:N `Note`, `Cart`, `Order`, `UserCoupon`, `UserWallet`, `ChatMessage`, `GroupMessage`, `Follow`.
- `Note` 1:N `NoteMedia`, `NoteProduct`, `NoteComment`, `NoteCollect`, and receives `ClickSummary`.
- `Product` N:M `Note` through `NoteProduct`, 1:N `Cart`, `OrderItem`, `ProductFootprint`, optional `UserCoupon`.
- `Order` 1:N `OrderItem`.
- `ChatThread` 1:N `ChatMessage`.
- `ChatGroup` 1:N members, messages, and join requests.

Legacy tables `tb_note_like`, `tb_note_like_agg`, and `tb_like_outbox` remain in the schema but are not primary migration targets because the current like implementation uses `click_records` and `click_summary`.

