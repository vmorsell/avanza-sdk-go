// Package trading provides trading functionality for the Avanza API.
package trading

import (
	"fmt"
	"strconv"
)

// Validate validates a PlaceOrderRequest and returns an error if any required fields are missing or invalid.
func (r *PlaceOrderRequest) Validate() error {
	if r.AccountID == "" {
		return fmt.Errorf("accountId is required")
	}
	if r.OrderbookID == "" {
		return fmt.Errorf("orderbookId is required")
	}
	if r.Price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	if r.Volume <= 0 {
		return fmt.Errorf("volume must be greater than 0")
	}
	if r.Side != OrderSideBuy && r.Side != OrderSideSell {
		return fmt.Errorf("side must be %s or %s", OrderSideBuy, OrderSideSell)
	}
	if r.Condition != OrderConditionNormal && r.Condition != OrderConditionFillOrKill {
		return fmt.Errorf("condition must be %s or %s", OrderConditionNormal, OrderConditionFillOrKill)
	}
	return nil
}

// Validate validates a ValidateOrderRequest and returns an error if any required fields are missing or invalid.
func (r *ValidateOrderRequest) Validate() error {
	if r.AccountID == "" {
		return fmt.Errorf("accountId is required")
	}
	if r.OrderbookID == "" {
		return fmt.Errorf("orderbookId is required")
	}
	if r.Price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	if r.Volume <= 0 {
		return fmt.Errorf("volume must be greater than 0")
	}
	if r.Side != OrderSideBuy && r.Side != OrderSideSell {
		return fmt.Errorf("side must be %s or %s", OrderSideBuy, OrderSideSell)
	}
	if r.Condition != OrderConditionNormal && r.Condition != OrderConditionFillOrKill {
		return fmt.Errorf("condition must be %s or %s", OrderConditionNormal, OrderConditionFillOrKill)
	}
	if r.ISIN == "" {
		return fmt.Errorf("isin is required")
	}
	if r.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if r.MarketPlace == "" {
		return fmt.Errorf("marketPlace is required")
	}
	return nil
}

// Validate validates a PreliminaryFeeRequest and returns an error if any required fields are missing or invalid.
func (r *PreliminaryFeeRequest) Validate() error {
	if r.AccountID == "" {
		return fmt.Errorf("accountId is required")
	}
	if r.OrderbookID == "" {
		return fmt.Errorf("orderbookId is required")
	}
	if r.Price == "" {
		return fmt.Errorf("price is required")
	}
	price, err := strconv.ParseFloat(r.Price, 64)
	if err != nil {
		return fmt.Errorf("price must be a valid number: %w", err)
	}
	if price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	if r.Volume == "" {
		return fmt.Errorf("volume is required")
	}
	volume, err := strconv.Atoi(r.Volume)
	if err != nil {
		return fmt.Errorf("volume must be a valid integer: %w", err)
	}
	if volume <= 0 {
		return fmt.Errorf("volume must be greater than 0")
	}
	if r.Side != OrderSideBuy && r.Side != OrderSideSell {
		return fmt.Errorf("side must be %s or %s", OrderSideBuy, OrderSideSell)
	}
	return nil
}

// Validate validates a PlaceStopLossRequest and returns an error if any required fields are missing or invalid.
func (r *PlaceStopLossRequest) Validate() error {
	if r.AccountID == "" {
		return fmt.Errorf("accountId is required")
	}
	if r.OrderbookID == "" {
		return fmt.Errorf("orderbookId is required")
	}
	if err := r.StopLossTrigger.Validate(); err != nil {
		return fmt.Errorf("stopLossTrigger: %w", err)
	}
	if err := r.StopLossOrderEvent.Validate(); err != nil {
		return fmt.Errorf("stopLossOrderEvent: %w", err)
	}
	return nil
}

// Validate validates a StopLossTrigger and returns an error if any required fields are missing or invalid.
func (t *StopLossTrigger) Validate() error {
	if t.Type != StopLossTriggerLessOrEqual && t.Type != StopLossTriggerGreaterOrEqual {
		return fmt.Errorf("type must be %s or %s", StopLossTriggerLessOrEqual, StopLossTriggerGreaterOrEqual)
	}
	if t.Value <= 0 {
		return fmt.Errorf("value must be greater than 0")
	}
	if t.ValueType != StopLossValueMonetary && t.ValueType != StopLossValuePercentage {
		return fmt.Errorf("valueType must be %s or %s", StopLossValueMonetary, StopLossValuePercentage)
	}
	return nil
}

// Validate validates a StopLossOrderEvent and returns an error if any required fields are missing or invalid.
func (e *StopLossOrderEvent) Validate() error {
	if e.Type != StopLossOrderEventBuy && e.Type != StopLossOrderEventSell {
		return fmt.Errorf("type must be %s or %s", StopLossOrderEventBuy, StopLossOrderEventSell)
	}
	if e.Price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	if e.Volume <= 0 {
		return fmt.Errorf("volume must be greater than 0")
	}
	if e.ValidDays <= 0 {
		return fmt.Errorf("validDays must be greater than 0")
	}
	if e.PriceType != StopLossPriceMonetary && e.PriceType != StopLossPricePercentage {
		return fmt.Errorf("priceType must be %s or %s", StopLossPriceMonetary, StopLossPricePercentage)
	}
	return nil
}
