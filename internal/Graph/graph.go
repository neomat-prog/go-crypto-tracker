package Graph

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type Candle struct {
	Open  float64
	Close float64
	High  float64
	Low   float64
}

const (
	targetLabels   = 6
	minChartHeight = 3

	maxRowsPerTick = 2

	axisLabelWidth = 9

	AxisWidth = 3 + axisLabelWidth

	minSlots = 20

	clearLine = "\x1b[K"

	epsilon = 1e-9
)

func RenderCandles(symbol string, candles []Candle, height, width int, st Style) string {
	if len(candles) == 0 {
		return "no candles\n"
	}
	if width < 1 {
		width = 1
	}
	if height < minChartHeight {
		height = minChartHeight
	}

	g := st.glyphs()

	bodyW, stride := candleGeometry(width)
	slots := max(width/stride, 1)
	if len(candles) > slots {
		candles = candles[len(candles)-slots:]
	}

	tick := inferTick(candles)
	lo, hi, step := candleBounds(candles, height)

	if maxH := maxRowsPerTick * math.Ceil((hi-lo)/tick); maxH < float64(height) {
		height = max(int(maxH), minChartHeight)
	}

	per := st.subRowsPerRow()
	subRows := height * per
	gridW := slots * stride

	body := newBitGrid(subRows, gridW)
	wick := newBitGrid(subRows, gridW)
	colOf := make([]string, gridW)

	offset := slots - len(candles)

	for i, c := range candles {
		x0 := (offset + i) * stride
		wickX := x0 + bodyW/2

		color := st.dirColor(c.Close >= c.Open)
		for dx := 0; dx < bodyW && x0+dx < gridW; dx++ {
			colOf[x0+dx] = color
		}

		if wickX < gridW {
			top := priceToSub(c.High, lo, hi, subRows)
			bottom := priceToSub(c.Low, lo, hi, subRows)
			for s := top; s <= bottom; s++ {
				wick[s][wickX] = true
			}
		}

		top := priceToSub(math.Max(c.Open, c.Close), lo, hi, subRows)
		bottom := priceToSub(math.Min(c.Open, c.Close), lo, hi, subRows)
		for s := top; s <= bottom; s++ {
			for dx := 0; dx < bodyW && x0+dx < gridW; dx++ {
				body[s][x0+dx] = true
			}
		}
	}

	priceDec := decimalsFor(tick, hi)
	labels := labelRows(lo, hi, step, subRows, per, decimalsFor(step, hi))

	last := candles[len(candles)-1]
	lastColor := st.dirColor(last.Close >= last.Open)
	lastRow := priceToSub(last.Close, lo, hi, subRows) / per

	var b strings.Builder

	first := candles[0].Open
	delta := last.Close - first
	pct := 0.0
	if first != 0 {
		pct = (delta / first) * 100
	}

	fmt.Fprintf(&b, "%s %s%.*f %+.*f (%+0.2f%%)%s%s\n",
		symbol, st.dirColor(delta >= 0), priceDec, last.Close, priceDec, delta, pct, st.reset(), clearLine)

	for y := 0; y < height; y++ {
		active := ""
		setColor := func(c string) {
			if c == active {
				return
			}
			if c == "" {
				b.WriteString(st.reset())
			} else {
				b.WriteString(c)
			}
			active = c
		}

		label, labelled := labels[y]
		onLast := y == lastRow

		topSub := y * per
		botSub := topSub + per - 1

		for x := range gridW {
			r := glyphFor(g, body[topSub][x], body[botSub][x], wick[topSub][x], wick[botSub][x])

			switch {
			case r != 0:
				setColor(colOf[x])
			case onLast:
				r = g.lastPrice
				setColor(lastColor)
			case labelled && x%2 == 0:
				r = g.gridDot
				setColor(st.grid())
			default:
				r = ' '
				setColor("")
			}

			b.WriteRune(r)
		}

		setColor("")
		b.WriteString(st.axis())
		b.WriteByte(' ')
		b.WriteRune(g.axisTick)

		switch {
		case onLast:
			b.WriteString(st.reset())
			b.WriteByte(' ')
			if st.Color != ColorNone {
				b.WriteString(sgrReverse)
				b.WriteString(lastColor)
			}
			fmt.Fprintf(&b, "%*.*f", axisLabelWidth, priceDec, last.Close)
			b.WriteString(st.reset())
		case labelled:
			b.WriteByte(' ')
			b.WriteString(label)
			b.WriteString(st.reset())
		default:
			b.WriteString(st.reset())
		}

		b.WriteString(clearLine)
		b.WriteByte('\n')
	}

	return b.String()
}

