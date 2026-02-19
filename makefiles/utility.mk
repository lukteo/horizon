##@ Utility

.PHONY: lint format format-client format-server format-openapi

lint: ## Lints the entire codebase
	@cd ./client && bun run check
	@cd ./server && golangci-lint run ./...

format: ## Formats the entire codebase
	@$(MAKE) format-client
	@$(MAKE) format-server

format-client: ## Formats client code only
	@cd ./client && bun run format

format-server: ## Formats server code only
	@cd ./server && go tool goimports -w .
	@cd ./server && go tool golines -w .
	@cd ./server && gofmt -w .

format-openapi: ## Formats openapi yaml
	@cd ./client && bun run format:openapi
