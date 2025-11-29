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

	// Get trading accounts to find account ID
	fmt.Println("Fetching trading accounts...")
	tradingAccounts, err := client.Accounts.GetTradingAccounts(ctx)
	if err != nil {
		log.Fatalf("Failed to get trading accounts: %v", err)
	}

	if len(tradingAccounts) == 0 {
		log.Fatal("No trading accounts found")
	}

	// Use the first available trading account
	account := tradingAccounts[0]
	accountID := account.AccountID

	fmt.Printf("Using account: %s (%s)\n", account.Name, account.AccountTypeName)
	fmt.Printf("Available for purchase: %.2f SEK\n", account.AvailableForPurchase)

	// Validate an order before placing it
	orderbookID := "5247" // Ericsson B
	price := 1.9998
	volume := 2
	side := avanza.OrderSideBuy

	fmt.Printf("\nValidating order: %s %d shares of orderbook ID %s at %.4f SEK...\n", string(side), volume, orderbookID, price)

	validateReq := &avanza.ValidateOrderRequest{
		IsDividendReinvestment: false,
		RequestID:              nil,
		OrderRequestParameters: nil,
		Price:                  price,
		Volume:                 volume,
		OpenVolume:             nil,
		AccountID:              accountID,
		Side:                   avanza.OrderSideBuy,
		OrderbookID:            orderbookID,
		ValidUntil:             nil,
		Metadata:               nil,
		Condition:              avanza.OrderConditionNormal,
		ISIN:                   "SE0015811963",
		Currency:               "SEK",
		MarketPlace:            "XSTO",
	}

	validateResp, err := client.Trading.ValidateOrder(ctx, validateReq)
	if err != nil {
		log.Fatalf("Failed to validate order: %v", err)
	}

	fmt.Println("\nOrder Validation Results:")
	fmt.Println("  Commission Warning:      ", formatValidation(validateResp.CommissionWarning.Valid))
	fmt.Println("  Employee Validation:    ", formatValidation(validateResp.EmployeeValidation.Valid))
	fmt.Println("  Large In Scale Warning: ", formatValidation(validateResp.LargeInScaleWarning.Valid))
	fmt.Println("  Order Value Limit:      ", formatValidation(validateResp.OrderValueLimitWarning.Valid))
	fmt.Println("  Price Ramping Warning:  ", formatValidation(validateResp.PriceRampingWarning.Valid))
	fmt.Println("  Canada Odd Lot Warning: ", formatValidation(validateResp.CanadaOddLotWarning.Valid))

	// Check if order is valid
	allValid := validateResp.CommissionWarning.Valid &&
		validateResp.EmployeeValidation.Valid &&
		validateResp.LargeInScaleWarning.Valid &&
		validateResp.OrderValueLimitWarning.Valid &&
		validateResp.PriceRampingWarning.Valid &&
		validateResp.CanadaOddLotWarning.Valid

	if allValid {
		fmt.Println("\nOrder validation passed! Order can be placed.")
	} else {
		fmt.Println("\nOrder validation failed! Check the warnings above.")
	}
}

func formatValidation(valid bool) string {
	if valid {
		return "Valid"
	}
	return "Invalid"
}
