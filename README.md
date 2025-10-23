# Avanza Go SDK

Reverse-engineered Go client for Avanza.

## READ THIS FIRST

**TL;DR: This is for learning, not for your trading bot that's going to make you rich. It probably violates their terms and shouldn't be used by anyone.**

- **Not official** - Avanza doesn't know it exists (?)
- **ToS violation** - You're probably breaking this terms
- **Educational only** - Experiment away, don't build your retirement fund on this
- **Rate limits** - Don't be that person who DDoSes their servers
- **No warranty** - Use at your own risk, no support provided
- **Reverse engineered** - This was built by sniffing network traffic and will likely break when Avanza changes their API
- **Incomplete** - Only covers a tiny fraction of Avanza's actual API surface

## Features

- **BankID Authentication** - QR code-based authentication with automatic session management
- **Account Overview** - Get categorized accounts, balances, and performance data
- **Trading Accounts** - List all trading accounts with detailed information
- **Account Positions** - Get detailed positions, stocks, and cash for specific accounts

## Installation

```bash
go get github.com/vmorsell/avanza
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/vmorsell/avanza"
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

    // Get account overview
    overview, err := client.Accounts.GetAccountOverview(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("You have %d accounts\n", len(overview.Accounts))
}
```

## Configuration

```go
// Default configuration
client := avanza.New()

// Custom HTTP timeout
httpClient := &http.Client{Timeout: 60 * time.Second}
client := avanza.New(avanza.WithHTTPClient(httpClient))

// Custom base URL (useful for testing)
client := avanza.New(avanza.WithBaseURL("https://test.example.com"))
```

## Legal Notice

This is an **unofficial** client library created for educational purposes. It is not affiliated with, endorsed by, or supported by Avanza Bank AB. The use of this library may violate Avanza's Terms of Service. Users are responsible for ensuring their use complies with all applicable terms and conditions.
