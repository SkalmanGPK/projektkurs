FROM golang:1.22-alpine

WORKDIR /app

COPY main.go .

ENV GO111MODULE=off

RUN go build -o ids main.go

EXPOSE 8080

CMD ["./ids"]
