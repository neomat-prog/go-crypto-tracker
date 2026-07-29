package market

import (
	"slices"
	"testing"
	"time"
)

func kl(min int, close float64) Kline {
	return Kline{
		OpenTime: time.Date(2026, 1, 1, 0, min, 0, 0, time.UTC),
		Close:    close,
	}
}

func TestRingPush(t *testing.T) {
	tests := []struct {
		name string
		size int
		in   []Kline
		want []float64
	}{
		{"empty", 3, nil, []float64{}},
		{"partial", 3, []Kline{kl(0, 1), kl(1, 2)}, []float64{1, 2}},
		{"exact", 3, []Kline{kl(0, 1), kl(1, 2), kl(2, 3)}, []float64{1, 2, 3}},
		{"wraps", 3, []Kline{kl(0, 1), kl(1, 2), kl(2, 3), kl(3, 4)}, []float64{2, 3, 4}},
		{"wraps twice", 2, []Kline{kl(0, 1), kl(1, 2), kl(2, 3), kl(3, 4), kl(4, 5)}, []float64{4, 5}},
		{"open candle overwrites", 3,
			[]Kline{kl(0, 1), kl(1, 2), kl(1, 99)},
			[]float64{1, 99}},
		{"open candle overwrites at wrap", 2,
			[]Kline{kl(0, 1), kl(1, 2), kl(2, 3), kl(2, 30)},
			[]float64{2, 30}},
		{"size 1", 1, []Kline{kl(0, 1), kl(1, 2)}, []float64{2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRing(tt.size)
			for _, k := range tt.in {
				r.Push(k)
			}
			got := r.Closes()
			if !slices.Equal(got, tt.want) {
				t.Errorf("Closes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRingClose(t *testing.T) {
	r := NewRing(3)
	r.Push(kl(0, 1))
	a := r.Closes()
	a[0] = 999
	if b := r.Closes(); b[0] != 1 {
		t.Fatalf("Closes() returned shared memory: got %v after mutating caller copy", b)
	}
}
