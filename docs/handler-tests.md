# HTTP handler tests

Guide for tests under `internal/interfaces/http/handler/test/`.

Architecture rules: [`.cursor/rules/architecture.mdc`](../.cursor/rules/architecture.mdc).

## Principles

- **Unit tests at the HTTP boundary** — Fiber `app.Test(req)` with mocked command/query handlers.
- **No infrastructure** — no real Postgres, RabbitMQ, Stripe, Clerk, or Docker.
- **One test package per resource** — never beside production handler files.
- **Ports for mocking** — handlers depend on unexported interfaces in `*_ports.go`.
- **Target full coverage** — aim for 100% statement coverage on each `*_handler.go` when adding or changing an endpoint.

## When tests are required

| Change | Required test work |
|--------|-------------------|
| New `*_handler.go` | Create `test/<resource>/`: mocks, ports, scenarios for every public HTTP method |
| New HTTP method | Add `Test<Resource>Handler_<Method>_<Scenario>` (success + error paths) |
| Changed behaviour | Update tests; keep affected handler fully covered |

Tests are part of the definition of done — not a follow-up task.

## Directory layout

### Production

```
internal/interfaces/http/handler/
  <resource>_handler.go
  <resource>_ports.go
  <resource>_test_exports.go   # optional, per resource
  quota_errors.go
  endpoint_import_read.go
```

### Tests

```
internal/interfaces/http/handler/test/
  <resource>/                  # package: <resource>test (e.g. endpointtest)
    handler_test.go
    helpers_test.go
    <method>_<concern>_test.go   # optional split (see step/)
  quota/
internal/interfaces/http/testutil/
  http.go
  fixtures.go
```

## Shared utilities

### `fixtures.go`

Stable UUIDs: `TestUserID`, `TestProjectID`, `TestWorkflowID`, … — do not duplicate per package.

### `http.go`

| Helper | Usage |
|--------|--------|
| `NewTestApp()` | Fiber + `validation.FiberErrorHandler` |
| `WithActiveProject(userID, projectID)` | User with active project |
| `WithUserWithoutProject(userID)` | → `400` on project-scoped routes |
| `WithLocal(key, value)` | Webhook payload in `c.Locals` |
| `JSONRequest`, `MultipartImportRequest` | Build requests |
| `DecodeJSON`, `DecodeJSONMap` | Parse responses |

## Auth scenarios

| Scenario | Helper | Expected |
|----------|--------|----------|
| Unauthorized | No user middleware | `401` |
| Missing active project | `WithUserWithoutProject` | `400` |
| Happy path | `WithActiveProject` | Depends on handler |

This is intentional — not a bug.

## Test naming

```
Test<Resource>Handler_<Method>_<Scenario>
```

Examples:

- `TestEndpointHandler_Create_Success`
- `TestWorkflowHandler_GetByID_HandlerError_WrongProject` (asserts `409` + `WRONG_ORGANIZATION` legacy code)

## Scenario checklist

1. **Success** — status, response shape, mock called with expected fields
2. **Unauthorized** or **MissingActiveProject**
3. **Invalid input** — bad JSON, invalid UUID, DTO validation
4. **Business error** — not found, wrong project, quota, `409` codes
5. **Internal error** — mock returns unexpected error → `500`

Before replicating the pattern on a new resource, run:

```bash
go test ./internal/interfaces/http/handler/test/... \
  -coverprofile=coverage.out \
  -coverpkg=./internal/interfaces/http/handler/...
go tool cover -func=coverage.out | grep <resource>_handler
```

A large test count does not guarantee full coverage.

## `*_test_exports.go`

Rare hooks when ports + HTTP cannot reach private code:

| File | Exports |
|------|---------|
| `endpoint_test_exports.go` | OpenAPI multipart read |
| `quota_test_exports.go` | `RespondQuotaErrorForTest` |
| `billing_webhook_test_exports.go` | Stripe invoice mappers |

Keep small; no business logic.

## Coverage & CI

- **Local floor:** run `go tool cover -func=coverage.out` after tests.
- **CI:** `.github/workflows/tests.yml` fails if handler coverage &lt; **98%**.
- **Codecov:** `codecov.yml` — flag `handler`, project target 98%, patch target 100%.

```bash
make tests          # Docker
make coverage
make coverage-html
```

## Test package index

| Folder | Handler |
|--------|---------|
| `activity_log/` | Activity list |
| `assertion/` | Assertions |
| `billing_webhook/` | Stripe webhooks |
| `connection/` | Connections |
| `endpoint/` | Endpoints + import |
| `invoice/` | Invoices |
| `plan/` | Plans |
| `project/` | Projects |
| `quota/` | Quota mapper |
| `realtime/` | Centrifugo |
| `step/` | Steps |
| `subscription/` | Subscription |
| `user/` | User |
| `user_webhook/` | Clerk |
| `variable/` | Variables |
| `workflow/` | Workflows |
| `workflow_run/` | Runs |

## Adding a new handler

1. Add `<resource>_handler.go` + `<resource>_ports.go`.
2. Wire in `cmd/api/di`.
3. Create `test/<resource>/` with mocks and scenarios.
4. Reach 100% on `*_handler.go`.
5. Add a row to the [Documentation](README.md#documentation) section in the root README and this index.
