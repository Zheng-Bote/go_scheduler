# Build Stage
FROM golang:1.25-bullseye AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN go build -o scheduler ./cmd/scheduler && \
    go build -o encrypt-config ./cmd/encrypt-config && \
    go build -o job1 ./cmd/job1 && \
    go build -o job2 ./cmd/job2

# Final Stage
FROM ubuntu:latest

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/scheduler .
COPY --from=builder /app/encrypt-config .
COPY --from=builder /app/job1 .
COPY --from=builder /app/job2 .
COPY --from=builder /app/migrations ./migrations

# Expose HTTP port (default is 8080)
EXPOSE 8080

# The scheduler expects the path to the encrypted config as an argument
# CMD ["./scheduler", "/app/config.json.enc"]
