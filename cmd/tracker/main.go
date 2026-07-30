package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/neomat-prog/internal/market"
	"github.com/neomat-prog/internal/tracker"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	tp := market.NewMockTransport("BTCUSDT", 100*time.Millisecond, 0)
	t := tracker.NewTracker(tracker.TrackerOpts{
		Symbols:  []string{"BTCUSDT"},
		RingSize: 100,
		Interval: 500 * time.Millisecond,
	}, tp)

	go func() {
		<-ctx.Done()
		tp.Close()
		t.Stop()
	}()

	t.Start()
}
