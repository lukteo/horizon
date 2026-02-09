.PHONY: help
help: ## Display this help message
	@echo "Horizon - Security Analytics Platform"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: setup
setup: ## Setup the development environment
	@echo "Setting up development environment..."
	go mod tidy
	cd client && bun install

.PHONY: build
build: ## Build the server binary
	@echo "Building server binary..."
	go build -o bin/horizon-server ./cmd/api/main.go

.PHONY: build-client
build-client: ## Build the client application
	@echo "Building client application..."
	cd client && bun run build

.PHONY: build-all
build-all: build build-client ## Build both server and client

.PHONY: run
run: ## Run the server locally
	@echo "Running server..."
	go run ./cmd/api/main.go

.PHONY: run-client
run-client: ## Run the client application
	@echo "Running client application..."
	cd client && bun run dev

.PHONY: run-docker
run-docker: ## Run the full application with Docker Compose
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

.PHONY: stop-docker
stop-docker: ## Stop Docker Compose services
	@echo "Stopping services with Docker Compose..."
	docker-compose down

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	go test ./...

.PHONY: test-v
test-v: ## Run tests with verbose output
	@echo "Running tests with verbose output..."
	go test -v ./...

.PHONY: migrate-up
migrate-up: ## Run database migrations
	@echo "Running database migrations..."
	cd server && go run tool github.com/pressly/goose/v3/cmd/goose -dir migrations postgres "$$DATABASE_URL" up

.PHONY: migrate-down
migrate-down: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	cd server && go run tool github.com/pressly/goose/v3/cmd/goose -dir migrations postgres "$$DATABASE_URL" down

.PHONY: gen-models
gen-models: ## Generate database models with gojet
	@echo "Generating database models with gojet..."
	cd server && go run tool github.com/go-jet/jet/v2/cmd/jet generate -dsn="$$DATABASE_URL" -schemas=public -path=./internal/modelz -tables=log_mappings,raw_logs,normalized_logs,sigma_rules,alerts,incidents

.PHONY: gen-openapi-models
gen-openapi-models: ## Generate OpenAPI models
	@echo "Generating OpenAPI models..."
	cd server && go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest -generate types,server -package models openapi.yaml > internal/generated/openapi.go

.PHONY: gen-client
gen-client: ## Generate client code from OpenAPI spec
	@echo "Generating client code..."
	cd server && go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest -generate client -package client openapi.yaml > ../client/src/generated/client.go

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

.PHONY: deps
deps: ## Install dependencies
	@echo "Installing Go dependencies..."
	go mod tidy
	@echo "Installing client dependencies..."
	cd client && bun install

.PHONY: docker-build
docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker-compose build

.PHONY: logs
logs: ## Show Docker logs
	@echo "Showing Docker logs..."
	docker-compose logs -f

.PHONY: reset-db
reset-db: ## Reset the database
	@echo "Resetting database..."
	docker-compose down -v
	docker-compose up -d postgres
	sleep 10
	make migrate-up