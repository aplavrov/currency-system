package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aplavrov/currency-system/gw-notification/internal/models"
	"github.com/segmentio/kafka-go"
)

type notificationService interface {
	HandleEvent(ctx context.Context, event models.TransactionEvent) error
}

type Consumer struct {
	service notificationService
	logger  *slog.Logger
	reader  *kafka.Reader
}

func NewConsumer(service notificationService, logger *slog.Logger, reader *kafka.Reader) *Consumer {
	return &Consumer{service: service, logger: logger, reader: reader}
}

func (c *Consumer) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.logger.Error("error reading message:", "error", err)
				continue
			}
			var event models.TransactionEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				c.logger.Error("error unmarshalling notification:", "error", err)
				continue
			}
			c.logger.Info("read message", "transaction_id", event.TransactionID)
			if err := c.service.HandleEvent(ctx, event); err != nil {
				c.logger.Error("error handling event", "error", err)
				continue
			}
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("error committing message:", "error", err)
			}
		}
	}
}
