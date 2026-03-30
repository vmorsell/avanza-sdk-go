// Package auth provides shared helpers for examples.
package auth

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vmorsell/avanza-sdk-go"
	"github.com/vmorsell/avanza-sdk-go/accounts"
)

// Authenticate creates an Avanza client and completes the full BankID flow.
// It uses a 5-minute timeout for the auth phase. Calls log.Fatal on any error.
func Authenticate() *avanza.Avanza {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

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

	fmt.Printf("Authenticated as %s\n", collectResp.Name)

	if err := client.Auth.EstablishSession(ctx, collectResp); err != nil {
		log.Fatalf("Failed to establish session: %v", err)
	}

	return client
}

// FirstTradingAccount fetches trading accounts and returns the first one.
// Calls log.Fatal if none exist.
func FirstTradingAccount(ctx context.Context, client *avanza.Avanza) accounts.TradingAccount {
	accts, err := client.Accounts.GetTradingAccounts(ctx)
	if err != nil {
		log.Fatalf("Failed to get trading accounts: %v", err)
	}
	if len(accts) == 0 {
		log.Fatal("No trading accounts found")
	}
	return accts[0]
}
