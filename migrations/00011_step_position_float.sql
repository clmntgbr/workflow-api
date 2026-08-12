-- +goose Up
-- +goose StatementBegin
ALTER TABLE steps
    ALTER COLUMN position_x TYPE DOUBLE PRECISION USING position_x::DOUBLE PRECISION,
    ALTER COLUMN position_y TYPE DOUBLE PRECISION USING position_y::DOUBLE PRECISION;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE steps
    ALTER COLUMN position_x TYPE INTEGER USING ROUND(position_x)::INTEGER,
    ALTER COLUMN position_y TYPE INTEGER USING ROUND(position_y)::INTEGER;
-- +goose StatementEnd
