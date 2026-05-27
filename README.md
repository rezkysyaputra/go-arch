# Go Clean Architecture Boilerplate

Scalable Golang backend boilerplate using Gin, RabbitMQ, PostgreSQL, GORM, Redis, Viper, Zerolog, golang-migrate, and strict manual dependency injection.

## Layout

```text
cmd/api                  application entrypoint and dependency wiring
internal/config          environment configuration
internal/domain          domain models and interfaces only
internal/delivery/http   Gin handlers and routes
internal/delivery/amqp   RabbitMQ publisher and background consumers
internal/repository      GORM repository implementations
internal/usecase         business logic
```

## Run

Start the full stack with Docker:

```bash
docker compose up --build
```

RabbitMQ Management UI is available at `http://localhost:15672` with `guest` / `guest`.
Swagger UI is available at `http://localhost:8080/swagger/index.html`.

## Local Development

1. Copy `.env.example` to `.env`.
2. Start PostgreSQL, RabbitMQ, and Redis.

```bash
docker compose up -d postgres rabbitmq redis
```

3. Run migrations:

```bash
go run ./cmd/migrate up
```

4. Run the API:

```bash
go run ./cmd/api
```

PostgreSQL can be opened from DBeaver with:

```text
Host: localhost
Port: 5432
Database: go_arch
Username: postgres
Password: postgres
```

## Tooling

Install the migration CLI if it is not available yet:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Check it:

```bash
migrate -version
```

## Endpoints

```text
GET  /health
GET  /health/live
GET  /health/ready
POST /users
GET  /users/:id
```

Example request:

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Ada Lovelace","email":"ada@example.com"}'
```

## Response Format

Success:

```json
{
  "success": true,
  "message": "user created successfully",
  "data": {},
  "request_id": "47b40210c3d4b4930e9b20ce70961206"
}
```

Error:

```json
{
  "success": false,
  "message": "request validation failed",
  "errors": [
    {
      "field": "email",
      "code": "email",
      "message": "email must be a valid email address"
    }
  ],
  "request_id": "47b40210c3d4b4930e9b20ce70961206"
}
```

## API Docs

Swagger UI:

```text
http://localhost:8080/swagger/index.html
```

Regenerate docs after changing handlers or annotations:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs --parseInternal
```

## Migrations

Create a new migration file pair:

```bash
migrate create -ext sql -dir migrations -seq create_products_table
```

Or with Make:

```bash
make migrate-create NAME=create_products_table
```

This creates:

```text
migrations/000002_create_products_table.up.sql
migrations/000002_create_products_table.down.sql
```

Run migrations:

```bash
go run ./cmd/migrate up
```

Rollback one migration:

```bash
go run ./cmd/migrate down
```

## Feature Generator

Create boilerplate files for a new feature:

```bash
make feature NAME=product
```

For custom plural route/table names:

```bash
make feature NAME=category PLURAL=categories
```

The generator creates:

```text
internal/domain/{name}.go
internal/usecase/{name}_usecase.go
internal/repository/{name}_repository.go
internal/delivery/http/{name}_handler.go
migrations/{next}_create_{plural}_table.up.sql
migrations/{next}_create_{plural}_table.down.sql
```

After generating, wire the new repository/usecase manually in `cmd/api/main.go`, pass the usecase into `internal/delivery/http.NewRouter`, and call `Register{Name}Routes` in `internal/delivery/http/router.go`.

## Tests

```bash
go test ./...
```

## Useful Commands

```bash
make help
make install-tools
make test
make tidy
make swagger
make feature NAME=product
make migrate-create NAME=create_products_table
make migrate-up
make migrate-down
make docker-up
make docker-deps
make docker-down
make docker-build
```

Direct commands:

```bash
go test ./...
go mod tidy
migrate create -ext sql -dir migrations -seq create_products_table
go run ./cmd/migrate up
go run ./cmd/migrate down
docker compose up --build
docker compose up -d postgres rabbitmq redis
docker compose down --remove-orphans
docker compose build
```
