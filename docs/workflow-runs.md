# Workflow runs

## Overview

A **workflow run** is one execution of a workflow graph. Runs can be triggered manually, via API, on a schedule, or by webhook (when configured).

## HTTP routes

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/workflows/:id/start` | Start run (optional `context` JSON body) |
| `POST` | `/api/workflows/:id/stop` | Cancel in-progress run |
| `GET` | `/api/workflows/:workflowId/runs` | List runs (paginated) |
| `GET` | `/api/workflows/:workflowId/runs/:id` | Run detail (step runs, insights) |
| `GET` | `/api/workflows/:workflowId/runs/analytics` | Aggregated stats (`from` / `to` RFC3339 query) |

## Start / stop

### Start

- `201` with run detail on success.
- `409` + `RUN_IN_PROGRESS` if a run is already active.
- `403` on workflow run or concurrent run quota exceeded.

### Stop

- Cancels the in-progress run and non-terminal step runs (including `waiting` delays).
- `409` + `NO_RUN_IN_PROGRESS` when nothing to cancel.

## Execution flow (simplified)

```
POST /start → workflowRun.started.v1
  → Orchestrator: root steps
      HTTP  → stepRun.queued → executor → succeeded | failed
      Delay → waiting + resumeAt → worker poller → succeeded
      Condition → inline eval → branch routing
  → workflowRun.finished.v1
```

## Step run statuses

HTTP and delay steps emit `stepRun.*` realtime events. API exposes `startedAt`, `finishedAt`, `resumeAt` on step runs.

## Analytics query

```
GET …/runs/analytics?from=2026-01-01T00:00:00Z&to=2026-01-31T23:59:59Z
```

Returns totals, success/failure rates, average duration, last run time.

## Binaries involved

| Binary | Role |
|--------|------|
| `cmd/api` | Start/stop HTTP |
| `cmd/worker` | Orchestration, delay poller, outbox consumer |
| `cmd/executor` | HTTP step execution |
| `cmd/scheduler` | Scheduled workflow starts |

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `internal/interfaces/http/handler/workflow_run_handler.go` |
| Commands | `internal/application/command/workflowrun/` |
| Queries | `internal/application/query/workflowrun/` |
| Domain | `internal/domain/workflowrun/` |

## Events

- `workflowRun.started.v1`, `workflowRun.finished.v1`, …
- `stepRun.queued.v1`, `stepRun.succeeded.v1`, `stepRun.failed.v1`, …

## Tests

`internal/interfaces/http/handler/test/workflow_run/`
