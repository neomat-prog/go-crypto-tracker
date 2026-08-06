package binance

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/neomat-prog/internal/market"
)

const testTimeout = 2 * time.Second

type fakeStream struct {
	subs chan map[string]any
	send chan []byte
}

func stubWS(t *testing.T) *fakeStream {
	t.Helper()

	fs := &fakeStream{
		subs: make(chan map[string]any, 8),
		send: make(chan []byte),
	}

	up := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		go func() {
			for {
				var m map[string]any
				if err := conn.ReadJSON(&m); err != nil {
					return
				}
				fs.subs <- m
			}
		}()

		for b := range fs.send {
			if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
				return
			}
		}
	}))

	old := wsURL
	wsURL = "ws" + strings.TrimPrefix(srv.URL, "http")
	t.Cleanup(func() {
		wsURL = old
		close(fs.send)
		srv.Close()
	})
	return fs
}

func (fs *fakeStream) nextSub(t *testing.T) map[string]any {
	t.Helper()
	select {
	case m := <-fs.subs:
		return m
	case <-time.After(testTimeout):
		t.Fatal("timed out waiting for SUBSCRIBE message")
		return nil
	}
}

func (fs *fakeStream) push(t *testing.T, raw string) {
	t.Helper()
	select {
	case fs.send <- []byte(raw):
	case <-time.After(testTimeout):
		t.Fatal("timed out pushing frame to client")
	}
}

func recvKline(t *testing.T, ch <-chan market.Kline) market.Kline {
	t.Helper()
	select {
	case k := <-ch:
		return k
	case <-time.After(testTimeout):
		t.Fatal("timed out waiting for kline")
		return market.Kline{}
	}
}

func params(t *testing.T, m map[string]any) []string {
	t.Helper()
	if m["method"] != "SUBSCRIBE" {
		t.Fatalf("method = %v; want SUBSCRIBE", m["method"])
	}
	raw, ok := m["params"].([]any)
	if !ok {
		t.Fatalf("params = %#v; want []any", m["params"])
	}
	out := make([]string, len(raw))
	for i, v := range raw {
		out[i] = v.(string)
	}
	return out
}

const btcKlineFrame = `{"e":"kline","E":1700000000000,"s":"BTCUSDT","k":{
	"t":1700000000000,"T":1700000059999,
	"o":"100.0","h":"101.0","l":"99.0","c":"100.5","x":false}}`

func TestWSSubscribesAndStreamsKlines(t *testing.T) {
	fs := stubWS(t)

	ws := NewWS(WSOpts{Symbol: "BTCUSDT"})
	klinech := make(chan market.Kline, 4)

	done := make(chan error, 1)
	go func() { done <- ws.Start(klinech) }()

	if got := params(t, fs.nextSub(t)); len(got) != 1 || got[0] != "btcusdt@kline_1m" {
		t.Fatalf("initial params = %v; want [btcusdt@kline_1m]", got)
	}

	if err := ws.Subscribe("ETHUSDT"); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if got := params(t, fs.nextSub(t)); len(got) != 1 || got[0] != "ethusdt@kline_1m" {
		t.Fatalf("second params = %v; want [ethusdt@kline_1m]", got)
	}

	fs.push(t, btcKlineFrame)

	k := recvKline(t, klinech)
	if k.Symbol != "BTCUSDT" || k.Close != 100.5 || k.Closed {
		t.Fatalf("kline = %+v; want BTCUSDT close 100.5 open candle", k)
	}

	ws.Close()
	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatal("Start did not return after Close")
	}
}

func TestWSIgnoresNonKlineFrames(t *testing.T) {
	fs := stubWS(t)

	ws := NewWS(WSOpts{Symbol: "BTCUSDT"})
	defer ws.Close()

	klinech := make(chan market.Kline, 4)
	go ws.Start(klinech)
	fs.nextSub(t)

	fs.push(t, `{"result":null,"id":1}`)
	fs.push(t, `{"e":"trade","s":"BTCUSDT","p":"1"}`)
	fs.push(t, btcKlineFrame)

	if k := recvKline(t, klinech); k.Symbol != "BTCUSDT" {
		t.Fatalf("kline = %+v; want the BTCUSDT kline", k)
	}
}

func TestWSSeedsBackfillBeforeStreaming(t *testing.T) {
	stubREST(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[[1699999940000,"98.0","99.0","97.0","98.5","1.0",1699999999999]]`))
	})
	fs := stubWS(t)

	ws := NewWS(WSOpts{Symbol: "BTCUSDT", Backfill: 1})
	defer ws.Close()

	klinech := make(chan market.Kline, 4)
	go ws.Start(klinech)

	if k := recvKline(t, klinech); k.Close != 98.5 || !k.Closed {
		t.Fatalf("seed kline = %+v; want closed candle with close 98.5", k)
	}

	fs.nextSub(t)
	fs.push(t, btcKlineFrame)
	if k := recvKline(t, klinech); k.Close != 100.5 {
		t.Fatalf("live kline = %+v; want close 100.5", k)
	}
}

func TestSubscribeAfterClose(t *testing.T) {
	ws := NewWS(WSOpts{})
	ws.Close()

	if err := ws.Subscribe("BTCUSDT"); err != errClosed {
		t.Fatalf("Subscribe after Close = %v; want %v", err, errClosed)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	ws := NewWS(WSOpts{})
	ws.Close()
	ws.Close()
}
