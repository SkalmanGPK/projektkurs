# Step 1: build the binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Since we don't use external libraries we're turning of modules for simplicity
ENV GO111MODULE=off

# Copy source code
COPY main.go .

# Build binary
RUN go build -o app1 main.go

# Step 2: create the final container
FROM alpine:latest

# Copys the binary from building stage
COPY --from=builder /app/app1 /app1

# Expose the port which the health server listens to
EXPOSE 8080

# Run application
CMD ["/app1"]
