// Package trading provides trading functionality for the Avanza API.
package trading

// OrderSide indicates whether to buy or sell.
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"  // Buy order
	OrderSideSell OrderSide = "SELL" // Sell order
)

// OrderCondition specifies how the order should be executed.
type OrderCondition string

const (
	OrderConditionNormal     OrderCondition = "NORMAL"       // Standard order execution
	OrderConditionFillOrKill OrderCondition = "FILL_OR_KILL" // Execute immediately or cancel
)

// OrderRequestStatus indicates the result of placing an order.
type OrderRequestStatus string

const (
	OrderRequestStatusSuccess OrderRequestStatus = "SUCCESS" // Order placed successfully
	OrderRequestStatusError   OrderRequestStatus = "ERROR"   // Order placement failed
)

// StopLossStatus indicates the result of placing a stop loss order.
type StopLossStatus string

const (
	StopLossStatusSuccess StopLossStatus = "SUCCESS" // Stop loss order placed successfully
	StopLossStatusError   StopLossStatus = "ERROR"   // Stop loss order placement failed
)

// OrderMetadata contains order entry details.
type OrderMetadata struct {
	OrderEntryMode  string `json:"orderEntryMode"`
	HasTouchedPrice string `json:"hasTouchedPrice"`
}

// PlaceOrderRequest contains all parameters needed to place an order.
// Use Validate() before sending to ensure all required fields are set.
type PlaceOrderRequest struct {
	IsDividendReinvestment bool           `json:"isDividendReinvestment"`
	RequestID              string         `json:"requestId"`
	OrderRequestParameters interface{}    `json:"orderRequestParameters"`
	Price                  float64        `json:"price"`
	Volume                 int            `json:"volume"`
	OpenVolume             interface{}    `json:"openVolume"`
	AccountID              string         `json:"accountId"`
	Side                   OrderSide      `json:"side"`
	OrderbookID            string         `json:"orderbookId"`
	ValidUntil             interface{}    `json:"validUntil"`
	Metadata               OrderMetadata  `json:"metadata"`
	Condition              OrderCondition `json:"condition"`
}

// PlaceOrderResponse contains the result of placing an order.
// Check OrderRequestStatus to determine success or failure.
type PlaceOrderResponse struct {
	OrderRequestStatus OrderRequestStatus `json:"orderRequestStatus"`
	Message            string             `json:"message"`
	Parameters         []string           `json:"parameters"`
	OrderID            string             `json:"orderId"`
}

// OrderAccount contains account details associated with an order.
type OrderAccount struct {
	AccountID string `json:"accountId"`
	Name      struct {
		Value string `json:"value"`
	} `json:"name"`
	Type struct {
		AccountType string `json:"accountType"`
	} `json:"type"`
	URLParameterID string `json:"urlParameterId"`
}

// OrderOrderbook contains instrument details for an order.
type OrderOrderbook struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	CountryCode    string `json:"countryCode"`
	Currency       string `json:"currency"`
	InstrumentType string `json:"instrumentType"`
	VolumeFactor   string `json:"volumeFactor"`
	ISIN           string `json:"isin"`
	MIC            string `json:"mic"`
}

// Order represents an active or completed order.
type Order struct {
	Account              OrderAccount           `json:"account"`
	OrderID              string                 `json:"orderId"`
	Volume               int                    `json:"volume"`
	OriginalVolume       int                    `json:"originalVolume"`
	Price                float64                `json:"price"`
	Amount               float64                `json:"amount"`
	OrderbookID          string                 `json:"orderbookId"`
	Side                 OrderSide              `json:"side"`
	ValidUntil           string                 `json:"validUntil"`
	Created              string                 `json:"created"`
	Deletable            bool                   `json:"deletable"`
	Modifiable           bool                   `json:"modifiable"`
	Message              string                 `json:"message"`
	State                string                 `json:"state"`
	StateText            string                 `json:"stateText"`
	StateMessage         string                 `json:"stateMessage"`
	Orderbook            OrderOrderbook         `json:"orderbook"`
	AdditionalParameters map[string]interface{} `json:"additionalParameters"`
	Condition            OrderCondition         `json:"condition"`
}

// GetOrdersResponse contains all orders for the authenticated user.
type GetOrdersResponse struct {
	Orders          []Order       `json:"orders"`
	FundOrders      []interface{} `json:"fundOrders"`
	CancelledOrders []interface{} `json:"cancelledOrders"`
}

// ValidateOrderRequest contains order parameters to validate before placing.
// Use Validate() before sending to ensure all required fields are set.
type ValidateOrderRequest struct {
	IsDividendReinvestment bool           `json:"isDividendReinvestment"`
	RequestID              *string        `json:"requestId"`
	OrderRequestParameters interface{}    `json:"orderRequestParameters"`
	Price                  float64        `json:"price"`
	Volume                 int            `json:"volume"`
	OpenVolume             interface{}    `json:"openVolume"`
	AccountID              string         `json:"accountId"`
	Side                   OrderSide      `json:"side"`
	OrderbookID            string         `json:"orderbookId"`
	ValidUntil             interface{}    `json:"validUntil"`
	Metadata               interface{}    `json:"metadata"`
	Condition              OrderCondition `json:"condition"`
	ISIN                   string         `json:"isin"`
	Currency               string         `json:"currency"`
	MarketPlace            string         `json:"marketPlace"`
}

