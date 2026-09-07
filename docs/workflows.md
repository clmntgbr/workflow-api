# Workflows

## Overview

A **workflow** is a named graph scoped to a project: steps, connections, variables, and assertions. Status is **`active`** or **`deleted`** (soft delete). Pause scheduling with `scheduleType: "none"`; manual runs stay allowed unless the workflow is deleted.

Canvas position drives `executionOrder` / `treeIndex`; connections define branches and skip logic.

## HTTP routes

Requires authentication and active project.

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/workflows` | Create workflow |
| `POST` | `/api/workflows/import` | Import from export JSON (Pro / Business) |
| `GET` | `/api/workflows` | List workflows in active project (paginated) |
| `GET` | `/api/workflows/:workflowId` | Get workflow detail |
| `GET` | `/api/workflows/:workflowId/export` | Export workflow JSON |
| `PUT` | `/api/workflows/:workflowId` | Update name, description, schedule, concurrency, notifications |
| `DELETE` | `/api/workflows/:workflowId` | Soft-delete |

## Cross-project access

If a workflow belongs to another project the user is a member of, `GET /api/workflows/:workflowId` returns:

```json
{
  "code": "WRONG_ORGANIZATION",
  "message": "…"
}
```

HTTP `409`. The code name is **legacy** (pre–project rename); behaviour is “wrong project — switch active project”.

## Export / import

`GET /api/workflows/:workflowId/export` returns a JSON document (`formatVersion: 1`) with local refs (no UUIDs). `POST /api/workflows/import` creates a **new active** workflow from that document (schedule comes from the payload; `scheduleType: "none"` means no cron).

Import is gated by plan quota `allowsWorkflowImport` (Pro and Business). Other plans get `403` (`workflow import is not available on your current plan`). Graph quotas (workflows, steps, variables, assertions) still apply.

## Scheduling

Update payload supports schedule fields (`scheduleType`, interval, `scheduleAt`, timezone). Validation errors include `invalid workflow schedule`, `schedule interval must be at least 1 minute`.

To pause a schedule without deleting the workflow, set `scheduleType` to `none`. The scheduler only claims workflows with `status = active` and a non-null `nextRunAt`.

A `once` schedule becomes `none` after it fires (same clear as a pause: `scheduleAt` and `nextRunAt` are dropped). Recurring schedules keep their next tick.

When a scheduled start hits the monthly run quota, the schedule is cleared (`scheduleType: none`) instead of changing status. Concurrent-run quota skips do not clear the schedule.

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

Clearing a schedule records `workflow.updated.v1` with `updateReason: schedule_cleared` (activity: `workflow.schedule_cleared`). Historical `workflow.activated` / `workflow.deactivated` activity rows remain for older events.

## Tests

`internal/interfaces/http/handler/test/workflow/`
