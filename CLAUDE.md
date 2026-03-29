# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test

```bash
go test -race ./...                       # all tests with race detector
go test -race -run TestPlaceOrder ./...   # single test
go vet ./...                              # static analysis
make lint                                 # golangci-lint (install: make install-lint)
make ci                                   # test-race + govulncheck + lint
```

## Architecture

Unofficial, reverse-engineered Go SDK for Avanza's trading platform. The API surface is discovered by inspecting browser traffic — there are no official docs.

**Entry point**: `avanza.go` defines the `Avanza` facade. `avanza.New(opts...)` collects functional options, builds a single `*client.Client`, and injects it into four service structs: `Auth`, `Accounts`, `Trading`, `Market`.

**Shared HTTP client** (`client/`): All services share one `*client.Client` which manages cookies, the `AZACSRF` security token, request headers, and rate limiting. Cookies and security token are protected by `sync.RWMutex`. The rate limiter uses projected-future `lastCall` under a `sync.Mutex` to avoid TOCTOU races between concurrent callers.

**Auth flow** (`auth/`): BankID QR-based auth. The flow is `StartBankID` → `DisplayQRCode` → `PollBankIDWithQRUpdates` (polls + refreshes QR) → `EstablishSession` (hits three endpoints to collect all session cookies). Session must be established before any other API call.

**SSE subscriptions** (`trading/subscription.go`, `market/subscription.go`): Order updates and order depth use Server-Sent Events with automatic reconnection via exponential backoff (base 3s, max 30s). Both follow the same pattern: `start()` loop → `connectAndStream()` → `processSSEStream()`. Channel sends use `select` with `ctx.Done()` to prevent deadlocks on full buffers. `Close()` cancels context, waits on `sync.WaitGroup`, then closes channels.

## Key Conventions

- **Error handling**: Non-OK HTTP → return `*client.HTTPError` directly. No redundant wrapping with function names. Only multi-step functions (e.g. `EstablishSession`) add step-level context.
- **Validation**: Service methods validate required inputs inline at function start — no separate validation layer.
- **Tests**: `httptest.NewServer` for HTTP mocking. `atomic` counters for handler state shared across goroutines. No timing-sensitive assertions below 50ms.
- **Functional options**: Two layers — `avanza.Option` wraps `client.Option` so consumers only import the root package.
