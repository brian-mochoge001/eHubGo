FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build tools
RUN apk add --no-cache git curl

# Copy modules first (best practice for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy all source files
COPY . .

# Build the main application
RUN CGO_ENABLED=0 GOOS=linux go build -v -o main .

# Build the seed script
RUN CGO_ENABLED=0 GOOS=linux go build -v -o seed ./cmd/seed/main.go

# Final stage
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates

# Copy the binaries from the builder
COPY --from=builder /app/main .
COPY --from=builder /app/seed .
COPY --from=builder /app/rbac_policy.csv .

EXPOSE 8080

# Run seed and then start the application
CMD ["/bin/sh", "-c", "./seed && ./main"]
