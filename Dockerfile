# Stage 1: Build
FROM golang:1.22-alpine AS builder

# Install git, needed for go modules
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary (adjust if your entrypoint is different)
RUN go build -o server ./cmd

# Stage 2: Minimal runtime image
FROM alpine:latest


WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy static frontend files (optional)
# COPY --from=builder /app/public ./public

# Copy production env file
COPY .env .env

# Expose the port your app listens on (change if needed)
EXPOSE 6666

# Run the Go binary
CMD ["./server"]
