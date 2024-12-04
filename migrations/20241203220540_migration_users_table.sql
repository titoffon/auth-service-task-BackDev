-- +goose Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255),
    last_ip_address VARCHAR(255)
);

-- +goose StatementBegin

-- +goose StatementEnd

-- +goose Down
DROP TABLE users;
-- +goose StatementBegin
-- +goose StatementEnd
