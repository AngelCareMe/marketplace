-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS customers (
    user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    first_name TEXT,
    last_name TEXT,
    phone TEXT,
    date_birth TIMESTAMP WITH TIME ZONE,
    address TEXT
);
-- +goose StatementEnd