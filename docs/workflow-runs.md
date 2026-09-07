# Workflow runs

## Overview

A **workflow run** is one execution of a workflow graph. Runs can be triggered manually, via API, on a schedule, or by webhook (when configured).

## HTTP routes

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/workflows/:workflowId/start` | Start run (optional `context` JSON body) |
| `POST` | `/api/workflows/:workflowId/stop` | Cancel in-progress run |
| `GET` | `/api/workflows/:workflowId/runs` | List runs (paginated) |
| `GET` | `/api/workflows/:workflowId/runs/:id` | Run detail (step runs, insights) |
| `GET` | `/api/workflows/:workflowId/runs/analytics` | Aggregated stats (`from` / `to` RFC3339 query) |
| `POST` | `/api/workflows/:workflowId/runs/export` | Request an XLSX run-history export (Pro/Business). `202` `{ id, status }`. Optional body `{ from, to }` RFC3339. |
| `GET` | `/api/workflows/:workflowId/runs/export/:jobId` | Export job status (`pending`, `processing`, `ready`, `failed`). File is emailed, not downloaded. |

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

Scheduler also polls stalled **HTTP/condition** step runs (not delays) every 30 minutes (clock-aligned, same pattern as due workflow claims). A `pending` step older than 30 minutes, or a `running` step past `timeout × attempts + retry delays + 5m` grace, fails that step, cancels siblings (including waiting delays), and fails the workflow run. Long delay steps never trigger this.

## Step run statuses

HTTP and delay steps emit `stepRun.*` realtime events. API exposes `startedAt`, `finishedAt`, `resumeAt` on step runs.

Workflow run payloads include `duration` (milliseconds): the sum of all step-run elapsed times (`finishedAt - startedAt`), including HTTP, condition, and delay waits. In-progress steps without `finishedAt` are not counted yet.

## Analytics query

```
GET …/runs/analytics?from=2026-01-01T00:00:00Z&to=2026-01-31T23:59:59Z
```

Returns totals, success/failure rates, average duration, last run time.

## Run history export

Pro and Business plans (`allows_data_export`) can request an XLSX of runs in the retention window:

```
POST /api/workflows/:workflowId/runs/export
{ "from": "2026-01-01T00:00:00Z", "to": "2026-01-31T23:59:59Z" }
```

Returns `202` `{ id, status: "pending" }`. The worker builds the workbook, redacts secrets, and emails it to the requester. Insight columns are included only when the plan allows insights. Files larger than ~15MB fail the job (no object storage). Status is available on `GET …/runs/export/:jobId` (`pending` | `processing` | `ready` | `failed`). Centrifugo also notifies the requester: `runExport.ready` / `runExport.failed`.

## Binaries involved

| Binary | Role |
|--------|------|
| `cmd/api` | Start/stop HTTP |
| `cmd/worker` | Orchestration, delay poller, outbox consumer, run-history export |
| `cmd/executor` | HTTP step execution |
| `cmd/scheduler` | Scheduled workflow starts (ticks aligned to `SCHEDULER_INTERVAL`, e.g. every minute at `:00`) and stalled HTTP/condition step runs (every 30 minutes, clock-aligned) |

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `internal/interfaces/http/handler/workflow_run_handler.go`, `run_export_handler.go` |
| Commands | `internal/application/command/workflowrun/`, `internal/application/command/runexport/`, `internal/application/command/steprun/` |
| Queries | `internal/application/query/workflowrun/` |
| Domain | `internal/domain/workflowrun/`, `internal/domain/steprun/` |

## Events

- `workflowRun.started.v1`, `workflowRun.finished.v1`, …
- `stepRun.queued.v1`, `stepRun.succeeded.v1`, `stepRun.failed.v1`, …
- `runExport.requested.v1`, `runExport.ready.v1`, `runExport.failed.v1`

## Tests

`internal/interfaces/http/handler/test/workflow_run/`
`internal/interfaces/http/handler/test/run_export/`
