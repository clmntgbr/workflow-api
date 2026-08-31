# Workflow API

[![Tests](https://github.com/clmntgbr/workflow-api/actions/workflows/tests.yml/badge.svg)](https://github.com/clmntgbr/workflow-api/actions/workflows/tests.yml)
[![codecov](https://codecov.io/gh/clmntgbr/workflow-api/graph/badge.svg)](https://codecov.io/gh/clmntgbr/workflow-api)

Backend for a visual HTTP workflow builder. Teams define reusable **endpoints**, compose them into **workflows** as a graph of **steps** and **connections**, inject **variables**, validate responses with **assertions**, and run workflows manually or on a schedule — with realtime collaboration across **projects**.

This service is the source of truth for that graph and its execution: persistence, auth, ordering (`executionOrder` / `treeIndex`), orchestration, quotas, billing, and event fan-out.

## What it does

### Builder

- **Projects** — multi-tenant workspaces, members, active project on the user
- **Workflows** — named graphs scoped to a project (soft delete, activate/deactivate, scheduling)
- **Endpoints** — reusable HTTP templates (method, URL, headers, query, body, retries) + OpenAPI import
- **Steps** — nodes on the canvas; three types:
  - **`http`** — snapshot of an endpoint (default)
  - **`delay`** — pause between steps (`delayDurationSeconds`), no HTTP call
  - **`condition`** — boolean expression (`expression`, e.g. `{{plan}} == "premium"`) evaluated at runtime; exactly two outgoing connections with `branch: "true"` / `"false"`
- **Connections** — directed edges between steps (`sourceStepId` → `targetStepId`); optional `branch` on edges from condition steps
- **Variables** — static values or JSON-path extracts from previous HTTP responses
- **Assertions** — post-response validation rules per HTTP step (status, headers, body)
- **Activity logs** — human-readable audit trail per workflow (builder actions + runs)

Canvas position drives execution order; connections drive branches and skip logic. A **delay step with no connections** (orphan on the canvas) is ignored at runtime.

Resources are scoped to the caller’s **active project**. Shared workflow URLs that belong to another project the user is a member of return `409 WRONG_ORGANIZATION` so the client can switch project.

### Execution

- **Workflow runs** — manual (`POST /workflows/:id/start`), API, schedule, webhook
- **Orchestration** — worker advances the graph after each step run succeeds/fails/skips
- **HTTP steps** — dedicated **executor** binary consumes `stepRun.queued`, calls the target API, runs assertions, records **insights** (timings)
- **Delay steps** — always `waiting` + `resumeAt`; worker polls (`STEP_RUN_WAITING_POLL_INTERVAL`, default `1s`) and resumes when due — executor handles HTTP only
- **Condition steps** — evaluated inline by the worker (no executor); taken branch is activated, the other branch subtree is skipped
- **Cancellation** — `POST /workflows/:id/stop` cancels the in-progress run and all non-terminal step runs (including `waiting`)
- **Scheduling** — **scheduler** binary claims due workflows and starts runs

Delay step runs emit the same realtime events as HTTP steps (`stepRun.started`, `stepRun.succeeded`). The API exposes `startedAt`, `finishedAt`, and `resumeAt` on step runs.

### Collaboration & billing

- **Realtime** — Centrifugo events so the canvas and run progress stay in sync
- **Subscriptions** — Stripe plans, quotas, invoices, billing portal

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
| Billing | Stripe |
| CLI | Cobra |
| Local | Docker Compose, Air, golangci-lint |

Architecture: **Clean Architecture + CQRS**. HTTP handlers call command/query handlers directly. Domain events go through a **transactional outbox**, then a **worker** publishes to RabbitMQ and dispatches handlers (dedup + Centrifugo + activity logs).

Conventions: [`.cursor/rules/architecture.mdc`](.cursor/rules/architecture.mdc).  
Product overview (French): [`docs/project-overview.md`](docs/project-overview.md).

## Layout

```
cmd/
  api/        HTTP server
  worker/     Outbox relay + RabbitMQ consumer + waiting delay poller
  executor/   HTTP step run execution
  scheduler/  Scheduled workflow runs
  cli/        Goose migrations + schema drift check
internal/
  domain/            Aggregates (user, project, workflow, endpoint, step, …)
  application/       command / query / event / realtime / registry
  infrastructure/    persistence, rabbitmq, clerk, centrifugo, stripe, config
  interfaces/http/
    handler/         HTTP handlers (production code)
      test/          Handler tests (one subfolder per resource)
        endpoint/    Endpoint handler tests (mocked command/query)
        quota/       Quota error mapping tests
    testutil/        Shared HTTP test helpers (Fiber app, JSON, multipart)
    dto/             Request/response DTOs, presenters, validation
migrations/          Schema source of truth
.github/workflows/   CI (tests + Codecov upload)
```

## Domain events

Versioned types on the bus (`*.v1`). Realtime types drop the version (`entity.action`).

| Domain | Realtime | When |
|---|---|---|
| `workflow.*.v1` | `workflow.*` | CRUD, activate/deactivate |
| `endpoint.*.v1` | `endpoint.*` | Endpoint CRUD |
| `step.created.v1` / `updated.v1` / `deleted.v1` / `position_updated.v1` | `step.*` | Step CRUD / move |
| `connection.*.v1` | `connection.*` | Link / unlink steps |
| `variable.*.v1` | `variable.*` | Variable CRUD |
| `assertion.*.v1` | `assertion.*` | Assertion CRUD |
| `workflowRun.*.v1` | `workflowRun.*` | Run lifecycle + `finished` |
| `stepRun.*.v1` | `stepRun.*` | Step execution (HTTP and delay) |

Plus user and project events (`user.*`, `project.*`).

On step create/move and connection create/delete, the API **synchronously** recalculates `index`, `executionOrder`, and `treeIndex` from canvas position and the directed graph. Deleting a step also deletes its connections.

Activity log messages are projected from domain events (natural language, structured IDs in API fields, user attribution on manual builder actions).

## Delay steps (API)

Create an HTTP step (default):

```json
POST /api/workflows/:workflowId/steps
{
  "endpointId": "…",
  "position": { "x": 100, "y": 200 }
}
```

Create a delay step:

```json
POST /api/workflows/:workflowId/steps
{
  "type": "delay",
  "name": "Wait 30 seconds",
  "delayDurationSeconds": 30,
  "position": { "x": 100, "y": 300 }
}
```

Update a delay step (`PUT …/steps/:id`):

```json
{
  "name": "Wait 1 minute",
  "description": "Optional",
  "delayDurationSeconds": 60
}
```

Rules:

- `http` — `endpointId` required; no `delayDurationSeconds` or `expression`
- `delay` — `delayDurationSeconds` required; no `endpointId` or `expression`; no variables or assertions
- `condition` — `expression` required; no `endpointId` or `delayDurationSeconds`; no variables or assertions; two outgoing connections with `branch: "true"` and `branch: "false"`
- orphan delay (no connections) — skipped at runtime

## Condition steps (API)

Create a condition step:

```json
POST /api/workflows/:workflowId/steps
{
  "type": "condition",
  "name": "Premium plan?",
  "expression": "{{plan}} == \"premium\"",
  "position": { "x": 100, "y": 400 }
}
```

Link branches (`POST /api/workflows/:workflowId/connections`):

```json
{ "sourceStepId": "<condition-step-id>", "targetStepId": "<true-branch-step-id>", "branch": "true" }
{ "sourceStepId": "<condition-step-id>", "targetStepId": "<false-branch-step-id>", "branch": "false" }
```

Update (`PUT …/steps/:id` on a condition step):

```json
{
  "name": "Premium plan?",
  "description": "Optional",
  "expression": "{{plan}} == \"premium\""
}
```

Expressions use [expr-lang](https://github.com/expr-lang/expr) with workflow variables in scope (`{{varName}}` placeholders are resolved before evaluation).

## Configuration

Copy [`.env.dist`](.env.dist) and fill required values (Clerk, Postgres, RabbitMQ, Centrifugo).

| Variable | Default | Description |
|---|---|---|
| `STEP_RUN_WAITING_POLL_INTERVAL` | `1s` | Worker poll interval for due `waiting` delay step runs |
| `STEP_RUN_WAITING_POLL_BATCH_SIZE` | `100` | Batch size per poll tick |
| `SCHEDULER_INTERVAL` | `1m` | How often scheduled workflows are claimed |
| `RABBITMQ_EXECUTOR_*` | `step_run.execute` | Executor queue topology |
| `OUTBOX_POLL_INTERVAL` | `2s` | Outbox relay poll interval |

See `.env.dist` for the full list (CORS, worker concurrency, Stripe, etc.).

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
| Worker | — | Outbox relay + consumer + delay poller |
| Executor | — | HTTP step runs |
| Scheduler | — | Cron-like workflow starts |
| Postgres | `9543` | |
| RabbitMQ | `5672` / UI `15672` | Credentials from `.env` |
| Centrifugo | `8000` | WebSocket |
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
| `make tests` | Handler HTTP tests via Docker (`handler/test/...`) |
| `make coverage` | Handler tests + `coverage.out` |
| `make coverage-html` | Coverage report → `coverage.html` |
| `make shell` | Shell into the API container |

## Testing

Handler tests live under `internal/interfaces/http/handler/test/`, one subfolder per resource. Each HTTP handler depends on **port interfaces** (`*_ports.go`) so command/query handlers can be mocked. Shared helpers and stable UUIDs live in `internal/interfaces/http/testutil/` (`fixtures.go`, `http.go`).

| Test package | Handler covered |
|---|---|
| `test/endpoint/` | Endpoints + OpenAPI import |
| `test/user/` | Current user, active project |
| `test/plan/` | Plans list |
| `test/invoice/` | Invoices list |
| `test/realtime/` | Centrifugo connection |
| `test/activity_log/` | Workflow activity |
| `test/project/` | Projects CRUD, members, activate |
| `test/workflow/` | Workflows CRUD, activate/deactivate |
| `test/connection/` | Step connections (incl. condition branches) |
| `test/step/` | Steps (http / delay / condition) |
| `test/variable/` | Workflow variables |
| `test/assertion/` | Step assertions |
| `test/workflow_run/` | Runs start/stop, list, detail, analytics |
| `test/subscription/` | Subscription, quota, billing portal |
| `test/billing_webhook/` | Stripe webhooks |
| `test/user_webhook/` | Clerk webhooks |
| `test/quota/` | Quota error mapping |

Naming: `Test<Resource>Handler_<Method>_<Scenario>`. Typical scenarios: success, unauthorized / missing active project, invalid input, business error, internal error.

**Locally** (no Docker):

```bash
go test ./internal/interfaces/http/handler/test/... -count=1 -race
go test ./internal/interfaces/http/handler/test/... -count=1 \
  -coverprofile=coverage.out \
  -coverpkg=./internal/interfaces/http/handler/...
go tool cover -html=coverage.out -o coverage.html
```

**Via Docker** (same paths inside the `api` container): `make tests`, `make coverage`, `make coverage-html`.

**CI** — [`.github/workflows/tests.yml`](.github/workflows/tests.yml) runs on push/PR to `main` / `master`: tests with `-race`, coverage on `internal/interfaces/http/handler/...`, artifact upload, and [Codecov](https://codecov.io/gh/clmntgbr/workflow-api) (flag `handler`). Set the `CODECOV_TOKEN` repository secret for uploads.

## HTTP surface

Health: `GET /livez`, `/readyz`, `/startupz`

Auth: Bearer Clerk JWT on `/api/*`. Webhooks: `POST /webhooks/clerk` (Svix), `POST /webhooks/stripe`.

| Area | Routes |
|---|---|
| User | `GET /api/users/me`, `PUT /api/users/me/active-project` |
| Projects | `GET/POST /api/projects`, `GET/PUT/DELETE /api/projects/:id`, members, activate |
| Workflows | `GET/POST /api/workflows`, `GET/PUT/DELETE /api/workflows/:id`, activate/deactivate |
| Endpoints | `GET/POST /api/endpoints`, import OpenAPI, `GET/PUT/DELETE /api/endpoints/:id` |
| Steps | nested under `/api/workflows/:workflowId/steps` (CRUD + `PUT …/position`) — `http`, `delay`, and `condition`; `lastRunStatus` reflects the **active run** when one is in progress, otherwise the latest completed run per step |
| Connections | nested under `/api/workflows/:workflowId/connections` (`branch` on edges from condition steps) |
| Variables | nested under `/api/workflows/:workflowId/variables` (+ paths search per step) |
| Assertions | nested under `/api/workflows/:workflowId/steps/:stepId/assertions` |
| Activity | `GET /api/workflows/:workflowId/activity` |
| Workflow runs | `POST /workflows/:id/start`, `POST /workflows/:id/stop`, list, detail, analytics |
| Billing | `/api/plans`, `/api/subscription`, `/api/quota`, `/api/subscriptions/*`, `/api/invoices` |
| Realtime | `GET /api/realtime/connection` |

## Messaging

- Commands persist the aggregate **and** outbox rows in the same transaction.
- Worker polls unpublished rows, publishes an envelope (`eventId`, `type`, `aggregateId`, `occurredAt`, `payload`), then sets `published_at`.
- Dedup key: `(event_id, handler_name)` in `processed_events`.
- Topology: main queue → TTL retry → main; poison / non-retryable → DLQ. Never `Nack(requeue=true)`.

### Execution flow (simplified)

```
StartWorkflowRun → workflowRun.started.v1
  → Orchestrator: root step runs (HTTP queued, delay waiting, condition inline)
  → stepRun.queued.v1 (HTTP only)
  → Executor: HTTP call
  → stepRun.succeeded.v1 | failed.v1
  → Orchestrator: enqueue next steps (branch routing for conditions), finalize run
  → Centrifugo + activity log
```

Delays: `waiting` + `resumeAt` → worker poller → `started` + `succeeded` → orchestrator continues.
