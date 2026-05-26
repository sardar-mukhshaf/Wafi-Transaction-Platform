.PHONY: help local-up local-down local-logs proto build test test-integration test-contract lint fmt clean

help: ## Show this help message
	@echo "Saudi Distributed Commerce & Settlement Fabric"
	@echo "================================================"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

local-up: ## Start local development stack (Docker Compose)
	cd scripts/local-setup && docker-compose up -d --build

local-down: ## Stop local development stack
	cd scripts/local-setup && docker-compose down -v

local-logs: ## Tail logs from local stack
	cd scripts/local-setup && docker-compose logs -f

proto: ## Generate code from protobuf definitions
	@echo "Generating Go protobuf code..."
	@mkdir -p libs/kafka-clients/go/proto
	@protoc --go_out=libs/kafka-clients/go/proto --go_opt=paths=source_relative \
		--go-grpc_out=libs/kafka-clients/go/proto --go-grpc_opt=paths=source_relative \
		-I libs/proto \
		libs/proto/*.proto

build: ## Build all service images
	cd apps/order-service && docker build -t fabric/order-service:latest .
	cd apps/payment-service && docker build -t fabric/payment-service:latest .
	cd apps/catalog-service && docker build -t fabric/catalog-service:latest .

test: ## Run unit tests for all services
	cd apps/order-service && go test ./...
	cd apps/payment-service && ./mvnw test
	cd apps/catalog-service && npm test

test-integration: ## Run integration tests against local stack
	@echo "Running integration tests..."
	cd apps/order-service && go test -tags=integration ./...

test-contract: ## Run Pact contract tests
	@echo "Running contract tests..."
	cd apps/order-service && go test -tags=contract ./...

lint: ## Lint all services
	cd apps/order-service && golangci-lint run ./...
	cd apps/payment-service && ./mvnw spotbugs:check
	cd apps/catalog-service && npm run lint

fmt: ## Format all code
	cd apps/order-service && gofmt -w .
	cd apps/catalog-service && npm run format

clean: ## Clean build artifacts and Docker volumes
	cd scripts/local-setup && docker-compose down -v --remove-orphans
	docker system prune -f
