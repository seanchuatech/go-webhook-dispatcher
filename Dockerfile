# Stage 1: Build the statically linked Go binary
FROM golang:alpine AS builder

# Install CA certificates to enable HTTPS requests to webhooks
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the go.mod and go.sum files first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary statically
# CGO_ENABLED=0 disables CGO, making the binary fully statically linked.
# -ldflags="-w -s" removes debugging info, reducing binary size drastically.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o webhook-dispatcher ./cmd/webhook-dispatcher

# Stage 2: Create the minimal production image
# 'scratch' is an empty image, providing maximum security and minimum size.
FROM scratch

# Copy root certificates so HTTPS calls work
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the pre-built binary
COPY --from=builder /app/webhook-dispatcher /webhook-dispatcher

# The application listens on port 8080 by default
EXPOSE 8080

# Specify the entrypoint command that runs when the container starts
ENTRYPOINT ["/webhook-dispatcher"]
