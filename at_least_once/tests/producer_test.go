package tests

import (
	"context"
	"kafka-app/internal/config"
	"kafka-app/internal/models"
	"kafka-app/internal/producer"
	"kafka-app/pkg/logger"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kfake"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestProducer(t *testing.T) {
	// Initialize logger for tests
	cfg := &config.ProducerConfig{
		LogLevel: "error", // Use error level to reduce test output
	}

	// Create fake Kafka cluster with pre-created topic
	cluster := kfake.MustCluster(kfake.SeedTopics(1, "test-topic"))
	defer cluster.Close()

	// Update config with cluster addresses
	cfg.Brokers = cluster.ListenAddrs()
	cfg.Topic = "test-topic"
	cfg.BatchSize = 1048576
	cfg.BatchTimeout = 100
	cfg.MaxRetries = 3

	// Initialize logger
	logger.Init(cfg.LogLevel)

	ctx := context.Background()

	// Create producer
	prod, err := producer.New(cfg)
	if err != nil {
		t.Fatalf("failed to create producer: %v", err)
	}
	defer prod.Close(context.Background())

	// Produce a test event
	event := &models.Event{
		ID:        uuid.New().String(),
		Type:      "test_event",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"test": "data",
		},
	}

	err = prod.ProduceSync(ctx, event)
	if err != nil {
		t.Fatalf("failed to produce event: %v", err)
	}

	// Verify message was produced
	client, _ := kgo.NewClient(
		kgo.SeedBrokers(cluster.ListenAddrs()...),
		kgo.ConsumeTopics("test-topic"),
	)
	defer client.Close()

	fetches := client.PollFetches(ctx)
	if fetches.NumRecords() != 1 {
		t.Fatalf("expected 1 record, got %d", fetches.NumRecords())
	}

	fetches.EachRecord(func(r *kgo.Record) {
		receivedEvent, err := models.FromJSON(r.Value)
		if err != nil {
			t.Fatalf("failed to deserialize event: %v", err)
		}
		if receivedEvent.ID != event.ID {
			t.Fatalf("expected event ID %s, got %s", event.ID, receivedEvent.ID)
		}
	})
}
