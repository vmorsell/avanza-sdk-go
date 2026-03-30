package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vmorsell/avanza-sdk-go/examples/internal/auth"
)

func main() {
	client := auth.Authenticate()

	orderbookID := "738784" // BULL OMX X2 AVA

	// Use a cancellable context (no timeout — runs until interrupted).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	fmt.Printf("Subscribing to order depth for %s (Ctrl-C to stop)...\n\n", orderbookID)

	sub, err := client.Market.SubscribeToOrderDepth(ctx, orderbookID)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub.Close()

	for {
		select {
		case event := <-sub.Events():
			switch event.Event {
			case "ORDER_DEPTH":
				for i, lvl := range event.Data.Levels {
					buyMM, sellMM := "", ""
					if i == event.Data.MarketMakerLevelInBid {
						buyMM = "*"
					}
					if i == event.Data.MarketMakerLevelInAsk {
						sellMM = "*"
					}
					fmt.Printf("%d: Buy %.0f @ %.2f%s | Sell %.0f @ %.2f%s\n",
						i, lvl.BuyVolume, lvl.BuyPrice, buyMM,
						lvl.SellVolume, lvl.SellPrice, sellMM)
				}
				fmt.Println()
			case "info":
				fmt.Println("[heartbeat]")
			}
		case err := <-sub.Errors():
			log.Printf("Subscription error: %v", err)
			return
		case <-ctx.Done():
			return
		}
	}
}
