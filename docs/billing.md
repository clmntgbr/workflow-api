# Billing

## Overview

Billing integrates **Stripe** for subscriptions, quotas, invoices, and the customer portal. Plans define limits; quotas gate resource creation across the builder.

## HTTP routes

### Plans (public)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/plans` | No | List available plans |

### Subscription & quota

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/subscriptions` | Current subscription |
| `GET` | `/api/quota` | Usage vs limits for active project |
| `POST` | `/api/subscriptions` | Create checkout session |
| `POST` | `/api/subscriptions/preview` | Preview proration |
| `GET` | `/api/subscriptions/portal` | Stripe billing portal URL |

### Invoices

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/invoices` | List invoices for user |

All subscription/invoice routes require authentication.

## Quotas

Enforced on create paths (workflows, endpoints, steps, runs, etc.). HTTP handlers map quota errors via `respondQuotaError`:

| Error | HTTP |
|-------|------|
| No subscription | `403` |
| Quota exceeded (workflow, endpoint, step, run, project, …) | `403` with message |
| Active project required | `400` |

See `internal/application/command/quota/` and `test/quota/`.

## Stripe webhooks

Inbound events update subscription and invoice state — see [Webhooks](webhooks.md#stripe-billing).

Handled event types include:

- `checkout.session.completed`
- `customer.subscription.updated` / `deleted`
- `invoice.payment_succeeded` / `payment_failed`

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `subscription_handler.go`, `plan_handler.go`, `invoice_handler.go`, `billing_webhook_handler.go` |
| Commands | `internal/application/command/subscription/` |
| Queries | `internal/application/query/subscription/` |
| Stripe adapter | `internal/infrastructure/stripe/` |
| Test exports | `billing_webhook_test_exports.go` |

## Tests

- `internal/interfaces/http/handler/test/subscription/`
- `internal/interfaces/http/handler/test/plan/`
- `internal/interfaces/http/handler/test/invoice/`
- `internal/interfaces/http/handler/test/billing_webhook/`
- `internal/interfaces/http/handler/test/quota/`
