# Workflows

## Overview

A **workflow** is a named graph scoped to a project: steps, connections, variables, and assertions. Workflows can be **activated** or **deactivated**, scheduled, and executed as [workflow runs](workflow-runs.md).

Canvas position drives `executionOrder` / `treeIndex`; connections define branches and skip logic.

## HTTP routes

Requires authentication and active project.

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/workflows` | Create workflow |
| `GET` | `/api/workflows` | List workflows in active project (paginated) |
| `GET` | `/api/workflows/:id` | Get workflow detail |
| `PUT` | `/api/workflows/:id` | Update name, description, status, schedule, concurrency, notifications |
| `POST` | `/api/workflows/:id/activate` | Activate |
| `POST` | `/api/workflows/:id/deactivate` | Deactivate |
| `DELETE` | `/api/workflows/:id` | Soft-delete |

## Cross-project access

If a workflow belongs to another project the user is a member of, `GET /api/workflows/:id` returns:

```json
{
  "code": "WRONG_ORGANIZATION",
  "message": "…"
}
```

HTTP `409`. The code name is **legacy** (pre–project rename); behaviour is “wrong project — switch active project”.

## Scheduling

Update payload supports schedule fields (`scheduleType`, interval, `scheduleAt`, timezone). Validation errors include `invalid workflow schedule`, `schedule interval must be at least 1 minute`.

Scheduled runs are claimed by the **scheduler** binary — see [Workflow runs](workflow-runs.md).

## Nested resources

| Resource | Base path |
|----------|-----------|
| Steps | `/api/workflows/:workflowId/steps` |
| Connections | `/api/workflows/:workflowId/connections` |
| Variables | `/api/workflows/:workflowId/variables` |
| Assertions | `/api/workflows/:workflowId/steps/:stepId/assertions` |
| Activity | `/api/workflows/:workflowId/activity` |
| Runs | `/api/workflows/:workflowId/runs` |

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `internal/interfaces/http/handler/workflow_handler.go` |
| Commands | `internal/application/command/workflow/` |
| Queries | `internal/application/query/workflow/` |
| Domain | `internal/domain/workflow/` |
| Scheduler | `cmd/scheduler/` |

## Events

| Domain (bus) | Realtime |
|--------------|----------|
| `workflow.created.v1` | `workflow.created` |
| `workflow.updated.v1` | `workflow.updated` |
| `workflow.deleted.v1` | `workflow.deleted` |
| activate/deactivate events | matching `workflow.*` |

## Tests

`internal/interfaces/http/handler/test/workflow/`
