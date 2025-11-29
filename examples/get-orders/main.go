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

	// Get all current orders
	fmt.Println("\n📋 Fetching current orders...")
	orders, err := client.Trading.GetOrders(ctx)
	if err != nil {
		log.Fatalf("Failed to get orders: %v", err)
	}

	fmt.Printf("✅ Found %d orders\n", len(orders.Orders))
	fmt.Printf("Fund orders: %d\n", len(orders.FundOrders))
	fmt.Printf("Cancelled orders: %d\n", len(orders.CancelledOrders))

	// Display orders
	if len(orders.Orders) > 0 {
		fmt.Println("\n📊 Current Orders:")
		for i, order := range orders.Orders {
			fmt.Printf("\n--- Order %d ---\n", i+1)
			fmt.Printf("Order ID: %s\n", order.OrderID)
			fmt.Printf("Side: %s\n", order.Side)
			fmt.Printf("Volume: %d shares\n", order.Volume)
			fmt.Printf("Price: %.2f %s\n", order.Price, order.Orderbook.Currency)
			fmt.Printf("Amount: %.2f %s\n", order.Amount, order.Orderbook.Currency)
			fmt.Printf("Instrument: %s (%s)\n", order.Orderbook.Name, order.Orderbook.ID)
			fmt.Printf("State: %s (%s)\n", order.State, order.StateText)
			fmt.Printf("Created: %s\n", order.Created)
			fmt.Printf("Valid Until: %s\n", order.ValidUntil)
			fmt.Printf("Account: %s (%s)\n", order.Account.Name.Value, order.Account.AccountID)

			if order.Message != "" {
				fmt.Printf("Message: %s\n", order.Message)
			}
			if order.StateMessage != "" {
				fmt.Printf("State Message: %s\n", order.StateMessage)
			}

			fmt.Printf("Deletable: %t\n", order.Deletable)
			fmt.Printf("Modifiable: %t\n", order.Modifiable)
		}
	} else {
		fmt.Println("\n📭 No current orders found")
	}
}
