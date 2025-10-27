package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	// Get session info for debugging
	sessionInfo, err := client.Auth.GetSessionInfo(ctx)
	if err != nil {
		log.Fatalf("Failed to get session info: %v", err)
	}
	fmt.Printf("Session info: Logged in as %s (ID: %s)\n", sessionInfo.User.GreetingName, sessionInfo.User.ID)

	// Debug: Print available cookies
	cookies := client.GetCookies()
	fmt.Printf("Available cookies: %v\n", cookies)

	// Check for specific cookies
	requiredCookies := []string{"AZAPERSISTENCE", "csid", "cstoken", "AZACSRF"}
	for _, cookie := range requiredCookies {
		if value, exists := cookies[cookie]; exists {
			fmt.Printf("✓ %s: %s\n", cookie, value)
		} else {
			fmt.Printf("✗ %s: missing\n", cookie)
		}
	}

	orderbookID := "2185403" // BEAR OMX X20 AVA 73

	// Create a new context for the subscription that can be cancelled
	subCtx, subCancel := context.WithCancel(ctx)
	defer subCancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		subCancel()
	}()

	fmt.Printf("\n📊 Subscribing to order depth for orderbook %s...\n", orderbookID)

	// Subscribe to order depth updates using channels (idiomatic Go)
	subscription, err := client.SubscribeToOrderDepth(subCtx, orderbookID)
	if err != nil {
		log.Fatalf("Failed to subscribe to order depth: %v", err)
	}
	defer subscription.Close()

	fmt.Println("✅ Order depth subscription active. Press Ctrl+C to stop.")

	// Process events using channels
	for {
		select {
		case event := <-subscription.Events():
			if event.Event == "ORDER_DEPTH" {
				fmt.Printf("\n📈 Order Depth Update for %s:\n", event.Data.OrderbookID)
				fmt.Printf("  Market Maker Level (Ask): %d\n", event.Data.MarketMakerLevelInAsk)
				fmt.Printf("  Market Maker Level (Bid): %d\n", event.Data.MarketMakerLevelInBid)
				fmt.Printf("  Price Levels:\n")

				for i, level := range event.Data.Levels {
					if level.BuyPrice > 0 || level.SellPrice > 0 {
						fmt.Printf("    Level %d: Buy %.2f@%.2f | Sell %.2f@%.2f\n",
							i+1, level.BuyVolume, level.BuyPrice, level.SellVolume, level.SellPrice)
					}
				}
				fmt.Println()
			} else if event.Event == "info" {
				fmt.Printf("ℹ️  Info: %s\n", event.Data.OrderbookID)
			}
		case err := <-subscription.Errors():
			log.Printf("❌ Subscription error: %v", err)
			return
		case <-subCtx.Done():
			fmt.Println("Order depth subscription ended.")
			return
		}
	}
}
