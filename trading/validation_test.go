package trading

import (
	"strings"
	"testing"
)

func TestPlaceOrderRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     PlaceOrderRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: PlaceOrderRequest{
				AccountID:   "account123",
				OrderbookID: "orderbook456",
				Price:       100.0,
				Volume:      10,
				Side:        OrderSideBuy,
				Condition:   OrderConditionNormal,
			},
			wantErr: false,
		},
		{
			name: "missing accountId",
			req: PlaceOrderRequest{
				OrderbookID: "orderbook456",
				Price:       100.0,
				Volume:      10,
				Side:        OrderSideBuy,
				Condition:   OrderConditionNormal,
			},
			wantErr: true,
			errMsg:  "accountId is required",
		},
		{
			name: "missing orderbookId",
			req: PlaceOrderRequest{
				AccountID: "account123",
				Price:     100.0,
				Volume:    10,
				Side:      OrderSideBuy,
				Condition: OrderConditionNormal,
			},
			wantErr: true,
			errMsg:  "orderbookId is required",
		},
		{
			name: "invalid price - zero",
			req: PlaceOrderRequest{
				AccountID:   "account123",
				OrderbookID: "orderbook456",
				Price:       0.0,
				Volume:      10,
				Side:        OrderSideBuy,
				Condition:   OrderConditionNormal,
			},
			wantErr: true,
			errMsg:  "price must be greater than 0",
		},
		{
			name: "invalid price - negative",
			req: PlaceOrderRequest{
				AccountID:   "account123",
				OrderbookID: "orderbook456",
				Price:       -10.0,
				Volume:      10,
				Side:        OrderSideBuy,
				Condition:   OrderConditionNormal,
			},
			wantErr: true,
			errMsg:  "price must be greater than 0",
		},
		{
			name: "invalid volume - zero",
			req: PlaceOrderRequest{
				AccountID:   "account123",
				OrderbookID: "orderbook456",
				Price:       100.0,
				Volume:      0,
				Side:        OrderSideBuy,
				Condition:   OrderConditionNormal,
			},
			wantErr: true,
			errMsg:  "volume must be greater than 0",
		},
		{
			name: "invalid volume - negative",
			req: PlaceOrderRequest{
				AccountID:   "account123",
				OrderbookID: "orderbook456",
				Price:       100.0,
				Volume:      -5,
				Side:        OrderSideBuy,
				Condition:   OrderConditionNormal,
			},
			wantErr: true,
			errMsg:  "volume must be greater than 0",
		},
		{
			name: "invalid side",
			req: PlaceOrderRequest{
				AccountID:   "account123",
				OrderbookID: "orderbook456",
				Price:       100.0,
				Volume:      10,
				Side:        OrderSide("INVALID"),
				Condition:   OrderConditionNormal,
			},
			wantErr: true,
			errMsg:  "side must be",
		},
		{
			name: "invalid condition",
			req: PlaceOrderRequest{
				AccountID:   "account123",
				OrderbookID: "orderbook456",
				Price:       100.0,
				Volume:      10,
				Side:        OrderSideBuy,
				Condition:   OrderCondition("INVALID"),
			},
			wantErr: true,
			errMsg:  "condition must be",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("Validate() expected error containing %q, got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %q, want containing %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateOrderRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     ValidateOrderRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: ValidateOrderRequest{
				AccountID:   "account123",
				OrderbookID: "orderbook456",
				Price:       100.0,
				Volume:      10,
				Side:        OrderSideBuy,
				Condition:   OrderConditionNormal,
				ISIN:        "SE0000000000",
				Currency:    "SEK",
				MarketPlace: "XSTO",
			},
			wantErr: false,
		},
		{
			name: "missing isin",
			req: ValidateOrderRequest{
				AccountID:   "account123",
				OrderbookID: "orderbook456",
				Price:       100.0,
				Volume:      10,
				Side:        OrderSideBuy,
				Condition:   OrderConditionNormal,
				Currency:    "SEK",
				MarketPlace: "XSTO",
			},
			wantErr: true,
		},
		{
			name: "missing currency",
			req: ValidateOrderRequest{
				AccountID:   "account123",
				OrderbookID: "orderbook456",
				Price:       100.0,
				Volume:      10,
				Side:        OrderSideBuy,
				Condition:   OrderConditionNormal,
				ISIN:        "SE0000000000",
				MarketPlace: "XSTO",
			},
			wantErr: true,
		},
		{
			name: "missing marketPlace",
			req: ValidateOrderRequest{
				AccountID:   "account123",
				OrderbookID: "orderbook456",
				Price:       100.0,
				Volume:      10,
				Side:        OrderSideBuy,
				Condition:   OrderConditionNormal,
				ISIN:        "SE0000000000",
				Currency:    "SEK",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPreliminaryFeeRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     PreliminaryFeeRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: PreliminaryFeeRequest{
				AccountID:   "account123",
				OrderbookID: "orderbook456",
				Price:       "100.0",
				Volume:      "10",
				Side:        OrderSideBuy,
			},
			wantErr: false,
		},
		{
			name: "invalid price - not a number",
			req: PreliminaryFeeRequest{
				AccountID:   "account123",
				OrderbookID: "orderbook456",
				Price:       "invalid",
				Volume:      "10",
				Side:        OrderSideBuy,
			},
			wantErr: true,
		},
		{
			name: "invalid volume - not a number",
			req: PreliminaryFeeRequest{
				AccountID:   "account123",
				OrderbookID: "orderbook456",
				Price:       "100.0",
				Volume:      "invalid",
				Side:        OrderSideBuy,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlaceStopLossRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     PlaceStopLossRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: PlaceStopLossRequest{
				AccountID:   "account123",
				OrderbookID: "orderbook456",
				StopLossTrigger: StopLossTrigger{
					Type:      StopLossTriggerLessOrEqual,
					Value:     200.0,
					ValueType: StopLossValueMonetary,
				},
				StopLossOrderEvent: StopLossOrderEvent{
					Type:      StopLossOrderEventBuy,
					Price:     100.0,
					Volume:    10,
					ValidDays: 30,
					PriceType: StopLossPriceMonetary,
				},
			},
			wantErr: false,
		},
		{
			name: "missing accountId",
			req: PlaceStopLossRequest{
				OrderbookID: "orderbook456",
				StopLossTrigger: StopLossTrigger{
					Type:      StopLossTriggerLessOrEqual,
					Value:     200.0,
					ValueType: StopLossValueMonetary,
				},
				StopLossOrderEvent: StopLossOrderEvent{
					Type:      StopLossOrderEventBuy,
					Price:     100.0,
					Volume:    10,
					ValidDays: 30,
					PriceType: StopLossPriceMonetary,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
