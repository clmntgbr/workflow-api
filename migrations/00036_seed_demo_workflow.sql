-- +goose Up
-- +goose StatementBegin
--
-- Demo / playground workflow exercising HTTP steps, static & extracted variables,
-- delay, condition branching, and assertions against JSONPlaceholder (public API).
--
-- Fixed IDs (prefix 0195…) make the seed easy to reference in tests and rollback.
--
-- Graph:
--   [Get user] → [Delay 3s] → [Condition plan == "premium"]
--                                    ├─ true  → [Get posts by user]
--                                    └─ false → [Get single post]
--
INSERT INTO projects (id, name, created_at, updated_at)
VALUES (
    '01950000-0000-7000-8000-000000000001',
    'Playground',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
);

INSERT INTO user_projects (user_id, project_id, created_at)
SELECT u.id, '01950000-0000-7000-8000-000000000001', NOW()
FROM users u
ON CONFLICT DO NOTHING;

INSERT INTO workflows (
    id, name, description, status, project_id,
    schedule_type, schedule_interval_value, schedule_interval_unit, schedule_timezone,
    concurrency,
    notifications_enabled, notify_on_success, notify_on_failure, notify_on_cancel,
    created_at, updated_at
) VALUES (
    '01950000-0000-7000-8000-000000000002',
    'Demo — E2E feature test',
    'Seeded workflow: HTTP + variables + 3s delay + condition branch + assertions. Uses JSONPlaceholder. Set static variable plan to "free" to exercise the false branch.',
    'active',
    '01950000-0000-7000-8000-000000000001',
    'none', 0, NULL, 'UTC',
    1,
    FALSE, FALSE, FALSE, FALSE,
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
);

INSERT INTO endpoints (
    id, name, description, url, method, headers, query_params, body,
    timeout_ms, retry_on_failure, retry_count, retry_delay_ms,
    status, project_id, created_at, updated_at
) VALUES
(
    '01950000-0000-7000-8000-000000000010',
    'JSONPlaceholder — Get user',
    'Public test API: fetch user #1',
    'https://jsonplaceholder.typicode.com/users/1',
    'GET',
    '{}'::jsonb,
    '{}'::jsonb,
    '{}'::jsonb,
    30000, FALSE, 0, 10000,
    'active',
    '01950000-0000-7000-8000-000000000001',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01950000-0000-7000-8000-000000000011',
    'JSONPlaceholder — List posts by user',
    'Public test API: list posts filtered by userId',
    'https://jsonplaceholder.typicode.com/posts',
    'GET',
    '{}'::jsonb,
    '{}'::jsonb,
    '{}'::jsonb,
    30000, FALSE, 0, 10000,
    'active',
    '01950000-0000-7000-8000-000000000001',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01950000-0000-7000-8000-000000000012',
    'JSONPlaceholder — Get post',
    'Public test API: fetch post #1',
    'https://jsonplaceholder.typicode.com/posts/1',
    'GET',
    '{}'::jsonb,
    '{}'::jsonb,
    '{}'::jsonb,
    30000, FALSE, 0, 10000,
    'active',
    '01950000-0000-7000-8000-000000000001',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
);

INSERT INTO steps (
    id, workflow_id, endpoint_id, project_id, type, delay_duration_seconds, expression,
    name, description, url, method, headers, query_params, body,
    timeout_ms, retry_on_failure, retry_count, retry_delay_ms,
    step_index, execution_order, tree_index, position_x, position_y,
    status, created_at, updated_at
) VALUES
(
    '01950000-0000-7000-8000-000000000101',
    '01950000-0000-7000-8000-000000000002',
    '01950000-0000-7000-8000-000000000010',
    '01950000-0000-7000-8000-000000000001',
    'http', NULL, NULL,
    'Get user',
    'Fetch user #1 and extract userId',
    'https://jsonplaceholder.typicode.com/users/1',
    'GET',
    '{}'::jsonb, '{}'::jsonb, '{}'::jsonb,
    30000, FALSE, 0, 10000,
    '1', 1, 0, 0, 0,
    'active', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z'
),
(
    '01950000-0000-7000-8000-000000000102',
    '01950000-0000-7000-8000-000000000002',
    NULL,
    '01950000-0000-7000-8000-000000000001',
    'delay', 3, NULL,
    'Wait 3 seconds',
    'Short delay before condition evaluation',
    '', '',
    '{}'::jsonb, '{}'::jsonb, '{}'::jsonb,
    30000, FALSE, 0, 10000,
    '2', 2, 0, 280, 0,
    'active', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z'
),
(
    '01950000-0000-7000-8000-000000000103',
    '01950000-0000-7000-8000-000000000002',
    NULL,
    '01950000-0000-7000-8000-000000000001',
    'condition', NULL, '{{plan}} == "premium"',
    'Check plan',
    'Routes to premium or free branch',
    '', '',
    '{}'::jsonb, '{}'::jsonb, '{}'::jsonb,
    30000, FALSE, 0, 10000,
    '3', 3, 0, 560, 0,
    'active', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z'
),
(
    '01950000-0000-7000-8000-000000000104',
    '01950000-0000-7000-8000-000000000002',
    '01950000-0000-7000-8000-000000000011',
    '01950000-0000-7000-8000-000000000001',
    'http', NULL, NULL,
    'Get posts (premium)',
    'Premium branch: list posts for extracted userId',
    'https://jsonplaceholder.typicode.com/posts?userId={{userId}}',
    'GET',
    '{}'::jsonb, '{}'::jsonb, '{}'::jsonb,
    30000, FALSE, 0, 10000,
    '4', 4, 1, 840, -120,
    'active', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z'
),
(
    '01950000-0000-7000-8000-000000000105',
    '01950000-0000-7000-8000-000000000002',
    '01950000-0000-7000-8000-000000000012',
    '01950000-0000-7000-8000-000000000001',
    'http', NULL, NULL,
    'Get post (free)',
    'Free branch: fetch a single post',
    'https://jsonplaceholder.typicode.com/posts/1',
    'GET',
    '{}'::jsonb, '{}'::jsonb, '{}'::jsonb,
    30000, FALSE, 0, 10000,
    '5', 4, 2, 840, 120,
    'active', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z'
);

