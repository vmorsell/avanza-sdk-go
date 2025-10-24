package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/vmorsell/avanza-sdk-go"
	"github.com/vmorsell/avanza-sdk-go/internal/trading"
)

func main() {
	client := avanza.New()

	// Authenticate with BankID
	fmt.Println("Starting BankID authentication...")
	startResp, err := client.Auth.StartBankID(context.Background())
	if err != nil {
		log.Fatalf("Failed to start BankID: %v", err)
	}

	// Display QR code
	if err := client.Auth.DisplayQRCode(startResp.QRToken); err != nil {
		log.Fatalf("Failed to display QR code: %v", err)
	}

	// Poll for authentication completion with automatic QR refresh
	collectResp, err := client.Auth.PollBankIDWithQRUpdates(context.Background())
	if err != nil {
		log.Fatalf("BankID authentication failed: %v", err)
	}

	fmt.Printf("\n✅ Authentication successful! Welcome %s\n", collectResp.Name)

	// Establish session for API calls
	fmt.Println("Establishing session...")
	if err := client.Auth.EstablishSession(context.Background(), collectResp); err != nil {
		log.Fatalf("Failed to establish session: %v", err)
	}
	fmt.Println("Session established successfully!")

	// Get trading accounts to find account ID
	fmt.Println("Fetching trading accounts...")
	tradingAccounts, err := client.Accounts.GetTradingAccounts(context.Background())
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

	// Place a buy order
	// Note: This is a real order! Make sure you want to execute it.
	orderReq := &trading.PlaceOrderRequest{
		IsDividendReinvestment: false,
		RequestID:              uuid.New().String(),
		Price:                  2.0, // Low price to avoid order being filled
		Volume:                 1,   // One share
		AccountID:              accountID,
		Side:                   trading.OrderSideBuy,
		OrderbookID:            "5247", // Orderbook ID for Investor B
		Condition:              trading.OrderConditionNormal,
		Metadata: trading.OrderMetadata{
			OrderEntryMode:  "ADVANCED",
			HasTouchedPrice: "true",
		},
	}

	fmt.Printf("\nPlacing order:\n")
	fmt.Printf("  Side:        %s\n", orderReq.Side)
	fmt.Printf("  OrderbookID: %s\n", orderReq.OrderbookID)
	fmt.Printf("  Price:       %.2f\n", orderReq.Price)
	fmt.Printf("  Volume:      %d\n", orderReq.Volume)
	fmt.Printf("  Account:     %s\n", orderReq.AccountID)

	orderResp, err := client.Trading.PlaceOrder(context.Background(), orderReq)
	if err != nil {
		log.Fatalf("Failed to place order: %v", err)
	}

	fmt.Printf("\n✅ Order placed successfully!\n")
	fmt.Printf("  Order ID: %s\n", orderResp.OrderID)
	fmt.Printf("  Status:   %s\n", orderResp.OrderRequestStatus)
	if orderResp.Message != "" {
		fmt.Printf("  Message:  %s\n", orderResp.Message)
	}
}
