package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/neomat-prog/internal"
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
		_ = tp.Close()
		t.Stop()
	}()

	go t.Start()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			closes := t.Snapshot("BTCUSDT")

			fmt.Print("\033[2J\033[H")

			fmt.Print(internal.Render(
				"BTCUSDT",
				closes,
				15,
				80,
			))
		}
	}
}
