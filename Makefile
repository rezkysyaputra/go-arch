GO ?= go
GOCACHE ?= $(CURDIR)/.gocache
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
	set GOCACHE=$(GOCACHE)&& $(GO) test ./...

tidy:
	set GOCACHE=$(GOCACHE)&& $(GO) mod tidy

swagger:
	set GOCACHE=$(GOCACHE)&& $(GO) run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init -g cmd/api/main.go -o docs --parseInternal

feature:
	set GOCACHE=$(GOCACHE)&& $(GO) run ./tools/featuregen $(FEATURE_ARGS)
	set GOCACHE=$(GOCACHE)&& $(GO) fmt ./...

migrate-create:
	migrate create -ext sql -dir migrations -seq $(MIGRATION_NAME)

migrate-up:
	set GOCACHE=$(GOCACHE)&& $(GO) run ./cmd/migrate up

migrate-down:
	set GOCACHE=$(GOCACHE)&& $(GO) run ./cmd/migrate down

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
