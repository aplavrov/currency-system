package server

import (
	"context"
	"errors"
	"log/slog"

	"github.com/aplavrov/currency-system/gw-exchanger/internal/service"
	"github.com/aplavrov/currency-system/proto-exchange/exchange"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type exchangeService interface {
	GetRates(ctx context.Context) (service.Rates, error)
	GetRate(ctx context.Context, from string, to string) (service.Rate, error)
}

type exchangeServer struct {
	exchangeService exchangeService
	logger          *slog.Logger
	exchange_grpc.UnimplementedExchangeServiceServer
}

func Register(gRPC *grpc.Server, s exchangeService, logger *slog.Logger) {
	exchange_grpc.RegisterExchangeServiceServer(gRPC, &exchangeServer{exchangeService: s, logger: logger})
}

func (s *exchangeServer) GetExchangeRates(ctx context.Context, empty *exchange_grpc.Empty) (*exchange_grpc.ExchangeRatesResponse, error) {
	logger := s.logger.With(
		"method", "GetExchangeRates",
	)
	logger.Info("get all exchange rates request")
	rates, err := s.exchangeService.GetRates(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &exchange_grpc.ExchangeRatesResponse{Rates: rates}, nil
}

func (s *exchangeServer) GetExchangeRateForCurrency(ctx context.Context, req *exchange_grpc.CurrencyRequest) (*exchange_grpc.ExchangeRateResponse, error) {
	logger := s.logger.With(
		"method", "GetExchangeRateForCurrency",
	)
	logger.Info("get exchange rate request", "from", req.GetFromCurrency(), "to", req.GetToCurrency())
	rate, err := s.exchangeService.GetRate(ctx, req.GetFromCurrency(), req.GetToCurrency())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCurrency):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, service.ErrInvalidRate):
			return nil, status.Error(codes.Internal, err.Error())
		case errors.Is(err, service.ErrWrongCurrency):
			return nil, status.Error(codes.Internal, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &exchange_grpc.ExchangeRateResponse{
		FromCurrency: rate.FromCurrency,
		ToCurrency:   rate.ToCurrency,
		Rate:         rate.Rate,
	}, nil
}
