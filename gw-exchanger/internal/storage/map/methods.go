package _map

import (
	"context"

	"github.com/aplavrov/currency-system/gw-exchanger/internal/storage"
)

type MapStorage struct {
	items map[string]float32
}

func New() *MapStorage {
	res := make(map[string]float32)
	res["USD"] = 77.18
	res["EUR"] = 91.23
	res["RUB"] = 1.0
	return &MapStorage{items: res}
}

func (s *MapStorage) GetAllRates(ctx context.Context) ([]storage.Rate, error) {
	var res []storage.Rate
	for currency, rate := range s.items {
		res = append(res, storage.Rate{Currency: currency, Rate: rate})
	}
	return res, nil
}

func (s *MapStorage) GetCurrencyRate(ctx context.Context, currency string) (storage.Rate, error) {
	return storage.Rate{Currency: currency, Rate: s.items[currency]}, nil
}
