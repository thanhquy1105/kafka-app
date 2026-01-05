package tests

import (
	"context"
	"kafka-app/internal/config"
	"kafka-app/internal/consumer"
	"kafka-app/internal/models"
	"kafka-app/pkg/logger"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kfake"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestConsumer(t *testing.T) {
	// Initialize logger for tests
	logger.Init("error") // Use error level to reduce test output

	// Create fake Kafka cluster
	cluster := kfake.MustCluster()
	defer cluster.Close()

	// Produce a test message first
	producerClient, _ := kgo.NewClient(
		kgo.SeedBrokers(cluster.ListenAddrs()...),
		kgo.DefaultProduceTopic("test-topic"),
	)
	defer producerClient.Close()

	event := &models.Event{
		ID:        uuid.New().String(),
		Type:      "test_event",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"test": "data",
		},
	}

	data, _ := event.ToJSON()
	producerClient.ProduceSync(context.Background(), &kgo.Record{
		Key:   []byte(event.ID),
		Value: data,
	})

	// Create consumer config
	cfg := &config.ConsumerConfig{
		Brokers:        cluster.ListenAddrs(),
		Topics:         []string{"test-topic"},
		GroupID:        "test-group",
		AutoCommit:     true,
		CommitInterval: 5000,
		LogLevel:       "info",
	}

	// Create consumer
	cons, err := consumer.New(cfg)
	if err != nil {
		t.Fatalf("failed to create consumer: %v", err)
	}
	defer cons.Close(context.Background())

	// Start consumer
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cons.Start(ctx)

	// Wait a bit for message to be consumed
	time.Sleep(2 * time.Second)

	// Test passes if no errors occurred
	t.Log("consumer test completed successfully")
}
