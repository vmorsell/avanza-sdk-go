// Package accounts provides account management functionality for the Avanza API.
package accounts

// AccountOverview contains all accounts, categorized and with loans.
type AccountOverview struct {
	Categories []Category `json:"categories"`
	Accounts   []Account  `json:"accounts"`
	Loans      []Loan     `json:"loans"`
}

// Category groups accounts by type (e.g., SPARANDE, BUFFERT).
type Category struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	TotalValue      Money       `json:"totalValue"`
	Performance     Performance `json:"performance"`
	SavingsGoalView interface{} `json:"savingsGoalView"`
}

// Account represents a single account (ISK, KF, AF, etc.).
type Account struct {
	ID                       string          `json:"id"`
	CategoryID               string          `json:"categoryId"`
	Balance                  Money           `json:"balance"`
	Profit                   Profit          `json:"profit"`
	Type                     string          `json:"type"`
	TotalValue               Money           `json:"totalValue"`
	BuyingPower              Money           `json:"buyingPower"`
	BuyingPowerWithoutCredit Money           `json:"buyingPowerWithoutCredit"`
	DepositInterestRate      Money           `json:"depositInterestRate"`
	LoanInterestRate         Money           `json:"loanInterestRate"`
	Credit                   *Money          `json:"credit"`
	Name                     AccountName     `json:"name"`
	Status                   string          `json:"status"`
	ErrorStatus              string          `json:"errorStatus"`
	Overmortgaged            interface{}     `json:"overmortgaged"`
	CurrencyBalances         []Money         `json:"currencyBalances"`
	Overdrawn                []interface{}   `json:"overdrawn"`
	Performance              Performance     `json:"performance"`
	Settings                 AccountSettings `json:"settings"`
	ClearingAccountNumber    string          `json:"clearingAccountNumber"`
	AccountType24            bool            `json:"accountType24"`
	DiscretionaryPortfolio   bool            `json:"discretionaryPortfolio"`
	URLParameterID           string          `json:"urlParameterId"`
	Owner                    bool            `json:"owner"`
}

// AccountName contains both the default and user-defined account name.
type AccountName struct {
	DefaultName     string `json:"defaultName"`
	UserDefinedName string `json:"userDefinedName"`
}

// AccountSettings contains account configuration.
type AccountSettings struct {
	IsHidden bool `json:"IS_HIDDEN"`
}

// Money represents a monetary value with currency and precision.
// Unit is typically a currency code (e.g., "SEK", "USD").
// DecimalPrecision indicates the number of decimal places for display.
type Money struct {
	Value            float64 `json:"value"`
	Unit             string  `json:"unit"`
	UnitType         string  `json:"unitType"`
	DecimalPrecision int     `json:"decimalPrecision"`
}

// Profit contains both absolute and relative profit values.
type Profit struct {
	Absolute Money `json:"absolute"`
	Relative Money `json:"relative"`
}

// Performance contains performance metrics for various time periods.
// Fields may be nil if data is not available for that period.
type Performance struct {
	OneWeek     *PerformanceData `json:"ONE_WEEK,omitempty"`
	ThisYear    *PerformanceData `json:"THIS_YEAR,omitempty"`
	OneMonth    *PerformanceData `json:"ONE_MONTH,omitempty"`
	ThreeMonths *PerformanceData `json:"THREE_MONTHS,omitempty"`
	OneYear     *PerformanceData `json:"ONE_YEAR,omitempty"`
	ThreeYears  *PerformanceData `json:"THREE_YEARS,omitempty"`
	AllTime     *PerformanceData `json:"ALL_TIME,omitempty"`
}

// PerformanceData contains absolute and relative performance for a time period.
type PerformanceData struct {
	Absolute Money `json:"absolute"`
	Relative Money `json:"relative"`
}

// Loan represents a loan account.
type Loan struct{}

