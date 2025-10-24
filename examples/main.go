package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vmorsell/avanza-sdk-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, cancelling...")
		cancel()
	}()

	avanza := avanza.New()

	fmt.Println("Starting BankID authentication...")

	startResp, err := avanza.Auth.StartBankID(ctx)
	if err != nil {
		log.Fatalf("failed to start bankid: %v", err)
	}

	if err := avanza.Auth.DisplayQRCode(startResp.QRToken); err != nil {
		log.Fatalf("failed to display qr code: %v", err)
	}

	collectResp, err := avanza.Auth.PollBankIDWithQRUpdates(ctx)
	if err != nil {
		log.Fatalf("auth failed: %v", err)
	}

	fmt.Printf("Authentication successful! Welcome %s\n", collectResp.Name)
}
