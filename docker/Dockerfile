# Build stage: compile frontend assets and Go binary.
FROM docker.m.daocloud.io/golang:1.24 AS builder

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends git nodejs npm \
    && npm install -g yarn \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN mkdir -p /out \
    && go generate ./... \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/multi-chain-faucet .

# Runtime stage: only ship the binary and default config.
FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/multi-chain-faucet /app/multi-chain-faucet
COPY multichain-config.json /app/multichain-config.json

EXPOSE 8080

CMD ["/app/multi-chain-faucet", "-multichain", "/app/multichain-config.json"]
