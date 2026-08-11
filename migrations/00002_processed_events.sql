-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS processed_events (
    event_id UUID NOT NULL,
    handler_name TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (event_id, handler_name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS processed_events;
-- +goose StatementEnd
