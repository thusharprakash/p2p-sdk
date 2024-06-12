# Base image
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# Install dependencies
RUN go mod download

# Build the application
RUN go build -o main

# Final image
FROM alpine:latest

# Copy application binary
COPY --from=builder /app/main /app/main

# Expose port (adjust if needed)
EXPOSE 62235

# Set command to run the application
CMD ["/app/main"]
