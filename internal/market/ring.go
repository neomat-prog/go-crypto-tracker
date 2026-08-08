package market

type Ring struct {
	buf  []Kline
	head int // head := idx of oldest candle
	n    int // count of valid candles currently stored
}

func NewRing(size int) *Ring {
	return &Ring{
		buf: make([]Kline, size),
		// head = 0
		// n = 0
	}
}

func (r *Ring) Push(k Kline) {

	if len(r.buf) == 0 {
		return
	}

	if r.n > 0 {
		last := (r.head + r.n - 1) % len(r.buf)

		if r.buf[last].OpenTime.Equal(k.OpenTime) {
			r.buf[last] = k
			return
		}
	}

	if r.n < len(r.buf) {
		idx := (r.head + r.n) % len(r.buf)
		r.buf[idx] = k
		r.n++
		return
	}

	r.buf[r.head] = k
	r.head = (r.head + 1) % len(r.buf)
}

func (r *Ring) Closes() []float64 {
	closes := make([]float64, r.n)

	for i := 0; i < r.n; i++ {
		idx := (r.head + i) % len(r.buf)
		closes[i] = r.buf[idx].Close
	}

	return closes
}

func (r *Ring) Klines() []Kline {
	klines := make([]Kline, r.n)

	for i := 0; i < r.n; i++ {
		idx := (r.head + i) % len(r.buf)
		klines[i] = r.buf[idx]
	}

	return klines
}
