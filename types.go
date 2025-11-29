// Package avanza provides a Go client library for the Avanza trading platform API.
package avanza

import (
	"github.com/vmorsell/avanza-sdk-go/internal/accounts"
	"github.com/vmorsell/avanza-sdk-go/internal/market"
	"github.com/vmorsell/avanza-sdk-go/internal/trading"
)

// Re-export types from internal packages for backward compatibility and convenience.

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

// Re-export constants
const (
	// OrderSide constants
	OrderSideBuy  = trading.OrderSideBuy
	OrderSideSell = trading.OrderSideSell

	// OrderCondition constants
	OrderConditionNormal     = trading.OrderConditionNormal
	OrderConditionFillOrKill = trading.OrderConditionFillOrKill

	// OrderRequestStatus constants
	OrderRequestStatusSuccess = trading.OrderRequestStatusSuccess
	OrderRequestStatusError   = trading.OrderRequestStatusError

	// StopLossStatus constants
	StopLossStatusSuccess = trading.StopLossStatusSuccess
	StopLossStatusError   = trading.StopLossStatusError

	// StopLossTriggerType constants
	StopLossTriggerLessOrEqual    = trading.StopLossTriggerLessOrEqual
	StopLossTriggerGreaterOrEqual = trading.StopLossTriggerGreaterOrEqual

	// StopLossValueType constants
	StopLossValueMonetary   = trading.StopLossValueMonetary
	StopLossValuePercentage = trading.StopLossValuePercentage

	// StopLossOrderEventType constants
	StopLossOrderEventBuy  = trading.StopLossOrderEventBuy
	StopLossOrderEventSell = trading.StopLossOrderEventSell

	// StopLossPriceType constants
	StopLossPriceMonetary   = trading.StopLossPriceMonetary
	StopLossPricePercentage = trading.StopLossPricePercentage
)

