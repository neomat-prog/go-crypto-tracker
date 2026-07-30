package market

import (
	"math"
	"sync"
	"time"
)

type MockTransport struct {
	Symbol string
	Delay  time.Duration
	Steps  int

	quitch chan struct{}
	once   sync.Once
}

func NewMockTransport(symbol string, delay time.Duration, steps int) *MockTransport {
	return &MockTransport{
		Symbol: symbol,
		Delay:  delay,
		Steps:  steps,
		quitch: make(chan struct{}),
	}
}

func (m *MockTransport) Start(ch chan<- Kline) error {
	base := time.Now().UTC().Truncate(time.Minute)
	for i := 0; m.Steps == 0 || i < m.Steps; i++ {
		mid := 100 + 10*math.Sin(float64(i)/5)
		k := Kline{
			Symbol:   m.Symbol,
			OpenTime: base.Add(time.Duration(i) * time.Minute),
			Open:     mid,
			High:     mid + 1,
			Low:      mid - 1,
			Close:    mid,
			Closed:   true,
		}

		select {
		case <-m.quitch:
			return nil
		case ch <- k:
		}

		if m.Delay > 0 {
			select {
			case <-m.quitch:
				return nil
			case <-time.After(m.Delay):
			}
		}
	}
	return nil
}

func (m *MockTransport) Subscribe(string) error { return nil }

func (m *MockTransport) Close() error {
	m.once.Do(func() {
		close(m.quitch)
	})
	return nil
}
