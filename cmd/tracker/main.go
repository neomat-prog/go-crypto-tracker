package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/neomat-prog/cmd/config"
	"github.com/neomat-prog/internal"
	"github.com/neomat-prog/internal/binance"
	"github.com/neomat-prog/internal/tracker"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.LoadDotEnv(".env")
	if err != nil {
		log.Fatal(err)
	}

	tp := binance.NewWS(binance.WSOpts{
		Symbol:   cfg.Symbol,
		Interval: cfg.Interval,
		Backfill: cfg.Backfill,
	})

	t := tracker.NewTracker(tracker.TrackerOpts{
		Symbol:   cfg.Symbol,
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
			closes := t.Snapshot(cfg.Symbol)

			fmt.Print("\033[2J\033[H")

			fmt.Print(internal.Render(
				cfg.Symbol,
				closes,
				15,
				80,
			))
		}
	}
}
