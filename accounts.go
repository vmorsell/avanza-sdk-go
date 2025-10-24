package avanza

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// AccountOverview represents the complete account overview response.
type AccountOverview struct {
	Categories []Category `json:"categories"`
	Accounts   []Account  `json:"accounts"`
	Loans      []Loan     `json:"loans"`
}

// Category represents an account category (e.g., SPARANDE, BUFFERT).
type Category struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	TotalValue      Money       `json:"totalValue"`
	Performance     Performance `json:"performance"`
	SavingsGoalView interface{} `json:"savingsGoalView"`
}

// Account represents a single account.
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

// AccountName represents the account name structure.
type AccountName struct {
	DefaultName     string `json:"defaultName"`
	UserDefinedName string `json:"userDefinedName"`
}

// AccountSettings represents account settings.
type AccountSettings struct {
	IsHidden bool `json:"IS_HIDDEN"`
}

// Money represents a monetary value with precision.
type Money struct {
	Value            float64 `json:"value"`
	Unit             string  `json:"unit"`
	UnitType         string  `json:"unitType"`
	DecimalPrecision int     `json:"decimalPrecision"`
}

// Profit represents profit information.
type Profit struct {
	Absolute Money `json:"absolute"`
	Relative Money `json:"relative"`
}

// Performance represents performance data for different time periods.
type Performance struct {
	OneWeek     *PerformanceData `json:"ONE_WEEK,omitempty"`
	ThisYear    *PerformanceData `json:"THIS_YEAR,omitempty"`
	OneMonth    *PerformanceData `json:"ONE_MONTH,omitempty"`
	ThreeMonths *PerformanceData `json:"THREE_MONTHS,omitempty"`
	OneYear     *PerformanceData `json:"ONE_YEAR,omitempty"`
	ThreeYears  *PerformanceData `json:"THREE_YEARS,omitempty"`
	AllTime     *PerformanceData `json:"ALL_TIME,omitempty"`
}

// PerformanceData represents performance data for a specific time period.
type PerformanceData struct {
	Absolute Money `json:"absolute"`
	Relative Money `json:"relative"`
}

// Loan represents a loan (currently empty in the API response).
type Loan struct{}

// GetAccountOverview retrieves the complete account overview including categories, accounts, and loans.
func (a *Avanza) GetAccountOverview(ctx context.Context) (*AccountOverview, error) {
	resp, err := a.client.Get(ctx, "/_api/account-overview/overview/categorizedAccounts")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var overview AccountOverview
	if err := json.NewDecoder(resp.Body).Decode(&overview); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &overview, nil
}

// TradingAccount represents a trading account with its details.
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

// CurrencyBalance represents the balance for a specific currency.
type CurrencyBalance struct {
	Currency    string  `json:"currency"`
	CountryCode string  `json:"countryCode"`
	Balance     float64 `json:"balance"`
}

// GetTradingAccounts retrieves all trading accounts for the authenticated user.
func (a *Avanza) GetTradingAccounts(ctx context.Context) ([]TradingAccount, error) {
	resp, err := a.client.Get(ctx, "/_api/trading-critical/rest/accounts")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("trading accounts request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var accounts []TradingAccount
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, fmt.Errorf("failed to decode trading accounts: %w", err)
	}

	return accounts, nil
}

// AccountPosition represents a position in an account.
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

// AccountInfo represents account information in a position.
type AccountInfo struct {
	ID                  string `json:"id"`
	Type                string `json:"type"`
	Name                string `json:"name"`
	URLParameterID      string `json:"urlParameterId"`
	HasCredit           bool   `json:"hasCredit"`
	HasAutoDistribution bool   `json:"hasAutoDistribution"`
}

// Instrument represents an instrument in a position.
type Instrument struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Name         string    `json:"name"`
	Orderbook    Orderbook `json:"orderbook"`
	Currency     string    `json:"currency"`
	ISIN         string    `json:"isin"`
	VolumeFactor float64   `json:"volumeFactor"`
}

// Orderbook represents orderbook information for an instrument.
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

// Quote represents quote information for an instrument.
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

// Turnover represents turnover information.
type Turnover struct {
	Volume Money `json:"volume"`
	Value  Money `json:"value"`
}

// LastDeal represents the last deal information.
type LastDeal struct {
	Date string `json:"date"`
	Time string `json:"time"`
}

// PositionPerformance represents performance information for a position.
type PositionPerformance struct {
	Absolute Money `json:"absolute"`
	Relative Money `json:"relative"`
}

// CashPosition represents a cash position in an account.
type CashPosition struct {
	Account      AccountInfo `json:"account"`
	TotalBalance Money       `json:"totalBalance"`
	ID           string      `json:"id"`
}

// AccountPositions represents the positions for a specific account.
type AccountPositions struct {
	WithOrderbook     []AccountPosition `json:"withOrderbook"`
	WithoutOrderbook  []interface{}     `json:"withoutOrderbook"`
	CashPositions     []CashPosition    `json:"cashPositions"`
	WithCreditAccount bool              `json:"withCreditAccount"`
}

// GetAccountPositions retrieves positions for a specific account using its URL parameter ID.
func (a *Avanza) GetAccountPositions(ctx context.Context, urlParameterID string) (*AccountPositions, error) {
	endpoint := fmt.Sprintf("/_api/position-data/positions/%s", urlParameterID)

	resp, err := a.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("account positions request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var positions AccountPositions
	if err := json.NewDecoder(resp.Body).Decode(&positions); err != nil {
		return nil, fmt.Errorf("failed to decode account positions: %w", err)
	}

	return &positions, nil
}
