# Build stage
FROM golang:1.26.4 AS builder

WORKDIR /app

# Copy dependency files  first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install CA certificates if your app makes HTTPS requests
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]
