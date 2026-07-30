package market

import "time"

type Kline struct {
	Symbol   string
	OpenTime time.Time
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Closed   bool // <- binance "x" field
}

type Transport interface {
	Start(ch chan<- Kline) error
	Subscribe(symbol string) error
	Close() error
}
