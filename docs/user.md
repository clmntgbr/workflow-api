# User & authentication

## Overview

Users are provisioned from **Clerk** via webhooks. The API authenticates requests with a **Clerk JWT** (Bearer token). The authenticated user may have an **active project** used to scope builder resources.

## Authentication

| Route group | Middleware |
|-------------|------------|
| `/api/*` (except plans) | `AuthenticateMiddleware` — validates JWT, loads user |
| `/webhooks/clerk` | Svix signature verification |
| `/webhooks/stripe` | Stripe webhook signature |

## HTTP routes

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/users/me` | Current user profile and active project |
| `PUT` | `/api/users/me/active-project` | Set active project (`projectId` UUID) |

## Active project

Most builder endpoints require `httpctx.GetActiveProjectID()`:

- Missing active project → `400` (`Active project is required`)
- Unauthenticated → `401` (`Unauthorized`)

Setting active project validates membership (`project not found`, `user is not a member of the project`).

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `internal/interfaces/http/handler/user_handler.go` |
| Commands | `internal/application/command/user/` |
| Queries | `internal/application/query/user/` |
| Domain | `internal/domain/user/` |
| Clerk adapter | `internal/infrastructure/clerk/` |
| Webhook | [Webhooks](webhooks.md#clerk-user-lifecycle) |

## Events

- `user.created.v1`, `user.updated.v1`, `user.deleted.v1` (from Clerk webhooks + domain)

## Tests

`internal/interfaces/http/handler/test/user/`
