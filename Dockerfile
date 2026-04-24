FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build tools and sqlc
RUN apk add --no-cache git curl
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Copy modules first (best practice for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy all source files
COPY . .

# Generate database bindings
RUN sqlc generate

# Build the main application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
