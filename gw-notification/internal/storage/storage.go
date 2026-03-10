package storage

import (
	"context"

	"github.com/aplavrov/currency-system/gw-notification/internal/models"
)

type Repository interface {
	StoreOperation(ctx context.Context, event models.TransactionEvent) error
}