// TradingAccount represents a trading account with balances and capabilities.
type TradingAccount struct {
	Name                              string            `json:"name"`
	AccountID                         string            `json:"accountId"`
	AccountTypeName                   string            `json:"accountTypeName"`
	AccountType                       string            `json:"accountType"`
	AvailableForPurchase              float64           `json:"availableForPurchase"`
	AvailableForPurchaseWithoutCredit float64           `json:"availableForPurchaseWithoutCredit"`
	AvailableCredit                   float64           `json:"availableCredit"`
	HasCredit                         bool              `json:"hasCredit"`
	IsTradable                        bool              `json:"isTradable"`
	IsShortSellable                   bool              `json:"isShortSellable"`
	IsOvermortgaged                   bool              `json:"isOvermortgaged"`
	IsOverdrawn                       bool              `json:"isOverdrawn"`
	IsHidden                          bool              `json:"isHidden"`
	Positions                         []interface{}     `json:"positions"`
	CurrencyBalances                  []CurrencyBalance `json:"currencyBalances"`
	URLParameterID                    string            `json:"urlParameterId"`
}

// CurrencyBalance contains the balance for a specific currency.
type CurrencyBalance struct {
	Currency    string  `json:"currency"`
	CountryCode string  `json:"countryCode"`
	Balance     float64 `json:"balance"`
}

// AccountPosition represents a holding (stock, fund, etc.) in an account.
type AccountPosition struct {
	Account                                AccountInfo         `json:"account"`
	Instrument                             Instrument          `json:"instrument"`
	LastTradingDayPerformance              PositionPerformance `json:"lastTradingDayPerformance"`
	ID                                     string              `json:"id"`
	SuperInterestApproved                  bool                `json:"superInterestApproved"`
	Volume                                 Money               `json:"volume"`
	Value                                  Money               `json:"value"`
	AverageAcquiredPrice                   Money               `json:"averageAcquiredPrice"`
	AverageAcquiredPriceInstrumentCurrency Money               `json:"averageAcquiredPriceInstrumentCurrency"`
	AcquiredValue                          Money               `json:"acquiredValue"`
	CollateralFactor                       Money               `json:"collateralFactor"`
}

// AccountInfo contains account details used in positions.
type AccountInfo struct {
	ID                  string `json:"id"`
	Type                string `json:"type"`
	Name                string `json:"name"`
	URLParameterID      string `json:"urlParameterId"`
	HasCredit           bool   `json:"hasCredit"`
	HasAutoDistribution bool   `json:"hasAutoDistribution"`
}

// Instrument represents a financial instrument (stock, fund, etc.).
type Instrument struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Name         string    `json:"name"`
	Orderbook    Orderbook `json:"orderbook"`
	Currency     string    `json:"currency"`
	ISIN         string    `json:"isin"`
	VolumeFactor float64   `json:"volumeFactor"`
}

// Orderbook contains market data for an instrument.
type Orderbook struct {
	ID          string   `json:"id"`
	FlagCode    string   `json:"flagCode"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	TradeStatus string   `json:"tradeStatus"`
	Quote       Quote    `json:"quote"`
	Turnover    Turnover `json:"turnover"`
	LastDeal    LastDeal `json:"lastDeal"`
}

// Quote contains current bid/ask prices and latest trade information.
type Quote struct {
	Highest       Money  `json:"highest"`
	Lowest        Money  `json:"lowest"`
	Buy           Money  `json:"buy"`
	Sell          Money  `json:"sell"`
	Latest        Money  `json:"latest"`
	Change        Money  `json:"change"`
	ChangePercent Money  `json:"changePercent"`
	Updated       string `json:"updated"`
}

// Turnover contains trading volume and value for a period.
type Turnover struct {
	Volume Money `json:"volume"`
	Value  Money `json:"value"`
}

// LastDeal contains the timestamp of the last trade.
type LastDeal struct {
	Date string `json:"date"`
	Time string `json:"time"`
}

// PositionPerformance contains profit/loss for a position.
type PositionPerformance struct {
	Absolute Money `json:"absolute"`
	Relative Money `json:"relative"`
}

// CashPosition represents uninvested cash in an account.
type CashPosition struct {
	Account      AccountInfo `json:"account"`
	TotalBalance Money       `json:"totalBalance"`
	ID           string      `json:"id"`
}

// AccountPositions contains all positions for an account.
type AccountPositions struct {
	WithOrderbook     []AccountPosition `json:"withOrderbook"`
	WithoutOrderbook  []interface{}     `json:"withoutOrderbook"`
	CashPositions     []CashPosition    `json:"cashPositions"`
	WithCreditAccount bool              `json:"withCreditAccount"`
}
