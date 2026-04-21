# avanza-sdk-go

[![CI](https://github.com/vmorsell/avanza-sdk-go/actions/workflows/ci.yml/badge.svg)](https://github.com/vmorsell/avanza-sdk-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/vmorsell/avanza-sdk-go.svg)](https://pkg.go.dev/github.com/vmorsell/avanza-sdk-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/vmorsell/avanza-sdk-go)](https://goreportcard.com/report/github.com/vmorsell/avanza-sdk-go)
[![Latest Release](https://img.shields.io/github/v/release/vmorsell/avanza-sdk-go?sort=semver)](https://github.com/vmorsell/avanza-sdk-go/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vmorsell/avanza-sdk-go)](go.mod)
[![License](https://img.shields.io/github/license/vmorsell/avanza-sdk-go)](LICENSE)

Go SDK for [Avanza Bank](https://www.avanza.se), the Swedish online broker. Unofficial, reverse-engineered from the web client, so the API surface here tracks whatever Avanza's own frontend uses. No public API exists.

Covers BankID login, account and position data, order placement (stocks, funds, stop-loss), instrument search, and real-time streams (order-book depth, own-order updates, stop-loss events) over SSE.

> 0.x software. The Avanza endpoints are undocumented and can change without notice. Pin exact versions, and expect minor bumps to occasionally break things.

## Install

```bash
go get github.com/vmorsell/avanza-sdk-go
```

Requires Go 1.23 or newer.

## Quick start

Authentication is BankID only. The flow starts a session, renders a QR in your terminal, polls until you scan it, then exchanges the result for session cookies.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/vmorsell/avanza-sdk-go"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    c := avanza.New()

    startResp, err := c.Auth.StartBankID(ctx)
    if err != nil {
        log.Fatal(err)
    }

    if err := c.Auth.DisplayQRCode(startResp.QRToken); err != nil {
        log.Fatal(err)
    }

    collectResp, err := c.Auth.PollBankIDWithQRUpdates(ctx)
    if err != nil {
        log.Fatal(err)
    }

    if err := c.Auth.EstablishSession(ctx, collectResp); err != nil {
        log.Fatal(err)
    }

    overview, err := c.Accounts.GetOverview(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("logged in as %s, %d accounts\n", collectResp.Name, len(overview.Accounts))
}
```

Hold on to the `Avanza` struct after that. Every service (`Auth`, `Accounts`, `Trading`, `Market`) hangs off it and they share one HTTP client, cookie jar, and rate limiter. Safe to share across goroutines.

## Placing an order

```go
resp, err := c.Trading.PlaceOrder(ctx, &trading.PlaceOrderRequest{
    RequestID:   uuid.New().String(),
    AccountID:   accountID,
    OrderbookID: "5247", // Investor B on Stockholmsbörsen
    Side:        trading.OrderSideBuy,
    Condition:   trading.OrderConditionNormal,
    Price:       245.50,
    Volume:      10,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("order %s: %s\n", resp.OrderID, resp.OrderRequestStatus)
```

For anything real, run `Trading.ValidateOrder` and `Trading.GetPreliminaryFee` first. Validation flags commission thresholds, price ramping, large-in-scale, etc. The fee call gives you the commission in the order's currency before you commit.

## Streaming

Order-book depth, own-order updates, and stop-loss events come over Server-Sent Events. Subscriptions reconnect automatically on transient failures with exponential backoff from 3s up to 30s.

```go
sub, err := c.Market.SubscribeToOrderDepth(ctx, "738784")
if err != nil {
    log.Fatal(err)
}
defer sub.Close()

for {
    select {
    case event := <-sub.Events():
        if event.Event == "ORDER_DEPTH" {
            for i, lvl := range event.Data.Levels {
                fmt.Printf("%d: %.0f @ %.2f / %.0f @ %.2f\n",
                    i, lvl.BuyVolume, lvl.BuyPrice, lvl.SellVolume, lvl.SellPrice)
            }
        }
    case err := <-sub.Errors():
        log.Printf("stream error: %v", err)
        return
    case <-ctx.Done():
        return
    }
}
```

`sub.Close()` cancels the underlying context and waits for goroutines to drain before returning.

## Configuration

Functional options on `avanza.New`:

```go
c := avanza.New(
    avanza.WithBaseURL("http://localhost:8080"),
    avanza.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
    avanza.WithUserAgent("my-trading-bot/1.0"),
    avanza.WithRateLimiter(&client.SimpleRateLimiter{Interval: 200 * time.Millisecond}),
)
```

Defaults: `https://www.avanza.se`, stdlib `http.Client`, minimum 100ms between requests. The rate limiter is an interface, so swap it for something smarter if you need token bucket or adaptive behavior.

## Errors

Non-2xx responses come back as `*client.HTTPError`:

```go
_, err := c.Accounts.GetOverview(ctx)
if err != nil {
    var httpErr *client.HTTPError
    if errors.As(err, &httpErr) {
        log.Printf("avanza %d: %s", httpErr.StatusCode, httpErr.Body)
    }
}
```

Bodies are left raw because Avanza's error shapes aren't consistent enough to model generically.

## Examples

Runnable end-to-end examples live under [`examples/`](examples/), grouped by feature. Each is a `main.go` you can run directly once you have a test account.

## License

[MIT](LICENSE).

## Disclaimer

Not affiliated with Avanza Bank AB. The endpoints are undocumented and may change or break at any point. Trading involves real money; bugs here could cause unintended orders. Test against small positions before letting anything loose.
