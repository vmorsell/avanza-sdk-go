package main

import (
	"context"
	"fmt"
	"log"

	"github.com/vmorsell/avanza-sdk-go/examples/internal/auth"
	"github.com/vmorsell/avanza-sdk-go/trading"
)

func main() {
	client := auth.Authenticate()
	ctx := context.Background()
	account := auth.FirstTradingAccount(ctx, client)

	resp, err := client.Trading.GetPreliminaryFee(ctx, &trading.PreliminaryFeeRequest{
		AccountID:   account.AccountID,
		OrderbookID: "5247", // Investor B
		Price:       "350.0",
		Volume:      "1",
		Side:        trading.OrderSideBuy,
	})
	if err != nil {
		log.Fatalf("Failed to get fee: %v", err)
	}

	fmt.Printf("Fee for BUY 1 x Investor B @ 350.0 %s:\n", resp.OrderbookCurrency)
	fmt.Printf("  Commission:  %s\n", resp.Commission)
	fmt.Printf("  Market fees: %s\n", resp.MarketFees)
	fmt.Printf("  Total fees:  %s\n", resp.TotalFees)
	fmt.Printf("  Order value: %s\n", resp.TotalSumWithoutFees)
	fmt.Printf("  Total cost:  %s\n", resp.TotalSum)
}
