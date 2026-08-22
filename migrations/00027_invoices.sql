-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    subscription_id UUID NULL REFERENCES subscriptions (id) ON DELETE SET NULL,
    stripe_invoice_id VARCHAR(255) NOT NULL,
    stripe_customer_id VARCHAR(255) NULL,
    stripe_subscription_id VARCHAR(255) NULL,
    number VARCHAR(255) NULL,
    status VARCHAR(64) NOT NULL,
    currency VARCHAR(16) NOT NULL,
    amount_due BIGINT NOT NULL DEFAULT 0,
    amount_paid BIGINT NOT NULL DEFAULT 0,
    total BIGINT NOT NULL DEFAULT 0,
    hosted_invoice_url TEXT NULL,
    invoice_pdf TEXT NULL,
    billing_reason VARCHAR(255) NULL,
    description TEXT NULL,
    attempt_count BIGINT NOT NULL DEFAULT 0,
    period_start TIMESTAMPTZ NULL,
    period_end TIMESTAMPTZ NULL,
    paid_at TIMESTAMPTZ NULL,
    stripe_created_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_stripe_invoice_id ON invoices (stripe_invoice_id);
CREATE INDEX IF NOT EXISTS idx_invoices_user_id ON invoices (user_id);
CREATE INDEX IF NOT EXISTS idx_invoices_subscription_id ON invoices (subscription_id);
CREATE INDEX IF NOT EXISTS idx_invoices_stripe_customer_id
    ON invoices (stripe_customer_id)
    WHERE stripe_customer_id IS NOT NULL AND stripe_customer_id <> '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_invoices_stripe_customer_id;
DROP INDEX IF EXISTS idx_invoices_subscription_id;
DROP INDEX IF EXISTS idx_invoices_user_id;
DROP INDEX IF EXISTS idx_invoices_stripe_invoice_id;
DROP TABLE IF EXISTS invoices;
-- +goose StatementEnd
