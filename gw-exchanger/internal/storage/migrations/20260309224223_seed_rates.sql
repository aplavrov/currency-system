-- +goose Up
-- +goose StatementBegin
INSERT INTO rates (currency, rate) VALUES
    ('USD',78.24),
    ('RUB',1.00),
    ('EUR',90.97);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
