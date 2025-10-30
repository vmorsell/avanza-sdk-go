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
	fmt.Printf("Authentication successful. Welcome %s\n", collectResp.Name)

	// Establish session for API calls
	fmt.Println("Establishing session...")
	if err := client.Auth.EstablishSession(ctx, collectResp); err != nil {
		log.Fatalf("Failed to establish session: %v", err)
	}

	orderbookID := "738784" // BULL OMX X2 AVA

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

	fmt.Printf("\nSubscribing to order depth for orderbook %s...\n", orderbookID)

	// Subscribe to order depth updates
	subscription, err := client.SubscribeToOrderDepth(subCtx, orderbookID)
	if err != nil {
		log.Fatalf("Failed to subscribe to order depth: %v", err)
	}
	defer subscription.Close()

	// Process events using channels
	for {
		select {
		case event := <-subscription.Events():
			switch event.Event {
			case "ORDER_DEPTH":
				for i, lvl := range event.Data.Levels {
					buyMM := ""
					sellMM := ""
					if i == event.Data.MarketMakerLevelInBid {
						buyMM = "*"
					}
					if i == event.Data.MarketMakerLevelInAsk {
						sellMM = "*"
					}
					fmt.Printf("%d: Buy: %.0f @ %.2f%s | Sell: %.0f @ %.2f%s\n",
						i,
						lvl.BuyVolume, lvl.BuyPrice, buyMM,
						lvl.SellVolume, lvl.SellPrice, sellMM,
					)
				}
				fmt.Println()

			case "info":
				fmt.Println("[got heartbeat]")
			}
		case err := <-subscription.Errors():
			log.Printf("Subscription error: %v", err)
			return
		case <-subCtx.Done():
			fmt.Println("Order depth subscription ended.")
			return
		}
	}
}
