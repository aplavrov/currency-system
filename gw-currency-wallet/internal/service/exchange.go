package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/storage"
	"github.com/google/uuid"
)

type RateProvider interface {
	GetExchangeRates(ctx context.Context) (Rates, error)
	GetExchangeRateForCurrency(ctx context.Context, fromCurrency string, toCurrency string) (float32, error)
}

type ExchangeService struct {
	walletStorage storage.WalletStorage
	txManager     txManager
	rateProvider  RateProvider
	logger        *slog.Logger
}

func NewExchangeService(walletStorage storage.WalletStorage, txManager txManager, rateProvider RateProvider, logger *slog.Logger) *ExchangeService {
	return &ExchangeService{walletStorage: walletStorage, txManager: txManager, rateProvider: rateProvider, logger: logger}
}

func (s *ExchangeService) Exchange(ctx context.Context, walletID uuid.UUID, FromCurrency string, ToCurrency string, Amount float64) (float64, Balance, error) {
	logger := s.logger.With("method", "Exchange")
	logger.Info("initiating exchange", "from", FromCurrency, "to", ToCurrency)

	if !isValidCurrency(FromCurrency) || !isValidCurrency(ToCurrency) {
		return 0, nil, ErrInvalidCurrency
	}

	rate, err := s.rateProvider.GetExchangeRateForCurrency(ctx, FromCurrency, ToCurrency)
	if err != nil {
		logger.Error("failed to get rate", "error", err)
		return 0, nil, err
	}

	if !isValidAmount(Amount) {
		return 0, nil, ErrInvalidAmount
	}

	exchangedAmount := Amount * float64(rate)
	result := make(Balance)

	err = s.txManager.RunSerializable(ctx, func(ctx context.Context) error {
		err := s.walletStorage.Withdraw(ctx, walletID, FromCurrency, Amount)
		if err != nil {
			if errors.Is(err, storage.ErrInsufficientFunds) {
				return ErrInsufficientFunds
			}
			if errors.Is(err, storage.ErrWalletNotFound) {
				return ErrWalletNotFound
			}
			return err
		}

		err = s.walletStorage.Deposit(ctx, walletID, ToCurrency, exchangedAmount)
		if err != nil {
			return err
		}

		rows, err := s.walletStorage.GetBalance(ctx, walletID)
		if err != nil {
			return err
		}
		for _, row := range rows {
			if row.Currency == FromCurrency || row.Currency == ToCurrency {
				result[row.Currency] = row.Amount
			}
		}
		return err
	})

	if err != nil {
		return 0, nil, err
	}

	return exchangedAmount, result, nil
}

func (s *ExchangeService) GetRates(ctx context.Context) (Rates, error) {
	return s.rateProvider.GetExchangeRates(ctx)
}
