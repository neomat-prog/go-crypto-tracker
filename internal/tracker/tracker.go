package tracker

import (
	"time"

	"github.com/neomat-prog/internal/market"
)

type TrackerOpts struct {
	Symbols  []string
	RingSize int
	Interval time.Duration
}

func (t *Tracker) Start() error {
	go t.transport.Start(t.klinech)
	close(t.readych)
	t.loop()
	return nil
}

func (t *Tracker) loop() {
	for {
		select {
		case <-t.quitch:
			return
		case k := <-t.klinech:
			r, ok := t.store[k.Symbol]
			if !ok {
				r = market.NewRing(t.RingSize)
				t.store[k.Symbol] = r
			}
			r.Push(k)
		case req := <-t.snapch:
			if r, ok := t.store[req.symbol]; ok {
				req.respch <- r.Closes()
			} else {
				req.respch <- nil
			}
		}
	}
}

func (t *Tracker) Snapshot(symbol string) []float64 {
	select {
	case <-t.quitch:
		return nil
	case <-t.readych:
	}
	req := snapReq{symbol: symbol, respch: make(chan []float64, 1)}
	select {
	case <-t.quitch:
		return nil
	case t.snapch <- req:
	}
	select {
	case <-t.quitch:
		return nil
	case closes := <-req.respch:
		return closes
	}
}
