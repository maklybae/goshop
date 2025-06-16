-- +goose Up
-- +goose StatementBegin
CREATE TABLE outbox (
    id UUID PRIMARY KEY,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    processed BOOLEAN NOT NULL DEFAULT FALSE,
    reserved_to TIMESTAMP DEFAULT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE outbox;
-- +goose StatementEnd