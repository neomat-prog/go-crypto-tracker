package tracker

import (
	"log"
	"time"
)

type TrackerOpts struct {
	Symbols     []string
	Interval    string
	RenderRate  time.Duration
	HistorySize int
	Logger      *log.Logger
}

type Tracker struct {
	TrackerOpts

	transport Transport
	store     map[string]*Ring

	klinech  chan Kline
	addSymch chan string
	quitch   chan struct{}
}

func NewTracker(opts TrackerOpts, tr Transport) *Tracker {
	return &Tracker{
		TrackerOpts: opts,
		transport:   tr,
		store:       make(map[string]*Ring),
		klinech:     make(chan Kline, 1024),
		addSymch:    make(chan string),
		quitch:      make(chan struct{}),
	}
}
