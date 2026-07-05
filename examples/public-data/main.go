// Command public-data demonstrates the market-data endpoints that need no login.
//
// Unlike the other examples, it never calls the BankID auth flow: a plain
// avanza.New() client can search and read instrument data straight away.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/vmorsell/avanza-sdk-go"
	"github.com/vmorsell/avanza-sdk-go/market"
)

func main() {
	// No authentication — these endpoints serve public market data.
	c := avanza.New()
	ctx := context.Background()

	// Search for an instrument.
	const query = "investor"
	search, err := c.Market.Search(ctx, &market.SearchRequest{Query: query, Size: 3})
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}
	fmt.Printf("Search %q — %d hits:\n", query, search.TotalNumberOfHits)
	for _, hit := range search.Hits {
		fmt.Printf("  %-6s %-30s %s %s\n", hit.Type, hit.Title, hit.Price.Last, hit.Price.Currency)
	}

	const orderbookID = "5247" // Investor B

	// Quote and key indicators.
	stock, err := c.Market.GetStock(ctx, orderbookID)
	if err != nil {
		log.Fatalf("GetStock failed: %v", err)
	}
	fmt.Printf("\n%s (%s): %.2f %s (%.2f%%)\n",
		stock.Name, stock.Listing.TickerSymbol,
		stock.Quote.Last, stock.Listing.Currency, stock.Quote.ChangePercent)

	// Extended company data.
	details, err := c.Market.GetStockDetails(ctx, orderbookID)
	if err != nil {
		log.Fatalf("GetStockDetails failed: %v", err)
	}
	fmt.Printf("  CEO: %s | Chairman: %s\n", details.Company.CEO, details.Company.Chairman)
	if len(details.CompanyOwners.Owners) > 0 {
		top := details.CompanyOwners.Owners[0]
		fmt.Printf("  Top owner: %s (%.1f%% capital, %.1f%% votes)\n",
			top.Name, top.PercentOfCapital, top.PercentOfVotes)
	}

	// Intraday OHLC price chart.
	chart, err := c.Market.GetStockPriceChart(ctx, orderbookID, market.TimePeriodToday)
	if err != nil {
		log.Fatalf("GetStockPriceChart failed: %v", err)
	}
	fmt.Printf("  Price chart: %d bars at %q resolution (prev close %.2f)\n",
		len(chart.OHLC), chart.Metadata.Resolution.ChartResolution, chart.PreviousClosingPrice)

	// Latest news headlines.
	news, err := c.Market.GetNews(ctx, orderbookID)
	if err != nil {
		log.Fatalf("GetNews failed: %v", err)
	}
	fmt.Printf("\nNews (%d articles):\n", len(news.Articles))
	for i, a := range news.Articles {
		if i >= 3 {
			break
		}
		fmt.Printf("  %s — %s (%s)\n", a.TimePublished, a.Headline, a.NewsSource)
	}

	// Community forum activity.
	forum, err := c.Market.GetForum(ctx, orderbookID)
	if err != nil {
		log.Fatalf("GetForum failed: %v", err)
	}
	fmt.Printf("\nForum (%d posts):\n", len(forum.Posts))
	for i, p := range forum.Posts {
		if i >= 3 {
			break
		}
		fmt.Printf("  %s: %q (%d likes, %d replies)\n", p.Author, p.Title, p.Likes, p.Replies)
	}
}
