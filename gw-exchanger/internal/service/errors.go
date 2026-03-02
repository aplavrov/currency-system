package service

import "errors"

var ErrInvalidRate = errors.New("invalid rate")
var ErrInvalidCurrency = errors.New("invalid currency")
var ErrWrongCurrency = errors.New("got wrong currency from db")
