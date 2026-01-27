package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vmorsell/avanza-sdk-go/client"
)

// newTestClient creates a client configured to use the test server URL.
func newTestClient(baseURL string) *client.Client {
	return client.NewClient(client.WithBaseURL(baseURL))
}

func TestNewAuthService(t *testing.T) {
	c := client.NewClient()
	service := NewAuthService(c)

	if service == nil {
		t.Fatal("expected service to be non-nil")
	}

	if service.client == nil {
		t.Error("expected client to be non-nil")
	}
}

func TestStartBankID_Success(t *testing.T) {
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := requestCount.Add(1)

		// First request should be GET to "/" for initial cookies
		if n == 1 {
			if r.Method != "GET" {
				t.Errorf("expected GET method for initial request, got %s", r.Method)
			}
			if r.URL.Path != "/" {
				t.Errorf("expected path / for initial request, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		// Second request should be POST to the BankID endpoint
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		if r.URL.Path != "/_api/authentication/v2/sessions/bankid" {
			t.Errorf("expected path /_api/authentication/v2/sessions/bankid, got %s", r.URL.Path)
		}

		var req BankIDStartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if req.Method != "QR_START" {
			t.Errorf("expected method QR_START, got %s", req.Method)
		}

		if req.ReturnScheme != "NULL" {
			t.Errorf("expected returnScheme NULL, got %s", req.ReturnScheme)
		}

		resp := BankIDStartResponse{
			TransactionID: "FOO",
			QRToken:       "BAR",
			Expires:       "42",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx := context.Background()
	resp, err := service.StartBankID(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.TransactionID != "FOO" {
		t.Errorf("expected transaction ID FOO, got %s", resp.TransactionID)
	}

	if resp.QRToken != "BAR" {
		t.Errorf("expected QR token BAR, got %s", resp.QRToken)
	}

	if resp.Expires != "42" {
		t.Errorf("expected expires 42, got %s", resp.Expires)
	}
}

func TestStartBankID_HTTPError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"bad request", http.StatusBadRequest},
		{"unauthorized", http.StatusUnauthorized},
		{"forbidden", http.StatusForbidden},
		{"internal error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(`{"error":"FOO"}`))
			}))
			defer server.Close()

			c := newTestClient(server.URL)
			service := NewAuthService(c)

			ctx := context.Background()
			_, err := service.StartBankID(ctx)
			if err == nil {
				t.Error("expected error for HTTP error status, got nil")
			}
		})
	}
}

func TestStartBankID_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx := context.Background()
	_, err := service.StartBankID(ctx)
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

func TestStartBankID_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := service.StartBankID(ctx)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestRestartBankID_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		if r.URL.Path != "/_api/authentication/v2/sessions/bankid/restart" {
			t.Errorf("expected path /_api/authentication/v2/sessions/bankid/restart, got %s", r.URL.Path)
		}

		resp := BankIDStartResponse{
			TransactionID: "FOO",
			QRToken:       "BAR",
			Expires:       "42",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx := context.Background()
	resp, err := service.RestartBankID(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.QRToken != "BAR" {
		t.Errorf("expected QR token BAR, got %s", resp.QRToken)
	}
}

func TestRestartBankID_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"FOO"}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx := context.Background()
	_, err := service.RestartBankID(ctx)
	if err == nil {
		t.Error("expected error for HTTP error status, got nil")
	}
}

func TestCollectBankID_AllStates(t *testing.T) {
	tests := []struct {
		name       string
		state      string
		hintCode   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "pending state",
			state:      "PENDING",
			hintCode:   "outstandingTransaction",
			statusCode: http.StatusAccepted,
			wantErr:    false,
		},
		{
			name:       "complete state",
			state:      "COMPLETE",
			hintCode:   "",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "failed state",
			state:      "FAILED",
			hintCode:   "userCancel",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "user sign hint",
			state:      "PENDING",
			hintCode:   "userSign",
			statusCode: http.StatusAccepted,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/_api/authentication/v2/sessions/bankid/collect" {
					t.Errorf("expected path /_api/authentication/v2/sessions/bankid/collect, got %s", r.URL.Path)
				}

				resp := BankIDCollectResponse{
					State:                "COMPLETE",
					HintCode:             tt.hintCode,
					Name:                 "FOO BAR",
					TransactionID:        "FOO",
					IdentificationNumber: "42",
					Logins: []Login{
						{
							CustomerID: "FOO",
							Username:   "BAR",
							LoginPath:  "/test",
							Accounts: []Account{
								{AccountName: "FOO", AccountType: "BAR"},
							},
						},
					},
				}

				w.WriteHeader(tt.statusCode)
				_ = json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			c := newTestClient(server.URL)
			service := NewAuthService(c)

			ctx := context.Background()
			resp, err := service.CollectBankID(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("CollectBankID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if resp.State != "COMPLETE" {
					t.Errorf("expected state COMPLETE, got %s", resp.State)
				}
				if resp.Name != "FOO BAR" {
					t.Errorf("expected name FOO BAR, got %s", resp.Name)
				}
				if len(resp.Logins) != 1 {
					t.Errorf("expected 1 login, got %d", len(resp.Logins))
				}
			}
		})
	}
}

