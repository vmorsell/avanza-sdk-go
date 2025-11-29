// Package avanza provides a Go client library for the Avanza trading platform API.
package avanza

import (
	"github.com/vmorsell/avanza-sdk-go/accounts"
	"github.com/vmorsell/avanza-sdk-go/market"
	"github.com/vmorsell/avanza-sdk-go/trading"
)

// Type aliases from subpackages. Import from root for consistency,
// or directly from subpackages if you only need specific types.

// Account types
type (
	AccountOverview     = accounts.AccountOverview
	Category            = accounts.Category
	Account             = accounts.Account
	AccountName         = accounts.AccountName
	AccountSettings     = accounts.AccountSettings
	Money               = accounts.Money
	Profit              = accounts.Profit
	Performance         = accounts.Performance
	PerformanceData     = accounts.PerformanceData
	Loan                = accounts.Loan
	TradingAccount      = accounts.TradingAccount
	CurrencyBalance     = accounts.CurrencyBalance
	AccountPosition     = accounts.AccountPosition
	AccountInfo         = accounts.AccountInfo
	Instrument          = accounts.Instrument
	Orderbook           = accounts.Orderbook
	Quote               = accounts.Quote
	Turnover            = accounts.Turnover
	LastDeal            = accounts.LastDeal
	PositionPerformance = accounts.PositionPerformance
	CashPosition        = accounts.CashPosition
	AccountPositions    = accounts.AccountPositions
)

// Trading types
type (
	OrderSide               = trading.OrderSide
	OrderCondition          = trading.OrderCondition
	OrderRequestStatus      = trading.OrderRequestStatus
	StopLossStatus          = trading.StopLossStatus
	OrderMetadata           = trading.OrderMetadata
	PlaceOrderRequest       = trading.PlaceOrderRequest
	PlaceOrderResponse      = trading.PlaceOrderResponse
	OrderAccount            = trading.OrderAccount
	OrderOrderbook          = trading.OrderOrderbook
	Order                   = trading.Order
	GetOrdersResponse       = trading.GetOrdersResponse
	ValidateOrderRequest    = trading.ValidateOrderRequest
	ValidateOrderResponse   = trading.ValidateOrderResponse
	ValidationResult        = trading.ValidationResult
	PreliminaryFeeRequest   = trading.PreliminaryFeeRequest
	PreliminaryFeeResponse  = trading.PreliminaryFeeResponse
	CurrencyExchangeFee     = trading.CurrencyExchangeFee
	StopLossTriggerType     = trading.StopLossTriggerType
	StopLossValueType       = trading.StopLossValueType
	StopLossOrderEventType  = trading.StopLossOrderEventType
	StopLossPriceType       = trading.StopLossPriceType
	StopLossTrigger         = trading.StopLossTrigger
	StopLossOrderEvent      = trading.StopLossOrderEvent
	PlaceStopLossRequest    = trading.PlaceStopLossRequest
	PlaceStopLossResponse   = trading.PlaceStopLossResponse
	StopLossAccount         = trading.StopLossAccount
	StopLossOrderbook       = trading.StopLossOrderbook
	StopLossTriggerResponse = trading.StopLossTriggerResponse
	StopLossOrderDetails    = trading.StopLossOrderDetails
	StopLossOrder           = trading.StopLossOrder
)

// Market types
type (
	OrderDepthLevel        = market.OrderDepthLevel
	OrderDepthData         = market.OrderDepthData
	OrderDepthEvent        = market.OrderDepthEvent
	OrderDepthSubscription = market.OrderDepthSubscription
)

const (
	OrderSideBuy  = trading.OrderSideBuy
	OrderSideSell = trading.OrderSideSell

	OrderConditionNormal     = trading.OrderConditionNormal
	OrderConditionFillOrKill = trading.OrderConditionFillOrKill

	OrderRequestStatusSuccess = trading.OrderRequestStatusSuccess
	OrderRequestStatusError   = trading.OrderRequestStatusError

	StopLossStatusSuccess = trading.StopLossStatusSuccess
	StopLossStatusError   = trading.StopLossStatusError

	StopLossTriggerLessOrEqual    = trading.StopLossTriggerLessOrEqual
	StopLossTriggerGreaterOrEqual = trading.StopLossTriggerGreaterOrEqual

	StopLossValueMonetary   = trading.StopLossValueMonetary
	StopLossValuePercentage = trading.StopLossValuePercentage

	StopLossOrderEventBuy  = trading.StopLossOrderEventBuy
	StopLossOrderEventSell = trading.StopLossOrderEventSell

	StopLossPriceMonetary   = trading.StopLossPriceMonetary
	StopLossPricePercentage = trading.StopLossPricePercentage
)
