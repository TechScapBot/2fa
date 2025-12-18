FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY *.go ./

RUN go build -ldflags="-s -w" -o 2fa-api .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/2fa-api .

EXPOSE 8080

CMD ["./2fa-api"]
