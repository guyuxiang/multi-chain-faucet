package config

import "strings"

type NetworkConfig struct {
	ChainID    int64
	Symbol     string
	Name       string
	IsTestnet  bool
	DefaultRPC string
}

// Network configurations with chain IDs and default settings
var NetworkConfigs = map[string]NetworkConfig{
	// Ethereum Networks
	"mainnet": {ChainID: 1, Symbol: "ETH", Name: "Ethereum Mainnet", IsTestnet: false, DefaultRPC: "https://eth.public-rpc.com"},
	"sepolia": {ChainID: 11155111, Symbol: "ETH", Name: "Ethereum Sepolia", IsTestnet: true, DefaultRPC: "https://sepolia.infura.io/v3/"},
	"holesky": {ChainID: 17000, Symbol: "ETH", Name: "Ethereum Holesky", IsTestnet: true, DefaultRPC: "https://ethereum-holesky.publicnode.com"},
	"goerli":  {ChainID: 5, Symbol: "ETH", Name: "Ethereum Goerli", IsTestnet: true, DefaultRPC: "https://goerli.infura.io/v3/"},

	// Polygon Networks
	"polygon":      {ChainID: 137, Symbol: "POL", Name: "Polygon Mainnet", IsTestnet: false, DefaultRPC: "https://polygon-rpc.com"},
	"polygon-amoy": {ChainID: 80002, Symbol: "POL", Name: "Polygon Amoy", IsTestnet: true, DefaultRPC: "https://rpc-amoy.polygon.technology"},

	// BSC Networks
	"bsc":         {ChainID: 56, Symbol: "BNB", Name: "BNB Smart Chain", IsTestnet: false, DefaultRPC: "https://bsc-dataseed.binance.org"},
	"bsc-testnet": {ChainID: 97, Symbol: "BNB", Name: "BNB Smart Chain Testnet", IsTestnet: true, DefaultRPC: "https://data-seed-prebsc-1-s1.binance.org:8545"},

	// Arbitrum Networks
	"arbitrum":         {ChainID: 42161, Symbol: "ETH", Name: "Arbitrum One", IsTestnet: false, DefaultRPC: "https://arb1.arbitrum.io/rpc"},
	"arbitrum-sepolia": {ChainID: 421614, Symbol: "ETH", Name: "Arbitrum Sepolia", IsTestnet: true, DefaultRPC: "https://sepolia-rollup.arbitrum.io/rpc"},

	// Optimism Networks
	"optimism":         {ChainID: 10, Symbol: "ETH", Name: "Optimism Mainnet", IsTestnet: false, DefaultRPC: "https://mainnet.optimism.io"},
	"optimism-sepolia": {ChainID: 11155420, Symbol: "ETH", Name: "Optimism Sepolia", IsTestnet: true, DefaultRPC: "https://sepolia.optimism.io"},

	// Avalanche Networks
	"avalanche":      {ChainID: 43114, Symbol: "AVAX", Name: "Avalanche C-Chain", IsTestnet: false, DefaultRPC: "https://api.avax.network/ext/bc/C/rpc"},
	"avalanche-fuji": {ChainID: 43113, Symbol: "AVAX", Name: "Avalanche Fuji", IsTestnet: true, DefaultRPC: "https://api.avax-test.network/ext/bc/C/rpc"},

	// Base Networks
	"base":         {ChainID: 8453, Symbol: "ETH", Name: "Base Mainnet", IsTestnet: false, DefaultRPC: "https://mainnet.base.org"},
	"base-sepolia": {ChainID: 84532, Symbol: "ETH", Name: "Base Sepolia", IsTestnet: true, DefaultRPC: "https://sepolia.base.org"},

	// Fantom Networks
	"fantom":         {ChainID: 250, Symbol: "FTM", Name: "Fantom Opera", IsTestnet: false, DefaultRPC: "https://rpc.ftm.tools"},
	"fantom-testnet": {ChainID: 4002, Symbol: "FTM", Name: "Fantom Testnet", IsTestnet: true, DefaultRPC: "https://rpc.testnet.fantom.network"},

	// Linea Networks
	"linea":         {ChainID: 59144, Symbol: "ETH", Name: "Linea Mainnet", IsTestnet: false, DefaultRPC: "https://rpc.linea.build"},
	"linea-sepolia": {ChainID: 59141, Symbol: "ETH", Name: "Linea Sepolia", IsTestnet: true, DefaultRPC: "https://rpc.sepolia.linea.build"},

	// zkSync Networks
	"zksync":         {ChainID: 324, Symbol: "ETH", Name: "zkSync Era", IsTestnet: false, DefaultRPC: "https://mainnet.era.zksync.io"},
	"zksync-sepolia": {ChainID: 300, Symbol: "ETH", Name: "zkSync Sepolia", IsTestnet: true, DefaultRPC: "https://sepolia.era.zksync.dev"},
}

// GetSupportedNetworks returns a list of all supported networks
func GetSupportedNetworks() map[string]NetworkConfig {
	return NetworkConfigs
}

// GetNetworkByName returns the network configuration for a given network name
func GetNetworkByName(name string) (NetworkConfig, bool) {
	config, exists := NetworkConfigs[strings.ToLower(name)]
	return config, exists
}

// GetChainIDMap returns a map of network names to chain IDs for backward compatibility
func GetChainIDMap() map[string]int {
	chainIDMap := make(map[string]int)
	for name, config := range NetworkConfigs {
		chainIDMap[name] = int(config.ChainID)
	}
	return chainIDMap
}
