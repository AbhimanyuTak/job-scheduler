# Job Scheduler Makefile

.PHONY: help build run test clean docker-build docker-run docker-dev docker-stop

# Default target
help:
	@echo "Job Scheduler - Available Commands:"
	@echo ""
	@echo "Development:"
	@echo "  make build          - Build the Go application"
	@echo "  make run            - Run the application locally"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run full application stack with Docker"
	@echo "  make docker-dev     - Start development database only"
	@echo "  make docker-stop    - Stop all Docker services"
	@echo ""
	@echo "Database:"
	@echo "  make db-setup       - Setup development database"
	@echo "  make db-reset       - Reset database (remove volumes)"
	@echo ""
	@echo "Testing:"
	@echo "  make test-api       - Test API endpoints"

# Development commands
build:
	@echo "Building application..."
	go build -o bin/scheduler ./cmd/scheduler

run:
	@echo "Running application..."
	go run ./cmd/scheduler

test:
	@echo "Running tests..."
	go test ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t job-scheduler .

docker-run:
	@echo "Starting full application stack..."
	./scripts/docker-run.sh

docker-dev:
	@echo "Starting development database..."
	./scripts/docker-setup.sh

docker-stop:
	@echo "Stopping Docker services..."
	docker compose down
	docker compose -f docker-compose.dev.yml down

# Database commands
db-setup:
	@echo "Setting up development database..."
	./scripts/docker-setup.sh

db-reset:
	@echo "Resetting database..."
	docker compose down -v
	docker compose -f docker-compose.dev.yml down -v

# API testing
test-api:
	@echo "Testing API endpoints..."
	./scripts/test-api.sh

# Development workflow
dev: docker-dev
	@echo "Development environment ready!"
	@echo "Run 'make run' to start the application"

# Production build
prod: docker-build
	@echo "Production build complete!"
