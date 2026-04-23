# Stage 1: Build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
RUN CGO_ENABLED=0 go build -o verifier main.go

# Stage 2: Final image
FROM alpine:latest
WORKDIR /
COPY --from=builder /app/verifier /verifier
# Run the binary
CMD ["/verifier"]
