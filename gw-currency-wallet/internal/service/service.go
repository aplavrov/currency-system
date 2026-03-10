package service

import "context"

type txManager interface {
	RunSerializable(ctx context.Context, fn func(ctxTx context.Context) error) error
	RunReadUncommitted(ctx context.Context, fn func(ctxTx context.Context) error) error
}

type messagesStorage interface {
	Send(ctx context.Context, key string, value any) error
}
