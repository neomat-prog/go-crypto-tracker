package binance

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func stubREST(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	old := restURL
	restURL = srv.URL
	t.Cleanup(func() {
		restURL = old
		srv.Close()
	})
	return srv
}

func TestBackfill(t *testing.T) {
	var gotQuery string
	stubREST(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Write([]byte(`[
			[1700000000000,"100.0","101.0","99.0","100.5","1.0",1700000059999],
			[1700000060000,"100.5","102.0","100.0","101.5","2.0",1700000119999]
		]`))
	})

	ks, err := backfill("btcusdt", "1m", 2)
	if err != nil {
		t.Fatalf("backfill: %v", err)
	}

	if want := "interval=1m&limit=2&symbol=BTCUSDT"; gotQuery != want {
		t.Errorf("query = %q; want %q", gotQuery, want)
	}
	if len(ks) != 2 {
		t.Fatalf("len = %d; want 2", len(ks))
	}

	k := ks[0]
	if k.Symbol != "BTCUSDT" {
		t.Errorf("Symbol = %q; want BTCUSDT", k.Symbol)
	}
	if want := time.UnixMilli(1700000000000).UTC(); !k.OpenTime.Equal(want) {
		t.Errorf("OpenTime = %v; want %v", k.OpenTime, want)
	}
	if k.Open != 100.0 || k.High != 101.0 || k.Low != 99.0 || k.Close != 100.5 {
		t.Errorf("OHLC = %v/%v/%v/%v; want 100/101/99/100.5", k.Open, k.High, k.Low, k.Close)
	}
	if !k.Closed {
		t.Error("Closed = false; want true for backfilled candles")
	}
}

func TestBackfillSkipsBadRows(t *testing.T) {
	stubREST(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[
			[1700000000000,"1.0","2.0"],
			["notatimestamp","1.0","2.0","0.5","1.5"],
			[1700000120000,"1.0","2.0","0.5","nope"],
			[1700000180000,"1.0","2.0","0.5","1.5"]
		]`))
	})

	ks, err := backfill("BTCUSDT", "1m", 4)
	if err != nil {
		t.Fatalf("backfill: %v", err)
	}
	if len(ks) != 1 {
		t.Fatalf("len = %d; want 1 (three malformed rows dropped): %+v", len(ks), ks)
	}
	if ks[0].Close != 1.5 {
		t.Errorf("Close = %v; want 1.5", ks[0].Close)
	}
}

func TestBackfillErrors(t *testing.T) {
	tests := []struct {
		name string
		h    http.HandlerFunc
	}{
		{
			name: "non-200",
			h: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, `{"code":-1121,"msg":"Invalid symbol."}`, http.StatusBadRequest)
			},
		},
		{
			name: "bad json",
			h: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"not":"an array"}`))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stubREST(t, tc.h)
			if ks, err := backfill("BTCUSDT", "1m", 1); err == nil {
				t.Fatalf("backfill = %v, nil; want error", ks)
			}
		})
	}
}
