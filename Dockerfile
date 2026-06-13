# Stage 1: Build the Go binary using the official Go Alpine image
FROM golang:1.25-alpine AS builder

# Install git and certificates (needed for secure outbound webhooks to Telegram)
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Leverage Docker cache for dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your application code
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main .

# Stage 2: Final lightweight runtime footprint
FROM alpine:latest  
RUN apk add --no-cache tzdata
ENV TZ=Europe/Warsaw
WORKDIR /root/

# Copy trusted CA certificates from the builder stage so secure HTTPS calls don't fail
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled production binary
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]