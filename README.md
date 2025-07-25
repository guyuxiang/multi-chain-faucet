# eth-faucet

[![Build](https://img.shields.io/github/actions/workflow/status/chainflag/eth-faucet/build.yml?branch=main)](https://github.com/chainflag/eth-faucet/actions/workflows/build.yml)
[![Release](https://img.shields.io/github/v/release/chainflag/eth-faucet)](https://github.com/chainflag/eth-faucet/releases)
[![Report](https://goreportcard.com/badge/github.com/chainflag/eth-faucet)](https://goreportcard.com/report/github.com/chainflag/eth-faucet)
[![Go](https://img.shields.io/github/go-mod/go-version/chainflag/eth-faucet)](https://go.dev/)
[![License](https://img.shields.io/github/license/chainflag/eth-faucet)](https://github.com/chainflag/eth-faucet/blob/main/LICENSE)

The faucet is a web application with the goal of distributing small amounts of Ether in private and test networks.

## Features

* Configure the funding account using a private key or keystore
* Implement CAPTCHA verification to prevent abuse
* Rate-limit requests by ETH address and IP address to prevent spam
* Prevent X-Forwarded-For spoofing by specifying the number of reverse proxies

## Get started

### Prerequisites

* Go (version 1.17 or later)
* Node.js

### Installation

1. Clone the repository and navigate to the app’s directory
```bash
git clone https://github.com/chainflag/eth-faucet.git
cd eth-faucet
```

2. Bundle frontend with Vite
```bash
go generate
```

3. Build Go project 
```bash
go build -o eth-faucet
```

## Usage

### Supported Networks

The faucet supports multiple blockchain networks. Use the `-list-networks` flag to see all available networks:

```bash
./eth-faucet -list-networks
```

Supported networks include:
- **Ethereum**: mainnet, sepolia, holesky, goerli
- **Polygon**: polygon, polygon-mumbai, polygon-amoy  
- **BSC**: bsc, bsc-testnet
- **Arbitrum**: arbitrum, arbitrum-sepolia
- **Optimism**: optimism, optimism-sepolia
- **Avalanche**: avalanche, avalanche-fuji
- **Base**: base, base-sepolia
- **Fantom**: fantom, fantom-testnet
- **Linea**: linea, linea-sepolia
- **zkSync**: zksync, zksync-sepolia

### Running the Faucet

**Use a private key**

```bash
./eth-faucet -httpport 8080 -wallet.provider http://localhost:8545 -wallet.privkey privkey -faucet.name sepolia
```

**Use a keystore**

```bash
./eth-faucet -httpport 8080 -wallet.provider http://localhost:8545 -wallet.keyjson keystore -wallet.keypass password.txt -faucet.name polygon-mumbai
```

**Network auto-configuration**

When using a supported network name, the faucet will automatically set the correct chain ID and token symbol:

```bash
# For Polygon Mumbai testnet - automatically sets MATIC symbol and chain ID 80001
./eth-faucet -faucet.name polygon-mumbai -wallet.privkey your_private_key

# For Base Sepolia - automatically sets ETH symbol and chain ID 84532  
./eth-faucet -faucet.name base-sepolia -wallet.privkey your_private_key
```

### Multi-Chain Mode

The faucet supports running multiple networks simultaneously, allowing users to select different blockchains from a web interface.

**Generate configuration template:**
```bash
./eth-faucet -generate-config
```

This creates a `multichain-config.json` file with sample configuration for multiple networks.

**Run multi-chain faucet:**
```bash
./eth-faucet -multichain multichain-config.json
```

**Multi-chain configuration example:**
```json
{
  "http_port": 8080,
  "proxy_count": 0,
  "hcaptcha_sitekey": "your_hcaptcha_sitekey",
  "hcaptcha_secret": "your_hcaptcha_secret",
  "default_network": "sepolia",
  "networks": [
    {
      "name": "sepolia",
      "provider": "",
      "private_key": "0x1234...your_sepolia_private_key",
      "payout": 1.0,
      "interval": 1440
    },
    {
      "name": "polygon-mumbai", 
      "provider": "",
      "private_key": "0x5678...your_mumbai_private_key",
      "payout": 1.0,
      "interval": 1440
    },
    {
      "name": "bsc-testnet",
      "provider": "",
      "private_key": "0x9abc...your_bsc_private_key", 
      "payout": 0.1,
      "interval": 720
    }
  ]
}
```

**Multi-chain features:**
- **Network Selection**: Users can choose from available networks in the web interface
- **Per-Network Configuration**: Each network has its own payout amount, rate limiting, and wallet
- **Automatic Provider**: Uses default RPC endpoints if provider field is empty
- **Mixed Testnet/Mainnet**: Can run both testnet and mainnet faucets simultaneously (be careful with mainnet!)
- **Independent Rate Limiting**: Each network has separate rate limiting rules

### Configuration

You can configure the funding account by using environment variables instead of command-line flags:
```bash
export WEB3_PROVIDER=rpc_endpoint
export PRIVATE_KEY=hex_private_key
```

or

```bash
export WEB3_PROVIDER=rpc_endpoint
export KEYSTORE=keystore_path
echo "your_keystore_password" > `pwd`/password.txt
```

Then run the faucet application without the wallet command-line flags:
```bash
./eth-faucet -httpport 8080
```

**Optional Flags**

The following are the available command-line flags(excluding above wallet flags):

| Flag              | Description                                      | Default Value |
|-------------------|--------------------------------------------------|---------------|
| -httpport         | Listener port to serve HTTP connection           | 8080          |
| -proxycount       | Count of reverse proxies in front of the server  | 0             |
| -faucet.amount    | Number of Ethers to transfer per user request    | 1.0           |
| -faucet.minutes   | Number of minutes to wait between funding rounds | 1440          |
| -faucet.name      | Network name (auto-configures chain ID & symbol) | testnet       |
| -faucet.symbol    | Token symbol to display on the frontend          | ETH           |
| -list-networks    | List all supported networks and exit             | false         |
| -multichain       | Path to multi-chain configuration file           |               |
| -generate-config  | Generate sample multi-chain configuration        | false         |
| -hcaptcha.sitekey | hCaptcha                                  |               |
| -hcaptcha.secret  | hCaptcha                                   |               |

### Docker deployment

```bash
docker run -d -p 8080:8080 -e WEB3_PROVIDER=rpc_endpoint -e PRIVATE_KEY=hex_private_key chainflag/eth-faucet:1.2.0
```

or

```bash
docker run -d -p 8080:8080 -e WEB3_PROVIDER=rpc_endpoint -e KEYSTORE=keystore_path -v `pwd`/keystore:/app/keystore -v `pwd`/password.txt:/app/password.txt chainflag/eth-faucet:1.2.0
```

## License

Distributed under the MIT License. See LICENSE for more information.
