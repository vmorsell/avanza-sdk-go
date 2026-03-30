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

	// Get account overview
	overview, err := client.Accounts.GetOverview(ctx)
	if err != nil {
		log.Fatalf("Failed to get overview: %v", err)
	}

	fmt.Printf("\nAccounts (%d):\n", len(overview.Accounts))
	for _, a := range overview.Accounts {
		fmt.Printf("  %s (%s): %.2f %s\n",
			a.Name.UserDefinedName, a.Type,
			a.TotalValue.Value, a.TotalValue.Unit)
	}

	// Get trading accounts
	tradingAccounts, err := client.Accounts.GetTradingAccounts(ctx)
	if err != nil {
		log.Fatalf("Failed to get trading accounts: %v", err)
	}

	fmt.Printf("\nTrading accounts (%d):\n", len(tradingAccounts))
	for _, a := range tradingAccounts {
		fmt.Printf("  %s (%s): %.2f SEK available\n",
			a.Name, a.AccountTypeName, a.AvailableForPurchase)
	}

	// Show positions for first account
	if len(tradingAccounts) == 0 {
		return
	}

	first := tradingAccounts[0]
	positions, err := client.Accounts.GetPositions(ctx, first.URLParameterID)
	if err != nil {
		log.Fatalf("Failed to get positions: %v", err)
	}

	fmt.Printf("\nPositions in %s:\n", first.Name)
	for _, p := range positions.WithOrderbook {
		fmt.Printf("  %s: %.0f shares, value %.2f %s\n",
			p.Instrument.Name, p.Volume.Value,
			p.Value.Value, p.Value.Unit)
	}
	for _, c := range positions.CashPositions {
		fmt.Printf("  Cash: %.2f %s\n", c.TotalBalance.Value, c.TotalBalance.Unit)
	}
}
