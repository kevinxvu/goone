# Stage 1: Build stage
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git make gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/server ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/migrate ./cmd/migration

# Stage 2: Runtime stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrate /app/migrate
COPY --from=builder /app/internal/migrations /app/internal/migrations
COPY --from=builder /app/.env /app/.env
COPY --from=builder /app/scripts/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app
USER appuser

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/server"]
