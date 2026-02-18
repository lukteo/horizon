##@ Docker
DC = docker compose
PROJECT = horizon

.PHONY: dup ddown drestart dps dlogs dshell-db dshell-client dclean

dup: ## Spin up all containers in background
	@$(DC) up -d

ddown: ## Stop and remove all containers
	@$(DC) down

drestart: ## Restart all containers
	@$(DC) restart

dps: ## List running containers and their status
	@$(DC) ps

dlogs: ## Follow logs for all services
	@$(DC) logs -f

dbuild: ## Rebuild images without using cache
	@$(DC) build --no-cache

dclean: ## Stop containers and wipe volumes (DESTRUCTIVE)
	@$(DC) down -v

# Service Specific Access
dshell-db: ## Enter the Postgres CLI inside the container
	@$(DC) exec postgres psql -U postgres -d horizon

dshell-client: ## Enter the Bun shell inside the client container
	@$(DC) exec client /bin/sh

dstats: ## Show real-time CPU/Memory usage of containers
	@docker stats
