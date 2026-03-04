package producer

import (
	"context"
	"fmt"
	"kafka-app/internal/config"
	"kafka-app/internal/models"
	"kafka-app/pkg/logger"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	client *kgo.Client
	config *config.ProducerConfig
	wg     sync.WaitGroup
	done   chan struct{}
}

// New creates a new producer
func New(cfg *config.ProducerConfig) (*Producer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.DefaultProduceTopic(cfg.Topic),
		kgo.ProducerBatchMaxBytes(int32(cfg.BatchSize)),
		kgo.ProducerLinger(time.Duration(cfg.BatchTimeout) * time.Millisecond),
		kgo.RecordPartitioner(kgo.StickyKeyPartitioner(nil)),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	return &Producer{
		client: client,
		config: cfg,
		done:   make(chan struct{}),
	}, nil
}

// Start begins producing messages
func (p *Producer) Start(ctx context.Context) {
	p.wg.Add(1)
	go p.produceLoop(ctx)
}

// produceLoop continuously produces messages
func (p *Producer) produceLoop(ctx context.Context) {
	defer p.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("producer loop stopping")
			return
		case <-p.done:
			return
		case <-ticker.C:
			if err := p.produceEvent(ctx); err != nil {
				logger.Log.Error("failed to produce event", "error", err)
			}
		}
	}
}

// produceEvent produces a single event
func (p *Producer) produceEvent(ctx context.Context) error {
	event := &models.Event{
		ID:        uuid.New().String(),
		Type:      "test_event",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"message": "Hello from Kafka producer!",
			"counter": time.Now().Unix(),
		},
	}

	data, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	record := &kgo.Record{
		Key:   []byte(event.ID),
		Value: data,
	}

	// Produce with callback
	p.client.Produce(ctx, record, func(r *kgo.Record, err error) {
		if err != nil {
			logger.Log.Error("produce failed",
				"error", err,
				"topic", r.Topic,
				"partition", r.Partition,
			)
		} else {
			logger.Log.Info("produced message",
				"topic", r.Topic,
				"partition", r.Partition,
				"offset", r.Offset,
				"event_id", event.ID,
			)
		}
	})

	return nil
}

// ProduceSync produces a message synchronously
func (p *Producer) ProduceSync(ctx context.Context, event *models.Event) error {
	data, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	record := &kgo.Record{
		Key:   []byte(event.ID),
		Value: data,
	}

	results := p.client.ProduceSync(ctx, record)
	if err := results.FirstErr(); err != nil {
		return fmt.Errorf("produce failed: %w", err)
	}

	logger.Log.Info("produced message synchronously",
		"event_id", event.ID,
		"offset", record.Offset,
	)

	return nil
}

// Close gracefully shuts down the producer
func (p *Producer) Close(ctx context.Context) error {
	logger.Log.Info("shutting down producer")
	close(p.done)
	p.wg.Wait()

	// Flush any pending messages
	if err := p.client.Flush(ctx); err != nil {
		logger.Log.Error("failed to flush messages", "error", err)
		return err
	}

	p.client.Close()
	logger.Log.Info("producer shutdown complete")
	return nil
}
