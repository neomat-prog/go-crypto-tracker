package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/neomat-prog/cmd/cli"
	"github.com/neomat-prog/cmd/config"
	"github.com/neomat-prog/internal/Graph"
	"github.com/neomat-prog/internal/binance"
	"github.com/neomat-prog/internal/tracker"
	"golang.org/x/term"
)

const (
	headerRows = 1

	minHeight = 8
	minWidth  = 20
)

func main() {
	color := flag.String("color", "auto", "color depth: auto, truecolor, 256, 16, none")
	ascii := flag.Bool("ascii", false, "draw with ASCII only, for terminals without Unicode")
	flag.Parse()

	style, err := resolveStyle(*color, *ascii)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.LoadDotEnv(".env")
	if err != nil {
		log.Fatal(err)
	}

	sc := cli.NewScreen(true)
	defer sc.Close()

	tp := binance.NewWS(binance.WSOpts{
		Symbol:   cfg.Symbol,
		Interval: cfg.Interval,
		Backfill: cfg.Backfill,
	})

	t := tracker.NewTracker(tracker.TrackerOpts{
		Symbol:   cfg.Symbol,
		RingSize: 150,
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
			klines := t.SnapshotKlines(cfg.Symbol)

			candles := make([]Graph.Candle, len(klines))
			for i, k := range klines {
				candles[i] = Graph.Candle{
					Open:  k.Open,
					High:  k.High,
					Low:   k.Low,
					Close: k.Close,
				}
			}

			height, width := chartSize()
			if height < minHeight || width < minWidth {
				sc.Frame("terminal too small\x1b[K\n")
				continue
			}

			sc.Frame(Graph.RenderCandles(
				cfg.Symbol,
				candles,
				height,
				width,
				style,
			))
		}
	}
}

func chartSize() (height, width int) {
	cols, rows, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || cols <= 0 || rows <= 0 {
		return 20, 150
	}

	height = rows - headerRows
	width = cols - Graph.AxisWidth

	if height < 2 {
		height = 2
	}
	if width < 1 {
		width = 1
	}
	return height, width
}

func resolveStyle(color string, ascii bool) (Graph.Style, error) {
	st := Graph.DetectStyle()

	switch color {
	case "auto":
	case "truecolor":
		st.Color = Graph.ColorTrue
	case "256":
		st.Color = Graph.Color256
	case "16":
		st.Color = Graph.Color16
	case "none":
		st.Color = Graph.ColorNone
	default:
		return st, fmt.Errorf("--color %q is not one of auto, truecolor, 256, 16, none", color)
	}

	if ascii {
		st.ASCII = true
	}
	return st, nil
}
