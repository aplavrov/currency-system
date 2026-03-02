package storage

import "context"

type ExchangeStorage interface {
	GetAllRates(ctx context.Context) ([]Rate, error)
	GetCurrencyRate(ctx context.Context, currency string) (Rate, error)
}
