package main

import (
	"context"
	"fmt"
	"log"

	"github.com/vmorsell/avanza-sdk-go/examples/internal/auth"
)

func main() {
	client := auth.Authenticate()
	ctx := context.Background()

	orders, err := client.Trading.GetStopLossOrders(ctx)
	if err != nil {
		log.Fatalf("Failed to get stop loss orders: %v", err)
	}

	fmt.Printf("Active stop loss orders: %d\n", len(orders))

	for _, o := range orders {
		fmt.Printf("  %s %s — trigger %s %.2f %s, %s %d @ %.2f, valid until %s\n",
			o.ID, o.Orderbook.Name,
			o.Trigger.Type, o.Trigger.Value, o.Orderbook.Currency,
			o.Order.Type, o.Order.Volume, o.Order.Price,
			o.Trigger.ValidUntil)
	}
}
