FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install required system packages
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the applications
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o client ./cmd/client

# Use a smaller image for the final build
FROM alpine:3.18

WORKDIR /app

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binaries from the builder stage
COPY --from=builder /app/server .
COPY --from=builder /app/client .
COPY --from=builder /app/db ./db

# Set the entry point
ENTRYPOINT ["/app/server"] 