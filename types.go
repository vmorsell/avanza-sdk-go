// Package avanza provides a Go client library for the Avanza trading platform API.
package avanza

import (
	"github.com/vmorsell/avanza-sdk-go/internal/accounts"
	"github.com/vmorsell/avanza-sdk-go/internal/market"
	"github.com/vmorsell/avanza-sdk-go/internal/trading"
)

// Re-export types from internal packages for convenience.
// These are type aliases that point to the internal package types.

// Account types
type (
	AccountOverview    = accounts.AccountOverview
	Category           = accounts.Category
	Account            = accounts.Account
	AccountName        = accounts.AccountName
	AccountSettings     = accounts.AccountSettings
	Money              = accounts.Money
	Profit             = accounts.Profit
	Performance        = accounts.Performance
	PerformanceData    = accounts.PerformanceData
	Loan               = accounts.Loan
	TradingAccount     = accounts.TradingAccount
	CurrencyBalance    = accounts.CurrencyBalance
	AccountPosition    = accounts.AccountPosition
	AccountInfo        = accounts.AccountInfo
	Instrument         = accounts.Instrument
	Orderbook          = accounts.Orderbook
	Quote              = accounts.Quote
	Turnover           = accounts.Turnover
	LastDeal           = accounts.LastDeal
	PositionPerformance = accounts.PositionPerformance
	CashPosition       = accounts.CashPosition
	AccountPositions   = accounts.AccountPositions
)

// Trading types
type (
	OrderSide              = trading.OrderSide
	OrderCondition         = trading.OrderCondition
	OrderRequestStatus     = trading.OrderRequestStatus
	StopLossStatus         = trading.StopLossStatus
	OrderMetadata          = trading.OrderMetadata
	PlaceOrderRequest      = trading.PlaceOrderRequest
	PlaceOrderResponse     = trading.PlaceOrderResponse
	OrderAccount           = trading.OrderAccount
	OrderOrderbook         = trading.OrderOrderbook
	Order                  = trading.Order
	GetOrdersResponse      = trading.GetOrdersResponse
	ValidateOrderRequest   = trading.ValidateOrderRequest
	ValidateOrderResponse  = trading.ValidateOrderResponse
	ValidationResult       = trading.ValidationResult
	PreliminaryFeeRequest  = trading.PreliminaryFeeRequest
	PreliminaryFeeResponse = trading.PreliminaryFeeResponse
	CurrencyExchangeFee    = trading.CurrencyExchangeFee
	StopLossTriggerType    = trading.StopLossTriggerType
	StopLossValueType      = trading.StopLossValueType
	StopLossOrderEventType = trading.StopLossOrderEventType
	StopLossPriceType      = trading.StopLossPriceType
	StopLossTrigger        = trading.StopLossTrigger
	StopLossOrderEvent     = trading.StopLossOrderEvent
	PlaceStopLossRequest   = trading.PlaceStopLossRequest
	PlaceStopLossResponse  = trading.PlaceStopLossResponse
	StopLossAccount        = trading.StopLossAccount
	StopLossOrderbook      = trading.StopLossOrderbook
	StopLossTriggerResponse = trading.StopLossTriggerResponse
	StopLossOrderDetails   = trading.StopLossOrderDetails
	StopLossOrder          = trading.StopLossOrder
)

// Market types
type (
	OrderDepthLevel        = market.OrderDepthLevel
	OrderDepthData         = market.OrderDepthData
	OrderDepthEvent        = market.OrderDepthEvent
	OrderDepthSubscription = market.OrderDepthSubscription
)

// Re-export constants for convenience.
const (
	// OrderSide constants
	OrderSideBuy  = trading.OrderSideBuy  // Buy order
	OrderSideSell = trading.OrderSideSell // Sell order

	// OrderCondition constants
	OrderConditionNormal     = trading.OrderConditionNormal     // Standard order execution
	OrderConditionFillOrKill = trading.OrderConditionFillOrKill // Execute immediately or cancel

	// OrderRequestStatus constants
	OrderRequestStatusSuccess = trading.OrderRequestStatusSuccess // Order placed successfully
	OrderRequestStatusError   = trading.OrderRequestStatusError   // Order placement failed

	// StopLossStatus constants
	StopLossStatusSuccess = trading.StopLossStatusSuccess // Stop loss order placed successfully
	StopLossStatusError   = trading.StopLossStatusError   // Stop loss order placement failed

	// StopLossTriggerType constants
	StopLossTriggerLessOrEqual    = trading.StopLossTriggerLessOrEqual    // Trigger when price drops to or below value
	StopLossTriggerGreaterOrEqual = trading.StopLossTriggerGreaterOrEqual // Trigger when price rises to or above value

	// StopLossValueType constants
	StopLossValueMonetary   = trading.StopLossValueMonetary   // Value is an absolute price
	StopLossValuePercentage = trading.StopLossValuePercentage // Value is a percentage change

	// StopLossOrderEventType constants
	StopLossOrderEventBuy  = trading.StopLossOrderEventBuy  // Place a buy order when triggered
	StopLossOrderEventSell = trading.StopLossOrderEventSell // Place a sell order when triggered

	// StopLossPriceType constants
	StopLossPriceMonetary   = trading.StopLossPriceMonetary   // Price is an absolute value
	StopLossPricePercentage = trading.StopLossPricePercentage // Price is a percentage of current price
)

