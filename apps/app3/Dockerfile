FROM golang:1.22-alpine AS builder
WORKDIR /app
ENV GO111MODULE=on
RUN go mod init app3 && go get github.com/go-sql-driver/mysql
COPY main.go .
RUN CGO_ENABLED=0 go build -o app3-filter main.go

FROM alpine:latest
COPY --from=builder /app/app3-filter /app3-filter
EXPOSE 8081
CMD ["/app3-filter"]
