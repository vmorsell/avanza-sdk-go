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

	fmt.Printf("\nAuthentication successful! Welcome %s\n", collectResp.Name)

	// Establish session for API calls
	fmt.Println("Establishing session...")
	if err := client.Auth.EstablishSession(ctx, collectResp); err != nil {
		log.Fatalf("Failed to establish session: %v", err)
	}
	fmt.Println("Session established successfully!")

	// Get all active stop loss orders
	fmt.Println("\nFetching active stop loss orders...")
	stopLossOrders, err := client.Trading.GetStopLossOrders(ctx)
	if err != nil {
		log.Fatalf("Failed to get stop loss orders: %v", err)
	}

	fmt.Printf("Found %d active stop loss orders\n", len(stopLossOrders))

	// Display stop loss orders
	if len(stopLossOrders) > 0 {
		fmt.Println("\nActive Stop Loss Orders:")
		for i, order := range stopLossOrders {
			fmt.Printf("\n--- Stop Loss Order %d ---\n", i+1)
			fmt.Printf("ID: %s\n", order.ID)
			fmt.Printf("Status: %s\n", order.Status)
			fmt.Printf("Instrument: %s (%s) - %s\n", order.Orderbook.Name, order.Orderbook.ShortName, order.Orderbook.ID)
			fmt.Printf("Currency: %s\n", order.Orderbook.Currency)
			fmt.Printf("Account: %s (%s)\n", order.Account.Name, order.Account.ID)

			// Trigger information
			fmt.Printf("\nTrigger:\n")
			fmt.Printf("  Type: %s\n", order.Trigger.Type)
			fmt.Printf("  Value: %.2f %s\n", order.Trigger.Value, order.Orderbook.Currency)
			fmt.Printf("  Value Type: %s\n", order.Trigger.ValueType)
			fmt.Printf("  Valid Until: %s\n", order.Trigger.ValidUntil)
			fmt.Printf("  Market Maker Quote: %t\n", order.Trigger.TriggerOnMarketMakerQuote)

			// Order details
			fmt.Printf("\nOrder Details:\n")
			fmt.Printf("  Type: %s\n", order.Order.Type)
			fmt.Printf("  Price: %.2f %s\n", order.Order.Price, order.Orderbook.Currency)
			fmt.Printf("  Volume: %d shares\n", order.Order.Volume)
			fmt.Printf("  Price Type: %s\n", order.Order.PriceType)
			fmt.Printf("  Valid Days: %d\n", order.Order.ValidDays)
			fmt.Printf("  Short Selling Allowed: %t\n", order.Order.ShortSellingAllowed)
			fmt.Printf("  Price Decimal Precision: %d\n", order.Order.PriceDecimalPrecision)

			// Permissions
			fmt.Printf("\nPermissions:\n")
			fmt.Printf("  Editable: %t\n", order.Editable)
			fmt.Printf("  Deletable: %t\n", order.Deletable)

			if order.Message != "" {
				fmt.Printf("\nMessage: %s\n", order.Message)
			}
		}
	} else {
		fmt.Println("\nNo active stop loss orders found")
	}
}
