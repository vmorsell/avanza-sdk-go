package avanza_test

import (
	"context"
	"log"
	"time"

	"github.com/vmorsell/avanza-sdk-go"
	"github.com/vmorsell/avanza-sdk-go/trading"
)

func ExampleNew() {
	client := avanza.New()
	_ = client
}

func ExampleAvanza_Auth() {
	client := avanza.New()
	ctx := context.Background()

	startResp, _ := client.Auth.StartBankID(ctx)
	_ = startResp
}

func ExampleAvanza_Accounts() {
	client := avanza.New()
	ctx := context.Background()

	overview, _ := client.Accounts.GetOverview(ctx)
	_ = overview
}

func ExampleWithBaseURL() {
	client := avanza.New(avanza.WithBaseURL("http://localhost:8080"))
	_ = client
}

func ExampleWithHTTPClient() {
	client := avanza.New(avanza.WithHTTPClient(nil))
	_ = client
}

func ExampleWithUserAgent() {
	client := avanza.New(avanza.WithUserAgent("MyApp/1.0"))
	_ = client
}

func ExampleAvanza_Trading() {
	client := avanza.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &trading.PlaceOrderRequest{
		Side:        trading.OrderSideBuy,
		OrderbookID: "5247",
		Price:       100.0,
		Volume:      1,
	}
	_ = req

	_, _ = client.Trading.PlaceOrder(ctx, req)
}

func Example_placeOrder() {
	client := avanza.New()
	ctx := context.Background()

	// Validate order first
	validateReq := &trading.ValidateOrderRequest{
		AccountID:   "account123",
		OrderbookID: "5247",
		Price:       100.0,
		Volume:      1,
		Side:        trading.OrderSideBuy,
		Condition:   trading.OrderConditionNormal,
		ISIN:        "SE0000108656",
		Currency:    "SEK",
		MarketPlace: "STO",
	}
	_, err := client.Trading.ValidateOrder(ctx, validateReq)
	if err != nil {
		log.Fatal(err)
	}

	// Place order
	req := &trading.PlaceOrderRequest{
		AccountID:   "account123",
		OrderbookID: "5247",
		Price:       100.0,
		Volume:      1,
		Side:        trading.OrderSideBuy,
		Condition:   trading.OrderConditionNormal,
	}
	resp, err := client.Trading.PlaceOrder(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	_ = resp
}

func Example_validateOrder() {
	client := avanza.New()
	ctx := context.Background()

	req := &trading.ValidateOrderRequest{
		AccountID:   "account123",
		OrderbookID: "5247",
		Price:       100.0,
		Volume:      1,
		Side:        trading.OrderSideBuy,
		Condition:   trading.OrderConditionNormal,
		ISIN:        "SE0000108656",
		Currency:    "SEK",
		MarketPlace: "STO",
	}

	resp, err := client.Trading.ValidateOrder(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	// Check validation results
	if !resp.CommissionWarning.Valid {
		log.Println("Commission warning")
	}
	_ = resp
}

func Example_getPreliminaryFee() {
	client := avanza.New()
	ctx := context.Background()

	req := &trading.PreliminaryFeeRequest{
		AccountID:   "account123",
		OrderbookID: "5247",
		Price:       "100.0",
		Volume:      "1",
		Side:        trading.OrderSideBuy,
	}

	fee, err := client.Trading.GetPreliminaryFee(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	_ = fee
}

func Example_placeStopLoss() {
	client := avanza.New()
	ctx := context.Background()

	req := &trading.PlaceStopLossRequest{
		AccountID:   "account123",
		OrderbookID: "5247",
		StopLossTrigger: trading.StopLossTrigger{
			Type:      trading.StopLossTriggerLessOrEqual,
			Value:     90.0,
			ValueType: trading.StopLossValueMonetary,
		},
		StopLossOrderEvent: trading.StopLossOrderEvent{
			Type:      trading.StopLossOrderEventSell,
			Price:     90.0,
			Volume:    1,
			PriceType: trading.StopLossPriceMonetary,
			ValidDays: 30,
		},
	}

	resp, err := client.Trading.PlaceStopLoss(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	_ = resp
}

func Example_subscribeToOrderDepth() {
	client := avanza.New()
	ctx := context.Background()

	sub, err := client.Market.SubscribeToOrderDepth(ctx, "5247")
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Close()

	// Process events
	for event := range sub.Events() {
		_ = event
		// Handle order depth update
	}
}

func Example_pollBankIDWithQRUpdates() {
	client := avanza.New()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	startResp, err := client.Auth.StartBankID(ctx)
	if err != nil {
		log.Fatal(err)
	}

	client.Auth.DisplayQRCode(startResp.QRToken)

	// Poll with automatic QR refresh
	collectResp, err := client.Auth.PollBankIDWithQRUpdates(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Establish session
	if err := client.Auth.EstablishSession(ctx, collectResp); err != nil {
		log.Fatal(err)
	}

	_ = collectResp
}
