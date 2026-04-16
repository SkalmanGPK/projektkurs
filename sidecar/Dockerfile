FROM golang:1.26-alpine AS builder
WORKDIR /app

# Kopiera båda mod-filerna först
COPY go.mod go.sum ./

# Ladda ner beroenden
RUN go mod download

# Kopiera källkoden
COPY main.go ./

# Bygg binären
RUN CGO_ENABLED=0 GOOS=linux go build -o pinger main.go

FROM alpine:latest
WORKDIR /
COPY --from=builder /app/pinger /pinger

EXPOSE 8080
ENTRYPOINT ["/pinger"]
