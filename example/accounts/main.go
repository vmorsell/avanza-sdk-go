package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vmorsell/avanza"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create client
	client := avanza.New()

	// First authenticate (you'll need to do this before accessing accounts)
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
	overview, err := client.Accounts.GetAccountOverview(ctx)
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
}
