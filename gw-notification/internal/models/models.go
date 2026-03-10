package models

import "time"

type TransactionEvent struct {
	TransactionID string    `json:"transaction_id"`
	WalletID      string    `json:"wallet_id"`
	Type          string    `json:"type"`
	Currency      string    `json:"currency"`
	Amount        float64   `json:"amount"`
	CreatedAt     time.Time `json:"created_at"`
}
