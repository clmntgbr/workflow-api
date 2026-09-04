# Headers

## Overview

**Header suggestions** reuse keys and values already stored on endpoints and steps in the **active project**. The raw template is returned (for example `Bearer {{token-id}}`), never an interpolated secret.

Both routes are paginated with `paginate.PaginateQuery` (`page`, `limit`, `search`).

## HTTP routes

Requires authentication and active project.

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/headers/suggest` | Distinct header **keys**, ranked by frequency |
| `GET` | `/api/headers/suggest-values` | Header **values** (`key` + `value` + `count`). `key` is optional; omit it to search every key |

`search` filters keys (suggest) or values (suggest-values) with a case-insensitive contains match.

## Response

Both endpoints return `paginate.PaginateResponse`:

```json
{
  "members": [{ "key": "Authorization", "count": 12 }],
  "total": 1,
  "page": 1,
  "limit": 20,
  "totalPages": 1
}
```

Value suggestions add `value` on each member:

```json
{
  "members": [{ "key": "Authorization", "value": "Bearer {{token-id}}", "count": 5 }]
}
```

An empty page serializes `members` as `[]`, not `null`.

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `internal/interfaces/http/handler/header_handler.go` |
| Presenter | `internal/interfaces/http/presenter/header.go` |
| Queries | `internal/application/query/header/` |
| Domain | `internal/domain/header/` |
| Read repository | `internal/infrastructure/persistence/read/header_query_repository.go` |

The repository unpacks JSONB headers with `LATERAL jsonb_each_text` / `jsonb_object_keys` and named SQL parameters (`@name`). Postgres operators that contain `?` must not be mixed with GORM positional placeholders.

## Tests

`internal/interfaces/http/handler/test/header/` — success, missing active project, invalid query, internal error, pagination clamp, optional `key` on suggest-values.
