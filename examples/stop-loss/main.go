package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vmorsell/avanza-sdk-go/examples/internal/auth"
	"github.com/vmorsell/avanza-sdk-go/trading"
)

func main() {
	client := auth.Authenticate()
	ctx := context.Background()
	account := auth.FirstTradingAccount(ctx, client)

	validUntil := time.Now().AddDate(0, 0, 30).Format("2006-01-02")

	req := &trading.PlaceStopLossRequest{
		ParentStopLossID: "0",
		AccountID:        account.AccountID,
		OrderbookID:      "5247", // Investor B
		StopLossTrigger: trading.StopLossTrigger{
			Type:       trading.StopLossTriggerLessOrEqual,
			Value:      200.0,
			ValueType:  trading.StopLossValueMonetary,
			ValidUntil: validUntil,
		},
		StopLossOrderEvent: trading.StopLossOrderEvent{
			Type:      trading.StopLossOrderEventBuy,
			Price:     200.0,
			Volume:    3,
			ValidDays: 8,
			PriceType: trading.StopLossPriceMonetary,
		},
	}

	fmt.Printf("Placing stop loss: trigger <= 200.00 SEK, BUY 3 @ 200.00, valid until %s\n", validUntil)

	resp, err := client.Trading.PlaceStopLoss(ctx, req)
	if err != nil {
		log.Fatalf("Failed to place stop loss: %v", err)
	}

	fmt.Printf("Stop loss placed: ID=%s, status=%s\n", resp.StopLossOrderID, resp.Status)
}
