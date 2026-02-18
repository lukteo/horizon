default: help

include makefiles/base.mk
include makefiles/utility.mk
include makefiles/generator.mk
include makefiles/migration.mk
include makefiles/docker.mk
include makefiles/init.mk

.PHONY: help
help: ## Display this help.
	@clear
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[.a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo ""
