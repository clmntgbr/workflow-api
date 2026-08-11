# Go API Template

Go/Fiber API template with **Clean Architecture**, **CQRS**, **Outbox pattern**, **RabbitMQ**, **PostgreSQL**, and **Clerk** authentication.

## Features

- CQRS: command / query / event handlers under `internal/application`
- Domain events (`user.created.v1`) + transactional outbox
- RabbitMQ worker (outbox relay + consumer + handler registry)
- Clerk JWT auth + webhook sync (Svix)
- Docker Compose local stack (API, worker, Postgres, RabbitMQ, ngrok)

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| HTTP | Fiber v3 |
| ORM | GORM + PostgreSQL |
| Auth | Clerk (JWKS + webhooks) |
| Messaging | RabbitMQ (topic exchange) |
| CLI | Cobra |
| Dev | Docker, Air, golangci-lint |

## Architecture

```
cmd/
  api/          HTTP server
  worker/       Outbox relay + RabbitMQ consumer
  cli/          Migrations
internal/
  domain/                 Aggregates, events, ports, paginate
  application/
    command/              Write-side handlers
    query/                Read-side handlers
    event/                Async event handlers
    registry/             Event handler registry
  infrastructure/
    persistence/write|read|outbox
    messaging/rabbitmq
    clerk|config
  interfaces/http/        Handlers, middleware, DTO, presenter
```

### Sync command + async event flow

```
HTTP / webhook
  → CreateUser command (same DB transaction)
      → save User aggregate
      → StoreEvents(outbox)
  → 201 response

worker (outbox relay)
  → poll unpublished outbox rows
  → publish to RabbitMQ
  → mark published_at

worker (consumer)
  → HandlerRegistry.Dispatch(event type)
  → UserCreatedHandler / NotifyUserOnCreatedHandler
  → UserUpdatedHandler / UserDeletedHandler
```

Supported domain events: `user.created.v1`, `user.updated.v1`, `user.deleted.v1` (all written to outbox on the matching command).

### Idempotence + retry / DLQ

- Each domain event has a stable `eventId` (set when recorded on the aggregate).
- Handlers are wrapped with dedup on `(event_id, handler_name)` via `processed_events`.
- RabbitMQ topology: `domain.events` → `domain.events.retry` (TTL) → back to main; non-retryable / max attempts → `domain.events.dlq`.
- Never `Nack(requeue=true)` — retries go through the TTL retry queue.

If queue declare fails after changing args, delete the old queues (or recreate the RabbitMQ volume) then restart the worker.

## Getting Started

```bash
cp .env.dist .env
# fill Clerk keys

make dev
make migrate
```

| Service | Host port | Notes |
|---|---|---|
| API | `4000` | Air hot reload |
| Worker | — | relay + consumer |
| Postgres | `9543` | |
| RabbitMQ | `5672` / UI `15672` | user/password from `.env` |
| ngrok | `4040` | webhook tunnel |

## Makefile

| Command | Description |
|---|---|
| `make dev` | Start compose.dev stack (build) |
| `make dev-down` | Stop stack |
| `make dev-logs` | Follow API + worker logs |
| `make worker-logs` | Follow worker logs only |
| `make migrate` | Apply SQL migrations + schema drift check |
| `make migrate-check` | Fail if persistence models ≠ DB columns |
| `make migrate-status` | Goose migration status |
| `make lint` | golangci-lint --fix |
| `make shell` | Shell into API container |

## API

### Health

- `GET /livez`, `/readyz`, `/startupz`

### Protected

- `GET /api/users/me` — Bearer Clerk JWT (query side / read repository)

### Webhooks

- `POST /webhooks/clerk` — Svix-verified; `user.created` / `updated` / `deleted`

## CQRS notes

- Commands write through aggregates (`NewUser` records `UserCreated`) and persist events in `outbox_events` in the **same** GORM transaction.
- Queries use `UserReadRepository` (SQL projection / `UserView`), not domain mutation.
- Event types are versioned (`user.created.v1`).
- One worker process runs outbox relay + consumer; scale the worker horizontally as needed.
- Start with a single queue (`domain.events`); split queues later if throughput requires it.

## Environment

See `.env.dist`. Required extras for messaging:

- `RABBITMQ_URL`
- `RABBITMQ_EXCHANGE` (default `domain.events`)
- `RABBITMQ_QUEUE` (default `domain.events`)
- `RABBITMQ_ROUTING_KEY` (default `user.#`)
- `OUTBOX_POLL_INTERVAL` (default `2s`)
- `WORKER_CONCURRENCY` (default `4`)

## Extending

1. Add aggregate under `internal/domain/<name>/` (`entity.go`, `events.go`, `repository.go`)
2. Add command/query handlers under `internal/application/...`
3. Implement write/read repos under `internal/infrastructure/persistence/`
4. Register event handlers in `cmd/worker/di`
5. Wire HTTP in `cmd/api/di` + `routes.go`
6. Add a SQL migration in `migrations/` and a persistence model under `internal/infrastructure/persistence/`
7. Run `make migrate` (applies SQL + fails on model/DB drift)

## Architecture rules

Conventions for CQRS, outbox, events, and RabbitMQ live in [`.cursor/rules/architecture.mdc`](.cursor/rules/architecture.mdc).
