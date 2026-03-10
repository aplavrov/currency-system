-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username VARCHAR(20) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    pass_hash TEXT NOT NULL,
    wallet_id UUID NOT NULL UNIQUE,

    CONSTRAINT wallet_id_fk FOREIGN KEY (wallet_id) REFERENCES wallets(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd