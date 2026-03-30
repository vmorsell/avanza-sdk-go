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

	req := &trading.ValidateOrderRequest{
		Price:       1.9998,
		Volume:      2,
		AccountID:   account.AccountID,
		Side:        trading.OrderSideBuy,
		OrderbookID: "5247", // Investor B
		Condition:   trading.OrderConditionNormal,
		ISIN:        "SE0015811963",
		Currency:    "SEK",
		MarketPlace: "XSTO",
	}

	fmt.Printf("Validating: BUY 2 x Investor B @ 1.9998 SEK\n")

	resp, err := client.Trading.ValidateOrder(ctx, req)
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	checks := []struct {
		name  string
		valid bool
	}{
		{"Commission warning", resp.CommissionWarning.Valid},
		{"Employee validation", resp.EmployeeValidation.Valid},
		{"Large-in-scale warning", resp.LargeInScaleWarning.Valid},
		{"Order value limit", resp.OrderValueLimitWarning.Valid},
		{"Price ramping warning", resp.PriceRampingWarning.Valid},
		{"Canada odd lot warning", resp.CanadaOddLotWarning.Valid},
	}

	allValid := true
	for _, c := range checks {
		status := "OK"
		if !c.valid {
			status = "FAILED"
			allValid = false
		}
		fmt.Printf("  %-25s %s\n", c.name, status)
	}

	if allValid {
		fmt.Println("\nAll checks passed.")
	} else {
		fmt.Println("\nSome checks failed — review warnings before placing.")
	}
}
