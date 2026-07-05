package market

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vmorsell/avanza-sdk-go/internal/schema"
)

// driftFixtures maps each golden capture in testdata/ to the type it must
// decode into with no unmapped fields. To check drift against a fresh response,
// save it here (e.g. `curl .../market-guide/stock/<id> > testdata/<name>.json`),
// add a row, and run `go test ./market -run Drift`. Any reported path is a JSON
// key the API returns that the Go type does not yet model.
var driftFixtures = map[string]func() any{
	"stock_nvidia.json":            func() any { return &Stock{} },
	"certificate_bull_nvidia.json": func() any { return &Certificate{} },
	"certificate_bear_nvidia.json": func() any { return &Certificate{} },
	"warrant_mini_l_nvidia.json":   func() any { return &Warrant{} },
	"warrant_mini_s_nvidia.json":   func() any { return &Warrant{} },
	"warrant_turbo_l_nvidia.json":  func() any { return &Warrant{} },
	"warrant_turbo_s_nvidia.json":  func() any { return &Warrant{} },

	// Stock-only sub-endpoints.
	"stock_quote.json":       func() any { return &Quote{} },
	"stock_orderdepth.json":  func() any { return &MarketDataOrderDepth{} },
	"stock_marketplace.json": func() any { return &MarketPlace{} },

	// Off-hours price, universal across instrument types (nil quote for derivatives).
	"stock_offhours_price.json":       func() any { return &OffHoursPrice{} },
	"certificate_offhours_price.json": func() any { return &OffHoursPrice{} },
	"warrant_offhours_price.json":     func() any { return &OffHoursPrice{} },

	// Details endpoints — one shape per instrument type.
	"stock_details.json":       func() any { return &StockDetails{} },
	"certificate_details.json": func() any { return &CertificateDetails{} },
	"warrant_details.json":     func() any { return &WarrantDetails{} },
}

func TestSchemaDrift(t *testing.T) {
	for name, newTarget := range driftFixtures {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata", name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			assertNoDrift(t, name, raw, newTarget())
		})
	}
}

// assertNoDrift verifies that raw decodes into target with no unmapped fields.
// A plain decode catches type-level drift (a field whose JSON type changed);
// UnknownFields catches the more common structural drift of new or renamed keys.
// label identifies the payload in failure messages (a fixture name or a URL).
func assertNoDrift(t *testing.T, label string, raw []byte, target any) {
	t.Helper()

	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("decode %s into type: %v\n%s", label, err, snippet(raw))
	}

	unknown, err := schema.UnknownFields(raw, target)
	if err != nil {
		t.Fatalf("UnknownFields: %v", err)
	}
	if len(unknown) > 0 {
		t.Errorf("%s returns fields not modelled by the Go type — add them:\n  %s",
			label, strings.Join(unknown, "\n  "))
	}
}

func snippet(raw []byte) string {
	const max = 500
	if len(raw) > max {
		return string(raw[:max]) + "…"
	}
	return string(raw)
}
