.DEFAULT_GOAL := help

# HOST is only used for API specs generation
HOST ?= localhost:8080

# Generates a help message. Borrowed from https://github.com/pydanny/cookiecutter-djangopackage.
help: ## Display this help message
	@echo "Please use \`make <target>' where <target> is one of"
	@perl -nle'print $& if m{^[\.a-zA-Z_-]+:.*?## .*$$}' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m  %-25s\033[0m %s\n", $$1, $$2}'

depends: ## Install & build dependencies
	go get ./...
	go build ./...
	go mod tidy

provision: depends ## Provision dev environment
	docker-compose up -d
	scripts/waitdb.sh
	@$(MAKE) migrate

dev: ## Bring up the server on dev environment with hot reload
	air

remove: ## Bring down the server on dev environment, remove all docker related stuffs as well
	docker-compose down -v --remove-orphans

migrate: ## Run database migrations
	go run cmd/migration/main.go

migrate.undo: ## Undo the last database migration
	go run cmd/migration/main.go --down

wire: ## Generate wire dependency injection code
	cd internal/di && GOFLAGS=-mod=mod wire

wire.check: ## Check if wire code is up to date
	cd internal/di && GOFLAGS=-mod=mod wire check

seed: ## Run database seeder
	echo "To be done!"

test: ## Run tests
	sh scripts/test.sh

test.cover: test ## Run tests and open coverage statistics page
	go tool cover -html=coverage-all.out

build: clean ## Build the server binary file on host machine
	sh scripts/build.sh

build.linux: ## Build the server binary file for Linux host
	@$(MAKE) GOOS=linux GOARCH=amd64 build

build.windows: ## Build the server binary file for Windows host
	@$(MAKE) GOOS=windows GOARCH=amd64 build

build.arm: clean ## Build the server binary file for ARM host
	GOOS=linux GOARCH=arm64 sh scripts/build-arm.sh

build.air:  ## Build the server binary file for air hot reload
	sh scripts/build-air.sh

clean: ## Clean up the built & test files
	rm -rf ./server ./*.out

specs: ## Generate swagger specs
	swag fmt -g /cmd/api/main.go
	swag fmt -d ./internal/api
	swag init --parseInternal --parseDependency --parseDepth 1 -g /cmd/api/main.go -o ./internal/api/docs
%: # prevent error for `up` target when passing arguments
ifeq ($(filter up,$(MAKECMDGOALS)),up)
	@:
else
	$(error No rule to make target `$@`.)
endif
