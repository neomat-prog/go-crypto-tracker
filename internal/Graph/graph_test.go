package Graph

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var ansiEscape = regexp.MustCompile("\x1b\\[[0-9;]*m")

func tightRangeCandles() []Candle {
	// Mimics the live ETHUSDT screenshot: ~20 candles within a few cents.
	candles := make([]Candle, 20)
	base := 1923.18
	deltas := []float64{0, 1, 2, 1, 0, -1, 0, 2, 3, 2, 1, 0, -1, 0, 1, 2, 1, 0, -1, 0}
	for i, d := range deltas {
		open := base + float64(d)*0.01
		close := open + 0.01
		candles[i] = Candle{
			Open:  open,
			Close: close,
			High:  close + 0.01,
			Low:   open - 0.01,
		}
	}
	return candles
}

func TestCandleBoundsFloorsStepAtTickSize(t *testing.T) {
	candles := tightRangeCandles()
	lo, hi := candleBounds(candles, 15)

	if span := hi - lo; span < tickSize*15 {
		t.Fatalf("candleBounds span = %v, want at least %v", span, tickSize*15)
	}
}

func TestRenderCandlesNoDuplicateRowLabels(t *testing.T) {
	out := RenderCandles("ETHUSDT", tightRangeCandles(), 15, 80)

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	seen := make(map[string]bool)
	for _, line := range lines[1:] { // skip header line
		plain := ansiEscape.ReplaceAllString(line, "")
		idx := strings.Index(plain, "┤")
		if idx < 0 {
			continue
		}
		label := strings.TrimSpace(plain[:idx])
		if _, err := strconv.ParseFloat(label, 64); err != nil {
			continue
		}
		if seen[label] {
			t.Fatalf("duplicate row label %q found in output:\n%s", label, out)
		}
		seen[label] = true
	}
}
