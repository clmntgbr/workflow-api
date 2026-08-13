# Workflow API

Backend for a visual HTTP workflow builder. Teams define reusable **endpoints**, compose them into **workflows** as a graph of **steps** and **connections**, then collaborate in realtime across organizations.

This service is the source of truth for that graph: it owns persistence, auth, ordering (`executionOrder` / `treeIndex`), and event fan-out.

## What it does

- **Organizations** — multi-tenant workspaces, members, active organization on the user
- **Workflows** — named graphs scoped to an organization (soft delete)
- **Endpoints** — reusable HTTP templates (method, URL, headers, query, body, retries)
- **Steps** — endpoint snapshot on a canvas; position drives execution order; connections drive branches
- **Connections** — directed edges between steps (`sourceStepId` → `targetStepId`)
- **Realtime** — Centrifugo events so the canvas stays in sync across clients

Resources are scoped to the caller’s **active organization**. Shared workflow URLs that belong to another org the user is a member of return `409 WRONG_ORGANIZATION` so the client can switch org.

## Tech stack

| Layer | Choice |
|---|---|
| Language | Go 1.25 |
| HTTP | Fiber v3 |
| Persistence | PostgreSQL + GORM (writes) / SQL views (reads) |
| Migrations | Goose (`migrations/*.sql`) |
| Auth | Clerk (JWT / JWKS + Svix webhooks) |
| Messaging | RabbitMQ (topic exchange, retry + DLQ) |
| Realtime | Centrifugo |
| CLI | Cobra |
| Local | Docker Compose, Air, golangci-lint |

Architecture: **Clean Architecture + CQRS**. HTTP handlers call command/query handlers directly. Domain events go through a **transactional outbox**, then a **worker** publishes to RabbitMQ and dispatches handlers (dedup + Centrifugo).

## Layout

```
cmd/
  api/       HTTP server
  worker/    Outbox relay + RabbitMQ consumer
  cli/       Goose migrations + schema drift check
internal/
  domain/            Aggregates (user, organization, workflow, endpoint, step, connection)
  application/       command / query / event / realtime / registry
  infrastructure/    persistence, rabbitmq, clerk, centrifugo, config
  interfaces/http/   handlers, DTOs, presenters, validation
migrations/          Schema source of truth
```

Conventions: [`.cursor/rules/architecture.mdc`](.cursor/rules/architecture.mdc).

## Domain events

Versioned types on the bus (`*.v1`). Realtime types drop the version (`entity.action`).

| Domain | Realtime | When |
|---|---|---|
| `workflow.created.v1` / `updated.v1` / `deleted.v1` | `workflow.*` | Workflow CRUD |
| `endpoint.created.v1` / `updated.v1` / `deleted.v1` | `endpoint.*` | Endpoint CRUD |
| `step.created.v1` / `updated.v1` / `deleted.v1` | `step.*` | Step CRUD / move |
| `connection.created.v1` / `deleted.v1` | `connection.*` | Link / unlink steps |

Plus user and organization events (`user.*`, `organization.*`).

On step create/move and connection create/delete, the API **synchronously** recalculates `index`, `executionOrder`, and `treeIndex` from canvas position and the directed graph. Deleting a step also deletes its connections.

## Getting started

```bash
cp .env.dist .env
# fill Clerk, Postgres, RabbitMQ, Centrifugo

make dev
make migrate
```

| Service | Port | Notes |
|---|---|---|
| API | `4000` | Air hot reload |
| Worker | — | Outbox relay + consumer |
| Postgres | `9543` | |
| RabbitMQ | `5672` / UI `15672` | Credentials from `.env` |
| ngrok | `4040` | Clerk webhook tunnel |

## Makefile

| Command | Description |
|---|---|
| `make dev` | Start compose.dev stack |
| `make dev-down` | Stop stack |
| `make dev-logs` | Follow API + worker |
| `make migrate` | Apply SQL + fail on model/DB drift |
| `make migrate-check` | Schema drift check only |
| `make lint` | golangci-lint --fix |
| `make shell` | Shell into the API container |

## HTTP surface

Health: `GET /livez`, `/readyz`, `/startupz`

Auth: Bearer Clerk JWT on `/api/*`. Webhook: `POST /webhooks/clerk` (Svix).

| Area | Routes |
|---|---|
| User | `GET /api/users/me`, `PUT /api/users/me/active-organization` |
| Organizations | `GET/POST /api/organizations`, `GET/PUT/DELETE /api/organizations/:id`, members, activate |
| Workflows | `GET/POST /api/workflows`, `GET/PUT/DELETE /api/workflows/:id` |
| Endpoints | `GET/POST /api/endpoints`, `GET/PUT/DELETE /api/endpoints/:id` |
| Steps | nested under `/api/workflows/:workflowId/steps` (CRUD + `PUT .../position`) |
| Connections | nested under `/api/workflows/:workflowId/connections` |
| Realtime | `GET /api/realtime/connection` |

## Messaging

- Commands persist the aggregate **and** outbox rows in the same transaction.
- Worker polls unpublished rows, publishes an envelope (`eventId`, `type`, `aggregateId`, `occurredAt`, `payload`), then sets `published_at`.
- Dedup key: `(event_id, handler_name)` in `processed_events`.
- Topology: main queue → TTL retry → main; poison / non-retryable → DLQ. Never `Nack(requeue=true)`.
