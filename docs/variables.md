# Variables

## Overview

**Variables** hold data used during workflow execution:

| Kind | Description |
|------|-------------|
| `static` | Fixed value set in the builder |
| `extracted` | JSON-path extract from a previous HTTP step response |

Variables are scoped to a **workflow**. HTTP steps can reference them in URLs, headers, and bodies via `{{key}}` placeholders.

## HTTP routes

Base: `/api/workflows/:workflowId/variables`

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `…/variables` | Create variable |
| `GET` | `…/variables` | List all variables in workflow |
| `GET` | `…/variables/:id` | Get by ID |
| `PUT` | `…/variables/:id` | Update |
| `DELETE` | `…/variables/:id` | Delete |
| `GET` | `…/steps/:stepId/variables` | Variables available at a step (upstream) |
| `GET` | `…/steps/:stepId/variable-paths` | Search JSON paths for extracts (paginated) |

## Create payload (examples)

```json
{
  "name": "User ID",
  "key": "userId",
  "kind": "extracted",
  "stepId": "…",
  "path": "$.data.id"
}
```

```json
{
  "name": "Plan",
  "key": "plan",
  "kind": "static",
  "value": "premium"
}
```

## Errors

| Case | HTTP |
|------|------|
| Duplicate key | `409` |
| Step not found | `404` |
| Variable in use (delete) | `409` |
| Wrong project / workflow | `404` |

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `internal/interfaces/http/handler/variable_handler.go` |
| Commands | `internal/application/command/variable/` |
| Queries | `internal/application/query/variable/` |
| Domain | `internal/domain/variable/` |

## Events

- `variable.created.v1`, `variable.updated.v1`, `variable.deleted.v1`

## Tests

`internal/interfaces/http/handler/test/variable/` — `handler_test.go`, `list_available_test.go`, `search_paths_test.go`
