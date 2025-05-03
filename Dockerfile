# --- Builder Stage ---
    FROM golang:1.24.2-bookworm AS builder

    RUN apt-get update && apt-get install -y \
        build-essential libpq-dev
    
    ENV CGO_ENABLED=1
    ENV CC=gcc
    
    WORKDIR /app
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    
    RUN go build -o webhook-service ./cmd/webhook-service
    
    
    # --- Runtime Stage ---
    FROM debian:bookworm-slim
    
    # Install minimal runtime dependencies
    RUN apt-get update && apt-get install -y \
        ca-certificates && \
        apt-get clean && \
        rm -rf /var/lib/apt/lists/*
    
    WORKDIR /app
    
    COPY --from=builder /app/webhook-service /app/webhook-service
    
    EXPOSE 8080
    
    CMD ["./webhook-service"]