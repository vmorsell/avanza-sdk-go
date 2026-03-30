package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/vmorsell/avanza-sdk-go/client"
	"github.com/vmorsell/avanza-sdk-go/examples/internal/auth"
	"github.com/vmorsell/avanza-sdk-go/trading"
)

func main() {
	az := auth.Authenticate()
	ctx := context.Background()
	account := auth.FirstTradingAccount(ctx, az)
	fmt.Printf("Using account: %s (%.2f SEK available)\n", account.Name, account.AvailableForPurchase)

	// Place a buy order with price far out of range so it won't fill.
	req := &trading.PlaceOrderRequest{
		RequestID:   uuid.New().String(),
		Price:       2.0,
		Volume:      1,
		AccountID:   account.AccountID,
		Side:        trading.OrderSideBuy,
		OrderbookID: "5247", // Investor B
		Condition:   trading.OrderConditionNormal,
		Metadata: trading.OrderMetadata{
			OrderEntryMode:  "ADVANCED",
			HasTouchedPrice: "true",
		},
	}

	resp, err := az.Trading.PlaceOrder(ctx, req)
	if err != nil {
		// Demonstrate HTTPError handling — useful for inspecting API error bodies.
		var httpErr *client.HTTPError
		if errors.As(err, &httpErr) {
			log.Fatalf("HTTP %d: %s", httpErr.StatusCode, httpErr.Body)
		}
		log.Fatalf("Failed to place order: %v", err)
	}

	fmt.Printf("Order placed: ID=%s, status=%s\n", resp.OrderID, resp.OrderRequestStatus)
	if resp.Message != "" {
		fmt.Printf("  Message: %s\n", resp.Message)
	}
}
