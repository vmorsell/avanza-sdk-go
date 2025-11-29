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

	// Authenticate with BankID
	fmt.Println("Starting BankID authentication...")
	startResp, err := client.Auth.StartBankID(ctx)
	if err != nil {
		log.Fatalf("Failed to start BankID: %v", err)
	}

	// Display QR code
	if err := client.Auth.DisplayQRCode(startResp.QRToken); err != nil {
		log.Fatalf("Failed to display QR code: %v", err)
	}

	// Poll for authentication completion with automatic QR refresh
	collectResp, err := client.Auth.PollBankIDWithQRUpdates(ctx)
	if err != nil {
		log.Fatalf("BankID authentication failed: %v", err)
	}

	fmt.Printf("\n✅ Authentication successful! Welcome %s\n", collectResp.Name)

	// Establish session for API calls
	fmt.Println("Establishing session...")
	if err := client.Auth.EstablishSession(ctx, collectResp); err != nil {
		log.Fatalf("Failed to establish session: %v", err)
	}
	fmt.Println("Session established successfully!")

	// Get trading accounts to find account ID
	fmt.Println("Fetching trading accounts...")
	tradingAccounts, err := client.Accounts.GetTradingAccounts(ctx)
	if err != nil {
		log.Fatalf("Failed to get trading accounts: %v", err)
	}

	if len(tradingAccounts) == 0 {
		log.Fatal("No trading accounts found")
	}

	// Use the first available trading account
	account := tradingAccounts[0]
	accountID := account.AccountID

	fmt.Printf("Using account: %s (%s)\n", account.Name, account.AccountTypeName)
	fmt.Printf("Available for purchase: %.2f SEK\n", account.AvailableForPurchase)

	// Place a stop loss order
	orderbookID := "5247" // Example: Ericsson B
	triggerValue := 200.0
	orderPrice := 200.0
	volume := 3

	fmt.Printf("\nPlacing stop loss order:")
	fmt.Printf("  Orderbook ID: %s\n", orderbookID)
	fmt.Printf("  Trigger: When price <= %.2f SEK\n", triggerValue)
	fmt.Printf("  Action: BUY %d shares at %.2f SEK\n", volume, orderPrice)
	fmt.Printf("  Valid until: 2025-11-23\n")

	stopLossReq := &avanza.PlaceStopLossRequest{
		ParentStopLossID: "0", // New stop loss order
		AccountID:        accountID,
		OrderBookID:      orderbookID,
		StopLossTrigger: avanza.StopLossTrigger{
			Type:                      avanza.StopLossTriggerLessOrEqual,
			Value:                     triggerValue,
			ValueType:                 avanza.StopLossValueMonetary,
			ValidUntil:                "2025-11-23",
			TriggerOnMarketMakerQuote: false,
		},
		StopLossOrderEvent: avanza.StopLossOrderEvent{
			Type:                avanza.StopLossOrderEventBuy,
			Price:               orderPrice,
			Volume:              volume,
			ValidDays:           8,
			PriceType:           avanza.StopLossPriceMonetary,
			ShortSellingAllowed: false,
		},
	}

	stopLossResp, err := client.Trading.PlaceStopLoss(ctx, stopLossReq)
	if err != nil {
		log.Fatalf("Failed to place stop loss order: %v", err)
	}

	fmt.Printf("\n✅ Stop loss order placed successfully!")
	fmt.Printf("  Status: %s\n", stopLossResp.Status)
	fmt.Printf("  Stop Loss Order ID: %s\n", stopLossResp.StopLossOrderID)

	fmt.Println("\n📋 Stop Loss Order Details:")
	fmt.Printf("  Trigger Condition: Price <= %.2f SEK\n", triggerValue)
	fmt.Printf("  Trigger Type: %s\n", stopLossReq.StopLossTrigger.Type)
	fmt.Printf("  Value Type: %s\n", stopLossReq.StopLossTrigger.ValueType)
	fmt.Printf("  Valid Until: %s\n", stopLossReq.StopLossTrigger.ValidUntil)
	fmt.Printf("  Order Action: %s %d shares at %.2f SEK\n",
		stopLossReq.StopLossOrderEvent.Type,
		stopLossReq.StopLossOrderEvent.Volume,
		stopLossReq.StopLossOrderEvent.Price)
	fmt.Printf("  Order Valid Days: %d\n", stopLossReq.StopLossOrderEvent.ValidDays)
	fmt.Printf("  Price Type: %s\n", stopLossReq.StopLossOrderEvent.PriceType)
}
