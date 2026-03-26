# Node stage: provide node/npm/yarn without apt installation.
FROM docker.m.daocloud.io/node:22-bookworm-slim AS node
# Build stage: compile frontend assets and Go binary.
FROM docker.m.daocloud.io/golang:1.24 AS builder

WORKDIR /app

COPY --from=node /usr/local/ /usr/local/

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

EXPOSE 80

CMD ["/app/multi-chain-faucet", "-multichain", "/app/multichain-config.json"]
