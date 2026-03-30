package main

import (
	"context"
	"fmt"
	"log"

	"github.com/vmorsell/avanza-sdk-go/examples/internal/auth"
	"github.com/vmorsell/avanza-sdk-go/market"
)

func main() {
	client := auth.Authenticate()
	ctx := context.Background()

	// Search all instrument types
	query := "investor"
	fmt.Printf("Searching for %q...\n", query)

	resp, err := client.Market.Search(ctx, &market.SearchRequest{
		Query: query,
		Size:  10,
	})
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	fmt.Printf("Found %d results:\n\n", resp.TotalNumberOfHits)
	for _, hit := range resp.Hits {
		fmt.Printf("  %-12s %s — %s %s (change %s%%)\n",
			hit.Type, hit.Title,
			hit.Price.Last, hit.Price.Currency, hit.Price.TodayChangePercent)
	}

	// Facets
	if len(resp.Facets.Types) > 0 {
		fmt.Println("\nBy type:")
		for _, f := range resp.Facets.Types {
			fmt.Printf("  %-25s %d\n", f.Type, f.Count)
		}
	}

	// Filtered search: stocks only
	fmt.Printf("\nSearching for %q (stocks only)...\n", query)

	stockResp, err := client.Market.Search(ctx, &market.SearchRequest{
		Query: query,
		Types: []string{"STOCK"},
		Size:  5,
	})
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	fmt.Printf("Found %d stocks:\n", stockResp.TotalNumberOfHits)
	for _, hit := range stockResp.Hits {
		fmt.Printf("  %s (%s) — %s %s\n",
			hit.Title, hit.FlagCode, hit.Price.Last, hit.Price.Currency)
	}
}
