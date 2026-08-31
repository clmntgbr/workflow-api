# Activity logs

## Overview

**Activity logs** provide a human-readable audit trail per workflow: builder actions (create/update/delete) and run lifecycle events. Messages are projected from domain events with user attribution on manual builder actions.

## HTTP routes

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/workflows/:workflowId/activity` | Paginated activity for workflow |

Requires authentication and active project. Workflow must belong to the active project.

## Response

Paginated list of activity entries (message, actor, timestamps, structured IDs in API fields). Exact shape: `presenter` package / OpenAPI if published.

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `internal/interfaces/http/handler/activity_log_handler.go` |
| Queries | `internal/application/query/activitylog/` |
| Projection | `internal/application/event/` handlers writing activity rows |
| Domain | `internal/domain/activitylog/` |

## Events (sources)

Activity rows are derived from workflow builder and run events, for example:

- Workflow, step, connection, variable, assertion CRUD
- Workflow run started / finished

Realtime updates may mirror run progress — see [Realtime](realtime.md).

## Tests

`internal/interfaces/http/handler/test/activity_log/`
