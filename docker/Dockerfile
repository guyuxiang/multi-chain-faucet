# Use golang base image
FROM reg.hrlyit.com/base/golang:1.23.10-alpine3.22-s
RUN apk add libstdc++ libgcc

# Install dependencies
RUN apk add --no-cache git nodejs npm

# Install yarn globally
RUN npm install -g yarn

# Set working directory
WORKDIR /app

# Copy entire project
COPY . .

# Generate frontend bundle
RUN go generate

# Build the application
RUN go build -o multi-chain-faucet

# Expose port
EXPOSE 8080

# Run the application with multichain config
CMD ["./multi-chain-faucet", "-multichain", "multichain-config.json"]