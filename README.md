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
- **Order Placement** - Place buy and sell orders for stocks and funds
- **Stop Loss Orders** - Place stop loss orders with trigger conditions
- **Order Validation** - Validate orders before placing them
- **Preliminary Fees** - Get fee estimates before placing orders
- **Order Depth Subscription** - Real-time order book updates

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

See the [examples](examples/) directory for complete working examples.
