# Endpoints

## Overview

An **endpoint** is a reusable HTTP template: method, URL, headers, query, body, timeout, and retry settings. **HTTP steps** snapshot an endpoint when placed on the canvas. Endpoints are scoped to the **active project**.

## HTTP routes

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/endpoints` | Create endpoint |
| `POST` | `/api/endpoints/import` | Bulk import from OpenAPI (multipart: `file` + `payload` JSON) |
| `GET` | `/api/endpoints` | List by active project (paginated) |
| `GET` | `/api/endpoints/:id` | Get by ID |
| `PUT` | `/api/endpoints/:id` | Update |
| `DELETE` | `/api/endpoints/:id` | Delete |

## OpenAPI import

- Multipart form: `file` (OpenAPI spec), `payload` (import options JSON).
- Max file size enforced at read boundary.
- Creates multiple endpoints in one command; returns created IDs.

## Quotas

Creating endpoints may return `403` when plan quota is exceeded (`endpoint quota exceeded for your current plan`).

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `internal/interfaces/http/handler/endpoint_handler.go`, `endpoint_helpers.go` |
| Test exports | `endpoint_test_exports.go` (multipart read hooks) |
| Commands | `internal/application/command/endpoint/` |
| Queries | `internal/application/query/endpoint/` |
| Domain | `internal/domain/endpoint/` |

## Events

- `endpoint.created.v1`, `endpoint.updated.v1`, `endpoint.deleted.v1`

## Tests

`internal/interfaces/http/handler/test/endpoint/` — CRUD, import, read helpers, coverage for import branches.
