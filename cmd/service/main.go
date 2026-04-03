package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/warpgr/test_task/app"
	"github.com/warpgr/test_task/configs"
)

func main() {
	ctx := context.Background()

	cfg, err := configs.GetRatesServiceConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	app := app.NewRatesServiceApp(*cfg)

	if err := app.Init(ctx); err != nil {
		log.Fatalf("failed to init app: %v", err)
	}

	errChan := app.Run(ctx)
	log.Println("Service started")

	wait(errChan)

	if err := app.Shutdown(ctx); err != nil {
		log.Printf("failed to shutdown app: %v", err)
	}

	log.Println("Service stopped")
}

func wait(errChan <-chan error) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Fatalf("application error: %v", err)
	case <-sigChan:
		log.Println("received shutdown signal")
	}
}
