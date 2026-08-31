APP_NAME := unik-api
BIN_DIR := bin
TEST_DATABASE_URL ?= postgresql://unik_test:unik_test@127.0.0.1:55434/unik_test
TEST_REDIS_URL ?= redis://127.0.0.1:56380/0

.PHONY: help run build test test-race test-integration integration-up integration-down fmt vet check tidy compose-up compose-down compose-logs seed clean

help:
	@echo "run          Run API locally"
	@echo "build        Build API and seed binaries"
	@echo "check        Format, vet and test"
	@echo "compose-up   Build and start the full stack"
	@echo "seed         Replace local data with demo data"

run:
	go run ./cmd/api

build:
	mkdir -p $(BIN_DIR)
	go build -trimpath -o $(BIN_DIR)/$(APP_NAME) ./cmd/api
	go build -trimpath -o $(BIN_DIR)/unik-seed ./cmd/seed

test:
	go test ./...

test-race:
	go test -race ./...

ifeq ($(OS),Windows_NT)
test-integration:
	set RUN_INTEGRATION_TESTS=1&& set TEST_DATABASE_URL=$(TEST_DATABASE_URL)&& set TEST_REDIS_URL=$(TEST_REDIS_URL)&& go test -count=1 -v ./tests/integration
else
test-integration:
	RUN_INTEGRATION_TESTS=1 TEST_DATABASE_URL=$(TEST_DATABASE_URL) TEST_REDIS_URL=$(TEST_REDIS_URL) go test -count=1 -v ./tests/integration
endif

integration-up:
	docker compose -f docker-compose.test.yml up -d --wait

integration-down:
	docker compose -f docker-compose.test.yml down

fmt:
	gofmt -w cmd internal

vet:
	go vet ./...

check: fmt vet test

tidy:
	go mod tidy

compose-up:
	docker compose up --build -d

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f app

seed:
	docker compose --profile tools run --rm seed

clean:
	go clean
