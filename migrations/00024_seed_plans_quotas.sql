-- +goose Up
-- +goose StatementBegin
DELETE FROM plans;
DELETE FROM quotas;

INSERT INTO quotas (
    id,
    name,
    max_project_members,
    max_workflows,
    max_steps_per_workflow,
    max_endpoints,
    max_variables_per_workflow,
    max_workflow_runs_per_month,
    max_concurrent_runs,
    min_schedule_interval_minutes,
    run_history_retention_days,
    max_step_timeout_seconds,
    max_retry_count_per_step,
    max_request_body_size_kb,
    max_response_body_size_kb,
    allows_openapi_import,
    allows_insights,
    allows_data_export,
    executor_priority,
    created_at,
    updated_at
) VALUES
(
    '01941f29-7c00-7558-b853-0fef65937505',
    'Free',
    2, 5, 10, 10, 5,
    100, 1, 60, 7,
    15, 1, 128, 128,
    FALSE, FALSE, FALSE, 0,
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01941f29-7c01-7ff3-b3eb-d5c787d32ba4',
    'Starter',
    5, 20, 50, 100, 50,
    1000, 2, 30, 30,
    30, 3, 256, 256,
    FALSE, TRUE, FALSE, 2,
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01941f29-7c02-798d-b4c6-9d777505cdbc',
    'Pro',
    10, 50, 100, 200, 100,
    5000, 5, 10, 90,
    60, 5, 1024, 1024,
    TRUE, TRUE, TRUE, 5,
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01941f29-7c03-7a99-937e-5307a3f91734',
    'Business',
    50, 100, 200, 500, 200,
    50000, 25, 1, 365,
    300, 10, 5120, 5120,
    TRUE, TRUE, TRUE, 10,
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
);

INSERT INTO plans (
    id,
    name,
    description,
    slug,
    stripe_price_id,
    is_active,
    billing_interval,
    price,
    currency,
    quota_id,
    created_at,
    updated_at
) VALUES
(
    '01941f29-7c04-7701-abeb-ce79cc69df64',
    'Free',
    'Pour découvrir et automatiser quelques appels HTTP simples.',
    'free',
    'price_1U79wb5vcjninxNh2Zl0CgJp',
    TRUE,
    'month',
    0.00,
    'EUR',
    '01941f29-7c00-7558-b853-0fef65937505',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01941f29-7c05-76cb-a3a0-2f7dea4e16d3',
    'Starter',
    'Pour les petites équipes qui automatisent leurs premiers workflows.',
    'starter-monthly',
    'price_1U79xV5vcjninxNhW3DIO6Vo',
    TRUE,
    'month',
    12.00,
    'EUR',
    '01941f29-7c01-7ff3-b3eb-d5c787d32ba4',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01941f29-7c06-74b7-b04e-d1c615cb6325',
    'Starter (annuel)',
    'Starter avec 2 mois offerts en facturation annuelle.',
    'starter-yearly',
    'price_1U79xp5vcjninxNhmIEzdOdI',
    TRUE,
    'year',
    120.00,
    'EUR',
    '01941f29-7c01-7ff3-b3eb-d5c787d32ba4',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01941f29-7c07-7306-9af3-6d704d78628b',
    'Pro',
    'Pour les équipes qui automatisent des workflows en production.',
    'pro-monthly',
    'price_1U79yE5vcjninxNhOpMnzjZn',
    TRUE,
    'month',
    29.00,
    'EUR',
    '01941f29-7c02-798d-b4c6-9d777505cdbc',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01941f29-7c08-7c39-8b73-8254661b6770',
    'Pro (annuel)',
    'Pro avec 2 mois offerts en facturation annuelle.',
    'pro-yearly',
    'price_1U79yP5vcjninxNhxaMMcHYY',
    TRUE,
    'year',
    290.00,
    'EUR',
    '01941f29-7c02-798d-b4c6-9d777505cdbc',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01941f29-7c09-7b95-9920-cf4116d057c2',
    'Business',
    'Pour les organisations avec des besoins de volume et de rétention étendus.',
    'business-monthly',
    'price_1U79yk5vcjninxNhzOBZidRz',
    TRUE,
    'month',
    99.00,
    'EUR',
    '01941f29-7c03-7a99-937e-5307a3f91734',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    '01941f29-7c0a-7cc5-823a-dcaef25c0d8e',
    'Business (annuel)',
    'Business avec 2 mois offerts en facturation annuelle.',
    'business-yearly',
    'price_1U79yv5vcjninxNhBFTNakpP',
    TRUE,
    'year',
    990.00,
    'EUR',
    '01941f29-7c03-7a99-937e-5307a3f91734',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM plans WHERE id IN (
    '01941f29-7c04-7701-abeb-ce79cc69df64',
    '01941f29-7c05-76cb-a3a0-2f7dea4e16d3',
    '01941f29-7c06-74b7-b04e-d1c615cb6325',
    '01941f29-7c07-7306-9af3-6d704d78628b',
    '01941f29-7c08-7c39-8b73-8254661b6770',
    '01941f29-7c09-7b95-9920-cf4116d057c2',
    '01941f29-7c0a-7cc5-823a-dcaef25c0d8e'
);

DELETE FROM quotas WHERE id IN (
    '01941f29-7c00-7558-b853-0fef65937505',
    '01941f29-7c01-7ff3-b3eb-d5c787d32ba4',
    '01941f29-7c02-798d-b4c6-9d777505cdbc',
    '01941f29-7c03-7a99-937e-5307a3f91734'
);
-- +goose StatementEnd
