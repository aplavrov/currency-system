package postgresql

import (
	"context"
	"log/slog"

	"github.com/aplavrov/currency-system/gw-exchanger/internal/storage"
)

type ExchangeStorage struct {
	db     DB
	logger *slog.Logger
}

func NewExchangeStorage(database DB, logger *slog.Logger) *ExchangeStorage {
	return &ExchangeStorage{db: database, logger: logger}
}

func (s *ExchangeStorage) GetAllRates(ctx context.Context) ([]storage.Rate, error) {
	logger := s.logger.With("method", "GetAllRates")
	logger.Info("get all rates request")
	var res []storage.Rate
	err := s.db.Select(ctx, &res, "SELECT * FROM rates")
	if err != nil {
		logger.Error("db error", "error", err)
		return nil, err
	}
	return res, nil
}
func (s *ExchangeStorage) GetCurrencyRate(ctx context.Context, currency string) (storage.Rate, error) {
	logger := s.logger.With("method", "GetCurrencyRate")
	logger.Info("get rate request", "currency", currency)
	var res storage.Rate
	err := s.db.Get(ctx, &res, "SELECT * FROM rates WHERE currency=$1", currency)
	if err != nil {
		logger.Error("db error", "error", err)
		return storage.Rate{}, err
	}
	return res, nil
}
