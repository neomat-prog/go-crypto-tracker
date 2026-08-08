<img src="assets/gopher.png" alt="Gopher" width="320" align="right"/>

<h1>go-crypto-tracker</h1>

<p>Terminal app that streams live price data from Binance and renders it as an UTF-8 chart.</p>

<br clear="all"/>

## Screenshot

![Candlestick chart](assets/graph.png)

## What it does

Connects to the Binance websocket kline stream for a configured symbol and interval, backfills recent history over REST on startup, keeps a rolling window of closing prices in memory, and redraws an ASCII line chart in the terminal twice a second. Reconnects automatically on dropped connections.

## Requirements

- Go 1.25+

## Configuration

Includes a custom `.env` parser and validator (no third-party dependency) that loads and validates configuration on startup.

Set values in `.env`, or as environment variables — those take precedence:

```dotenv
SYMBOL="ETHUSDT"
INTERVAL="1m"
BACKFILL="100"
```

| Variable | Description | Accepted values |
| --- | --- | --- |
| `SYMBOL` | Trading pair | e.g. `ETHUSDT` |
| `INTERVAL` | Kline interval | `1m` `3m` `5m` `15m` `30m` `1h` `2h` `4h` `6h` `8h` `12h` `1d` `3d` `1w` `1M` |
| `BACKFILL` | Historical candles fetched on startup | `0`–`1000` |

## Usage

```sh
make tracker
```

Or directly:

```sh
go run ./cmd/tracker
```

Press <kbd>Ctrl</kbd>+<kbd>C</kbd> to stop.

## Testing

```sh
make test
```

Coverage report:

```sh
make cover
```

## Layout

```
cmd/tracker       entry point
cmd/config        .env loading and validation
internal          ASCII chart rendering
internal/binance  REST backfill and websocket client
internal/market   kline model and ring buffer store
internal/tracker  orchestrates the transport and serves snapshots
```
