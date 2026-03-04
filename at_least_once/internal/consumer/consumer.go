package consumer

import (
	"context"
	"fmt"
	"kafka-app/internal/config"
	"kafka-app/internal/models"
	"kafka-app/pkg/logger"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
)

type Consumer struct {
	client *kgo.Client
	config *config.ConsumerConfig
	wg     sync.WaitGroup
	done   chan struct{}
}

// New creates a new consumer
func New(cfg *config.ConsumerConfig) (*Consumer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumerGroup(cfg.GroupID),
		kgo.ConsumeTopics(cfg.Topics...),
		kgo.FetchMinBytes(1),
		kgo.FetchMaxWait(5 * time.Second),
	}

	// Configure auto-commit if enabled
	if cfg.AutoCommit {
		opts = append(opts,
			kgo.AutoCommitInterval(time.Duration(cfg.CommitInterval)*time.Millisecond),
			kgo.AutoCommitCallback(
				// func(*Client, *kmsg.OffsetCommitRequest, *kmsg.OffsetCommitResponse, error) {
				// func(ctx context.Context, commits []kgo) error {
				// logger.Log.Info("auto-commit callback", "commits", commits)
				// return nil
				// },
				func(client *kgo.Client, req *kmsg.OffsetCommitRequest, resp *kmsg.OffsetCommitResponse, err error) {
					logger.Log.Info("auto-commit callback", "commits", req)
					return
				},
			),
		)
	} else {
		opts = append(opts, kgo.DisableAutoCommit())
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	return &Consumer{
		client: client,
		config: cfg,
		done:   make(chan struct{}),
	}, nil
}

// Start begins consuming messages
func (c *Consumer) Start(ctx context.Context) {
	c.wg.Add(1)
	go c.consumeLoop(ctx)
}

// consumeLoop continuously consumes messages
func (c *Consumer) consumeLoop(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("consumer loop stopping")
			return
		case <-c.done:
			return
		default:
			fetches := c.client.PollFetches(ctx)
			logger.Log.Info("fetched messages", "count", fetches.NumRecords())
			if errs := fetches.Errors(); len(errs) > 0 {
				for _, err := range errs {
					logger.Log.Error("fetch error",
						"error", err.Err,
						"topic", err.Topic,
						"partition", err.Partition,
					)
				}
				continue
			}

			fetches.EachRecord(func(record *kgo.Record) {
				if err := c.processRecord(ctx, record); err != nil {
					logger.Log.Error("failed to process record",
						"error", err,
						"topic", record.Topic,
						"partition", record.Partition,
						"offset", record.Offset,
					)
				}
			})

			// Manual commit if auto-commit is disabled
			if !c.config.AutoCommit {
				if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
					logger.Log.Error("failed to commit offsets", "error", err)
				}
			}
		}
	}
}

// processRecord processes a single record
func (c *Consumer) processRecord(ctx context.Context, record *kgo.Record) error {
	event, err := models.FromJSON(record.Value)
	if err != nil {
		return fmt.Errorf("failed to deserialize event: %w", err)
	}

	logger.Log.Info("consumed message",
		"topic", record.Topic,
		"partition", record.Partition,
		"offset", record.Offset,
		"event_id", event.ID,
		"event_type", event.Type,
		"timestamp", event.Timestamp,
	)

	// Process the event (add your business logic here)
	if err := c.handleEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to handle event: %w", err)
	}

	return nil
}

// handleEvent contains business logic for processing events
func (c *Consumer) handleEvent(ctx context.Context, event *models.Event) error {
	// Add your business logic here
	// For now, just log the event data
	time.Sleep(1 * time.Second)
	logger.Log.Debug("processing event", "data", event.Data)
	return nil
}

// Close gracefully shuts down the consumer
func (c *Consumer) Close(ctx context.Context) error {
	logger.Log.Info("shutting down consumer")
	close(c.done)
	c.wg.Wait()

	// Commit any uncommitted offsets
	if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
		logger.Log.Error("failed to commit final offsets", "error", err)
	}

	c.client.Close()
	logger.Log.Info("consumer shutdown complete")
	return nil
}
