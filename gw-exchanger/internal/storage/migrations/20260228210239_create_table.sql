-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS rates (
    currency CHAR(3) PRIMARY KEY NOT NULL,
    rate FLOAT NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wallets;
-- +goose StatementEnd
