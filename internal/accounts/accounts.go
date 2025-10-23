// Package accounts provides account overview functionality for the Avanza API.
package accounts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/vmorsell/avanza/internal/client"
)

// AccountsService handles account-related operations with Avanza.
type AccountsService struct {
	client *client.Client
}

// NewAccountsService creates a new accounts service with the given HTTP client.
func NewAccountsService(client *client.Client) *AccountsService {
	return &AccountsService{
		client: client,
	}
}

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
func (a *AccountsService) GetAccountOverview(ctx context.Context) (*AccountOverview, error) {
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
