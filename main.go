package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/meu/go-ether/api"
	"github.com/meu/go-ether/client"
	"github.com/meu/go-ether/config"
	"github.com/meu/go-ether/service"
	"github.com/meu/go-ether/store"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	cfg := config.Load()

	nodeURL := cfg.GetNodeURL()
	if nodeURL == "" {
		log.Fatal("ETH_WS_URL or ETH_RPC_URL must be set")
	}

	if cfg.ERC20Contract == "" {
		log.Fatal("ERC20_CONTRACT must be set")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ethClient, err := client.New(ctx, nodeURL)
	if err != nil {
		log.Fatalf("failed to create eth client: %v", err)
	}
	defer ethClient.Close()

	eventStore := store.NewEventStore(100)

	blockService := service.NewBlockService(ethClient)
	txService := service.NewTxService(ethClient)
	eventService, err := service.NewEventService(ethClient, eventStore, cfg.ERC20Contract)
	if err != nil {
		log.Fatalf("failed to create event service: %v", err)
	}

	go eventService.StartListening(ctx)

	handlers := api.NewHandlers(blockService, txService, eventStore)
	server := api.NewServer(handlers, ":8080")

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	fmt.Printf("\nreceived signal %s, shutting down...\n", sig.String())

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("http server shutdown error: %v", err)
	}

	cancel()
	log.Println("shutdown complete")
}
