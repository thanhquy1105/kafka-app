# Kafka Producer/Consumer Application

A production-ready Kafka producer and consumer application built with [franz-go](https://github.com/twmb/franz-go).

## Features

### Producer
- ✅ Batch producing with configurable batch size and timeout
- ✅ Async and sync produce methods
- ✅ Automatic retry logic
- ✅ Graceful shutdown with message flushing
- ✅ Structured JSON logging

### Consumer
- ✅ Consumer group support with auto-rebalancing
- ✅ Configurable offset commit (auto/manual)
- ✅ Error handling and recovery
- ✅ Graceful shutdown with offset commit
- ✅ Structured JSON logging

## Project Structure

```
kafka-app/
├── cmd/
│   ├── producer/          # Producer binary
│   └── consumer/          # Consumer binary
├── internal/
│   ├── config/            # Configuration loader
│   ├── producer/          # Producer logic
│   ├── consumer/          # Consumer logic
│   └── models/            # Data models
├── pkg/
│   └── logger/            # Shared logger
├── configs/               # YAML configs
├── tests/                 # Unit tests
└── Makefile              # Build commands
```

## Prerequisites

- Go 1.21+
- Kafka cluster (or use Docker Compose)

## Quick Start

### 1. Install Dependencies

```bash
make deps
```

### 2. Configure

Edit `configs/producer.yaml` and `configs/consumer.yaml` with your Kafka broker addresses.

### 3. Build

```bash
make build
```

### 4. Run Producer

```bash
make run-producer
```

### 5. Run Consumer (in another terminal)

```bash
make run-consumer
```

## Configuration

### Producer (`configs/producer.yaml`)

```yaml
brokers:
  - localhost:19092
  - localhost:29092
  - localhost:39092
topic: events
batch_size: 1048576        # 1MB
batch_timeout_ms: 100
max_retries: 3
log_level: info
```

### Consumer (`configs/consumer.yaml`)

```yaml
brokers:
  - localhost:19092
  - localhost:29092
  - localhost:39092
topics:
  - events
group_id: kafka-app-consumer-group
auto_commit: true
commit_interval_ms: 5000
log_level: info
```

## Development

### Run Tests

```bash
make test
```

Tests use [kfake](https://pkg.go.dev/github.com/twmb/franz-go/pkg/kfake) for testing without a real Kafka cluster.

### Test with Coverage

```bash
make test-coverage
```

### Format Code

```bash
make fmt
```

### Clean Build Artifacts

```bash
make clean
```

## Usage Examples

### Producer

The producer automatically generates and sends events every second:

```go
event := &models.Event{
    ID:        uuid.New().String(),
    Type:      "test_event",
    Timestamp: time.Now(),
    Data: map[string]interface{}{
        "message": "Hello from Kafka!",
    },
}
```

### Consumer

The consumer processes events and logs them:

```json
{
  "level": "INFO",
  "msg": "consumed message",
  "topic": "events",
  "partition": 0,
  "offset": 42,
  "event_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

## Graceful Shutdown

Both producer and consumer handle `SIGINT` and `SIGTERM` signals gracefully:

- **Producer**: Flushes pending messages before shutdown
- **Consumer**: Commits offsets before shutdown

Press `Ctrl+C` to trigger graceful shutdown.

## Monitoring

Logs are output in JSON format for easy parsing:

```json
{
  "time": "2026-01-03T15:00:00Z",
  "level": "INFO",
  "msg": "produced message",
  "topic": "events",
  "partition": 2,
  "offset": 100,
  "event_id": "abc-123"
}
```

## Deployment

### Build Binaries

```bash
make build
```

Binaries will be in `bin/`:
- `bin/producer`
- `bin/consumer`

### Run in Production

```bash
# Producer
./bin/producer -config /path/to/producer.yaml

# Consumer
./bin/consumer -config /path/to/consumer.yaml
```

## Troubleshooting

### Connection Issues

Check broker addresses in config files match your Kafka cluster.

### Consumer Not Receiving Messages

1. Verify topic exists
2. Check consumer group ID
3. Verify producer is running
4. Check logs for errors

### Producer Errors

1. Check broker connectivity
2. Verify topic exists or auto-create is enabled
3. Check batch size and timeout settings

## License

MIT

## Resources

- [franz-go Documentation](https://pkg.go.dev/github.com/twmb/franz-go)
- [franz-go Examples](https://github.com/twmb/franz-go/tree/master/examples)
- [Kafka Documentation](https://kafka.apache.org/documentation/)
