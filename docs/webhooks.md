# Webhooks

## Overview

External providers push lifecycle events to the API. Signatures are verified in middleware before handlers run.

| Provider | Path | Middleware |
|----------|------|------------|
| Clerk | `POST /webhooks/clerk` | Svix (`UserWebhookMiddleware`) |
| Stripe | `POST /webhooks/stripe` | Stripe signing secret (`BillingWebhookMiddleware`) |

No Clerk JWT on webhook routes.

## Clerk (user lifecycle)

Handler: `UserWebhookHandler.Execute`

Verified payload is set on `c.Locals("payload", dto.ClerkEvent)` before the handler runs.

| Event | Action |
|-------|--------|
| `user.created` | Create user if not exists (`201`) |
| `user.updated` | Update profile (`204`) |
| `user.deleted` | Delete by Clerk ID (`204`) |
| Unknown | `200` (ignored) |

### Idempotence

`user.created` skips creation if the Clerk ID already exists.

### Errors

| Case | HTTP |
|------|------|
| Invalid JSON | `400` |
| User not found on update | `404` |
| Handler failure | `500` |

## Stripe (billing)

Handler: `BillingWebhookHandler.Execute`

Payload: `c.Locals("payload", stripe.Event)`.

| Event | Action |
|-------|--------|
| `checkout.session.completed` | Link Stripe customer/subscription to user |
| `customer.subscription.updated` | Sync subscription state |
| `customer.subscription.deleted` | Mark subscription deleted |
| `invoice.payment_succeeded` | Upsert invoice + payment succeeded command |
| `invoice.payment_failed` | Upsert invoice + payment failed command |
| Unknown | `200` (logged, ignored) |

### Special errors

| Case | HTTP |
|------|------|
| Subscription not linked yet | `409` |
| Other handler errors | `500` |

Invoice mapping helpers are tested via `billing_webhook_test_exports.go`.

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `user_webhook_handler.go`, `billing_webhook_handler.go` |
| Middleware | `internal/interfaces/http/middleware/` |
| Commands | `internal/application/command/user/`, `…/subscription/` |

## Tests

- `internal/interfaces/http/handler/test/user_webhook/`
- `internal/interfaces/http/handler/test/billing_webhook/` (HTTP + `invoice_helpers_test.go`)

## Security

- Never trust raw webhook bodies — middleware validates signatures first.
- Handlers validate unmarshalled DTOs (`validate.Struct` on Clerk payloads).
