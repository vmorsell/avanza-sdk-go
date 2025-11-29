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

	// Get preliminary fees for a potential order
	fmt.Println("\nCalculating preliminary fees...")

	orderbookID := "5247" // Investor B
	price := "350.0"
	volume := "1"
	side := avanza.OrderSideBuy

	feeReq := &avanza.PreliminaryFeeRequest{
		AccountID:   accountID,
		OrderbookID: orderbookID,
		Price:       price,
		Volume:      volume,
		Side:        side,
	}

	feeResp, err := client.Trading.GetPreliminaryFee(ctx, feeReq)
	if err != nil {
		log.Fatalf("Failed to get preliminary fee: %v", err)
	}

	// Display fee breakdown
	fmt.Printf("\nOrder Fee Breakdown:\n")
	fmt.Printf("  Order Details:\n")
	fmt.Printf("    Side:        %s\n", string(side))
	fmt.Printf("    OrderbookID: %s\n", orderbookID)
	fmt.Printf("    Price:       %s %s per share\n", price, feeResp.OrderbookCurrency)
	fmt.Printf("    Volume:      %s shares\n", volume)
	fmt.Printf("    Account:     %s\n", accountID)

	fmt.Printf("\n  Fee Breakdown:\n")
	fmt.Printf("    Commission:          %s %s\n", feeResp.Commission, feeResp.OrderbookCurrency)
	fmt.Printf("    Market Fees:         %s %s\n", feeResp.MarketFees, feeResp.OrderbookCurrency)
	fmt.Printf("    Total Fees:          %s %s\n", feeResp.TotalFees, feeResp.OrderbookCurrency)

	fmt.Printf("\n  Cost Summary:\n")
	fmt.Printf("    Order Value:         %s %s\n", feeResp.TotalSumWithoutFees, feeResp.OrderbookCurrency)
	fmt.Printf("    Total Cost:          %s %s\n", feeResp.TotalSum, feeResp.OrderbookCurrency)

	if feeResp.TransactionTax != nil {
		fmt.Printf("    Transaction Tax:     %s %s\n", *feeResp.TransactionTax, feeResp.OrderbookCurrency)
	}

	if feeResp.Campaign != nil {
		fmt.Printf("    Campaign Discount:   %s %s\n", *feeResp.Campaign, feeResp.OrderbookCurrency)
	}

	// Show currency exchange info if applicable
	if feeResp.CurrencyExchangeFee.Rate != "" || feeResp.CurrencyExchangeFee.Sum != "" {
		fmt.Printf("\n  Currency Exchange:\n")
		fmt.Printf("    Exchange Rate:      %s\n", feeResp.CurrencyExchangeFee.Rate)
		fmt.Printf("    Exchange Fee:       %s\n", feeResp.CurrencyExchangeFee.Sum)
	}

	fmt.Printf("\nThis shows the fees you would pay if you placed this order.\n")
	fmt.Printf("The actual order placement would use the PlaceOrder function.\n")
}
