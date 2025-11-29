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

	// Create client
	client := avanza.New()

	fmt.Println("Starting BankID authentication...")

	startResp, err := client.Auth.StartBankID(ctx)
	if err != nil {
		log.Fatalf("Failed to start BankID: %v", err)
	}

	if err := client.Auth.DisplayQRCode(startResp.QRToken); err != nil {
		log.Fatalf("Failed to display QR code: %v", err)
	}

	collectResp, err := client.Auth.PollBankIDWithQRUpdates(ctx)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Printf("Authentication successful! Welcome %s\n", collectResp.Name)

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

	// Now get account overview
	fmt.Println("\nFetching account overview...")
	overview, err := client.Accounts.GetOverview(ctx)
	if err != nil {
		log.Fatalf("Failed to get account overview: %v", err)
	}

	// Display categories
	fmt.Println("\nAccount Categories:")
	for _, category := range overview.Categories {
		fmt.Printf("- %s (%s): %.2f %s\n",
			category.Name,
			category.ID,
			category.TotalValue.Value,
			category.TotalValue.Unit)
	}

	// Display accounts
	fmt.Println("\nAccounts:")
	for _, account := range overview.Accounts {
		fmt.Printf("- %s (%s): %.2f %s\n",
			account.Name.UserDefinedName,
			account.Type,
			account.TotalValue.Value,
			account.TotalValue.Unit)
	}

	fmt.Printf("\nTotal accounts: %d\n", len(overview.Accounts))

	// Get trading accounts for detailed information
	fmt.Println("\nFetching trading accounts...")
	tradingAccounts, err := client.Accounts.GetTradingAccounts(ctx)
	if err != nil {
		log.Fatalf("Failed to get trading accounts: %v", err)
	}

	fmt.Printf("Found %d trading accounts:\n", len(tradingAccounts))
	for _, account := range tradingAccounts {
		fmt.Printf("- %s (%s): %.2f SEK available\n",
			account.Name,
			account.AccountTypeName,
			account.AvailableForPurchase)
	}

	// Get positions for the first trading account
	if len(tradingAccounts) > 0 {
		firstAccount := tradingAccounts[0]
		fmt.Printf("\nFetching positions for account: %s\n", firstAccount.Name)

		positions, err := client.Accounts.GetPositions(ctx, firstAccount.URLParameterID)
		if err != nil {
			log.Fatalf("Failed to get account positions: %v", err)
		}

		fmt.Printf("Positions in %s:\n", firstAccount.Name)
		fmt.Printf("- Stocks/Funds: %d positions\n", len(positions.WithOrderbook))
		fmt.Printf("- Cash: %d positions\n", len(positions.CashPositions))

		// Show first few positions
		for i, position := range positions.WithOrderbook {
			if i >= 3 { // Limit to first 3 positions
				fmt.Printf("  ... and %d more positions\n", len(positions.WithOrderbook)-3)
				break
			}
			fmt.Printf("  - %s: %.0f shares @ %.2f %s (Value: %.2f %s)\n",
				position.Instrument.Name,
				position.Volume.Value,
				position.AverageAcquiredPrice.Value,
				position.AverageAcquiredPrice.Unit,
				position.Value.Value,
				position.Value.Unit)
		}

		// Show cash positions
		for _, cash := range positions.CashPositions {
			fmt.Printf("  - Cash: %.2f %s\n",
				cash.TotalBalance.Value,
				cash.TotalBalance.Unit)
		}
	}
}
