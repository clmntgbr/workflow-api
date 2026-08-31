# Assertions

## Overview

**Assertions** validate HTTP step responses after execution: status code, headers, or body (JSON path). Failed assertions mark the step run as failed.

Assertions belong to an **HTTP step** only (not delay or condition steps).

## HTTP routes

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `…/steps/:stepId/assertions` | Create assertion |
| `GET` | `…/steps/:stepId/assertions` | List assertions for step |
| `GET` | `…/steps/:stepId/assertion-paths` | Search body paths (paginated) |
| `GET` | `…/assertions/:id` | Get by ID |
| `PUT` | `…/assertions/:id` | Update |
| `DELETE` | `…/assertions/:id` | Delete |

Base prefix: `/api/workflows/:workflowId`

## Create payload (example)

```json
{
  "description": "Status is 200",
  "source": "status",
  "path": "",
  "operator": "equals",
  "expectedValue": "200"
}
```

### Source & operator

- **Source:** `status`, `header`, `body` (domain-validated).
- **Operator:** `equals`, `not_equals`, `contains`, `not_null`, etc. (see domain `assertion` package).
- `expectedValue` may be omitted for operators like `not_null`.

## Access control

All routes verify:

1. Active project matches workflow.
2. Step belongs to workflow (for step-scoped routes).

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `internal/interfaces/http/handler/assertion_handler.go` |
| Commands | `internal/application/command/assertion/` |
| Queries | `internal/application/query/assertion/` |
| Domain | `internal/domain/assertion/` |

## Events

- `assertion.created.v1`, `assertion.updated.v1`, `assertion.deleted.v1`

## Tests

`internal/interfaces/http/handler/test/assertion/` — `handler_test.go`, `search_paths_test.go`
