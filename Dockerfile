FROM golang:1.24-bullseye AS builder

ARG ROCKSDB_VERSION=10.2.1

# Install build dependencies
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential \
        cmake \
        git \
        libsnappy-dev \
        zlib1g-dev \
        libbz2-dev \
        liblz4-dev \
        libzstd-dev \
        libgflags-dev \
    && rm -rf /var/lib/apt/lists/*

# Build RocksDB
RUN git clone --branch v${ROCKSDB_VERSION} --depth 1 https://github.com/facebook/rocksdb.git /tmp/rocksdb \
    && cd /tmp/rocksdb \
    && cmake -DWITH_SNAPPY=ON -DWITH_ZLIB=ON -DWITH_LZ4=ON -DWITH_ZSTD=ON -DWITH_BZ2=ON -DCMAKE_BUILD_TYPE=Release . \
    && make -j$(nproc) \
    && make install \
    && ldconfig \
    && rm -rf /tmp/rocksdb

# Build service
WORKDIR /build/app
COPY app/go.mod app/go.sum ./
RUN go mod download
COPY app/ ./
RUN CGO_ENABLED=1 go build -o /bin/service ./cmd/server

# Build CLI
WORKDIR /build/cli
COPY cli/ ./
RUN go build -o /bin/cli ./

FROM debian:bullseye-slim
COPY --from=builder /usr/local/lib/ /usr/local/lib/
COPY --from=builder /usr/local/include/ /usr/local/include/
COPY --from=builder /bin/service /usr/local/bin/service
COPY --from=builder /bin/cli /usr/local/bin/cli
COPY app/configs/config.yml /etc/multi-cache/config.yml

RUN ldconfig

EXPOSE 8080
CMD ["service"]
