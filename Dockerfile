# ---------- Stage 1: build Go on Alpine Edge --------------------
FROM golang:1.24-alpine3.20 AS builder
# Используем edge-репозиторий для rocksdb 10.4
RUN echo "https://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories \
 && apk update && apk add --no-cache build-base rocksdb-dev

WORKDIR /app
COPY app/go.mod app/go.sum ./
RUN go mod download
COPY app/ ./
RUN CGO_ENABLED=1 go build -o /bin/service ./cmd/server

WORKDIR /cli
COPY cli/ ./
RUN go build -o /bin/cli .

# ---------- Stage 2: runtime on Alpine Edge ----------------------
FROM alpine:3.20
# Добавляем edge-community так же
RUN echo "https://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories \
 && apk update && apk add --no-cache curl libstdc++ rocksdb

COPY --from=builder /bin/service /usr/local/bin/service
COPY --from=builder /bin/cli     /usr/local/bin/cli

EXPOSE 8080
CMD ["service"]
