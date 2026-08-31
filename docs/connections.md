# Connections

## Overview

A **connection** is a directed edge between two steps: `sourceStepId` → `targetStepId`. Connections define execution flow after a step completes. **Condition steps** require labelled branches.

## HTTP routes

Base: `/api/workflows/:workflowId/connections`

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `…/connections` | Create connection |
| `GET` | `…/connections` | List all connections in workflow |
| `DELETE` | `…/connections/:id` | Delete connection |

## Request body

```json
{
  "sourceStepId": "…",
  "targetStepId": "…",
  "branch": "true"
}
```

- `branch` is optional for normal edges.
- For edges from a **condition** step: `branch` must be `"true"` or `"false"`.
- One condition step must have exactly one outgoing `true` and one `false` connection (enforced at domain/command level).

## Validation errors (examples)

| Error | HTTP |
|-------|------|
| `source step not found` / `target step not found` | `404` |
| `source and target steps must be different` | `400` |
| Condition branch rules (`ErrConditionRequiresBranch`, etc.) | `400` |

## Side effects

Creating or deleting a connection triggers **execution order recalculation** on the workflow graph (same as step position updates).

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `internal/interfaces/http/handler/connection_handler.go` |
| Commands | `internal/application/command/connection/` |
| Queries | `internal/application/query/connection/` |
| Domain | `internal/domain/connection/` |

## Events

- `connection.created.v1`, `connection.deleted.v1`

## Tests

`internal/interfaces/http/handler/test/connection/` — includes branch `true`/`false` and domain error paths.
