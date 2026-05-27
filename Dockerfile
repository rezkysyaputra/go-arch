FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /out/api /app/api
COPY --from=builder --chown=nonroot:nonroot /out/migrate /app/migrate
COPY --chown=nonroot:nonroot migrations /app/migrations

USER nonroot:nonroot

EXPOSE 8080

CMD ["/app/api"]
