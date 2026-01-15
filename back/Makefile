BINARY_NAME=bounty-server

GO=go
GO_BUILD=$(GO) build
GO_RUN=$(GO) run
GO_TEST=$(GO) test
GO_MOD=$(GO) mod

.PHONY: all build run test clean run_build tidy dev help fmt lint docker-up docker-down docker-restart docker-logs docker-ps

all: build

# Build the application
build:
	$(GO_BUILD) -o $(BINARY_NAME) ./cmd/main.go

# Run the application directly (requires PostgreSQL running)
run:
	$(GO_RUN) ./cmd/main.go

# Run tests
test:
	$(GO_TEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GO_TEST) -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out

# Build and run the binary
run_build: build
	./$(BINARY_NAME)

# Tidy up go modules
tidy:
	$(GO_MOD) tidy

# Development mode with auto-reload (requires air)
# Install air: go install github.com/air-verse/air@latest
dev:
	air

# Format code
fmt:
	$(GO) fmt ./...

# Run linter (requires golangci-lint)
# Install: brew install golangci-lint or go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
lint:
	golangci-lint run

# Docker commands (run from project root)
docker-up:
	cd .. && docker-compose up -d

docker-down:
	cd .. && docker-compose down

docker-restart:
	cd .. && docker-compose restart

docker-logs:
	cd .. && docker-compose logs -f

docker-ps:
	cd .. && docker-compose ps

# Display help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application (requires PostgreSQL)"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make run_build      - Build and run the binary"
	@echo "  make tidy           - Tidy up go modules"
	@echo "  make dev            - Run in development mode (requires air)"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up      - Start all services with Docker Compose"
	@echo "  make docker-down    - Stop all services"
	@echo "  make docker-restart - Restart all services"
	@echo "  make docker-logs    - View logs from all services"
	@echo "  make docker-ps      - List running containers"
	@echo ""
	@echo "  make help           - Display this help message"
