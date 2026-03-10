package delivery

import "github.com/aplavrov/currency-system/gw-currency-wallet/internal/service"

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BalanceResponse struct {
	Balance service.Balance `json:"balance"`
}

type WalletOperationResponse struct {
	Message    string          `json:"message"`
	NewBalance service.Balance `json:"new_balance"`
}

type DepositRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type WithdrawRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type ExchangeRequest struct {
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	Amount       float64 `json:"amount"`
}

type RatesResponse struct {
	Rates service.Rates `json:"rates"`
}

type ExchangeResponse struct {
	Message         string          `json:"message"`
	ExchangedAmount float64         `json:"exchanged_amount"`
	NewBalance      service.Balance `json:"new_balance"`
}
