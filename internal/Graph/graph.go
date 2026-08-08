package Graph

import (
	"fmt"
	"math"
	"strings"

	"github.com/guptarohit/asciigraph"
)

type Candle struct {
	Open  float64
	Close float64
	High  float64
	Low   float64
}

const tickSize = 0.01

func RenderGraph(symbol string, closes []float64, height, width int) string {
	if len(closes) < 2 {
		return fmt.Sprintf("need at least 2 closing prices, got %d\n", len(closes))
	}

	first := closes[0]
	last := closes[len(closes)-1]
	delta := last - first
	pct := 0.0

	if first != 0 {
		pct = (delta / first) * 100
	}

	header := fmt.Sprintf(
		"%s %.2f %+0.2f (%+0.2f%%)\n",
		symbol,
		last,
		delta,
		pct,
	)

	lo, hi := niceBounds(closes, height)

	graph := asciigraph.Plot(
		closes,
		asciigraph.Height(height),
		asciigraph.Width(width),
		asciigraph.Precision(2),
		asciigraph.LowerBound(lo),
		asciigraph.UpperBound(hi),
	)

	return header + graph + "\n"
}

func niceBounds(vs []float64, height int) (float64, float64) {
	lo, hi := vs[0], vs[0]
	for _, v := range vs {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}

	return boundsFromRange(lo, hi, height)
}

func boundsFromRange(lo, hi float64, height int) (float64, float64) {
	span := hi - lo
	if span <= 0 {
		span = tickSize * float64(height)
	}

	rows := float64(height)
	step := niceStep(span / rows)
	bound := math.Floor(lo/step) * step

	for bound+step*rows < hi {
		step = niceStep(step * 1.5)
		bound = math.Floor(lo/step) * step
	}

	return bound, bound + step*rows
}

func candleBounds(candles []Candle, height int) (float64, float64) {
	lo, hi := candles[0].Low, candles[0].High
	for _, c := range candles {
		lo = math.Min(lo, c.Low)
		hi = math.Max(hi, c.High)
	}

	return boundsFromRange(lo, hi, height)
}

func priceToRow(price, lo, hi float64, height int) int {
	if hi == lo {
		return height / 2
	}

	normalized := (hi - price) / (hi - lo)
	row := int(math.Round(normalized * float64(height-1)))

	if row < 0 {
		return 0
	}
	if row >= height {
		return height - 1
	}
	return row
}

const (
	candleBody = '█'

	// TradingView's default candle palette.
	colorGreen = "\x1b[38;2;38;166;154m"
	colorRed   = "\x1b[38;2;239;83;80m"
	colorAxis  = "\x1b[38;2;120;123;134m"
	colorReset = "\x1b[0m"

	candleWidth = 1
	candleGap   = 0
)

func RenderCandles(symbol string, candles []Candle, height, width int) string {
	if len(candles) == 0 {
		return "no candles\n"
	}

	stride := candleWidth + candleGap
	if maxCandles := width / stride; maxCandles > 0 && len(candles) > maxCandles {
		candles = candles[len(candles)-maxCandles:]
	}

	lo, hi := candleBounds(candles, height)

	grid := make([][]rune, height)
	colors := make([][]string, height)
	for y := range grid {
		grid[y] = make([]rune, len(candles)*stride)
		colors[y] = make([]string, len(candles)*stride)
		for x := range grid[y] {
			grid[y][x] = ' '
		}
	}

	for i, c := range candles {
		x0 := i * stride
		wickX := x0 + candleWidth/2

		high := priceToRow(c.High, lo, hi, height)
		low := priceToRow(c.Low, lo, hi, height)
		open := priceToRow(c.Open, lo, hi, height)
		close := priceToRow(c.Close, lo, hi, height)

		color := colorRed
		if c.Close >= c.Open {
			color = colorGreen
		}

		// wick: thin line through the middle of the candle's width
		for y := high; y <= low; y++ {
			grid[y][wickX] = '│'
			colors[y][wickX] = color
		}

		// body: full candle width, so it reads as a block, not a line
		top := min(open, close)
		bottom := max(open, close)

		for y := top; y <= bottom; y++ {
			for dx := 0; dx < candleWidth; dx++ {
				grid[y][x0+dx] = candleBody
				colors[y][x0+dx] = color
			}
		}
	}

	var b strings.Builder

	first := candles[0].Open
	last := candles[len(candles)-1]
	delta := last.Close - first
	pct := 0.0
	if first != 0 {
		pct = (delta / first) * 100
	}

	headerColor := colorRed
	if delta >= 0 {
		headerColor = colorGreen
	}

	fmt.Fprintf(&b, "%s %s%.2f %+0.2f (%+0.2f%%)%s\n", symbol, headerColor, last.Close, delta, pct, colorReset)

	for y := range grid {
		price := hi - (float64(y)/float64(height-1))*(hi-lo)

		fmt.Fprintf(&b, "%s%8.2f ┤%s", colorAxis, price, colorReset)
		for x, r := range grid[y] {
			if colors[y][x] == "" {
				b.WriteRune(r)
				continue
			}
			b.WriteString(colors[y][x])
			b.WriteRune(r)
			b.WriteString(colorReset)
		}
		b.WriteByte('\n')
	}

	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func niceStep(raw float64) float64 {
	if raw <= tickSize {
		return tickSize
	}

	exp := math.Pow(10, math.Floor(math.Log10(raw)))
	switch f := raw / exp; {
	case f <= 1:
		return exp
	case f <= 2:
		return 2 * exp
	case f <= 5:
		return 5 * exp
	default:
		return 10 * exp
	}
}
