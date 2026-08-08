package tracker

import (
	"sync"

	"github.com/neomat-prog/internal/market"
)

type Tracker struct {
	TrackerOpts

	transport market.Transport
	store     map[string]*market.Ring

	klinech      chan market.Kline
	addSymch     chan string
	stopOnce     sync.Once
	readych      chan struct{}
	snapch       chan snapReq
	snapKlinesch chan snapKlinesReq
	quitch       chan struct{}
}

type snapReq struct {
	symbol string
	respch chan []float64
}

type snapKlinesReq struct {
	symbol string
	respch chan []market.Kline
}

func NewTracker(opts TrackerOpts, tr market.Transport) *Tracker {
	return &Tracker{
		TrackerOpts:  opts,
		transport:    tr,
		store:        make(map[string]*market.Ring),
		klinech:      make(chan market.Kline, 1024),
		addSymch:     make(chan string),
		readych:      make(chan struct{}),
		snapch:       make(chan snapReq),
		snapKlinesch: make(chan snapKlinesReq),
		quitch:       make(chan struct{}),
	}
}

func (t *Tracker) Stop() {
	t.stopOnce.Do(func() { close(t.quitch) })
}
