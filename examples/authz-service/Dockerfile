# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Copy source code
COPY . .

# Download dependencies and generate go.sum
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o authz-service .

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/authz-service .

EXPOSE 8002

CMD ["./authz-service"]
