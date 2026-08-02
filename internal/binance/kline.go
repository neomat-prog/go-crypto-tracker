package binance

import (
	"strconv"
	"strings"
	"time"

	"github.com/neomat-prog/internal/market"
)

type klineEvent struct {
	Event     string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	K         struct {
		OpenTime    int64  `json:"t"`
		CloseTime   int64  `json:"T"`
		Open        string `json:"o"`
		High        string `json:"h"`
		Low         string `json:"l"`
		LastTradeID int64  `json:"L"`
		Close       string `json:"c"`
		Closed      bool   `json:"x"`
	} `json:"k"`
}

func (e klineEvent) kline() (market.Kline, error) {
	v, err := parseFloats(e.K.Open, e.K.High, e.K.Low, e.K.Close)
	if err != nil {
		return market.Kline{}, err
	}
	return market.Kline{
		Symbol:   e.Symbol,
		OpenTime: time.UnixMilli(e.K.OpenTime).UTC(),
		Open:     v[0], High: v[1], Low: v[2], Close: v[3],
		Closed: e.K.Closed,
	}, nil
}

func parseFloats(ss ...string) ([]float64, error) {
	out := make([]float64, len(ss))
	for i, s := range ss {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
		out[i] = f
	}
	return out, nil
}

func streamName(symbol, interval string) string {
	return strings.ToLower(symbol) + "@kline_" + interval
}

func subscribeMsg(streams ...string) map[string]any {
	return map[string]any{"method": "SUBSCRIBE", "params": streams, "id": 1}
}
