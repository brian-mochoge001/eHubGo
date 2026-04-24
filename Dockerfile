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

# Build the seed script
RUN CGO_ENABLED=0 GOOS=linux go build -o seed ./cmd/seed/main.go

# Final stage
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates

# Copy the binaries from the builder
COPY --from=builder /app/main .
COPY --from=builder /app/seed .

EXPOSE 8080

# By default, run the main app. 
# For seeding, use Render's "Pre-deploy command" or run ./seed manually
CMD ["./main"]
