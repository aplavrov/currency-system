package mongodb

import (
	"context"
	"log/slog"

	"github.com/aplavrov/currency-system/gw-notification/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
)

type OperationStorage struct {
	Collection *mongo.Collection
	logger     *slog.Logger
}

func NewOperationStorage(client *mongo.Client, dbName string, collectionName string, logger *slog.Logger) *OperationStorage {
	return &OperationStorage{Collection: client.Database(dbName).Collection(collectionName), logger: logger}
}

func (s *OperationStorage) StoreOperation(ctx context.Context, event models.TransactionEvent) error {
	logger := s.logger.With("method", "StoreOperation")
	_, err := s.Collection.InsertOne(ctx, event)
	if err != nil {
		logger.Error("error storing transaction in MongoDB", "transaction_id", event.TransactionID, "error", err)
		return err
	}
	logger.Info("transaction successfully stored", "transaction_id", event.TransactionID)
	return nil
}
