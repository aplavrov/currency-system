package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/storage"
	"github.com/google/uuid"
)

type WalletService struct {
	walletStorage   storage.WalletStorage
	messagesStorage messagesStorage
	txManager       txManager
	logger          *slog.Logger
}

func NewWalletService(walletStorage storage.WalletStorage, messagesStorage messagesStorage, txManager txManager, logger *slog.Logger) *WalletService {
	return &WalletService{walletStorage: walletStorage, messagesStorage: messagesStorage, txManager: txManager, logger: logger}
}

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (Balance, error) {
	logger := s.logger.With("method", "GetBalance")
	logger.Info("attempting to get balance", "walletID", walletID)
	rows, err := s.walletStorage.GetBalance(ctx, walletID)
	if err != nil {
		if errors.Is(err, storage.ErrWalletNotFound) {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}

	result := make(Balance)
	for _, row := range rows {
		result[row.Currency] = row.Amount
	}
	logger.Info("balance retrieved", "walletID", walletID)
	return result, nil
}

func isValidCurrency(currency string) bool {
	return currency == "USD" || currency == "RUB" || currency == "EUR"
}

func isValidAmount(amount float64) bool {
	return amount > 0
}

func isHugeAmount(amount float64) bool {
	return amount >= 30000
}

func (s *WalletService) Deposit(ctx context.Context, walletID uuid.UUID, currency string, amount float64) (Balance, error) {
	logger := s.logger.With("method", "Deposit")
	logger.Info("attempting to deposit", "walletID", walletID)

	if !isValidCurrency(currency) {
		return nil, ErrInvalidCurrency
	}

	if !isValidAmount(amount) {
		return nil, ErrInvalidAmount
	}

	var res Balance
	err := s.txManager.RunSerializable(ctx, func(ctx context.Context) error {
		if err := s.walletStorage.Deposit(ctx, walletID, currency, amount); err != nil {
			if errors.Is(err, storage.ErrWalletNotFound) {
				return ErrWalletNotFound
			}
			return err
		}

		balance, err := s.GetBalance(ctx, walletID)
		if err != nil {
			return err
		}

		res = balance
		return nil
	})

	if err == nil && isHugeAmount(amount) {
		go func() {
			ctxKafka, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := s.messagesStorage.Send(ctxKafka, walletID.String(), TransactionEvent{
				TransactionID: uuid.New().String(),
				WalletID:      walletID.String(),
				Type:          "Deposit",
				Currency:      currency,
				Amount:        amount,
				CreatedAt:     time.Now(),
			})
			if err != nil {
				logger.Error("failed to send kafka event", "error", err)
			} else {
				logger.Info("message sent to kafka")
			}
		}()
	}

	return res, err
}

func (s *WalletService) Withdraw(ctx context.Context, walletID uuid.UUID, currency string, amount float64) (Balance, error) {
	logger := s.logger.With("method", "Withdraw")
	logger.Info("attempting to withdraw", "walletID", walletID)

	if !isValidCurrency(currency) {
		return nil, ErrInvalidCurrency
	}

	if !isValidAmount(amount) {
		return nil, ErrInvalidAmount
	}

	var res Balance
	err := s.txManager.RunSerializable(ctx, func(ctx context.Context) error {
		if err := s.walletStorage.Withdraw(ctx, walletID, currency, amount); err != nil {
			if errors.Is(err, storage.ErrInsufficientFunds) {
				return ErrInsufficientFunds
			}
			if errors.Is(err, storage.ErrWalletNotFound) {
				return ErrWalletNotFound
			}
			return err
		}

		balance, err := s.GetBalance(ctx, walletID)
		if err != nil {
			return err
		}

		res = balance
		return nil
	})

	if err == nil && isHugeAmount(amount) {
		go func() {
			ctxKafka, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := s.messagesStorage.Send(ctxKafka, walletID.String(), TransactionEvent{
				TransactionID: uuid.New().String(),
				WalletID:      walletID.String(),
				Type:          "Withdraw",
				Currency:      currency,
				Amount:        amount,
				CreatedAt:     time.Now(),
			})
			if err != nil {
				logger.Error("failed to send kafka event", "error", err)
			} else {
				logger.Info("message sent to kafka")
			}
		}()
	}

	return res, err
}
