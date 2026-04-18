##@ Testing

.PHONY: test-server test-db-up test-db-reset

test-db-up: ## Ensures the test database exists and has the latest migrations applied
	@$(call use_env,local) \
		&& docker exec horizon-postgres psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$$TEST_DATABASE_NAME'" | grep -q 1 \
		|| docker exec horizon-postgres psql -U postgres -c "CREATE DATABASE $$TEST_DATABASE_NAME"
	@$(call use_env,local) \
		&& cd ./server \
		&& go tool goose -dir migrations postgres "$$TEST_DATABASE_URL" up

test-server: test-db-up ## Runs the server Go test suite against the test database
	@$(call use_env,local) \
		&& cd ./server \
		&& go test -p 1 ./...

test-db-reset: ## Drops and recreates the test database (WARNING - destructive)
	@$(call use_env,local) \
		&& docker exec horizon-postgres psql -U postgres -c "DROP DATABASE IF EXISTS $$TEST_DATABASE_NAME"
	@$(MAKE) test-db-up
