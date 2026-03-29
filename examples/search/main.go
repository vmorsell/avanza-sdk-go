package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vmorsell/avanza-sdk-go"
	"github.com/vmorsell/avanza-sdk-go/market"
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

	if err := client.Auth.DisplayQRCode(startResp.QRToken); err != nil {
		log.Fatalf("Failed to display QR code: %v", err)
	}

	collectResp, err := client.Auth.PollBankIDWithQRUpdates(ctx)
	if err != nil {
		log.Fatalf("BankID authentication failed: %v", err)
	}

	fmt.Printf("\nAuthentication successful! Welcome %s\n", collectResp.Name)

	if err := client.Auth.EstablishSession(ctx, collectResp); err != nil {
		log.Fatalf("Failed to establish session: %v", err)
	}

	// Search for instruments
	query := "investor"
	fmt.Printf("\nSearching for %q...\n", query)

	resp, err := client.Market.Search(ctx, &market.SearchRequest{
		Query: query,
		Size:  10,
	})
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	fmt.Printf("Found %d results (showing first %d):\n\n", resp.TotalNumberOfHits, len(resp.Hits))

	for _, hit := range resp.Hits {
		fmt.Printf("  %-12s %s\n", hit.Type, hit.Title)
		fmt.Printf("             Price: %s %s  Change: %s%%\n",
			hit.Price.Last, hit.Price.Currency, hit.Price.TodayChangePercent)
		fmt.Printf("             Market: %s  Orderbook ID: %s\n\n",
			hit.MarketPlaceName, hit.OrderbookID)
	}

	// Show facets (result counts by type)
	if len(resp.Facets.Types) > 0 {
		fmt.Println("Results by type:")
		for _, facet := range resp.Facets.Types {
			fmt.Printf("  %-25s %d\n", facet.Type, facet.Count)
		}
	}

	// Search with type filter
	fmt.Printf("\nSearching for %q (stocks only)...\n", query)

	stockResp, err := client.Market.Search(ctx, &market.SearchRequest{
		Query: query,
		Types: []string{"STOCK"},
		Size:  5,
	})
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	fmt.Printf("Found %d stocks:\n\n", stockResp.TotalNumberOfHits)
	for _, hit := range stockResp.Hits {
		fmt.Printf("  %s (%s) - %s %s\n",
			hit.Title, hit.FlagCode, hit.Price.Last, hit.Price.Currency)
	}
}
