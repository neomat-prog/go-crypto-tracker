package binance

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/neomat-prog/internal/market"
)

var wsURL = "wss://stream.binance.com:9443/ws"

const (
	readTimeout    = time.Minute
	reconnectDelay = 2 * time.Second
	subQueue       = 64
)

func (ws *WS) Start(klinech chan<- market.Kline) error {
	ws.seed(klinech)

	reconnect := time.NewTimer(0)
	defer reconnect.Stop()

	for {
		select {
		case <-ws.quitch:
			return ws.disconnect()

		case <-reconnect.C:
			if err := ws.dial(klinech); err != nil {
				log.Println("binance: dial:", err)
				reconnect.Reset(reconnectDelay)
			}

		case ce := <-ws.errch:
			if ce.conn != ws.conn {
				continue
			}
			log.Println("binance: read:", ce.err)
			ws.disconnect()
			reconnect.Reset(reconnectDelay)

		case name := <-ws.subch:
			ws.subs = append(ws.subs, name)
			if ws.conn == nil {
				continue
			}
			if err := ws.conn.WriteJSON(subscribeMsg(name)); err != nil {
				log.Println("binance: subscribe:", err)
			}
		}
	}
}

func (ws *WS) Subscribe(symbol string) error {
	select {
	case <-ws.quitch:
		return errClosed
	default:
	}

	select {
	case ws.subch <- streamName(symbol, ws.Interval):
		return nil
	case <-ws.quitch:
		return errClosed
	}
}

func (ws *WS) seed(klinech chan<- market.Kline) {
	if ws.Backfill <= 0 {
		return
	}

	for _, sym := range ws.Symbols {
		ks, err := backfill(sym, ws.Interval, ws.Backfill)
		if err != nil {
			log.Println("binance: backfill:", err)
			continue
		}

		for _, k := range ks {
			select {
			case klinech <- k:
			case <-ws.quitch:
				return
			}
		}
	}
}

func (ws *WS) dial(klinech chan<- market.Kline) error {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return err
	}

	conn.SetReadDeadline(time.Now().Add(readTimeout))
	conn.SetPingHandler(func(data string) error {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		return conn.WriteControl(websocket.PongMessage, []byte(data), time.Now().Add(5*time.Second))
	})

	if len(ws.subs) > 0 {
		if err := conn.WriteJSON(subscribeMsg(ws.subs...)); err != nil {
			conn.Close()
			return err
		}
	}

	ws.conn = conn
	go ws.readLoop(conn, klinech)
	return nil
}

func (ws *WS) disconnect() error {
	if ws.conn == nil {
		return nil
	}
	err := ws.conn.Close()
	ws.conn = nil
	return err
}

func (ws *WS) readLoop(conn *websocket.Conn, klinech chan<- market.Kline) {
	for {
		var ev klineEvent
		if err := conn.ReadJSON(&ev); err != nil {
			select {
			case ws.errch <- connErr{conn: conn, err: err}:
			case <-ws.quitch:
			}
			return
		}
		conn.SetReadDeadline(time.Now().Add(readTimeout))

		if ev.Event != "kline" {
			continue
		}
		k, err := ev.kline()
		if err != nil {
			continue
		}

		select {
		case klinech <- k:
		case <-ws.quitch:
			return
		}
	}
}
