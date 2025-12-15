# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o machineserver ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/machineserver .
COPY --from=builder /app/configs ./configs

# Create necessary directories
RUN mkdir -p data logs data/storage/programs data/storage/snapshots

# Expose ports
EXPOSE 8080 9090

CMD ["./machineserver"]
