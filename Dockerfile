FROM golang:1.25-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o rates-service ./cmd/service/main.go

FROM alpine:latest

RUN apk add --no-cache curl

WORKDIR /app

COPY --from=builder /build/rates-service /app/rates-service

COPY repository/db/migrations ./repository/db/migrations
COPY configs/pairs.json ./configs/pairs.json

EXPOSE 8080 5001

ENV MIGRATIONS_DIR=/app/repository/db/migrations
ENV GRINEX_PAIRS_JSON=/app/configs/pairs.json

CMD ["/app/rates-service"]
