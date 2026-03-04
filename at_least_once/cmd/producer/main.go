package main

import (
	"context"
	"flag"
	"kafka-app/internal/config"
	"kafka-app/internal/producer"
	"kafka-app/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var configPath = flag.String("config", "configs/producer.yaml", "path to config file")

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadProducerConfig(*configPath)
	if err != nil {
		panic(err)
	}

	// Initialize logger
	logger.Init(cfg.LogLevel)
	logger.Log.Info("starting producer", "config", *configPath)

	// Create producer
	prod, err := producer.New(cfg)
	if err != nil {
		logger.Log.Error("failed to create producer", "error", err)
		os.Exit(1)
	}

	// Start producing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	prod.Start(ctx)

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Log.Info("received shutdown signal")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := prod.Close(shutdownCtx); err != nil {
		logger.Log.Error("error during shutdown", "error", err)
		os.Exit(1)
	}

	logger.Log.Info("producer stopped successfully")
}
