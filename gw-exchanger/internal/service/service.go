package service

import (
	"context"
	"log/slog"

	"github.com/aplavrov/currency-system/gw-exchanger/internal/storage"
)

type ExchangeService struct {
	storage storage.ExchangeStorage
	logger  *slog.Logger
}

func New(s storage.ExchangeStorage, logger *slog.Logger) *ExchangeService {
	return &ExchangeService{storage: s, logger: logger}
}

func (s *ExchangeService) GetRates(ctx context.Context) (Rates, error) {
	logger := s.logger.With("method", "GetRates")
	logger.Info("get all rates request")
	rates, err := s.storage.GetAllRates(ctx)
	if err != nil {
		return Rates{}, err
	}

	res := Rates{}
	for _, item := range rates {
		res[item.Currency] = item.Rate
	}

	return res, nil
}

func isValidCurrency(currency string) bool {
	valid := map[string]struct{}{"USD": {}, "RUB": {}, "EUR": {}}
	_, ok := valid[currency]
	return ok
}

func isValidRate(rate float32) bool {
	return rate > 0
}

func (s *ExchangeService) GetRate(ctx context.Context, from string, to string) (Rate, error) {
	logger := s.logger.With("method", "GetRate")
	logger.Info("get rate request", "from", from, "to", to)
	if !isValidCurrency(from) || !isValidCurrency(to) {
		logger.Error("invalid currency", "error", ErrInvalidCurrency)
		return Rate{}, ErrInvalidCurrency
	}

	fromRate, err := s.storage.GetCurrencyRate(ctx, from)
	if err != nil {
		return Rate{}, err
	}

	toRate, err := s.storage.GetCurrencyRate(ctx, to)
	if err != nil {
		return Rate{}, err
	}

	if !isValidRate(fromRate.Rate) || !isValidRate(toRate.Rate) {
		logger.Error("invalid rate", "error", ErrInvalidRate)
		return Rate{}, ErrInvalidRate
	}

	if from != fromRate.Currency || to != toRate.Currency {
		logger.Error("wrong currency", "error", ErrWrongCurrency)
		return Rate{}, ErrWrongCurrency
	}

	return Rate{
		FromCurrency: fromRate.Currency,
		ToCurrency:   toRate.Currency,
		Rate:         fromRate.Rate / toRate.Rate,
	}, nil
}
