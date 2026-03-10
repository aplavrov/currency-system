-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY
);
CREATE TABLE IF NOT EXISTS balances (
    wallet_id UUID NOT NULL,
    currency CHAR(3) NOT NULL,
    amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    PRIMARY KEY (wallet_id, currency),

    CONSTRAINT wallet_id_fk FOREIGN KEY (wallet_id) REFERENCES wallets(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS balances;
DROP TABLE IF EXISTS wallets;
-- +goose StatementEnd
