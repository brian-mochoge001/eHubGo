FROM golang:1.26-alpine AS builder

# Install necessary build tools and sqlc
RUN apk add --no-cache git curl
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate database bindings
RUN sqlc generate

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
