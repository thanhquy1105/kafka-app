.PHONY: all build test clean run-producer run-consumer

# Build variables
BINARY_DIR=bin
PRODUCER_BINARY=$(BINARY_DIR)/producer
CONSUMER_BINARY=$(BINARY_DIR)/consumer

all: build

# Build both binaries
build:
	@echo "Building producer..."
	@mkdir -p $(BINARY_DIR)
	@go build -o $(PRODUCER_BINARY) ./cmd/producer
	@echo "Building consumer..."
	@go build -o $(CONSUMER_BINARY) ./cmd/consumer
	@echo "Build complete!"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./tests/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./tests/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Run producer
run-producer: build
	@echo "Starting producer..."
	@$(PRODUCER_BINARY) -config configs/producer.yaml

# Run consumer
run-consumer: build
	@echo "Starting consumer..."
	@$(CONSUMER_BINARY) -config configs/consumer.yaml

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Run both producer and consumer (in separate terminals)
run-all:
	@echo "Run 'make run-producer' in one terminal"
	@echo "Run 'make run-consumer' in another terminal"