// ValidateOrderResponse contains validation results for various checks.
// Each field indicates whether that validation passed (Valid=true).
type ValidateOrderResponse struct {
	CommissionWarning      ValidationResult `json:"commissionWarning"`
	EmployeeValidation     ValidationResult `json:"employeeValidation"`
	LargeInScaleWarning    ValidationResult `json:"largeInScaleWarning"`
	OrderValueLimitWarning ValidationResult `json:"orderValueLimitWarning"`
	PriceRampingWarning    ValidationResult `json:"priceRampingWarning"`
	CanadaOddLotWarning    ValidationResult `json:"canadaOddLotWarning"`
}

// ValidationResult indicates whether a validation check passed.
type ValidationResult struct {
	Valid bool `json:"valid"`
}

// PreliminaryFeeRequest contains order parameters to calculate fees.
// Use Validate() before sending to ensure all required fields are set.
type PreliminaryFeeRequest struct {
	AccountID   string    `json:"accountId"`
	OrderbookID string    `json:"orderbookId"`
	Price       string    `json:"price"`
	Volume      string    `json:"volume"`
	Side        OrderSide `json:"side"`
}

// PreliminaryFeeResponse contains fee calculations for an order.
// All monetary values are strings in the orderbook currency.
type PreliminaryFeeResponse struct {
	Commission          string              `json:"commission"`
	MarketFees          string              `json:"marketFees"`
	TotalFees           string              `json:"totalFees"`
	TotalSum            string              `json:"totalSum"`
	TotalSumWithoutFees string              `json:"totalSumWithoutFees"`
	OrderbookCurrency   string              `json:"orderbookCurrency"`
	TransactionTax      *string             `json:"transactionTax"`
	CurrencyExchangeFee CurrencyExchangeFee `json:"currencyExchangeFee"`
	Campaign            *string             `json:"campaign"`
}

// CurrencyExchangeFee contains exchange rate and fee for currency conversion.
type CurrencyExchangeFee struct {
	Rate string `json:"rate"`
	Sum  string `json:"sum"`
}

// StopLossTriggerType determines when the stop loss triggers.
type StopLossTriggerType string

const (
	StopLossTriggerLessOrEqual    StopLossTriggerType = "LESS_OR_EQUAL"    // Trigger when price drops to or below value
	StopLossTriggerGreaterOrEqual StopLossTriggerType = "GREATER_OR_EQUAL" // Trigger when price rises to or above value
)

// StopLossValueType specifies how the trigger value is interpreted.
type StopLossValueType string

const (
	StopLossValueMonetary   StopLossValueType = "MONETARY"   // Value is an absolute price
	StopLossValuePercentage StopLossValueType = "PERCENTAGE" // Value is a percentage change
)

// StopLossOrderEventType indicates what action to take when triggered.
type StopLossOrderEventType string

const (
	StopLossOrderEventBuy  StopLossOrderEventType = "BUY"  // Place a buy order when triggered
	StopLossOrderEventSell StopLossOrderEventType = "SELL" // Place a sell order when triggered
)

// StopLossPriceType specifies how the order price is determined.
type StopLossPriceType string

const (
	StopLossPriceMonetary   StopLossPriceType = "MONETARY"   // Price is an absolute value
	StopLossPricePercentage StopLossPriceType = "PERCENTAGE" // Price is a percentage of current price
)

// StopLossTrigger defines when the stop loss order should activate.
type StopLossTrigger struct {
	Type                      StopLossTriggerType `json:"type"`
	Value                     float64             `json:"value"`
	ValueType                 StopLossValueType   `json:"valueType"`
	ValidUntil                string              `json:"validUntil"`
	TriggerOnMarketMakerQuote bool                `json:"triggerOnMarketMakerQuote"`
}

// StopLossOrderEvent defines the order to place when the trigger activates.
type StopLossOrderEvent struct {
	Type                StopLossOrderEventType `json:"type"`
	Price               float64                `json:"price"`
	Volume              int                    `json:"volume"`
	ValidDays           int                    `json:"validDays"`
	PriceType           StopLossPriceType      `json:"priceType"`
	ShortSellingAllowed bool                   `json:"shortSellingAllowed"`
}

// PlaceStopLossRequest contains all parameters needed to place a stop loss order.
// Use Validate() before sending to ensure all required fields are set.
type PlaceStopLossRequest struct {
	ParentStopLossID   string             `json:"parentStopLossId"`
	AccountID          string             `json:"accountId"`
	OrderbookID        string             `json:"orderbookId"`
	StopLossTrigger    StopLossTrigger    `json:"stopLossTrigger"`
	StopLossOrderEvent StopLossOrderEvent `json:"stopLossOrderEvent"`
}

