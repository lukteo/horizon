##@ Initialization

.PHONY: init init-server init-client

init: ## Set up all required development tools for local use
	@echo "Starting development environment initialization..."
	@echo "[1/2] Initializing database..."
	@$(MAKE) init-db
	@echo "[2/3] Initializing server dependencies..."
	@$(MAKE) init-server
	@echo "[3/3] Initializing client dependencies..."
	# @$(MAKE) init-client
	@echo "Initialization complete!"

init-client: ## Initializes client dependencies
	@cd ./mobile && \
		if [ ! -f .env.local ]; then \
			echo "• Creating .env.local file..."; \
			cp .env.local.example .env.local; \
		else \
			echo "• Client .env.local already exists, skipping .env.local creation..."; \
		fi; \
		echo "• Installing mobile dependencies..."; \
		bun install --frozen-lockfile
	@echo "• Client initialization complete."

init-server: ## Initializes server dependencies
	@cd ./server && \
		if [ ! -f .env.local ]; then \
			echo "• Creating .env.local file..."; \
			cp .env.local.example .env.local; \
		else \
			echo "• Server .env.local already exists, skipping .env.local creation..."; \
		fi; \
		echo "• Installing go dependencies..."; \
		go mod tidy
	@echo "• Server initialization complete."

.PHONY: init-db
init-db: ## Initializes database (pgsql in docker)
	@echo ">> [1/3] Starting docker containers..."
	@$(MAKE) dup
	@echo ">> [2/3] Creating database..."
	@$(MAKE) mreset
	@echo ">> [3/3] Applying migrations..."
	@$(MAKE) mup
	@echo ">> Database initialization complete."
