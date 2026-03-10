package postgresql

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/storage"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type UserStorage struct {
	tx     storage.TxManager
	logger *slog.Logger
}

func NewUserStorage(tx storage.TxManager, logger *slog.Logger) *UserStorage {
	return &UserStorage{tx: tx, logger: logger}
}

func (s *UserStorage) CreateUser(ctx context.Context, params storage.CreateUserParams) error {
	logger := s.logger.With("method", "CreateUser")
	logger.Info("attempting to create user", "username", params.Username)
	_, err := s.tx.GetQueryEngine(ctx).Exec(ctx, "INSERT INTO users (id, username, email, pass_hash, wallet_id) VALUES ($1, $2, $3, $4, $5)", params.UserID, params.Username, params.Email, params.PassHash, params.WalletID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				if strings.Contains(pgErr.ConstraintName, "username") {
					logger.Error("username already exists", "error", err)
					return storage.ErrUsernameExists
				}
				if strings.Contains(pgErr.ConstraintName, "email") {
					logger.Error("email already exists", "error", err)
					return storage.ErrEmailExists
				}
			}
		}
		logger.Error("failed to create user", "error", err)
		return err
	}

	return nil
}

func (s *UserStorage) GetUser(ctx context.Context, username string) (storage.User, error) {
	logger := s.logger.With("method", "GetUser")
	logger.Info("getting user info", "username", username)
	var user storage.User
	err := s.tx.GetQueryEngine(ctx).Get(ctx, &user, "SELECT id, wallet_id, pass_hash FROM users WHERE username=($1)", username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error("user not found", "error", err)
			return storage.User{}, storage.ErrUsernameNotFound
		}
		logger.Error("failed to get user", "error", err)
		return storage.User{}, err
	}
	return user, nil
}