// PlaceStopLossResponse contains the result of placing a stop loss order.
// Check Status to determine success or failure.
type PlaceStopLossResponse struct {
	Status          StopLossStatus `json:"status"`
	StopLossOrderID string         `json:"stoplossOrderId"`
}

// StopLossAccount contains account details for a stop loss order.
type StopLossAccount struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	URLParameterID string `json:"urlParameterId"`
}

// StopLossOrderbook contains instrument details for a stop loss order.
type StopLossOrderbook struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	CountryCode              string `json:"countryCode"`
	Currency                 string `json:"currency"`
	ShortName                string `json:"shortName"`
	Type                     string `json:"type"`
	StoplossMarketMakerQuote bool   `json:"stoplossMarketMakerQuote"`
}

// StopLossTriggerResponse contains trigger configuration from the API.
type StopLossTriggerResponse struct {
	Value                     float64             `json:"value"`
	Type                      StopLossTriggerType `json:"type"`
	ValidUntil                string              `json:"validUntil"`
	ValueType                 StopLossValueType   `json:"valueType"`
	TriggerOnMarketMakerQuote bool                `json:"triggerOnMarketMakerQuote"`
}

// StopLossOrderDetails contains the order configuration for a stop loss.
type StopLossOrderDetails struct {
	Type                  StopLossOrderEventType `json:"type"`
	Price                 float64                `json:"price"`
	Volume                int                    `json:"volume"`
	ShortSellingAllowed   bool                   `json:"shortSellingAllowed"`
	ValidDays             int                    `json:"validDays"`
	PriceType             StopLossPriceType      `json:"priceType"`
	PriceDecimalPrecision int                    `json:"priceDecimalPrecision"`
}

// StopLossOrder represents an active stop loss order.
type StopLossOrder struct {
	ID        string                  `json:"id"`
	Status    StopLossStatus          `json:"status"`
	Account   StopLossAccount         `json:"account"`
	Orderbook StopLossOrderbook       `json:"orderbook"`
	Message   string                  `json:"message"`
	Trigger   StopLossTriggerResponse `json:"trigger"`
	Order     StopLossOrderDetails    `json:"order"`
	Editable  bool                    `json:"editable"`
	Deletable bool                    `json:"deletable"`
}

// OrderAction indicates the type of order event.
type OrderAction string

const (
	OrderActionNew     OrderAction = "NEW"     // New order created
	OrderActionDeleted OrderAction = "DELETED" // Order deleted/cancelled
)

// OrderStateName indicates the current state of an order.
type OrderStateName string

const (
	OrderStateActivePending OrderStateName = "ACTIVE_PENDING" // Order pending market open
	OrderStateDeleted       OrderStateName = "DELETED"        // Order has been deleted
)

// OrderEventOrderbook contains instrument details in an order event.
type OrderEventOrderbook struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	TickerSymbol    string `json:"tickerSymbol"`
	MarketplaceName string `json:"marketplaceName"`
	CountryCode     string `json:"countryCode"`
	InstrumentType  string `json:"instrumentType"`
	Tradable        bool   `json:"tradable"`
	VolumeFactor    int    `json:"volumeFactor"`
	CurrencyCode    string `json:"currencyCode"`
	FlagCode        string `json:"flagCode"`
}

// OrderEventState contains order state information.
type OrderEventState struct {
	Value       string         `json:"value"`
	Description string         `json:"description"`
	Name        OrderStateName `json:"name"`
}

// OrderEventData contains order data from an SSE event.
type OrderEventData struct {
	ID                   string               `json:"id"`
	AccountID            string               `json:"accountId"`
	Orderbook            OrderEventOrderbook  `json:"orderbook"`
	CurrentVolume        float64              `json:"currentVolume"`
	OriginalVolume       float64              `json:"originalVolume"`
	OpenVolume           *float64             `json:"openVolume"`
	Price                float64              `json:"price"`
	ValidDate            *string              `json:"validDate"`
	Type                 OrderSide            `json:"type"`
	State                OrderEventState      `json:"state"`
	Action               OrderAction          `json:"action"`
	Modifiable           bool                 `json:"modifiable"`
	Deletable            bool                 `json:"deletable"`
	Sum                  float64              `json:"sum"`
	VisibleDate          *string              `json:"visibleDate"`
	OrderDateTime        int64                `json:"orderDateTime"`
	EventTimeStamp       int64                `json:"eventTimeStamp"`
	UniqueID             string               `json:"uniqueId"`
	AdditionalParameters map[string]any       `json:"additionalParameters"`
	DetailedCancelStatus *string              `json:"detailedCancelStatus"`
	Condition            OrderCondition       `json:"condition"`
}

// OrderEvent is a single event from the orders subscription stream.
type OrderEvent struct {
	Event string         `json:"event"`
	Data  OrderEventData `json:"data"`
	ID    string         `json:"id"`
	Retry int            `json:"retry"`
}
