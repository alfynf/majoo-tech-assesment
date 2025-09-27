# --- Build Stage ---
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
# -ldflags="-w -s" is used to make the binary smaller
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/main ./cmd/api

# --- Migration Tool Stage ---
# This stage downloads the migrate tool
FROM alpine:latest AS migrator
ARG MIGRATE_VERSION=v4.17.1
WORKDIR /
RUN apk add --no-cache curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz | tar xvz

# --- Final Stage ---
FROM alpine:latest

WORKDIR /app

# Install netcat for the migration script's wait-for-db logic
RUN apk add --no-cache netcat-openbsd

# Copy the Pre-built binary from the "builder" stage
COPY --from=builder /app/main .
# Copy the migrate tool from the "migrator" stage
COPY --from=migrator /migrate /usr/local/bin/migrate

# Copy the migration script
COPY run-migrations.sh .
# Copy migration files
COPY internal/migration ./internal/migration

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["/app/main"]
