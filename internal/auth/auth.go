package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/vmorsell/avanza/internal/client"
)

type AuthService struct {
	client *client.Client
}

func NewAuthService(client *client.Client) *AuthService {
	return &AuthService{
		client: client,
	}
}

type BankIDStartRequest struct {
	Method       string `json:"method"`
	ReturnScheme string `json:"returnScheme"`
}

type BankIDStartResponse struct {
	TransactionID string `json:"transactionId"`
	Expires       string `json:"expires"`
	QRToken       string `json:"qrToken"`
}

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

type Login struct {
	CustomerID string    `json:"customerId"`
	Username   string    `json:"username"`
	Accounts   []Account `json:"accounts"`
	LoginPath  string    `json:"loginPath"`
}

type Account struct {
	AccountName string `json:"accountName"`
	AccountType string `json:"accountType"`
}

type BankIDRestartRequest struct{}

func (a *AuthService) StartBankID(ctx context.Context) (*BankIDStartResponse, error) {
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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response BankIDStartResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func (a *AuthService) RestartBankID(ctx context.Context) (*BankIDStartResponse, error) {
	resp, err := a.client.Post(ctx, "/_api/authentication/v2/sessions/bankid/restart", BankIDRestartRequest{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response BankIDStartResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func (a *AuthService) CollectBankID(ctx context.Context) (*BankIDCollectResponse, error) {
	resp, err := a.client.Post(ctx, "/_api/authentication/v2/sessions/bankid/collect", BankIDRestartRequest{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response BankIDCollectResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

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

func (a *AuthService) ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

func (a *AuthService) DisplayQRCode(qrCodeData string) error {
	if qrCodeData == "" {
		return fmt.Errorf("empty qr code data")
	}

	a.ClearScreen()
	fmt.Println("Scan QR code with BankID app to authenticate to Avanza...")
	qrterminal.GenerateHalfBlock(qrCodeData, qrterminal.L, os.Stdout)
	return nil
}
