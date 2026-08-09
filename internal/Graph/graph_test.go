package Graph

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var ansiEscape = regexp.MustCompile("\x1b\\[[0-9;]*m")

const (
	bodyGlyphs = "█▀▄"
	wickGlyphs = "│╵╷"
)

func tightRangeCandles() []Candle {
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

func TestCandleBoundsIndependentOfHeight(t *testing.T) {
	candles := tightRangeCandles()

	shortLo, shortHi, step := candleBounds(candles, 15)
	tallLo, tallHi, _ := candleBounds(candles, 54)

	if shortLo != tallLo || shortHi != tallHi {
		t.Fatalf("bounds changed with height: [%v..%v] at 15 rows, [%v..%v] at 54",
			shortLo, shortHi, tallLo, tallHi)
	}

	dataLo, dataHi := dataRange(candles)

	if shortLo > dataLo || shortHi < dataHi {
		t.Fatalf("bounds [%v..%v] do not cover data [%v..%v]", shortLo, shortHi, dataLo, dataHi)
	}
	if span := shortHi - shortLo; span > (dataHi-dataLo)+2*step {
		t.Fatalf("bounds span %v is more than the data span %v plus one step either side",
			span, dataHi-dataLo)
	}
}

func chartBodies(t *testing.T, out string) []string {
	t.Helper()

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")[1:]

	bodies := make([]string, 0, len(lines))
	for _, l := range lines {
		plain := strings.TrimSuffix(ansiEscape.ReplaceAllString(l, ""), clearLine)

		body, _, ok := strings.Cut(plain, string(unicodeGlyphs.axisTick))
		if !ok {
			t.Fatalf("chart row has no price scale: %q", plain)
		}
		bodies = append(bodies, body)
	}
	return bodies
}

func chartRows(t *testing.T, out string) (rows, drawn int) {
	t.Helper()

	bodies := chartBodies(t, out)
	for _, body := range bodies {
		if strings.ContainsAny(body, bodyGlyphs+wickGlyphs) {
			drawn++
		}
	}
	return len(bodies), drawn
}

func TestQuietMarketShrinksChart(t *testing.T) {
	const height = 54

	rows, drawn := chartRows(t, RenderCandles("ETHUSDT", tightRangeCandles(), height, 80, Style{}))

	if rows >= height {
		t.Errorf("chart used %d of %d rows for a 4 cent range, want fewer", rows, height)
	}
	if drawn < rows*3/4 {
		t.Errorf("only %d of %d chart rows hold a candle, want at least three quarters", drawn, rows)
	}
}

func TestActiveMarketUsesFullHeight(t *testing.T) {
	const height = 54

	candles := make([]Candle, 40)
	price := 1900.0
	for i := range candles {
		open := price + float64(i%7)*0.25
		close := open + 0.25
		candles[i] = Candle{Open: open, Close: close, High: close + 0.1, Low: open - 0.1}
	}

	if rows, _ := chartRows(t, RenderCandles("ETHUSDT", candles, height, 80, Style{})); rows != height {
		t.Errorf("chart used %d of %d rows for a wide range, want all of them", rows, height)
	}
}

func rowLabels(t *testing.T, out string) []string {
	t.Helper()

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")[1:]

	var labels []string
	for _, line := range lines {
		plain := strings.TrimSuffix(ansiEscape.ReplaceAllString(line, ""), clearLine)

		_, after, ok := strings.Cut(plain, string(unicodeGlyphs.axisTick))
		if !ok {
			continue
		}

		label := strings.TrimSpace(after)
		if _, err := strconv.ParseFloat(label, 64); err != nil {
			continue
		}
		labels = append(labels, label)
	}
	return labels
}

func TestRenderCandlesNoDuplicateRowLabels(t *testing.T) {
	out := RenderCandles("ETHUSDT", tightRangeCandles(), 15, 80, Style{})

	seen := make(map[string]bool)
	for _, label := range rowLabels(t, out) {
		if seen[label] {
			t.Fatalf("duplicate row label %q found in output:\n%s", label, out)
		}
		seen[label] = true
	}
}

func TestWicksAreVisible(t *testing.T) {
	candles := make([]Candle, 20)
	for i := range candles {
		candles[i] = Candle{Open: 100.00, Close: 100.10, High: 101.00, Low: 99.00}
	}

	out := RenderCandles("TESTUSDT", candles, 20, 120, Style{})

	wickOnly := 0
	for _, body := range chartBodies(t, out) {
		if strings.ContainsAny(body, wickGlyphs) && !strings.ContainsAny(body, bodyGlyphs) {
			wickOnly++
		}
	}

	if wickOnly == 0 {
		t.Fatalf("no row holds a wick without a body; wicks are being overdrawn:\n%s", out)
	}
}

func TestWickIsThinnerThanBody(t *testing.T) {
	if bodyW, stride := candleGeometry(160); bodyW != 3 || stride != 4 {
		t.Fatalf("candleGeometry(160) = (%d, %d), want (3, 4)", bodyW, stride)
	}

	candles := []Candle{{Open: 100.00, Close: 100.10, High: 101.00, Low: 99.00}}
	out := RenderCandles("TESTUSDT", candles, 20, 120, Style{})

	var widestWick, widestBody int
	for _, body := range chartBodies(t, out) {
		wick, bodyCells := 0, 0
		for _, r := range body {
			switch {
			case strings.ContainsRune(wickGlyphs, r):
				wick++
			case strings.ContainsRune(bodyGlyphs, r):
				bodyCells++
			}
		}
		widestWick = max(widestWick, wick)
		widestBody = max(widestBody, bodyCells)
	}

	if widestWick != 1 {
		t.Errorf("wick is %d columns wide, want 1", widestWick)
	}
	if widestBody != 3 {
		t.Errorf("body is %d columns wide, want 3", widestBody)
	}
}

func TestNewestCandleIsFlushRight(t *testing.T) {
	candles := make([]Candle, 5)
	for i := range candles {
		open := 100.0 + float64(i)*0.5
		candles[i] = Candle{Open: open, Close: open + 0.5, High: open + 0.6, Low: open - 0.1}
	}

	out := RenderCandles("TESTUSDT", candles, 20, 160, Style{})

	bodies := chartBodies(t, out)
	width := len([]rune(bodies[0]))

	leftmost, rightmost := -1, -1
	for _, body := range bodies {
		for i, r := range []rune(body) {
			if !strings.ContainsRune(bodyGlyphs+wickGlyphs, r) {
				continue
			}
			if leftmost == -1 || i < leftmost {
				leftmost = i
			}
			if i > rightmost {
				rightmost = i
			}
		}
	}

	const span = 5 * 4
	if rightmost < width-span || leftmost < width-span-1 {
		t.Errorf("candles occupy columns %d..%d of %d; they are not flush right",
			leftmost, rightmost, width)
	}
}

func TestLowPricedSymbol(t *testing.T) {
	candles := make([]Candle, 20)
	for i := range candles {
		open := 0.08340 + float64(i%5)*0.00001
		candles[i] = Candle{Open: open, Close: open + 0.00001, High: open + 0.00002, Low: open - 0.00001}
	}

	out := RenderCandles("DOGEUSDT", candles, 20, 120, Style{})

	labels := rowLabels(t, out)
	seen := make(map[string]bool)
	for _, label := range labels {
		if seen[label] {
			t.Fatalf("duplicate label %q on a sub-cent symbol:\n%s", label, out)
		}
		seen[label] = true

		if _, frac, ok := strings.Cut(label, "."); !ok || len(frac) < 4 {
			t.Errorf("label %q does not resolve a sub-cent price", label)
		}
	}

	if len(labels) < 3 {
		t.Fatalf("only %d price labels on a sub-cent symbol, want at least 3:\n%s", len(labels), out)
	}

	if rows, drawn := chartRows(t, out); drawn < 2 {
		t.Errorf("only %d of %d rows hold a candle; the range collapsed", drawn, rows)
	}
}

func TestStyleDegradation(t *testing.T) {
	candles := tightRangeCandles()

	t.Run("mono emits no SGR", func(t *testing.T) {
		out := RenderCandles("ETHUSDT", candles, 20, 120, Style{Color: ColorNone})
		if m := ansiEscape.FindString(out); m != "" {
			t.Errorf("mono output contains SGR %q", m)
		}
	})

	t.Run("ascii emits no multibyte runes", func(t *testing.T) {
		out := RenderCandles("ETHUSDT", candles, 20, 120, Style{ASCII: true})
		for i := 0; i < len(out); i++ {
			if out[i] >= 0x80 {
				t.Fatalf("ascii output contains non-ASCII byte %#x at %d", out[i], i)
			}
		}
	})

	t.Run("color modes render identical text", func(t *testing.T) {
		var want string
		for _, mode := range []ColorMode{ColorNone, Color16, Color256, ColorTrue} {
			out := RenderCandles("ETHUSDT", candles, 20, 120, Style{Color: mode})
			plain := ansiEscape.ReplaceAllString(out, "")

			if mode == ColorNone {
				want = plain
				continue
			}
			if plain != want {
				t.Fatalf("color mode %d changed the rendered text:\n%q\nwant:\n%q", mode, plain, want)
			}
		}
	})

	t.Run("rows fit the axis budget", func(t *testing.T) {
		const gridW = 120

		out := RenderCandles("ETHUSDT", candles, 20, gridW, Style{Color: ColorTrue})
		for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n")[1:] {
			plain := strings.TrimSuffix(ansiEscape.ReplaceAllString(line, ""), clearLine)
			if got := len([]rune(plain)); got > gridW+AxisWidth {
				t.Fatalf("row is %d columns, want at most %d:\n%q", got, gridW+AxisWidth, plain)
			}
		}
	})
}

func TestInferTick(t *testing.T) {
	cases := []struct {
		name    string
		candles []Candle
		want    float64
	}{
		{"cents", tightRangeCandles(), 0.01},
		{"sub cent", []Candle{
			{Open: 0.08340, Close: 0.08341, High: 0.08342, Low: 0.08339},
			{Open: 0.08341, Close: 0.08342, High: 0.08343, Low: 0.08340},
		}, 0.00001},
		{"flat feed falls back", []Candle{
			{Open: 100, Close: 100, High: 100, Low: 100},
		}, 0.001},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := inferTick(tc.candles); math.Abs(got-tc.want) > tc.want*1e-6 {
				t.Errorf("inferTick = %v, want %v", got, tc.want)
			}
		})
	}
}
