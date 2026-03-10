package service

import "errors"

var ErrUsernameOrEmailExists = errors.New("username or email already exists")
var ErrInvalidCredentials = errors.New("invalid username or password")
var ErrWalletNotFound = errors.New("wallet not found")
var ErrInvalidCurrency = errors.New("invalid currency")
var ErrInvalidAmount = errors.New("invalid amount")
var ErrInsufficientFunds = errors.New("insufficient funds")
