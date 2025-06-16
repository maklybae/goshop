-- +goose Up
-- +goose StatementBegin
CREATE TYPE order_status AS ENUM (
    'new',
    'finished',
    'cancelled'
);

CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    description TEXT,
    amount BIGINT NOT NULL,
    status order_status NOT NULL DEFAULT 'new'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders;

DROP TYPE order_status;
-- +goose StatementEnd