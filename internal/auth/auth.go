// Package auth provides BankID authentication functionality for the Avanza API.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/vmorsell/avanza-sdk-go/internal/client"
)

// AuthService handles authentication operations with Avanza using BankID.
type AuthService struct {
	client *client.Client
}

// NewAuthService creates a new authentication service with the given HTTP client.
func NewAuthService(client *client.Client) *AuthService {
	return &AuthService{
		client: client,
	}
}

// BankIDStartRequest is sent to initiate a BankID authentication session.
type BankIDStartRequest struct {
	Method       string `json:"method"`
	ReturnScheme string `json:"returnScheme"`
}

// BankIDStartResponse contains the QR token and transaction details.
// Use QRToken with DisplayQRCode() to show the QR code to the user.
type BankIDStartResponse struct {
	TransactionID string `json:"transactionId"`
	Expires       string `json:"expires"`
	QRToken       string `json:"qrToken"`
}

// BankIDCollectResponse contains authentication status and user information.
// State can be "PENDING", "COMPLETE", or "FAILED".
// When State is "COMPLETE", Logins contains available user accounts.
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

// Login represents a user account available after authentication.
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

// BankIDRestartRequest is sent to refresh an expiring QR code.
type BankIDRestartRequest struct{}

// StartBankID initiates a new BankID authentication session with QR code support.
// Returns transaction details including a QR token that can be displayed to the user.
//
// For automatic QR refresh, use PollBankIDWithQRUpdates instead.
//
// See also: PollBankIDWithQRUpdates, DisplayQRCode
func (a *AuthService) StartBankID(ctx context.Context) (*BankIDStartResponse, error) {
	// Get initial cookies (AZAPERSISTENCE, etc.)
	_, err := a.client.Get(ctx, "/")
	if err != nil {
		return nil, fmt.Errorf("failed to get initial cookies: %w", err)
	}

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
		return nil, fmt.Errorf("start bankid: %w", client.NewHTTPError(resp))
	}

	var response BankIDStartResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// RestartBankID refreshes an existing BankID session, generating a new QR code.
// This prevents the QR code from expiring during the authentication process.
func (a *AuthService) RestartBankID(ctx context.Context) (*BankIDStartResponse, error) {
	resp, err := a.client.Post(ctx, "/_api/authentication/v2/sessions/bankid/restart", BankIDRestartRequest{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("start bankid: %w", client.NewHTTPError(resp))
	}

	var response BankIDStartResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// CollectBankID checks the current status of the BankID authentication.
// Returns the current state and user information if authentication is complete.
func (a *AuthService) CollectBankID(ctx context.Context) (*BankIDCollectResponse, error) {
	resp, err := a.client.Post(ctx, "/_api/authentication/v2/sessions/bankid/collect", BankIDRestartRequest{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("start bankid: %w", client.NewHTTPError(resp))
	}

	var response BankIDCollectResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// PollBankID continuously polls the authentication status until completion or failure.
// Checks every second until the context is cancelled or authentication completes.
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

// PollBankIDWithQRUpdates polls for authentication completion while automatically
// refreshing the QR code every second to prevent expiration.
// This is the recommended method for QR-based authentication.
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
				restartResp, err := a.RestartBankID(ctx)
				if err != nil {
					fmt.Printf("restart: %v\n", err)
					continue
				}
				if err := a.DisplayQRCode(restartResp.QRToken); err != nil {
					fmt.Printf("display qr: %v\n", err)
				}
			}
		}
	}()

	return a.PollBankID(ctx)
}

// ClearScreen clears the terminal using ANSI escape codes.
func (a *AuthService) ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// DisplayQRCode renders a QR code in the terminal for BankID scanning.
// The screen is cleared first. Typically used with QRToken from StartBankID
// or RestartBankID. For automatic QR refresh, use PollBankIDWithQRUpdates instead.
func (a *AuthService) DisplayQRCode(qrCodeData string) error {
	if qrCodeData == "" {
		return fmt.Errorf("empty qr code data")
	}

	a.ClearScreen()
	fmt.Println("Scan QR code with BankID app to authenticate to Avanza...")
	qrterminal.GenerateHalfBlock(qrCodeData, qrterminal.L, os.Stdout)
	return nil
}

// EstablishSession establishes a session after successful BankID authentication.
// This is required before making other API calls.
func (a *AuthService) EstablishSession(ctx context.Context, collectResp *BankIDCollectResponse) error {
	if collectResp == nil || len(collectResp.Logins) == 0 {
		return fmt.Errorf("no logins available in authentication response")
	}

	login := collectResp.Logins[0]
	userEndpoint := fmt.Sprintf("/_api/authentication/v2/sessions/bankid/collect/%s", login.CustomerID)

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

// GetSessionInfo returns the current session state and user information.
func (a *AuthService) GetSessionInfo(ctx context.Context) (*SessionInfo, error) {
	resp, err := a.client.Get(ctx, "/_api/authentication/session/info/session")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get session info: %w", client.NewHTTPError(resp))
	}

	var sessionInfo SessionInfo
	if err := json.NewDecoder(resp.Body).Decode(&sessionInfo); err != nil {
		return nil, fmt.Errorf("failed to decode session info: %w", err)
	}

	return &sessionInfo, nil
}
