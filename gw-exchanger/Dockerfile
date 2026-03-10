FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app .

COPY config.env .
COPY internal/storage/migrations ./internal/storage/migrations

CMD ["./app"]