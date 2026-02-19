##@ Utility

.PHONY: lint format format-client format-server

lint: ## Lints the entire codebase
	# @cd ./mobile && bun run lint
	@cd ./server && golangci-lint run ./...

format: ## Formats the entire codebase
	# @$(MAKE) format-client
	@$(MAKE) format-server

format-client: ## Formats client code only
	@cd ./client && bun run format

format-server: ## Formats server code only
	@cd ./server && go tool goimports -w .
	@cd ./server && go tool golines -w .
	@cd ./server && gofmt -w .
