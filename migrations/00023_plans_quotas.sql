-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,

    max_project_members INTEGER NOT NULL DEFAULT 0,
    max_workflows INTEGER NOT NULL DEFAULT 0,
    max_steps_per_workflow INTEGER NOT NULL DEFAULT 0,
    max_endpoints INTEGER NOT NULL DEFAULT 0,
    max_variables_per_workflow INTEGER NOT NULL DEFAULT 0,

    max_workflow_runs_per_month INTEGER NOT NULL DEFAULT 0,
    max_concurrent_runs INTEGER NOT NULL DEFAULT 0,
    min_schedule_interval_minutes INTEGER NOT NULL DEFAULT 0,

    run_history_retention_days INTEGER NOT NULL DEFAULT 0,

    max_step_timeout_seconds INTEGER NOT NULL DEFAULT 0,
    max_retry_count_per_step INTEGER NOT NULL DEFAULT 0,
    max_request_body_size_kb INTEGER NOT NULL DEFAULT 0,
    max_response_body_size_kb INTEGER NOT NULL DEFAULT 0,

    allows_openapi_import BOOLEAN NOT NULL DEFAULT FALSE,
    allows_insights BOOLEAN NOT NULL DEFAULT FALSE,
    allows_data_export BOOLEAN NOT NULL DEFAULT FALSE,
    executor_priority INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    slug VARCHAR(255) NOT NULL,
    stripe_price_id VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    billing_interval VARCHAR(32) NOT NULL DEFAULT 'month',
    price NUMERIC(10, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(16) NOT NULL DEFAULT 'EUR',
    quota_id UUID NOT NULL REFERENCES quotas (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_plans_slug UNIQUE (slug),
    CONSTRAINT chk_plans_billing_interval CHECK (billing_interval IN ('month', 'year')),
    CONSTRAINT chk_plans_currency CHECK (currency IN ('EUR', 'USD'))
);

CREATE INDEX IF NOT EXISTS idx_plans_quota_id ON plans (quota_id);
CREATE INDEX IF NOT EXISTS idx_plans_is_active ON plans (is_active);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS plans;
DROP TABLE IF EXISTS quotas;
-- +goose StatementEnd
