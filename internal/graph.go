package internal

import (
	"fmt"

	"github.com/guptarohit/asciigraph"
)

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

	graph := asciigraph.Plot(
		closes,
		asciigraph.Height(height),
		asciigraph.Width(width),
	)

	return header + graph + "\n"
}
