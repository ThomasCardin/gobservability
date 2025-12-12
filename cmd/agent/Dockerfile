FROM golang:1.24-alpine AS builder

ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o agent ./cmd/agent

# Use Ubuntu for perf tools and flamegraph dependencies
FROM ubuntu:22.04

# Install perf tools and flamegraph dependencies
RUN apt-get update && apt-get install -y \
    linux-tools-generic \
    linux-tools-common \
    git \
    perl \
    && rm -rf /var/lib/apt/lists/*

# Install FlameGraph tools
RUN git clone https://github.com/brendangregg/FlameGraph.git /opt/FlameGraph
RUN ln -s /opt/FlameGraph/stackcollapse-perf.pl /usr/local/bin/stackcollapse-perf.pl
RUN ln -s /opt/FlameGraph/flamegraph.pl /usr/local/bin/flamegraph.pl

# Create symlink for perf (the exact perf version depends on kernel)
RUN ln -sf /usr/lib/linux-tools/*/perf /usr/local/bin/perf

WORKDIR /root/

COPY --from=builder /app/agent .

CMD ["./agent"]