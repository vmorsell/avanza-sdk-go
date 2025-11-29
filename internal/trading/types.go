// Package trading provides trading functionality for the Avanza API.
package trading

// OrderSide represents the side of an order (buy or sell).
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderCondition represents the condition type for an order.
type OrderCondition string

const (
	OrderConditionNormal     OrderCondition = "NORMAL"
	OrderConditionFillOrKill OrderCondition = "FILL_OR_KILL"
)

// OrderRequestStatus represents the status of an order request.
type OrderRequestStatus string

const (
	OrderRequestStatusSuccess OrderRequestStatus = "SUCCESS"
	OrderRequestStatusError   OrderRequestStatus = "ERROR"
)

// StopLossStatus represents the status of a stop loss order.
type StopLossStatus string

const (
	StopLossStatusSuccess StopLossStatus = "SUCCESS"
	StopLossStatusError   StopLossStatus = "ERROR"
)

// OrderMetadata contains additional metadata about the order.
type OrderMetadata struct {
	OrderEntryMode  string `json:"orderEntryMode"`
	HasTouchedPrice string `json:"hasTouchedPrice"`
}

// PlaceOrderRequest represents a request to place a new order.
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

// PlaceOrderResponse represents the response from placing an order.
type PlaceOrderResponse struct {
	OrderRequestStatus OrderRequestStatus `json:"orderRequestStatus"`
	Message            string             `json:"message"`
	Parameters         []string           `json:"parameters"`
	OrderID            string             `json:"orderId"`
}

// OrderAccount represents account information for an order.
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

// OrderOrderbook represents orderbook information for an order.
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

// Order represents a single order.
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

// GetOrdersResponse represents the response from getting orders.
type GetOrdersResponse struct {
	Orders          []Order       `json:"orders"`
	FundOrders      []interface{} `json:"fundOrders"`
	CancelledOrders []interface{} `json:"cancelledOrders"`
}

// ValidateOrderRequest represents a request to validate an order before placing it.
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

// ValidateOrderResponse represents the response from order validation.
type ValidateOrderResponse struct {
	CommissionWarning      ValidationResult `json:"commissionWarning"`
	EmployeeValidation     ValidationResult `json:"employeeValidation"`
	LargeInScaleWarning    ValidationResult `json:"largeInScaleWarning"`
	OrderValueLimitWarning ValidationResult `json:"orderValueLimitWarning"`
	PriceRampingWarning    ValidationResult `json:"priceRampingWarning"`
	CanadaOddLotWarning    ValidationResult `json:"canadaOddLotWarning"`
}

// ValidationResult represents the result of a validation check.
type ValidationResult struct {
	Valid bool `json:"valid"`
}

// PreliminaryFeeRequest represents a request to get preliminary fees for an order.
type PreliminaryFeeRequest struct {
	AccountID   string    `json:"accountId"`
	OrderbookID string    `json:"orderbookId"`
	Price       string    `json:"price"`
	Volume      string    `json:"volume"`
	Side        OrderSide `json:"side"`
}

// PreliminaryFeeResponse represents the response from getting preliminary fees.
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

// CurrencyExchangeFee represents currency exchange fee information.
type CurrencyExchangeFee struct {
	Rate string `json:"rate"`
	Sum  string `json:"sum"`
}

// StopLossTriggerType represents the type of stop loss trigger.
type StopLossTriggerType string

const (
	StopLossTriggerLessOrEqual    StopLossTriggerType = "LESS_OR_EQUAL"
	StopLossTriggerGreaterOrEqual StopLossTriggerType = "GREATER_OR_EQUAL"
)

// StopLossValueType represents the type of stop loss value.
type StopLossValueType string

const (
	StopLossValueMonetary   StopLossValueType = "MONETARY"
	StopLossValuePercentage StopLossValueType = "PERCENTAGE"
)

// StopLossOrderEventType represents the type of stop loss order event.
type StopLossOrderEventType string

const (
	StopLossOrderEventBuy  StopLossOrderEventType = "BUY"
	StopLossOrderEventSell StopLossOrderEventType = "SELL"
)

// StopLossPriceType represents the type of stop loss price.
type StopLossPriceType string

const (
	StopLossPriceMonetary   StopLossPriceType = "MONETARY"
	StopLossPricePercentage StopLossPriceType = "PERCENTAGE"
)

// StopLossTrigger represents the trigger conditions for a stop loss order.
type StopLossTrigger struct {
	Type                      StopLossTriggerType `json:"type"`
	Value                     float64             `json:"value"`
	ValueType                 StopLossValueType   `json:"valueType"`
	ValidUntil                string              `json:"validUntil"`
	TriggerOnMarketMakerQuote bool                `json:"triggerOnMarketMakerQuote"`
}

// StopLossOrderEvent represents the order event that will be triggered.
type StopLossOrderEvent struct {
	Type                StopLossOrderEventType `json:"type"`
	Price               float64                `json:"price"`
	Volume              int                    `json:"volume"`
	ValidDays           int                    `json:"validDays"`
	PriceType           StopLossPriceType      `json:"priceType"`
	ShortSellingAllowed bool                   `json:"shortSellingAllowed"`
}

// PlaceStopLossRequest represents a request to place a stop loss order.
type PlaceStopLossRequest struct {
	ParentStopLossID   string             `json:"parentStopLossId"`
	AccountID          string             `json:"accountId"`
	OrderBookID        string             `json:"orderBookId"`
	StopLossTrigger    StopLossTrigger    `json:"stopLossTrigger"`
	StopLossOrderEvent StopLossOrderEvent `json:"stopLossOrderEvent"`
}

// PlaceStopLossResponse represents the response from placing a stop loss order.
type PlaceStopLossResponse struct {
	Status          StopLossStatus `json:"status"`
	StopLossOrderID string         `json:"stoplossOrderId"`
}

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
	Value                     float64             `json:"value"`
	Type                      StopLossTriggerType `json:"type"`
	ValidUntil                string              `json:"validUntil"`
	ValueType                 StopLossValueType   `json:"valueType"`
	TriggerOnMarketMakerQuote bool                `json:"triggerOnMarketMakerQuote"`
}

// StopLossOrderDetails represents order details for a stop loss order.
type StopLossOrderDetails struct {
	Type                  StopLossOrderEventType `json:"type"`
	Price                 float64                `json:"price"`
	Volume                int                    `json:"volume"`
	ShortSellingAllowed   bool                   `json:"shortSellingAllowed"`
	ValidDays             int                    `json:"validDays"`
	PriceType             StopLossPriceType      `json:"priceType"`
	PriceDecimalPrecision int                    `json:"priceDecimalPrecision"`
}

// StopLossOrder represents a single stop loss order.
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
