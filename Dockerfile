FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY main.go ./
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -o caddy-gen .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/caddy-gen .

CMD ["./caddy-gen"]
