GO ?= go
GOCACHE ?= .gocache
MIGRATION_NAME ?= $(NAME)
FEATURE_NAME ?= $(NAME)
FEATURE_PLURAL ?= $(PLURAL)
FEATURE_ARGS = -name $(FEATURE_NAME) $(if $(FEATURE_PLURAL),-plural $(FEATURE_PLURAL),)
SWAG_VERSION ?= v1.16.6

.PHONY: help install-tools test tidy swagger feature migrate-create migrate-up migrate-down docker-up docker-deps docker-down docker-build clean

help:
	@echo "Available commands:"
	@echo "  make install-tools                         Install local development tools"
	@echo "  make test                                  Run Go tests"
	@echo "  make tidy                                  Run go mod tidy"
	@echo "  make swagger                               Regenerate Swagger/OpenAPI docs"
	@echo "  make feature NAME=product                  Create Clean Architecture feature files"
	@echo "  make migrate-create NAME=create_table      Create SQL migration files"
	@echo "  make migrate-up                            Run database migrations"
	@echo "  make migrate-down                          Roll back one migration"
	@echo "  make docker-up                             Start full Docker stack"
	@echo "  make docker-deps                           Start PostgreSQL, RabbitMQ, and Redis only"
	@echo "  make docker-down                           Stop Docker stack"
	@echo "  make docker-build                          Build Docker images"
	@echo "  make clean                                 Remove local Go build cache"

install-tools:
	$(GO) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

test:
	GOCACHE=$(GOCACHE) $(GO) test ./...

tidy:
	GOCACHE=$(GOCACHE) $(GO) mod tidy

swagger:
	GOCACHE=$(GOCACHE) $(GO) run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init -g cmd/api/main.go -o docs --parseInternal

feature:
	@if [ -z "$(FEATURE_NAME)" ]; then echo "Usage: make feature NAME=product [PLURAL=products]"; exit 1; fi
	GOCACHE=$(GOCACHE) $(GO) run ./tools/featuregen $(FEATURE_ARGS)
	GOCACHE=$(GOCACHE) $(GO) fmt ./...

migrate-create:
	@if [ -z "$(MIGRATION_NAME)" ]; then echo "Usage: make migrate-create NAME=create_products_table"; exit 1; fi
	migrate create -ext sql -dir migrations -seq $(MIGRATION_NAME)

migrate-up:
	GOCACHE=$(GOCACHE) $(GO) run ./cmd/migrate up

migrate-down:
	GOCACHE=$(GOCACHE) $(GO) run ./cmd/migrate down

docker-up:
	docker compose up --build

docker-deps:
	docker compose up -d postgres rabbitmq redis

docker-down:
	docker compose down --remove-orphans

docker-build:
	docker compose build

clean:
	rm -rf $(GOCACHE)
