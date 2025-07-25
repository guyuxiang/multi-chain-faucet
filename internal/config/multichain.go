package config

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
)

// ChainInstance represents a configured blockchain instance
type ChainInstance struct {
	Network    string
	Config     NetworkConfig
	PrivateKey *ecdsa.PrivateKey
	Provider   string
	Payout     float64
	Interval   int
}

// MultiChainConfig holds configuration for multiple blockchain networks
type MultiChainConfig struct {
	Chains          map[string]*ChainInstance
	DefaultChain    string
	HTTPPort        int
	ProxyCount      int
	HcaptchaSiteKey string
	HcaptchaSecret  string
}

// ChainConfigInput represents input configuration for a single chain
type ChainConfigInput struct {
	Network    string
	Provider   string
	PrivateKey string
	Keystore   string
	KeyPass    string
	Payout     float64
	Interval   int
}

// NewMultiChainConfig creates a new multi-chain configuration
func NewMultiChainConfig() *MultiChainConfig {
	return &MultiChainConfig{
		Chains:     make(map[string]*ChainInstance),
		HTTPPort:   8080,
		ProxyCount: 0,
	}
}

// AddChainWithKey adds a blockchain network with a parsed private key
func (mc *MultiChainConfig) AddChainWithKey(input ChainConfigInput, privateKey *ecdsa.PrivateKey) error {
	// Get network configuration
	networkConfig, exists := GetNetworkByName(input.Network)
	if !exists {
		return fmt.Errorf("unsupported network: %s", input.Network)
	}

	// Set provider (use default if not specified)
	provider := input.Provider
	if provider == "" {
		provider = networkConfig.DefaultRPC
		if provider == "" {
			return fmt.Errorf("no provider specified and no default RPC for network %s", input.Network)
		}
	}

	// Set default values
	payout := input.Payout
	if payout == 0 {
		payout = 1.0
	}

	interval := input.Interval
	if interval == 0 {
		interval = 1440 // 24 hours default
	}

	// Create chain instance
	chainInstance := &ChainInstance{
		Network:    input.Network,
		Config:     networkConfig,
		PrivateKey: privateKey,
		Provider:   provider,
		Payout:     payout,
		Interval:   interval,
	}

	mc.Chains[input.Network] = chainInstance

	// Set as default if it's the first chain
	if mc.DefaultChain == "" {
		mc.DefaultChain = input.Network
	}

	return nil
}

// GetChain returns a specific chain instance
func (mc *MultiChainConfig) GetChain(network string) (*ChainInstance, bool) {
	chain, exists := mc.Chains[network]
	return chain, exists
}

// GetActiveChains returns all active chain instances
func (mc *MultiChainConfig) GetActiveChains() map[string]*ChainInstance {
	return mc.Chains
}

// GetChainNetworks returns a list of active network names
func (mc *MultiChainConfig) GetChainNetworks() []string {
	networks := make([]string, 0, len(mc.Chains))
	for network := range mc.Chains {
		networks = append(networks, network)
	}
	return networks
}

// ParsePrivateKey parses a hex private key
func ParsePrivateKey(hexkey string) (*ecdsa.PrivateKey, error) {
	if strings.HasPrefix(hexkey, "0x") {
		hexkey = hexkey[2:]
	}

	// This will be implemented by importing crypto from cmd package
	return nil, fmt.Errorf("private key parsing should be done in cmd package")
}

// ParseKeystore parses a keystore file
func ParseKeystore(keystorePath, passwordPath string) (*ecdsa.PrivateKey, error) {
	// This will be implemented by importing chain functions from cmd package
	return nil, fmt.Errorf("keystore parsing should be done in cmd package")
}
