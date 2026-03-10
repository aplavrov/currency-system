package tx_manager

import (
	"context"

	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/storage"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"

	postgresql "github.com/aplavrov/currency-system/gw-currency-wallet/internal/storage/postgres"
)

type txManagerKey struct{}

type TxManager struct {
	pool *postgresql.Database
}

func NewTxManager(pool *postgresql.Database) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) RunSerializable(ctx context.Context, fn func(ctxTx context.Context) error) error {
	opts := pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	}
	return m.beginFunc(ctx, opts, fn)
}

func (m *TxManager) RunReadUncommitted(ctx context.Context, fn func(ctxTx context.Context) error) error {
	opts := pgx.TxOptions{
		IsoLevel:   pgx.ReadUncommitted,
		AccessMode: pgx.ReadOnly,
	}
	return m.beginFunc(ctx, opts, fn)
}

func (m *TxManager) beginFunc(ctx context.Context, opts pgx.TxOptions, fn func(ctxTx context.Context) error) error {
	tx, err := m.pool.GetPool().BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	ctx = context.WithValue(ctx, txManagerKey{}, &txQueryEngine{tx: tx})
	if err := fn(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (m *TxManager) GetQueryEngine(ctx context.Context) storage.QueryEngine {
	v, ok := ctx.Value(txManagerKey{}).(storage.QueryEngine)
	if ok && v != nil {
		return v
	}
	return m.pool
}

type txQueryEngine struct {
	tx pgx.Tx
}

func (t *txQueryEngine) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return t.tx.Exec(ctx, query, args...)
}

func (t *txQueryEngine) ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return t.tx.QueryRow(ctx, query, args...)
}

func (t *txQueryEngine) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, t.tx, dest, query, args...)
}

func (t *txQueryEngine) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, t.tx, dest, query, args...)
}
