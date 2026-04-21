package avanza_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/vmorsell/avanza-sdk-go"
	"github.com/vmorsell/avanza-sdk-go/client"
	"github.com/vmorsell/avanza-sdk-go/trading"
)

func ExampleNew() {
	c := avanza.New()
	ctx := context.Background()

	overview, err := c.Accounts.GetOverview(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d accounts\n", len(overview.Accounts))
}

func ExampleWithBaseURL() {
	c := avanza.New(avanza.WithBaseURL("http://localhost:8080"))
	_, err := c.Accounts.GetOverview(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleWithHTTPClient() {
	httpClient := &http.Client{Timeout: 60 * time.Second}
	c := avanza.New(avanza.WithHTTPClient(httpClient))
	_, err := c.Accounts.GetOverview(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleWithUserAgent() {
	c := avanza.New(avanza.WithUserAgent("my-trading-bot/1.0"))
	_, err := c.Accounts.GetOverview(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleWithRateLimiter() {
	limiter := &client.SimpleRateLimiter{Interval: 200 * time.Millisecond}
	c := avanza.New(avanza.WithRateLimiter(limiter))
	_, err := c.Accounts.GetOverview(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleAvanza_Auth() {
	c := avanza.New()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	startResp, err := c.Auth.StartBankID(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if err := c.Auth.DisplayQRCode(startResp.QRToken); err != nil {
		log.Fatal(err)
	}

	collectResp, err := c.Auth.PollBankIDWithQRUpdates(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if err := c.Auth.EstablishSession(ctx, collectResp); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("authenticated as %s\n", collectResp.Name)
}

func ExampleAvanza_Accounts() {
	c := avanza.New()
	ctx := context.Background()

	overview, err := c.Accounts.GetOverview(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, acct := range overview.Accounts {
		fmt.Printf("%s: %s\n", acct.Name, acct.ID)
	}
}

func ExampleAvanza_Trading_placeOrder() {
	c := avanza.New()
	ctx := context.Background()

	resp, err := c.Trading.PlaceOrder(ctx, &trading.PlaceOrderRequest{
		RequestID:   uuid.New().String(),
		AccountID:   "account-id",
		OrderbookID: "5247",
		Side:        trading.OrderSideBuy,
		Condition:   trading.OrderConditionNormal,
		Price:       245.50,
		Volume:      10,
	})
	if err != nil {
		var httpErr *client.HTTPError
		if errors.As(err, &httpErr) {
			log.Fatalf("avanza %d: %s", httpErr.StatusCode, httpErr.Body)
		}
		log.Fatal(err)
	}
	fmt.Printf("order %s: %s\n", resp.OrderID, resp.OrderRequestStatus)
}

func ExampleAvanza_Trading_validateOrder() {
	c := avanza.New()
	ctx := context.Background()

	resp, err := c.Trading.ValidateOrder(ctx, &trading.ValidateOrderRequest{
		AccountID:   "account-id",
		OrderbookID: "5247",
		Side:        trading.OrderSideBuy,
		Condition:   trading.OrderConditionNormal,
		Price:       245.50,
		Volume:      10,
		ISIN:        "SE0000108656",
		Currency:    "SEK",
		MarketPlace: "STO",
	})
	if err != nil {
		log.Fatal(err)
	}

	if !resp.CommissionWarning.Valid {
		log.Println("commission warning triggered")
	}
}

func ExampleAvanza_Trading_preliminaryFee() {
	c := avanza.New()
	ctx := context.Background()

	fee, err := c.Trading.GetPreliminaryFee(ctx, &trading.PreliminaryFeeRequest{
		AccountID:   "account-id",
		OrderbookID: "5247",
		Side:        trading.OrderSideBuy,
		Price:       "245.50",
		Volume:      "10",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("commission: %s %s\n", fee.Commission, fee.OrderbookCurrency)
}

func ExampleAvanza_Trading_placeStopLoss() {
	c := avanza.New()
	ctx := context.Background()

	resp, err := c.Trading.PlaceStopLoss(ctx, &trading.PlaceStopLossRequest{
		AccountID:   "account-id",
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
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("stop-loss %s: %s\n", resp.StopLossOrderID, resp.Status)
}

func ExampleAvanza_Market_subscribeToOrderDepth() {
	c := avanza.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := c.Market.SubscribeToOrderDepth(ctx, "5247")
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Close()

	for {
		select {
		case event := <-sub.Events():
			if event.Event == "ORDER_DEPTH" {
				for i, lvl := range event.Data.Levels {
					fmt.Printf("%d: %.0f @ %.2f / %.0f @ %.2f\n",
						i, lvl.BuyVolume, lvl.BuyPrice, lvl.SellVolume, lvl.SellPrice)
				}
			}
		case err := <-sub.Errors():
			log.Printf("stream error: %v", err)
			return
		case <-ctx.Done():
			return
		}
	}
}
