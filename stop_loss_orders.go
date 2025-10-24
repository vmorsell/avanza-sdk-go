package avanza

import (
	"context"
	"encoding/json"
	"fmt"
)

// StopLossAccount represents account information for a stop loss order.
type StopLossAccount struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	URLParameterID string `json:"urlParameterId"`
}

// StopLossOrderbook represents orderbook information for a stop loss order.
type StopLossOrderbook struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	CountryCode              string `json:"countryCode"`
	Currency                 string `json:"currency"`
	ShortName                string `json:"shortName"`
	Type                     string `json:"type"`
	StoplossMarketMakerQuote bool   `json:"stoplossMarketMakerQuote"`
}

// StopLossTriggerResponse represents trigger information from the API response.
type StopLossTriggerResponse struct {
	Value                     float64 `json:"value"`
	Type                      string  `json:"type"`
	ValidUntil                string  `json:"validUntil"`
	ValueType                 string  `json:"valueType"`
	TriggerOnMarketMakerQuote bool    `json:"triggerOnMarketMakerQuote"`
}

// StopLossOrderDetails represents order details for a stop loss order.
type StopLossOrderDetails struct {
	Type                  string  `json:"type"`
	Price                 float64 `json:"price"`
	Volume                int     `json:"volume"`
	ShortSellingAllowed   bool    `json:"shortSellingAllowed"`
	ValidDays             int     `json:"validDays"`
	PriceType             string  `json:"priceType"`
	PriceDecimalPrecision int     `json:"priceDecimalPrecision"`
}

// StopLossOrder represents a single stop loss order.
type StopLossOrder struct {
	ID        string                  `json:"id"`
	Status    string                  `json:"status"`
	Account   StopLossAccount         `json:"account"`
	Orderbook StopLossOrderbook       `json:"orderbook"`
	Message   string                  `json:"message"`
	Trigger   StopLossTriggerResponse `json:"trigger"`
	Order     StopLossOrderDetails    `json:"order"`
	Editable  bool                    `json:"editable"`
	Deletable bool                    `json:"deletable"`
}

// GetStopLossOrders retrieves all active stop loss orders.
func (a *Avanza) GetStopLossOrders(ctx context.Context) ([]StopLossOrder, error) {
	httpResp, err := a.client.Get(ctx, "/_api/trading/stoploss/")
	if err != nil {
		return nil, fmt.Errorf("failed to get stop loss orders: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	var orders []StopLossOrder
	if err := json.NewDecoder(httpResp.Body).Decode(&orders); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return orders, nil
}
