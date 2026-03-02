package service

type Rate struct {
	FromCurrency string
	ToCurrency   string
	Rate         float32
}

type Rates map[string]float32
