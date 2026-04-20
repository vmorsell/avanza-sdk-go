# avanza-sdk-go

[![CI](https://github.com/vmorsell/avanza-sdk-go/actions/workflows/ci.yml/badge.svg)](https://github.com/vmorsell/avanza-sdk-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/vmorsell/avanza-sdk-go.svg)](https://pkg.go.dev/github.com/vmorsell/avanza-sdk-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/vmorsell/avanza-sdk-go)](https://goreportcard.com/report/github.com/vmorsell/avanza-sdk-go)
[![Latest Release](https://img.shields.io/github/v/release/vmorsell/avanza-sdk-go?include_prereleases&sort=semver)](https://github.com/vmorsell/avanza-sdk-go/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vmorsell/avanza-sdk-go)](go.mod)
[![License](https://img.shields.io/github/license/vmorsell/avanza-sdk-go)](LICENSE)

Unofficial Go SDK for [Avanza Bank](https://www.avanza.se), Sweden's largest online stockbroker. Covers BankID authentication, account and position data, order placement and management (including stop-loss), transaction history, and real-time streaming of order-book depth and own-order updates over Server-Sent Events. Reverse-engineered from the Avanza web client — there is no official public API.

## Status

**`v0.x` — API surface may change between minor versions.** The SDK is in active use but the Avanza endpoints themselves are undocumented and can change without notice. Pin exact versions in production.

## Features

- **BankID authentication** — QR-code flow with automatic QR refresh and session establishment.
- **Accounts** — overview, account lists, positions, cash, transactions, aggregated values.
- **Trading** — place, modify, delete orders for stocks and funds. Order validation and preliminary fee quotes before submission.
- **Stop-loss orders** — full lifecycle (place, list, modify, delete) with trigger and order-event configuration.
- **Market data** — instrument search, quotes, order depth, orderbook trading parameters.
- **Streaming (SSE)** — subscribe to real-time order-book depth, stop-loss order updates, and your own order updates. Automatic reconnect with exponential backoff.
- **Thread-safe** — a single client can be shared across goroutines. Cookies and security tokens are protected under RWMutex.
- **Rate limiting** — pluggable `RateLimiter` interface with a 100ms-interval default, TOCTOU-safe under concurrency.

## Install

```bash
go get github.com/vmorsell/avanza-sdk-go
```

Go 1.21+.

## Quick start

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

    // 1. Start BankID — returns a QR token.
    startResp, err := c.Auth.StartBankID(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // 2. Render the QR in the terminal.
    if err := c.Auth.DisplayQRCode(startResp.QRToken); err != nil {
        log.Fatal(err)
    }

    // 3. Poll until the user scans it, auto-refreshing the QR as it expires.
    collectResp, err := c.Auth.PollBankIDWithQRUpdates(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // 4. Exchange the auth result for session cookies.
    if err := c.Auth.EstablishSession(ctx, collectResp); err != nil {
        log.Fatal(err)
    }

    overview, err := c.Accounts.GetOverview(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Logged in as %s — %d accounts\n", collectResp.Name, len(overview.Accounts))
}
```

## Placing an order

```go
import (
    "github.com/google/uuid"
    "github.com/vmorsell/avanza-sdk-go/trading"
)

resp, err := c.Trading.PlaceOrder(ctx, &trading.PlaceOrderRequest{
    RequestID:   uuid.New().String(),
    AccountID:   accountID,
    OrderbookID: "5247", // Investor B
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

For safer flows, call `Trading.ValidateOrder` and `Trading.GetPreliminaryFee` before `PlaceOrder`.

## Streaming order-book depth

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

The subscription reconnects automatically with exponential backoff (3s → 30s) on transient failures. `sub.Close()` cancels the underlying context and drains goroutines cleanly.

## Configuration

```go
import (
    "net/http"
    "time"

    "github.com/vmorsell/avanza-sdk-go"
    "github.com/vmorsell/avanza-sdk-go/client"
)

c := avanza.New(
    avanza.WithBaseURL("http://localhost:8080"),
    avanza.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
    avanza.WithUserAgent("MyApp/1.0"),
    avanza.WithRateLimiter(&client.SimpleRateLimiter{Interval: 200 * time.Millisecond}),
)
```

All options are optional and composable. Defaults: `https://www.avanza.se`, standard `http.Client`, 100ms minimum interval between requests.

## Error handling

Non-2xx responses return `*client.HTTPError` with the status code and raw body:

```go
import (
    "errors"
    "github.com/vmorsell/avanza-sdk-go/client"
)

_, err := c.Accounts.GetOverview(ctx)
if err != nil {
    var httpErr *client.HTTPError
    if errors.As(err, &httpErr) {
        log.Printf("avanza %d: %s", httpErr.StatusCode, httpErr.Body)
    }
}
```

## Examples

Runnable end-to-end examples (authenticate, list accounts, place/validate orders, subscribe to streams) live under [`examples/`](examples/).

## Related projects

Unofficial Avanza SDKs also exist in other languages — search GitHub for "avanza" if you need Python, JavaScript, or other runtimes.

## Contributing

Issues and PRs welcome. Run `make ci` (tests with `-race`, `govulncheck`, `golangci-lint`) before submitting.

## License

[MIT](LICENSE).

---

<details>
<summary><strong>Disclaimer</strong></summary>

This is an unofficial, reverse-engineered SDK. It is **not affiliated with, endorsed by, or supported by Avanza Bank AB**. The underlying endpoints are undocumented and may change or break without notice. Trading involves financial risk; bugs in this library could cause unintended orders. You are solely responsible for any use of this SDK and any resulting trades. Test thoroughly against small positions before automating anything meaningful.

</details>
