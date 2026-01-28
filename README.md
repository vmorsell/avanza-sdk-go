# Avanza Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/vmorsell/avanza-sdk-go.svg)](https://pkg.go.dev/github.com/vmorsell/avanza-sdk-go)

Reverse-engineered Go client for Avanza.

## READ THIS FIRST

**Unofficial, reverse-engineered SDK. Use at your own risk. Not affiliated with Avanza.**

## Features

- **BankID Authentication** - QR code-based authentication with automatic session management
- **Account Overview** - Get categorized accounts, balances, and performance data
- **Trading Accounts** - List all trading accounts with detailed information
- **Account Positions** - Get detailed positions, stocks, and cash for specific accounts
- **Transactions** - Fetch transaction history with date filtering
- **Aggregated Values** - Get total account values on specific dates
- **Order Management** - Place, modify, and delete orders for stocks and funds
- **Stop Loss Orders** - Place stop loss orders with trigger conditions
- **Order Validation** - Validate orders before placing them
- **Preliminary Fees** - Get fee estimates before placing orders
- **Order Depth Subscription** - Real-time order book updates
- **Orders Subscription** - Real-time updates on your own orders

## Installation

```bash
go get github.com/vmorsell/avanza-sdk-go
```

## Quick Start

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

    client := avanza.New()

    // Start BankID authentication
    startResp, err := client.Auth.StartBankID(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // Display QR code in terminal
    if err := client.Auth.DisplayQRCode(startResp.QRToken); err != nil {
        log.Fatal(err)
    }

    // Poll for authentication completion
    collectResp, err := client.Auth.PollBankIDWithQRUpdates(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Welcome %s\n", collectResp.Name)

    // Establish session before making API calls
    if err := client.Auth.EstablishSession(ctx, collectResp); err != nil {
        log.Fatal(err)
    }

    // Get account overview
    overview, err := client.Accounts.GetOverview(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("You have %d accounts\n", len(overview.Accounts))
}
```

## Configuration

```go
client := avanza.New(
    avanza.WithBaseURL("http://localhost:8080"),
    avanza.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
    avanza.WithUserAgent("MyApp/1.0"),
    avanza.WithRateLimiter(&client.SimpleRateLimiter{Interval: 200 * time.Millisecond}),
)
```

All options are composable and optional. Defaults: `https://www.avanza.se`, standard `http.Client`, 100ms rate limit.

## Error Handling

API errors are returned as `*client.HTTPError` with the status code and response body:

```go
import "github.com/vmorsell/avanza-sdk-go/client"

overview, err := c.Accounts.GetOverview(ctx)
if err != nil {
    var httpErr *client.HTTPError
    if errors.As(err, &httpErr) {
        fmt.Printf("status=%d body=%s\n", httpErr.StatusCode, httpErr.Body)
    }
}
```

## Examples

See the [examples](examples/) directory for complete working examples.
