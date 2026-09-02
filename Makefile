.PHONY: up down logs proto-gen migrate build run-trip run-driver run-pricing run-worker run-gateway test lint clean

# this above is commands 
# ============================================================
# Infrastructure
# ============================================================

## Start all infrastructure (PostgreSQL, Redis, Temporal)
up:
	docker compose up -d
	@echo "✅ Infrastructure is up"
	@echo "   PostgreSQL:   localhost:5432"
	@echo "   Redis:        localhost:6379"
	@echo "   Temporal:     localhost:7233"
	@echo "   Temporal UI:  http://localhost:8233"

## Stop all infrastructure
down:
	docker compose down

## Show logs for all containers
logs:
	docker compose logs -f

## Show status of all containers
ps:
	docker compose ps

# ============================================================
# Protobuf / Buf
# ============================================================

## Generate Go code from Protobuf definitions
proto-gen:
	cd proto && buf generate

## Lint Protobuf definitions
proto-lint:
	cd proto && buf lint

# ============================================================
# Database Migrations
# ============================================================

## Run all SQL migrations against PostgreSQL
migrate:
	@echo "Running migrations..."
	@for file in migrations/*.sql; do \
		echo "  Applying $$file..."; \
		PGPASSWORD=ridedispatch psql -h localhost -U ridedispatch -d ridedispatch -f $$file; \
	done
	@echo "✅ Migrations complete"

# ============================================================
# Build
# ============================================================

## Build all service binaries
build:
	go build -o bin/trip-service ./cmd/trip-service
	go build -o bin/driver-service ./cmd/driver-service
	go build -o bin/pricing-service ./cmd/pricing-service
	go build -o bin/worker ./cmd/worker
	go build -o bin/gateway ./cmd/gateway
	@echo "✅ All binaries built in ./bin/"

# ============================================================
# Run Services (for local development)
# ============================================================

## Run trip service
run-trip:
	go run ./cmd/trip-service

## Run driver service
run-driver:
	go run ./cmd/driver-service

## Run pricing service
run-pricing:
	go run ./cmd/pricing-service

## Run Temporal worker
run-worker:
	go run ./cmd/worker

## Run API gateway
run-gateway:
	go run ./cmd/gateway

# ============================================================
# Testing & Quality
# ============================================================

## Run all tests
test:
	go test ./... -v -race -count=1

## Run tests with coverage
test-cover:
	go test ./... -v -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

## Run linter
lint:
	golangci-lint run ./...

# ============================================================
# Cleanup
# ============================================================

## Remove build artifacts
clean:
	rm -rf bin/ coverage.* gen/
