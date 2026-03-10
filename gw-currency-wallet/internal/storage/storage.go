package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type UserStorage interface {
	CreateUser(ctx context.Context, params CreateUserParams) error
	GetUser(ctx context.Context, username string) (User, error)
}
type WalletStorage interface {
	CreateWallet(ctx context.Context, id uuid.UUID) error
	GetBalance(ctx context.Context, walletID uuid.UUID) ([]WalletBalance, error)
	Deposit(ctx context.Context, walletID uuid.UUID, currency string, amount float64) error
	Withdraw(ctx context.Context, walletID uuid.UUID, currency string, amount float64) error
}

type TxManager interface {
	GetQueryEngine(ctx context.Context) QueryEngine
}

type QueryEngine interface {
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
}
