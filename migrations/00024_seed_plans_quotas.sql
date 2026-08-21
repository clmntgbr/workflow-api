-- +goose Up
-- +goose StatementBegin
DELETE FROM plans;
DELETE FROM quotas;

INSERT INTO quotas (
    id,
    name,
    max_organization_members,
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
    'b2e1a1b0-0001-4a10-8c1a-000000000001',
    'Free',
    2, 3, 5, 10, 5,
    100, 1, 60, 7,
    15, 1, 128, 128,
    FALSE, FALSE, FALSE, 0,
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    'b2e1a1b0-0002-4a10-8c1a-000000000002',
    'Starter',
    5, 10, 10, 30, 10,
    1000, 2, 30, 30,
    30, 3, 256, 256,
    FALSE, TRUE, FALSE, 2,
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    'b2e1a1b0-0003-4a10-8c1a-000000000003',
    'Pro',
    10, 25, 25, 100, 25,
    5000, 5, 10, 90,
    60, 5, 1024, 1024,
    TRUE, TRUE, TRUE, 5,
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    'b2e1a1b0-0004-4a10-8c1a-000000000004',
    'Business',
    50, 200, 100, 1000, 100,
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
    'c3f2b2c1-0001-4b20-9d2b-000000000011',
    'Free',
    'Pour découvrir FlowForge et automatiser quelques appels HTTP simples.',
    'free',
    '',
    TRUE,
    'month',
    0.00,
    'EUR',
    'b2e1a1b0-0001-4a10-8c1a-000000000001',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    'c3f2b2c1-0002-4b20-9d2b-000000000012',
    'Starter',
    'Pour les petites équipes qui automatisent leurs premiers workflows.',
    'starter-monthly',
    'price_starter_monthly_REPLACE_ME',
    TRUE,
    'month',
    12.00,
    'EUR',
    'b2e1a1b0-0002-4a10-8c1a-000000000002',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    'c3f2b2c1-0003-4b20-9d2b-000000000013',
    'Starter (annuel)',
    'Starter avec 2 mois offerts en facturation annuelle.',
    'starter-yearly',
    'price_starter_yearly_REPLACE_ME',
    TRUE,
    'year',
    120.00,
    'EUR',
    'b2e1a1b0-0002-4a10-8c1a-000000000002',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    'c3f2b2c1-0004-4b20-9d2b-000000000014',
    'Pro',
    'Pour les équipes qui automatisent des workflows en production.',
    'pro-monthly',
    'price_pro_monthly_REPLACE_ME',
    TRUE,
    'month',
    29.00,
    'EUR',
    'b2e1a1b0-0003-4a10-8c1a-000000000003',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    'c3f2b2c1-0005-4b20-9d2b-000000000015',
    'Pro (annuel)',
    'Pro avec 2 mois offerts en facturation annuelle.',
    'pro-yearly',
    'price_pro_yearly_REPLACE_ME',
    TRUE,
    'year',
    290.00,
    'EUR',
    'b2e1a1b0-0003-4a10-8c1a-000000000003',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    'c3f2b2c1-0006-4b20-9d2b-000000000016',
    'Business',
    'Pour les organisations avec des besoins de volume et de rétention étendus.',
    'business-monthly',
    'price_business_monthly_REPLACE_ME',
    TRUE,
    'month',
    99.00,
    'EUR',
    'b2e1a1b0-0004-4a10-8c1a-000000000004',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
),
(
    'c3f2b2c1-0007-4b20-9d2b-000000000017',
    'Business (annuel)',
    'Business avec 2 mois offerts en facturation annuelle.',
    'business-yearly',
    'price_business_yearly_REPLACE_ME',
    TRUE,
    'year',
    990.00,
    'EUR',
    'b2e1a1b0-0004-4a10-8c1a-000000000004',
    '2026-01-01T00:00:00Z',
    '2026-01-01T00:00:00Z'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM plans WHERE id IN (
    'c3f2b2c1-0001-4b20-9d2b-000000000011',
    'c3f2b2c1-0002-4b20-9d2b-000000000012',
    'c3f2b2c1-0003-4b20-9d2b-000000000013',
    'c3f2b2c1-0004-4b20-9d2b-000000000014',
    'c3f2b2c1-0005-4b20-9d2b-000000000015',
    'c3f2b2c1-0006-4b20-9d2b-000000000016',
    'c3f2b2c1-0007-4b20-9d2b-000000000017'
);

DELETE FROM quotas WHERE id IN (
    'b2e1a1b0-0001-4a10-8c1a-000000000001',
    'b2e1a1b0-0002-4a10-8c1a-000000000002',
    'b2e1a1b0-0003-4a10-8c1a-000000000003',
    'b2e1a1b0-0004-4a10-8c1a-000000000004'
);
-- +goose StatementEnd
