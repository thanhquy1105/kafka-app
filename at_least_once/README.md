# Kafka Producer/Consumer Application (Boilerplate)

A production-ready Kafka application built with [franz-go](https://github.com/twmb/franz-go), designed as a clean, structured boilerplate for Go-based Kafka services.

---

## 🏗️ Architecture & Semantics

### Message Delivery: **At-least-once**
This application currently implements **At-least-once** semantics. This is the most common pattern for high-performance applications where speed is a priority and the business logic can handle occasional duplicate messages (idempotency).

- **Producer:** Retries sending messages until it receives an acknowledgement (ACK) from the Kafka broker.
- **Consumer:** Manual offset committing is performed *after* the message has been successfully processed. If the consumer crashes during processing, it will re-read the last uncommitted messages upon restart.

> [!TIP]
> If you need **Exactly-Once Semantics (EOS)**, refer to the transactional example in `../examples/transactions/eos/`.

### Tech Stack
- **Language:** Go 1.21+
- **Library:** `franz-go` (High-performance, feature-rich Kafka client).
- **Format:** Structured JSON Logging for ELK/Grafana integration.

---

## 📂 Project Structure

```text
kafka-app/
├── cmd/
│   ├── producer/          # Entry point for the Producer service
│   └── consumer/          # Entry point for the Consumer service
├── internal/
│   ├── config/            # YAML configuration mapping and loading
│   ├── producer/          # Core Producer logic (connection handling, async sending)
│   ├── consumer/          # Core Consumer logic (polling, record processing)
│   └── models/            # Shared data structures (Event schema)
├── pkg/
│   └── logger/            # Zap-based structured logger
├── configs/               # Configuration files (.yaml)
├── tests/                 # Integration and unit tests using kfake
└── Makefile              # Automation for build, run, and test
```

---

## 🚀 Getting Started

### 1. Prerequisites
- Docker & Docker Compose (for running a local Kafka cluster)
- Go 1.21+ installed locally

### 2. Spin up Kafka
Use the provided `docker-compose.yml` to start a 3-node Kafka cluster:
```bash
docker-compose up -d
```

### 3. Build & Run
**Build the binaries:**
```bash
make build
```

**Run the Producer:** (Generates mock events every second)
```bash
./bin/producer -config configs/producer.yaml
```

**Run the Consumer:** (Processes and logs events)
```bash
./bin/consumer -config configs/consumer.yaml
```

---

## 🔧 Configuration

The application is highly configurable via YAML files in the `configs/` directory.

### Key Producer Settings
- `batch_size`: Maximum bytes before sending a batch (default 1MB).
- `batch_timeout_ms`: Maximum time to wait before sending a partial batch.

### Key Consumer Settings
- `group_id`: Identifies the consumer group for balanced processing.
- `auto_commit`: Toggles between automatic or manual offset committing.

---

## 🧪 Testing
We use `kfake` to simulate Kafka without needing a real cluster for unit tests.

```bash
make test          # Run all tests
make test-coverage # Check code coverage
```

---

## 💡 Best Practices Implemented
- **Graceful Shutdown:** Handles `SIGINT`/`SIGTERM` to flush the producer's buffer and commit consumer offsets before exiting.
- **Structured Logging:** All logs are in JSON format, making them easy to parse in production environments.
- **Clean Architecture:** Separates business logic from the Kafka transport layer.

---

## 📄 License
MIT
