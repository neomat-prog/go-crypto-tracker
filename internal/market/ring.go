package market

type Ring struct {
	buf  []Kline
	head int
	n    int
}

func NewRing(size int) *Ring {
	return &Ring{
		buf: make([]Kline, size),
		// head = 0
		// n = 0
	}
}

// push add k, oldest out when full
func (r *Ring) Push(k Kline) {
	if r.n < len(r.buf) {
		idx := (r.head + r.n) % len(r.buf)
		r.buf[idx] = k
		r.n++
		return
	}

	r.buf[r.head] = k
	r.head = (r.head + 1) % len(r.buf)
}
