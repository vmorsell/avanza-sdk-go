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

	// Fetch stock details
	stock, err := client.Market.GetStock(ctx, "5247")
	if err != nil {
		log.Fatalf("GetStock failed: %v", err)
	}

	fmt.Printf("Stock: %s (%s)\n", stock.Name, stock.Listing.TickerSymbol)
	fmt.Printf("  ISIN:     %s\n", stock.ISIN)
	fmt.Printf("  Price:    %.2f %s (%.2f%%)\n",
		stock.Quote.Last, stock.Listing.Currency, stock.Quote.ChangePercent)
	fmt.Printf("  P/E:      %.2f\n", stock.KeyIndicators.PriceEarningsRatio)
	fmt.Printf("  Owners:   %d\n", stock.KeyIndicators.NumberOfOwners)
	if stock.KeyIndicators.Dividend != nil {
		fmt.Printf("  Dividend: %.2f %s (ex %s)\n",
			stock.KeyIndicators.Dividend.Amount,
			stock.KeyIndicators.Dividend.CurrencyCode,
			stock.KeyIndicators.Dividend.ExDate)
	}

	// Fetch certificate details
	cert, err := client.Market.GetCertificate(ctx, "2321838")
	if err != nil {
		log.Fatalf("GetCertificate failed: %v", err)
	}

	fmt.Printf("\nCertificate: %s\n", cert.Name)
	fmt.Printf("  Price:      %.2f %s (%.2f%%)\n",
		cert.Quote.Last, cert.Listing.Currency, cert.Quote.ChangePercent)
	fmt.Printf("  Leverage:   %.0fx\n", cert.KeyIndicators.Leverage)
	fmt.Printf("  Underlying: %s (%.2f)\n",
		cert.Underlying.Name, cert.Underlying.Quote.Last)

	// Fetch warrant details
	warrant, err := client.Market.GetWarrant(ctx, "564075")
	if err != nil {
		log.Fatalf("GetWarrant failed: %v", err)
	}

	fmt.Printf("\nWarrant: %s\n", warrant.Name)
	fmt.Printf("  Price:      %.2f %s (%.2f%%)\n",
		warrant.Quote.Last, warrant.Listing.Currency, warrant.Quote.ChangePercent)
	fmt.Printf("  Direction:  %s\n", warrant.KeyIndicators.Direction)
	fmt.Printf("  Leverage:   %.2fx\n", warrant.KeyIndicators.Leverage)
	fmt.Printf("  Barrier:    %.2f\n", warrant.KeyIndicators.BarrierLevel)
	fmt.Printf("  Underlying: %s (%.2f)\n",
		warrant.Underlying.Name, warrant.Underlying.Quote.Last)
}
