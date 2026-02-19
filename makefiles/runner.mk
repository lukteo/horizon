##@ Runners

.PHONY: rclient rserver

rclient: ## Starts client servers
	@echo "Starting mobile..."
	@cd ./client && bun run dev

rserver: ## Starts the server
	@echo "Starting server, press Ctrl + C to stop..."
	@$(call use_env,local) \
		&& cd ./server && \
		go run ./cmd/server/main.go
