package market

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/vmorsell/avanza-sdk-go/client"
)

// livePublicEndpoints enumerates the public (no-auth) endpoints checked for
// schema drift against live responses. Each row pairs a live URL with the Go
// type it must fully decode into. Authenticated endpoints belong here too once
// this check grows a logged-in client — keep them in a separate table so the
// public sweep can run without credentials.
var livePublicEndpoints = []struct {
	name   string
	path   string
	target func() any
}{
	{"stock", "/_api/market-guide/stock/4478", func() any { return &Stock{} }},
	{"stock/details", "/_api/market-guide/stock/4478/details", func() any { return &StockDetails{} }},
	{"stock/quote", "/_api/market-guide/stock/4478/quote", func() any { return &Quote{} }},
	{"stock/orderdepth", "/_api/market-guide/stock/4478/orderdepth", func() any { return &MarketDataOrderDepth{} }},
	{"stock/marketplace", "/_api/market-guide/stock/4478/marketplace", func() any { return &MarketPlace{} }},
	{"price-chart/stock", "/_api/price-chart/stock/4478?timePeriod=today", func() any { return &StockPriceChart{} }},
	{"news", "/_api/market-guide/news/4478", func() any { return &News{} }},
	{"forum", "/_api/market-guide/forum/4478", func() any { return &Forum{} }},
	{"offhours/stock", "/_push/market-offhours-price/latest/4478", func() any { return &OffHoursPrice{} }},

	{"certificate/bull", "/_api/market-guide/certificate/1612107", func() any { return &Certificate{} }},
	{"certificate/bear", "/_api/market-guide/certificate/1834856", func() any { return &Certificate{} }},
	{"certificate/bull/details", "/_api/market-guide/certificate/1612107/details", func() any { return &CertificateDetails{} }},
	{"certificate/bear/details", "/_api/market-guide/certificate/1834856/details", func() any { return &CertificateDetails{} }},
	{"offhours/certificate", "/_push/market-offhours-price/latest/1612107", func() any { return &OffHoursPrice{} }},

	{"warrant/mini-l", "/_api/market-guide/warrant/2044027", func() any { return &Warrant{} }},
	{"warrant/mini-l/details", "/_api/market-guide/warrant/2044027/details", func() any { return &WarrantDetails{} }},
	{"warrant/turbo-l/details", "/_api/market-guide/warrant/1586984/details", func() any { return &WarrantDetails{} }},
	{"warrant/mini-s", "/_api/market-guide/warrant/2191656", func() any { return &Warrant{} }},
	{"warrant/turbo-l", "/_api/market-guide/warrant/1586984", func() any { return &Warrant{} }},
	{"warrant/turbo-s", "/_api/market-guide/warrant/2521719", func() any { return &Warrant{} }},
	{"offhours/warrant", "/_push/market-offhours-price/latest/2044027", func() any { return &OffHoursPrice{} }},
}

// TestSchemaDriftLive fetches live responses from Avanza's public endpoints and
// reports any JSON fields the Go types do not model — the actual drift monitor,
// as opposed to TestSchemaDrift which only re-checks captured fixtures.
//
// It hits the network and depends on Avanza being up, so it is skipped unless
// AVANZA_LIVE=1 and therefore never runs in `make ci`:
//
//	AVANZA_LIVE=1 go test ./market -run TestSchemaDriftLive -v
func TestSchemaDriftLive(t *testing.T) {
	if os.Getenv("AVANZA_LIVE") != "1" {
		t.Skip("set AVANZA_LIVE=1 to run live schema-drift checks against avanza.se")
	}

	c := client.NewClient()
	ctx := context.Background()

	for _, ep := range livePublicEndpoints {
		t.Run(ep.name, func(t *testing.T) {
			raw, err := liveGetRaw(ctx, c, ep.path)
			if err != nil {
				t.Fatalf("fetch %s: %v", ep.path, err)
			}
			assertNoDrift(t, ep.path, raw, ep.target())
		})
	}
}

func liveGetRaw(ctx context.Context, c *client.Client, endpoint string) ([]byte, error) {
	resp, err := c.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
