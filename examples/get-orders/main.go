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

	orders, err := client.Trading.GetOrders(ctx)
	if err != nil {
		log.Fatalf("Failed to get orders: %v", err)
	}

	fmt.Printf("Orders: %d active, %d fund, %d cancelled\n",
		len(orders.Orders), len(orders.FundOrders), len(orders.CancelledOrders))

	for _, o := range orders.Orders {
		fmt.Printf("  %s %s %d @ %.2f %s — %s (%s)\n",
			o.OrderID, o.Side, o.Volume, o.Price, o.Orderbook.Currency,
			o.Orderbook.Name, o.StateText)
	}
}
