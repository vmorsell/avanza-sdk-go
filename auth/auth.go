// Package auth provides BankID authentication functionality for the Avanza API.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/vmorsell/avanza-sdk-go/client"
)

// AuthService handles BankID authentication.
type AuthService struct {
	client *client.Client
}

// NewAuthService creates a new authentication service.
func NewAuthService(client *client.Client) *AuthService {
	return &AuthService{
		client: client,
	}
}

// BankIDStartRequest initiates a BankID authentication session.
type BankIDStartRequest struct {
	Method       string `json:"method"`
	ReturnScheme string `json:"returnScheme"`
}

// BankIDStartResponse contains the QR token and transaction details.
type BankIDStartResponse struct {
	TransactionID string `json:"transactionId"`
	Expires       string `json:"expires"`
	QRToken       string `json:"qrToken"`
}

// BankIDCollectResponse contains authentication status.
// State is "PENDING", "COMPLETE", or "FAILED". Logins is populated when State is "COMPLETE".
type BankIDCollectResponse struct {
	Name                       string        `json:"name"`
	TransactionID              string        `json:"transactionId"`
	State                      string        `json:"state"`
	HintCode                   string        `json:"hintCode"`
	IdentificationNumber       string        `json:"identificationNumber"`
	Logins                     []Login       `json:"logins"`
	RecommendedTargetCustomers []interface{} `json:"recommendedTargetCustomers"`
	Poa                        []interface{} `json:"poa"`
}

// Login represents a user account available after authentication completes.
type Login struct {
	CustomerID string    `json:"customerId"`
	Username   string    `json:"username"`
	Accounts   []Account `json:"accounts"`
	LoginPath  string    `json:"loginPath"`
}

// Account represents an Avanza account type (ISK, KF, AF, etc.).
type Account struct {
	AccountName string `json:"accountName"`
	AccountType string `json:"accountType"`
}

// BankIDRestartRequest refreshes an expiring QR code.
type BankIDRestartRequest struct{}

// StartBankID initiates a BankID authentication session. Returns a QR token.
// For automatic QR refresh, use PollBankIDWithQRUpdates.
func (a *AuthService) StartBankID(ctx context.Context) (*BankIDStartResponse, error) {
	// Get initial cookies (AZAPERSISTENCE, etc.)
	initResp, err := a.client.Get(ctx, "/")
	if err != nil {
		return nil, fmt.Errorf("failed to get initial cookies: %w", err)
	}
	_ = initResp.Body.Close()

	reqBody := BankIDStartRequest{
		Method:       "QR_START",
		ReturnScheme: "NULL",
	}

	resp, err := a.client.Post(ctx, "/_api/authentication/v2/sessions/bankid", reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, client.NewHTTPError(resp)
	}

	var response BankIDStartResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// RestartBankID refreshes the BankID session with a new QR code.
func (a *AuthService) RestartBankID(ctx context.Context) (*BankIDStartResponse, error) {
	resp, err := a.client.Post(ctx, "/_api/authentication/v2/sessions/bankid/restart", BankIDRestartRequest{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, client.NewHTTPError(resp)
	}

	var response BankIDStartResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// CollectBankID checks the BankID authentication status.
func (a *AuthService) CollectBankID(ctx context.Context) (*BankIDCollectResponse, error) {
	resp, err := a.client.Post(ctx, "/_api/authentication/v2/sessions/bankid/collect", BankIDRestartRequest{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, client.NewHTTPError(resp)
	}

	var response BankIDCollectResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// PollBankID polls authentication status every second until completion or failure.
func (a *AuthService) PollBankID(ctx context.Context) (*BankIDCollectResponse, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			collectResp, err := a.CollectBankID(ctx)
			if err != nil {
				return nil, err
			}

			if collectResp.State == "COMPLETE" {
				return collectResp, nil
			}

			if collectResp.State == "FAILED" {
				return nil, fmt.Errorf("bankid authentication failed: %s", collectResp.HintCode)
			}
		}
	}
}

// PollBankIDWithQRUpdates polls authentication and refreshes the QR code every second.
// Recommended for QR-based authentication.
func (a *AuthService) PollBankIDWithQRUpdates(ctx context.Context) (*BankIDCollectResponse, error) {
	qrCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-qrCtx.Done():
				return
			case <-ticker.C:
				restartResp, err := a.RestartBankID(qrCtx)
				if err != nil {
					continue
				}
				_ = a.DisplayQRCode(restartResp.QRToken)
			}
		}
	}()

	return a.PollBankID(ctx)
}

// ClearScreen clears the terminal using ANSI escape codes.
func (a *AuthService) ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// DisplayQRCode renders a QR code in the terminal. Clears the screen first.
func (a *AuthService) DisplayQRCode(qrCodeData string) error {
	if qrCodeData == "" {
		return fmt.Errorf("empty qr code data")
	}

	a.ClearScreen()
	fmt.Println("Scan QR code with BankID app to authenticate to Avanza...")
	qrterminal.GenerateHalfBlock(qrCodeData, qrterminal.L, os.Stdout)
	return nil
}

// EstablishSession establishes a session after BankID authentication.
// Required before making other API calls.
func (a *AuthService) EstablishSession(ctx context.Context, collectResp *BankIDCollectResponse) error {
	if collectResp == nil || len(collectResp.Logins) == 0 {
		return fmt.Errorf("no logins available in authentication response")
	}

	login := collectResp.Logins[0]
	userEndpoint := fmt.Sprintf("/_api/authentication/v2/sessions/bankid/collect/%s", url.PathEscape(login.CustomerID))

	resp, err := a.client.Get(ctx, userEndpoint)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("select user: %w", client.NewHTTPError(resp))
	}

	// Get additional session cookies
	tradingResp, err := a.client.Get(ctx, "/handla/order.html")
	if err != nil {
		return fmt.Errorf("failed to visit trading page: %w", err)
	}
	defer tradingResp.Body.Close()

	// Verify session is active
	sessionResp, err := a.client.Get(ctx, "/_api/authentication/session/info/session")
	if err != nil {
		return fmt.Errorf("failed to verify session: %w", err)
	}
	defer sessionResp.Body.Close()

	if sessionResp.StatusCode != http.StatusOK {
		return fmt.Errorf("verify session: %w", client.NewHTTPError(sessionResp))
	}

	return nil
}

// SessionInfo contains the current session state and user details.
type SessionInfo struct {
	InvalidSessionID string `json:"invalidSessionId"`
	User             User   `json:"user"`
}

// User contains authenticated user information.
type User struct {
	LoggedIn           bool   `json:"loggedIn"`
	GreetingName       string `json:"greetingName"`
	PushSubscriptionID string `json:"pushSubscriptionId"`
	PushBaseURL        string `json:"pushBaseUrl"`
	SecurityToken      string `json:"securityToken"`
	Company            bool   `json:"company"`
	Minor              bool   `json:"minor"`
	Start              bool   `json:"start"`
	CustomerGroup      string `json:"customerGroup"`
	ID                 string `json:"id"`
}

// GetSessionInfo returns the current session state.
func (a *AuthService) GetSessionInfo(ctx context.Context) (*SessionInfo, error) {
	resp, err := a.client.Get(ctx, "/_api/authentication/session/info/session")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(resp)
	}

	var sessionInfo SessionInfo
	if err := json.NewDecoder(resp.Body).Decode(&sessionInfo); err != nil {
		return nil, fmt.Errorf("failed to decode session info: %w", err)
	}

	return &sessionInfo, nil
}
