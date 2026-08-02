package binance

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseFloats(t *testing.T) {
	tests := []struct {
		name    string
		in      []string
		want    []float64
		wantErr bool
	}{
		{name: "ok", in: []string{"1.5", "2", "0.25"}, want: []float64{1.5, 2, 0.25}},
		{name: "empty", in: nil, want: []float64{}},
		{name: "not a number", in: []string{"1.0", "abc"}, wantErr: true},
		{name: "empty string", in: []string{""}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseFloats(tc.in...)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseFloats(%v) = %v, nil; want error", tc.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseFloats(%v): %v", tc.in, err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("parseFloats(%v) = %v; want %v", tc.in, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("parseFloats(%v) = %v; want %v", tc.in, got, tc.want)
				}
			}
		})
	}
}

func TestStreamName(t *testing.T) {
	if got := streamName("BTCUSDT", "1m"); got != "btcusdt@kline_1m" {
		t.Fatalf("streamName = %q; want btcusdt@kline_1m", got)
	}
}

func TestSubscribeMsg(t *testing.T) {
	b, err := json.Marshal(subscribeMsg("btcusdt@kline_1m", "ethusdt@kline_1m"))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"id":1,"method":"SUBSCRIBE","params":["btcusdt@kline_1m","ethusdt@kline_1m"]}`
	if string(b) != want {
		t.Fatalf("subscribeMsg = %s; want %s", b, want)
	}
}

func TestKlineEventToKline(t *testing.T) {
	raw := `{"e":"kline","E":1700000000000,"s":"BTCUSDT","k":{
		"t":1700000000000,"T":1700000059999,
		"o":"100.5","h":"101.0","l":"99.5","c":"100.0","x":true}}`

	var ev klineEvent
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatal(err)
	}

	k, err := ev.kline()
	if err != nil {
		t.Fatalf("kline(): %v", err)
	}

	if k.Symbol != "BTCUSDT" {
		t.Errorf("Symbol = %q; want BTCUSDT", k.Symbol)
	}
	if want := time.UnixMilli(1700000000000).UTC(); !k.OpenTime.Equal(want) {
		t.Errorf("OpenTime = %v; want %v", k.OpenTime, want)
	}
	if k.Open != 100.5 || k.High != 101.0 || k.Low != 99.5 || k.Close != 100.0 {
		t.Errorf("OHLC = %v/%v/%v/%v; want 100.5/101/99.5/100", k.Open, k.High, k.Low, k.Close)
	}
	if !k.Closed {
		t.Error("Closed = false; want true")
	}
}

func TestKlineEventBadPrice(t *testing.T) {
	var ev klineEvent
	ev.K.Open = "not-a-price"
	if _, err := ev.kline(); err == nil {
		t.Fatal("kline() = nil error; want parse error")
	}
}
