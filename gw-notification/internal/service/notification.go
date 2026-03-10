package service

import (
	"context"
	"log/slog"

	"github.com/aplavrov/currency-system/gw-notification/internal/models"
	"github.com/aplavrov/currency-system/gw-notification/internal/storage"
)

type NotificationService struct {
	storage storage.Repository
	logger  *slog.Logger
}

func NewNotificationService(storage storage.Repository, logger *slog.Logger) *NotificationService {
	return &NotificationService{storage: storage, logger: logger}
}

func (s *NotificationService) HandleEvent(ctx context.Context, event models.TransactionEvent) error {
	logger := s.logger.With("method", "HandleEvent")
	if err := s.storage.StoreOperation(ctx, event); err != nil {
		logger.Error("failed to store transaction", "transaction_id", event.TransactionID, "error", err)
		return err
	}

	logger.Info("transaction stored", "transaction_id", event.TransactionID)
	return nil
}
