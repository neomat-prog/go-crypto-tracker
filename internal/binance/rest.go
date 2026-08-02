package binance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/neomat-prog/internal/market"
)

var restURL = "https://api.binance.com/api/v3/klines"

var restClient = &http.Client{Timeout: 10 * time.Second}

func backfill(symbol, interval string, limit int) ([]market.Kline, error) {
	q := url.Values{
		"symbol":   {strings.ToUpper(symbol)},
		"interval": {interval},
		"limit":    {strconv.Itoa(limit)},
	}

	resp, err := restClient.Get(restURL + "?" + q.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance: klines: %s", resp.Status)
	}

	var rows [][]any
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}

	out := make([]market.Kline, 0, len(rows))
	for _, r := range rows {
		if len(r) < 5 {
			continue
		}

		ms, ok := r[0].(float64)
		if !ok {
			continue
		}

		s := make([]string, 4)
		for i := range s {
			str, ok := r[i+1].(string)
			if !ok {
				continue
			}
			s[i] = str
		}

		v, err := parseFloats(s...)
		if err != nil {
			continue
		}

		out = append(out, market.Kline{
			Symbol:   strings.ToUpper(symbol),
			OpenTime: time.UnixMilli(int64(ms)).UTC(),
			Open:     v[0], High: v[1], Low: v[2], Close: v[3],
			Closed: true,
		})
	}

	return out, nil
}
