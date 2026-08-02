package internal

import (
	"fmt"
	"math"

	"github.com/guptarohit/asciigraph"
)

const tickSize = 0.01

func Render(symbol string, closes []float64, height, width int) string {
	if len(closes) < 2 {
		return fmt.Sprintf("need atleast 2 closing prices, got %d\n", len(closes))
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
	min, max := vs[0], vs[0]
	for _, v := range vs {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	span := max - min
	if span <= 0 {
		span = tickSize * float64(height)
	}

	rows := float64(height)
	step := niceStep(span / rows)
	lo := math.Floor(min/step) * step

	for lo+step*rows < max {
		step = niceStep(step * 1.5)
		lo = math.Floor(min/step) * step
	}

	return lo, lo + step*rows
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
