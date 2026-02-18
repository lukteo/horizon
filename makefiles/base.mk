SHELL := bash

# Environment loading function to source appropriate .env file
define use_env
	@echo "Using .env.$(1) as ENV"
	@set -o allexport; source ./server/.env.$(1); set +o allexport
endef
