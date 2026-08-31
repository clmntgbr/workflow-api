# Steps

## Overview

A **step** is a node on the workflow canvas. Three types:

| Type | Purpose | Key fields |
|------|---------|------------|
| `http` (default) | Call an external API | `endpointId` — snapshot of [endpoint](endpoints.md) |
| `delay` | Pause between steps | `delayDurationSeconds` (> 0) |
| `condition` | Branch on expression | `expression` (expr-lang, `{{var}}` placeholders) |

## HTTP routes

Base: `/api/workflows/:workflowId/steps`

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `…/steps` | Create step (type inferred from body) |
| `GET` | `…/steps` | List steps (`lastRunStatus` from active or latest run) |
| `GET` | `…/steps/:id` | Get step |
| `PUT` | `…/steps/:id` | Update (body shape depends on type) |
| `PUT` | `…/steps/:id/position` | Update canvas position only |
| `DELETE` | `…/steps/:id` | Delete step (cascades connections) |

## Rules by type

### HTTP

```json
POST /api/workflows/:workflowId/steps
{
  "endpointId": "…",
  "position": { "x": 100, "y": 200 }
}
```

- `endpointId` required.
- Supports [variables](variables.md) and [assertions](assertions.md).

### Delay

```json
{
  "type": "delay",
  "name": "Wait 30 seconds",
  "delayDurationSeconds": 30,
  "position": { "x": 100, "y": 300 }
}
```

- No variables or assertions.
- Runtime: `waiting` + `resumeAt`; worker poller resumes — see [Workflow runs](workflow-runs.md).
- **Orphan delay** (no outgoing connections): skipped at runtime.

### Condition

```json
{
  "type": "condition",
  "name": "Premium plan?",
  "expression": "{{plan}} == \"premium\"",
  "position": { "x": 100, "y": 400 }
}
```

- Exactly two outgoing [connections](connections.md) with `branch: "true"` and `branch: "false"`.
- Evaluated inline by the worker (no executor binary).

## Ordering

On create, move, or connection change, the API recalculates `index`, `executionOrder`, and `treeIndex` from canvas position and the directed graph.

## Quotas

Step creation may return `403` for step/workflow quota errors.

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `internal/interfaces/http/handler/step_handler.go` |
| Commands | `internal/application/command/step/` |
| Queries | `internal/application/query/step/` |
| Domain | `internal/domain/step/` |
| Executor | `cmd/executor/` (HTTP steps only) |

## Events

- `step.created.v1`, `step.updated.v1`, `step.deleted.v1`, `step.position_updated.v1`

## Tests

`internal/interfaces/http/handler/test/step/` — split by type: `create_http_test.go`, `create_delay_test.go`, `create_condition_test.go`, etc.
