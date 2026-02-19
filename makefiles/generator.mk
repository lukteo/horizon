##@ Code Generation

.PHONY: gen-openapi

gen-openapi: ## Generates code based on OpenAPI specification
	@echo "Generating client code from ~/openapi/openapi.yaml"
	@cd ./client && bun run gen:openapi
	@echo "Generating server code from ~/openapi/openapi.yaml"
	@cd ./server && go tool oapi-codegen -config ./generated/oapi/config.yml ../openapi/openapi.yml
	@echo "OpenAPI generation complete"
