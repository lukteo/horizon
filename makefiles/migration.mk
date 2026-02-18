##@ Database Migration

.PHONY: mcreate mup mdown mstatus mreset

mcreate: ## Creates a sequential migration in the server directory using goose
	@read -p "Enter migration name: " name; \
		cd ./server/migrations && go tool goose -s create "$$name" sql

mup: ## Runs the latest migrations that have yet to be ran
	@$(call use_env,local) \
		&& cd ./server && \
		go tool goose -dir migrations postgres "$$DATABASE_URL" up && \
		go tool jet -dsn="$$DATABASE_URL" -schema=public -path=./generated
	@$(MAKE) format-server

mdown: ## Rollback database migrations by 1
	@$(call use_env,local) \
		&& cd ./server && \
		go tool goose -dir migrations postgres "$$DATABASE_URL" down

mstatus: ## Gets the migration status with goose
	@$(call use_env,local) \
		&& cd ./server && \
		go tool goose -dir migrations postgres "$$DATABASE_URL" status

mreset: ## Recreates the database (WARNING - This command drops the database)
	@$(call use_env,local) \
		&& cd ./server && \
		docker-compose exec postgres psql -U postgres -c "DROP DATABASE IF EXISTS $$DATABASE_NAME;" && \
		docker-compose exec postgres psql -U postgres -c "CREATE DATABASE $$DATABASE_NAME;"
