package storage

import "errors"

var ErrUsernameNotFound = errors.New("username not found")
var ErrUsernameExists = errors.New("username already exists")
var ErrEmailExists = errors.New("email already exists")
var ErrWalletNotFound = errors.New("wallet not found")
var ErrInsufficientFunds = errors.New("insufficient funds")
