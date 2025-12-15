# Multi-stage build: compile frontend assets via go generate, then produce slim runtime image.
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

# System deps for node/yarn and CGO-less build
RUN apk add --no-cache git nodejs npm
RUN npm install -g yarn

# Pre-cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy sources
COPY . .

# Build frontend bundle (go:generate runs npm run build) and compile binary
RUN mkdir -p /out \
  && go generate ./... \
  && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/multi-chain-faucet .

FROM alpine:3.19
WORKDIR /app

RUN apk add --no-cache ca-certificates

# Copy binary and default config (override at runtime as needed)
COPY --from=builder /out/multi-chain-faucet /app/multi-chain-faucet
COPY multichain-config.json /app/multichain-config.json

EXPOSE 8080

# Use multichain mode by default; pass -httpport/-proxycount/other flags as needed.
CMD ["/app/multi-chain-faucet", "-multichain", "/app/multichain-config.json"]
