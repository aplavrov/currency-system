package grpcclient

import (
	"context"
	"log/slog"

	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/service"
	"github.com/aplavrov/currency-system/proto-exchange/exchange"
)

type rateCache interface {
	Set(from string, to string, rate float32)
	Get(from string, to string) (float32, bool)
}
type ExchangeClient struct {
	client exchange_grpc.ExchangeServiceClient
	cache  rateCache
	logger *slog.Logger
}

func NewExchangeClient(client exchange_grpc.ExchangeServiceClient, cache rateCache, logger *slog.Logger) *ExchangeClient {
	return &ExchangeClient{client: client, cache: cache, logger: logger}
}

func (c *ExchangeClient) GetExchangeRates(ctx context.Context) (service.Rates, error) {
	logger := c.logger.With("method", "GetExchangeRates")
	logger.Info("sending request through gRPC")
	resp, err := c.client.GetExchangeRates(ctx, &exchange_grpc.Empty{})
	if err != nil {
		logger.Error("internal error", "error", err)
		return nil, err
	}

	return resp.Rates, nil
}

func (c *ExchangeClient) GetExchangeRateForCurrency(ctx context.Context, fromCurrency string, toCurrency string) (float32, error) {
	logger := c.logger.With("method", "GetExchangeRateForCurrency")

	if rate, ok := c.cache.Get(fromCurrency, toCurrency); ok {
		logger.Info("getting data from the cache")
		return rate, nil
	}

	logger.Info("sending request through gRPC")
	resp, err := c.client.GetExchangeRateForCurrency(ctx, &exchange_grpc.CurrencyRequest{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
	})
	if err != nil {
		return 0, err
	}
	c.cache.Set(fromCurrency, toCurrency, resp.Rate)

	return resp.Rate, nil
}