func candleGeometry(width int) (bodyW, stride int) {
	switch {
	case width/4 >= minSlots:
		return 3, 4
	case width/2 >= minSlots:
		return 1, 2
	default:
		return 1, 1
	}
}

func glyphFor(g glyphSet, topBody, botBody, topWick, botWick bool) rune {
	switch {
	case topBody && botBody:
		return g.bodyFull
	case topBody:
		if botWick {
			return g.bodyFull
		}
		return g.bodyUpper
	case botBody:
		if topWick {
			return g.bodyFull
		}
		return g.bodyLower
	case topWick && botWick:
		return g.wickFull
	case topWick:
		return g.wickUpper
	case botWick:
		return g.wickLower
	}
	return 0
}

func priceToSub(price, lo, hi float64, subRows int) int {
	if subRows < 1 {
		return 0
	}
	if hi <= lo {
		return subRows / 2
	}

	s := int(math.Floor((hi - price) / (hi - lo) * float64(subRows)))
	if s < 0 {
		return 0
	}
	if s >= subRows {
		return subRows - 1
	}
	return s
}

func labelCount(height int) int {
	if height < 2*targetLabels {
		return max(height/2, 1)
	}
	return targetLabels
}

func candleBounds(candles []Candle, height int) (float64, float64, float64) {
	lo, hi := dataRange(candles)
	return boundsFromRange(lo, hi, labelCount(height), inferTick(candles))
}

func dataRange(candles []Candle) (float64, float64) {
	lo, hi := candles[0].Low, candles[0].High
	for _, c := range candles {
		lo = math.Min(lo, c.Low)
		hi = math.Max(hi, c.High)
	}
	return lo, hi
}

func boundsFromRange(lo, hi float64, labels int, tick float64) (float64, float64, float64) {
	if labels < 1 {
		labels = 1
	}

	span := hi - lo
	if span <= 0 {
		span = tick
	}

	step := niceStep(span/float64(labels), tick)

	lo = math.Floor(lo/step+epsilon) * step
	hi = math.Ceil(hi/step-epsilon) * step
	if hi-lo < step {
		hi = lo + step
	}

	return lo, hi, step
}

func labelRows(lo, hi, step float64, subRows, per, dec int) map[int]string {
	rows := make(map[int]string)

	for i := 0; ; i++ {
		price := lo + float64(i)*step
		if price > hi+step/2 {
			break
		}

		row := priceToSub(price, lo, hi, subRows) / per
		if _, taken := rows[row]; !taken {
			rows[row] = fmt.Sprintf("%*.*f", axisLabelWidth, dec, price)
		}
	}

	return rows
}

func inferTick(candles []Candle) float64 {
	prices := make([]float64, 0, len(candles)*4)
	maxPrice := 0.0

	for _, c := range candles {
		prices = append(prices, c.Open, c.High, c.Low, c.Close)
		maxPrice = math.Max(maxPrice, math.Abs(c.High))
	}
	if maxPrice <= 0 {
		maxPrice = 1
	}

	sort.Float64s(prices)

	gap := math.Inf(1)
	for i := 1; i < len(prices); i++ {
		if d := prices[i] - prices[i-1]; d > 0 && d < gap {
			gap = d
		}
	}
	if math.IsInf(gap, 1) {
		gap = maxPrice * 1e-5
	}

	tick := niceStep(gap, 0)
	return math.Min(math.Max(tick, maxPrice*1e-9), maxPrice*1e-2)
}

func decimalsFor(step, hi float64) int {
	dec := 0
	if step > 0 && step < 1 {
		dec = int(math.Ceil(-math.Log10(step) - epsilon))
	}
	if dec < 0 {
		dec = 0
	}
	if dec > 8 {
		dec = 8
	}

	for dec > 0 && len(fmt.Sprintf("%.*f", dec, hi)) > axisLabelWidth {
		dec--
	}
	return dec
}

func niceStep(raw, floor float64) float64 {
	if raw <= floor {
		return floor
	}
	if raw <= 0 {
		return floor
	}

	exp := math.Pow(10, math.Floor(math.Log10(raw)))
	switch f := raw / exp; {
	case f <= 1+epsilon:
		return exp
	case f <= 2+epsilon:
		return 2 * exp
	case f <= 5+epsilon:
		return 5 * exp
	default:
		return 10 * exp
	}
}

func newBitGrid(rows, cols int) [][]bool {
	grid := make([][]bool, rows)
	buf := make([]bool, rows*cols)
	for i := range grid {
		grid[i], buf = buf[:cols:cols], buf[cols:]
	}
	return grid
}
