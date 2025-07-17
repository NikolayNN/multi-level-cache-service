# ---------- Stage 1: build Go (Alpine) --------------------
FROM golang:1.24-alpine AS builder
# инструменты для CGO + заголовки/библиотеки RocksDB
RUN apk add --no-cache build-base rocksdb-dev>=10.2

WORKDIR /app
COPY app/go.mod app/go.sum ./
RUN go mod download
COPY app/ .
RUN CGO_ENABLED=1 go build -o /bin/service ./cmd/server

WORKDIR /cli
COPY cli/ .
RUN go build -o /bin/cli .

# ---------- Stage 2: runtime ------------------------------
FROM alpine:edge

RUN apk add --no-cache curl # для работы healthcheck

RUN apk add --no-cache libstdc++ rocksdb>=10.2
COPY --from=builder /bin/service /usr/local/bin/service
COPY --from=builder /bin/cli     /usr/local/bin/cli

EXPOSE 8080
CMD ["service"]