func TestCollectBankID_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx := context.Background()
	_, err := service.CollectBankID(ctx)
	if err == nil {
		t.Error("expected error for HTTP error status, got nil")
	}
}

func TestCollectBankID_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx := context.Background()
	_, err := service.CollectBankID(ctx)
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

func TestPollBankID_Complete(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)

		var state string
		if n < 3 {
			state = "PENDING"
		} else {
			state = "COMPLETE"
		}

		resp := BankIDCollectResponse{
			State: state,
			Name:  "FOO BAR",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := service.PollBankID(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.State != "COMPLETE" {
		t.Errorf("expected state COMPLETE, got %s", resp.State)
	}

	if cc := callCount.Load(); cc < 3 {
		t.Errorf("expected at least 3 calls, got %d", cc)
	}
}

func TestPollBankID_Failed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := BankIDCollectResponse{
			State:    "FAILED",
			HintCode: "userCancel",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := service.PollBankID(ctx)
	if err == nil {
		t.Error("expected error for failed authentication, got nil")
	}

	if err != nil && err.Error() != "bankid authentication failed: userCancel" {
		t.Errorf("expected specific error message, got: %v", err)
	}
}

func TestPollBankID_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := BankIDCollectResponse{
			State: "PENDING",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := service.PollBankID(ctx)
	if err == nil {
		t.Error("expected context deadline exceeded error, got nil")
	}
}

func TestPollBankID_ImmediateComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := BankIDCollectResponse{
			State: "COMPLETE",
			Name:  "FOO BAR",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx := context.Background()
	resp, err := service.PollBankID(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.State != "COMPLETE" {
		t.Errorf("expected state COMPLETE, got %s", resp.State)
	}
}

func TestPollBankIDWithQRUpdates_Complete(t *testing.T) {
	var mu sync.Mutex
	collectCalls := 0
	restartCalls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch r.URL.Path {
		case "/_api/authentication/v2/sessions/bankid/collect":
			collectCalls++
			var state string
			if collectCalls < 3 {
				state = "PENDING"
			} else {
				state = "COMPLETE"
			}

			resp := BankIDCollectResponse{
				State: state,
				Name:  "FOO BAR",
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
		case "/_api/authentication/v2/sessions/bankid/restart":
			restartCalls++

			resp := BankIDStartResponse{
				QRToken: "BAR",
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := service.PollBankIDWithQRUpdates(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.State != "COMPLETE" {
		t.Errorf("expected state COMPLETE, got %s", resp.State)
	}

	mu.Lock()
	cc := collectCalls
	rc := restartCalls
	mu.Unlock()

	if cc < 3 {
		t.Errorf("expected at least 3 collect calls, got %d", cc)
	}

	// QR should be updated at least once
	if rc < 1 {
		t.Errorf("expected at least 1 restart call, got %d", rc)
	}
}

func TestPollBankIDWithQRUpdates_Failed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_api/authentication/v2/sessions/bankid/collect":
			resp := BankIDCollectResponse{
				State:    "FAILED",
				HintCode: "userCancel",
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
		case "/_api/authentication/v2/sessions/bankid/restart":
			resp := BankIDStartResponse{
				QRToken: "BAR",
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := service.PollBankIDWithQRUpdates(ctx)
	if err == nil {
		t.Error("expected error for failed authentication, got nil")
	}
}

func TestPollBankIDWithQRUpdates_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_api/authentication/v2/sessions/bankid/collect":
			resp := BankIDCollectResponse{
				State: "PENDING",
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
		case "/_api/authentication/v2/sessions/bankid/restart":
			resp := BankIDStartResponse{
				QRToken: "BAR",
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := service.PollBankIDWithQRUpdates(ctx)
	if err == nil {
		t.Error("expected context deadline exceeded error, got nil")
	}
}

func TestDisplayQRCode_EmptyData(t *testing.T) {
	c := client.NewClient()
	service := NewAuthService(c)

	err := service.DisplayQRCode("")
	if err == nil {
		t.Error("expected error for empty QR code data, got nil")
	}

	expectedMsg := "empty qr code data"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestDisplayQRCode_ValidData(t *testing.T) {
	c := client.NewClient()
	service := NewAuthService(c)

	// We can't easily test the actual QR code output, but we can verify no error
	err := service.DisplayQRCode("FOO")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBankIDStartRequest_JSONMarshaling(t *testing.T) {
	req := BankIDStartRequest{
		Method:       "QR_START",
		ReturnScheme: "NULL",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded BankIDStartRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Method != req.Method {
		t.Errorf("expected method %s, got %s", req.Method, decoded.Method)
	}
}

func TestBankIDCollectResponse_CompleteStructure(t *testing.T) {
	jsonData := `{
		"name": "FOO BAR",
		"transactionId": "FOO",
		"state": "COMPLETE",
		"hintCode": "",
		"identificationNumber": "42",
		"logins": [
			{
				"customerId": "FOO",
				"username": "BAR",
				"loginPath": "/test",
				"accounts": [
					{
						"accountName": "FOO",
						"accountType": "BAR"
					}
				]
			}
		],
		"recommendedTargetCustomers": [],
		"poa": []
	}`

	var resp BankIDCollectResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Name != "FOO BAR" {
		t.Errorf("expected name FOO BAR, got %s", resp.Name)
	}

	if resp.State != "COMPLETE" {
		t.Errorf("expected state COMPLETE, got %s", resp.State)
	}

	if len(resp.Logins) != 1 {
		t.Fatalf("expected 1 login, got %d", len(resp.Logins))
	}

	if resp.Logins[0].CustomerID != "FOO" {
		t.Errorf("expected customer ID FOO, got %s", resp.Logins[0].CustomerID)
	}

	if len(resp.Logins[0].Accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(resp.Logins[0].Accounts))
	}
}

func TestCollectBankID_EmptyResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Empty body
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx := context.Background()
	_, err := service.CollectBankID(ctx)
	if err == nil {
		t.Error("expected error for empty response body, got nil")
	}
}

func TestStartBankID_EmptyResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "")
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx := context.Background()
	_, err := service.StartBankID(ctx)
	if err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestEstablishSession_Success(t *testing.T) {
	requestPaths := []string{}
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestPaths = append(requestPaths, r.URL.Path)
		mu.Unlock()

		switch r.URL.Path {
		case "/_api/authentication/v2/sessions/bankid/collect/customer-123":
			if r.Method != http.MethodGet {
				t.Errorf("expected GET for select user, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
		case "/handla/order.html":
			if r.Method != http.MethodGet {
				t.Errorf("expected GET for trading page, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
		case "/_api/authentication/session/info/session":
			if r.Method != http.MethodGet {
				t.Errorf("expected GET for session info, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	collectResp := &BankIDCollectResponse{
		State: "COMPLETE",
		Logins: []Login{
			{
				CustomerID: "customer-123",
				Username:   "testuser",
			},
		},
	}

	err := service.EstablishSession(context.Background(), collectResp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(requestPaths) != 3 {
		t.Fatalf("expected 3 requests, got %d: %v", len(requestPaths), requestPaths)
	}
	if requestPaths[0] != "/_api/authentication/v2/sessions/bankid/collect/customer-123" {
		t.Errorf("first request path = %q, want select user endpoint", requestPaths[0])
	}
	if requestPaths[1] != "/handla/order.html" {
		t.Errorf("second request path = %q, want trading page", requestPaths[1])
	}
	if requestPaths[2] != "/_api/authentication/session/info/session" {
		t.Errorf("third request path = %q, want session info", requestPaths[2])
	}
}

func TestEstablishSession_NilResponse(t *testing.T) {
	c := client.NewClient()
	service := NewAuthService(c)

	err := service.EstablishSession(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil response, got nil")
	}
}

func TestEstablishSession_NoLogins(t *testing.T) {
	c := client.NewClient()
	service := NewAuthService(c)

	collectResp := &BankIDCollectResponse{
		State:  "COMPLETE",
		Logins: []Login{},
	}

	err := service.EstablishSession(context.Background(), collectResp)
	if err == nil {
		t.Fatal("expected error for empty logins, got nil")
	}
}

func TestEstablishSession_SelectUserHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	collectResp := &BankIDCollectResponse{
		Logins: []Login{{CustomerID: "customer-123"}},
	}

	err := service.EstablishSession(context.Background(), collectResp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestEstablishSession_SessionVerifyHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_api/authentication/v2/sessions/bankid/collect/customer-123":
			w.WriteHeader(http.StatusOK)
		case "/handla/order.html":
			w.WriteHeader(http.StatusOK)
		case "/_api/authentication/session/info/session":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("session expired"))
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	collectResp := &BankIDCollectResponse{
		Logins: []Login{{CustomerID: "customer-123"}},
	}

	err := service.EstablishSession(context.Background(), collectResp)
	if err == nil {
		t.Fatal("expected error for session verification failure, got nil")
	}
}

func TestEstablishSession_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	collectResp := &BankIDCollectResponse{
		Logins: []Login{{CustomerID: "customer-123"}},
	}

	err := service.EstablishSession(ctx, collectResp)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}

func TestGetSessionInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/authentication/session/info/session" {
			t.Errorf("expected session info path, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SessionInfo{
			User: User{
				LoggedIn:      true,
				GreetingName:  "Test User",
				SecurityToken: "token-123",
				ID:            "user-1",
				CustomerGroup: "STANDARD",
			},
		})
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	info, err := service.GetSessionInfo(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !info.User.LoggedIn {
		t.Error("expected User.LoggedIn to be true")
	}
	if got, want := info.User.GreetingName, "Test User"; got != want {
		t.Errorf("User.GreetingName = %q, want %q", got, want)
	}
	if got, want := info.User.ID, "user-1"; got != want {
		t.Errorf("User.ID = %q, want %q", got, want)
	}
}

func TestGetSessionInfo_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	_, err := service.GetSessionInfo(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetSessionInfo_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{invalid"))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	_, err := service.GetSessionInfo(context.Background())
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestGetSessionInfo_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	service := NewAuthService(c)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.GetSessionInfo(ctx)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}