INSERT INTO connections (
    id, workflow_id, project_id, source_step_id, target_step_id, branch
) VALUES
(
    '01950000-0000-7000-8000-000000000201',
    '01950000-0000-7000-8000-000000000002',
    '01950000-0000-7000-8000-000000000001',
    '01950000-0000-7000-8000-000000000101',
    '01950000-0000-7000-8000-000000000102',
    NULL
),
(
    '01950000-0000-7000-8000-000000000202',
    '01950000-0000-7000-8000-000000000002',
    '01950000-0000-7000-8000-000000000001',
    '01950000-0000-7000-8000-000000000102',
    '01950000-0000-7000-8000-000000000103',
    NULL
),
(
    '01950000-0000-7000-8000-000000000203',
    '01950000-0000-7000-8000-000000000002',
    '01950000-0000-7000-8000-000000000001',
    '01950000-0000-7000-8000-000000000103',
    '01950000-0000-7000-8000-000000000104',
    'true'
),
(
    '01950000-0000-7000-8000-000000000204',
    '01950000-0000-7000-8000-000000000002',
    '01950000-0000-7000-8000-000000000001',
    '01950000-0000-7000-8000-000000000103',
    '01950000-0000-7000-8000-000000000105',
    'false'
);

INSERT INTO variables (
    id, name, key, description, kind, path, value, step_id, workflow_id, created_at, updated_at
) VALUES
(
    '01950000-0000-7000-8000-000000000301',
    'Plan tier',
    'plan',
    'Static plan used by the condition step (premium = true branch, free = false branch)',
    'static',
    NULL,
    '"premium"'::jsonb,
    NULL,
    '01950000-0000-7000-8000-000000000002',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01950000-0000-7000-8000-000000000302',
    'User ID',
    'userId',
    'Extracted from Get user response body',
    'extracted',
    '$.id',
    NULL,
    '01950000-0000-7000-8000-000000000101',
    '01950000-0000-7000-8000-000000000002',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
);

INSERT INTO assertions (
    id, description, source, path, operator, expected_value, step_id, workflow_id, created_at, updated_at
) VALUES
(
    '01950000-0000-7000-8000-000000000401',
    'User request returns 200',
    'status',
    NULL,
    'equals',
    '200',
    '01950000-0000-7000-8000-000000000101',
    '01950000-0000-7000-8000-000000000002',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01950000-0000-7000-8000-000000000402',
    'User body contains id',
    'body',
    '$.id',
    'not_null',
    NULL,
    '01950000-0000-7000-8000-000000000101',
    '01950000-0000-7000-8000-000000000002',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01950000-0000-7000-8000-000000000403',
    'Posts request returns 200',
    'status',
    NULL,
    'equals',
    '200',
    '01950000-0000-7000-8000-000000000104',
    '01950000-0000-7000-8000-000000000002',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01950000-0000-7000-8000-000000000404',
    'Posts response is an array',
    'body',
    '$[0]',
    'not_null',
    NULL,
    '01950000-0000-7000-8000-000000000104',
    '01950000-0000-7000-8000-000000000002',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01950000-0000-7000-8000-000000000405',
    'Post request returns 200',
    'status',
    NULL,
    'equals',
    '200',
    '01950000-0000-7000-8000-000000000105',
    '01950000-0000-7000-8000-000000000002',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01950000-0000-7000-8000-000000000406',
    'Post body contains id',
    'body',
    '$.id',
    'not_null',
    NULL,
    '01950000-0000-7000-8000-000000000105',
    '01950000-0000-7000-8000-000000000002',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM assertions WHERE workflow_id = '01950000-0000-7000-8000-000000000002';
DELETE FROM variables WHERE workflow_id = '01950000-0000-7000-8000-000000000002';
DELETE FROM connections WHERE workflow_id = '01950000-0000-7000-8000-000000000002';
DELETE FROM steps WHERE workflow_id = '01950000-0000-7000-8000-000000000002';
DELETE FROM workflows WHERE id = '01950000-0000-7000-8000-000000000002';
DELETE FROM endpoints WHERE project_id = '01950000-0000-7000-8000-000000000001';
DELETE FROM user_projects WHERE project_id = '01950000-0000-7000-8000-000000000001';
DELETE FROM projects WHERE id = '01950000-0000-7000-8000-000000000001';
-- +goose StatementEnd
