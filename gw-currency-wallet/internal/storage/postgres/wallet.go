package postgresql

import (
	"context"
	"errors"
	"log/slog"

	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

type WalletStorage struct {
	tx     storage.TxManager
	logger *slog.Logger
}

func NewWalletStorage(tx storage.TxManager, logger *slog.Logger) *WalletStorage {
	return &WalletStorage{tx: tx, logger: logger}
}

func (s *WalletStorage) CreateWallet(ctx context.Context, id uuid.UUID) error {
	logger := s.logger.With("method", "CreateWallet")
	logger.Info("attempting to create wallet", "id", id)

	_, err := s.tx.GetQueryEngine(ctx).Exec(ctx, "INSERT INTO wallets (id) VALUES ($1)", id)
	if err != nil {
		logger.Error("failed to create wallet", "error", err)
		return err
	}

	err = s.CreateWalletBalances(ctx, id, []string{"USD", "RUB", "EUR"}...)
	if err != nil {
		return err
	}

	return nil
}

func (s *WalletStorage) CreateWalletBalances(ctx context.Context, id uuid.UUID, currencies ...string) error {
	logger := s.logger.With("method", "CreateWalletBalances")
	logger.Info("attempting to create balances", "id", id, "currencies", currencies)
	for _, currency := range currencies {
		_, err := s.tx.GetQueryEngine(ctx).Exec(ctx, "INSERT INTO balances (wallet_id, currency) VALUES ($1, $2)", id, currency)
		if err != nil {
			logger.Error("failed to create balance", "currency", currency, "error", err)
			return err
		}
	}
	return nil
}

func (s *WalletStorage) GetBalance(ctx context.Context, walletID uuid.UUID) ([]storage.WalletBalance, error) {
	logger := s.logger.With("method", "GetBalance")
	logger.Info("attempting to get balances", "wallet_id", walletID)
	var res []storage.WalletBalance
	err := s.tx.GetQueryEngine(ctx).Select(ctx, &res, "SELECT wallet_id, currency, amount FROM balances WHERE wallet_id=$1", walletID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error("wallet not found")
			return nil, storage.ErrWalletNotFound
		}
		return nil, err
	}
	logger.Info("balances were retrieved")
	return res, nil
}

func (s *WalletStorage) Deposit(ctx context.Context, walletID uuid.UUID, currency string, amount float64) error {
	logger := s.logger.With("method", "Deposit")
	logger.Info("attempting to deposit", "wallet_id", walletID)
	tag, err := s.tx.GetQueryEngine(ctx).Exec(ctx, "UPDATE balances SET amount = amount + $1 WHERE wallet_id = $2 AND currency = $3", amount, walletID, currency)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return storage.ErrWalletNotFound
	}
	logger.Info("deposit successful", "wallet_id", walletID)
	return nil
}

func (s *WalletStorage) Withdraw(ctx context.Context, walletID uuid.UUID, currency string, amount float64) error {
	logger := s.logger.With("method", "Withdraw")
	logger.Info("attempting to withdraw", "wallet_id", walletID)
	tag, err := s.tx.GetQueryEngine(ctx).Exec(ctx, "UPDATE balances SET amount = amount - $1 WHERE wallet_id = $2 AND currency = $3 AND amount >= $1", amount, walletID, currency)
	if err != nil {
		return err
	}

	if tag.RowsAffected() > 0 {
		return nil
	}

	var exists bool
	err = s.tx.GetQueryEngine(ctx).Get(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM wallets WHERE id = $1)`,
		walletID,
	)
	if err != nil {
		return err
	}

	if !exists {
		return storage.ErrWalletNotFound
	}

	return storage.ErrInsufficientFunds
}
