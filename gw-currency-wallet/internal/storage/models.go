package storage

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `db:"id"`
	WalletID uuid.UUID `db:"wallet_id"`
	PassHash []byte    `db:"pass_hash"`
}

type CreateUserParams struct {
	UserID   uuid.UUID
	Username string
	Email    string
	PassHash []byte
	WalletID uuid.UUID
}

type WalletBalance struct {
	WalletID uuid.UUID
	Currency string
	Amount   float64
}
