package tracker

import (
	"fmt"
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
	go t.renderLoop()
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

func (t *Tracker) renderLoop() {
	tk := time.NewTicker(t.Interval)
	defer tk.Stop()
	for {
		select {
		case <-t.quitch:
			return
		case <-tk.C:
			fmt.Print("\033[H\033[2J")
			for _, s := range t.Symbols {
				fmt.Println(s, t.Snapshot(s))
			}
		}
	}
}
