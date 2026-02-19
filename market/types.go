// Package market provides market data functionality for the Avanza API.
package market

import "encoding/json"

// OrderDepthLevel contains bid/ask prices and volumes at a single price level.
type OrderDepthLevel struct {
	BuyPrice   float64 `json:"buyPrice"`
	BuyVolume  float64 `json:"buyVolume"`
	SellPrice  float64 `json:"sellPrice"`
	SellVolume float64 `json:"sellVolume"`
}

// UnmarshalJSON divides BuyPrice and SellPrice by 10, converting them from
// SEK to USD.
func (o *OrderDepthLevel) UnmarshalJSON(data []byte) error {
	type OrderDepthLevelAlias OrderDepthLevel
	var alias OrderDepthLevelAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*o = OrderDepthLevel(alias)
	o.BuyPrice /= 10
	o.SellPrice /= 10
	return nil
}

// OrderDepthData contains the complete order book snapshot.
type OrderDepthData struct {
	OrderbookID           string            `json:"orderbookId"`
	Levels                []OrderDepthLevel `json:"levels"`
	MarketMakerLevelInAsk int               `json:"marketMakerLevelInAsk"`
	MarketMakerLevelInBid int               `json:"marketMakerLevelInBid"`
}

// OrderDepthEvent is a single event from the order depth subscription stream.
type OrderDepthEvent struct {
	Event string         `json:"event"`
	Data  OrderDepthData `json:"data"`
	ID    string         `json:"id"`
	Retry int            `json:"retry"`
}
