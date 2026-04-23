FROM golang:1.26-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder
COPY --from=builder /app/main .
COPY --from=builder /app/schema.sql .
COPY --from=builder /app/cmd/migrate ./cmd/migrate
COPY --from=builder /app/cmd/seed ./cmd/seed

EXPOSE 8080

# Run migrations and seed, then start the server
CMD ["sh", "-c", "go run cmd/migrate/main.go && go run cmd/seed/main.go && ./main"]
