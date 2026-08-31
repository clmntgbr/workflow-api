# Projects

## Overview

A **project** is a multi-tenant workspace. Users belong to one or more projects; the **active project** on the user scopes workflows, endpoints, and related builder data.

## HTTP routes

All routes require authentication.

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/projects` | List projects for the current user (paginated) |
| `POST` | `/api/projects` | Create project |
| `GET` | `/api/projects/:id` | Get project by ID |
| `PUT` | `/api/projects/:id` | Update project |
| `POST` | `/api/projects/:id/activate` | Set as active project for the user |
| `DELETE` | `/api/projects/:id` | Soft-delete project |
| `DELETE` | `/api/projects/:id/members/:userId` | Remove member |

## Behaviour

- Listing and access are membership-scoped.
- `POST …/activate` is equivalent to `PUT /api/users/me/active-project` with that project ID.
- Deletes are soft deletes where applicable at the domain level.

## Code map

| Layer | Location |
|-------|----------|
| HTTP | `internal/interfaces/http/handler/project_handler.go` |
| Commands | `internal/application/command/project/` |
| Queries | `internal/application/query/project/` |
| Domain | `internal/domain/project/` |

## Events

- `project.created.v1`, `project.updated.v1`, `project.deleted.v1`, member events

## Realtime

Project-scoped channels feed canvas collaboration — see [Realtime](realtime.md).

## Tests

`internal/interfaces/http/handler/test/project/`
