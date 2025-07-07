FROM golang:1.23 as builder

# Set environment variables for Go
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o main ./cmd/server/main.go

# Use a minimal base image for running the application
FROM alpine:latest

# Set up certificates (if needed)
RUN apk add --no-cache ca-certificates

# Set the working directory inside the container
WORKDIR /root/

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/main .

# Copy other required files (e.g., configuration files, .env)
COPY --from=builder /app/config ./config

# Expose the port the app runs on
EXPOSE 8000

# Run the binary
CMD ["./main"]
