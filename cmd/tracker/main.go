package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/neomat-prog/internal"
	"github.com/neomat-prog/internal/binance"
	"github.com/neomat-prog/internal/tracker"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	tp := binance.NewWS(binance.WSOpts{
		Symbols:  []string{"ETHUSDT"},
		Interval: "1s",
		Backfill: 100,
	})

	t := tracker.NewTracker(tracker.TrackerOpts{
		Symbols:  []string{"ETHUSDT"},
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
			closes := t.Snapshot("ETHUSDT")

			fmt.Print("\033[2J\033[H")

			fmt.Print(internal.Render(
				"ETHUSDT",
				closes,
				15,
				80,
			))
		}
	}
}
