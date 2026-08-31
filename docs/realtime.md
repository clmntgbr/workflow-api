# Realtime

## Overview

**Centrifugo** delivers WebSocket updates so the canvas and run progress stay in sync across clients. The API does not publish realtime events directly from HTTP handlers — the preferred path is:

```
outbox → RabbitMQ → worker handler → Centrifugo publisher
```

Realtime event `type` uses `entity.action` (e.g. `workflow.created`), not the versioned domain type (`workflow.created.v1`).

## HTTP routes

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/realtime/connection` | Connection token / URL for Centrifugo |

Requires authentication.

## Event examples

| Realtime type | When |
|---------------|------|
| `workflow.*` | Workflow CRUD, activate/deactivate |
| `step.*` | Step CRUD, position |
| `connection.*` | Link / unlink |
| `variable.*` | Variable CRUD |
| `assertion.*` | Assertion CRUD |
| `workflowRun.*` | Run lifecycle |
| `stepRun.*` | Step execution (HTTP and delay) |

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `internal/interfaces/http/handler/realtime_handler.go` |
| Realtime helpers | `internal/application/realtime/` |
| Centrifugo adapter | `internal/infrastructure/centrifugo/` |
| Publish path | `internal/application/event/` handlers |

## Tests

`internal/interfaces/http/handler/test/realtime/`
