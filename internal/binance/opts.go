package binance

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

var errClosed = errors.New("binance: closed")

type WSOpts struct {
	Symbols  []string
	Interval string
	Backfill int
}

type WS struct {
	WSOpts

	conn *websocket.Conn
	subs []string

	subch    chan string
	errch    chan connErr
	quitch   chan struct{}
	stopOnce sync.Once
}

type connErr struct {
	conn *websocket.Conn
	err  error
}

func NewWS(opts WSOpts) *WS {
	if opts.Interval == "" {
		opts.Interval = "1m"
	}
	ws := &WS{
		WSOpts: opts,
		subch:  make(chan string, subQueue),
		errch:  make(chan connErr),
		quitch: make(chan struct{}),
	}
	for _, s := range opts.Symbols {
		ws.subs = append(ws.subs, streamName(s, opts.Interval))
	}
	return ws
}

func (ws *WS) Close() error {
	ws.stopOnce.Do(func() { close(ws.quitch) })
	return nil
}
