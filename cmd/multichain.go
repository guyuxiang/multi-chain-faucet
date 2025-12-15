package cmd

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/guyuxiang/multi-chain-faucet/internal/chain"
	"github.com/guyuxiang/multi-chain-faucet/internal/config"
	"github.com/guyuxiang/multi-chain-faucet/internal/server"
)

// ExecuteMultiChain starts the multi-chain faucet server
func ExecuteMultiChain(configPath string) {
	// Load configuration
	multiConfig, err := loadMultiChainConfig(configPath)
	if err != nil {
		panic(fmt.Errorf("failed to load multi-chain config: %w", err))
	}

	// Create and start server
	server, err := server.NewMultiChainServer(multiConfig)
	if err != nil {
		panic(fmt.Errorf("failed to create multi-chain server: %w", err))
	}

	go server.Run()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

// loadMultiChainConfig loads configuration from JSON file
func loadMultiChainConfig(configPath string) (*config.MultiChainConfig, error) {
	fileConfig, err := config.LoadMultiChainConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	// Create multi-chain config
	multiConfig := config.NewMultiChainConfig()
	multiConfig.HTTPPort = fileConfig.HTTPPort
	multiConfig.ProxyCount = fileConfig.ProxyCount
	multiConfig.HcaptchaSiteKey = fileConfig.HcaptchaSiteKey
	multiConfig.HcaptchaSecret = fileConfig.HcaptchaSecret

	// Add networks
	for _, netConfig := range fileConfig.Networks {
		networkConfig, err := config.BuildNetworkConfigFromFile(netConfig)
		if err != nil {
			return nil, fmt.Errorf("invalid network %s: %w", netConfig.Name, err)
		}

		chainInput := config.ChainConfigInput{
			Network:  netConfig.Name,
			Provider: netConfig.Provider,
			Payout:   netConfig.Payout,
			Interval: netConfig.Interval,
			Config:   networkConfig,
		}

		// Parse private key or keystore
		var privateKey *ecdsa.PrivateKey

		if netConfig.PrivateKey != "" {
			privateKey, err = parsePrivateKeyHex(netConfig.PrivateKey)
			if err != nil {
				return nil, fmt.Errorf("failed to parse private key for %s: %w", netConfig.Name, err)
			}
		} else if netConfig.Keystore != "" {
			privateKey, err = parseKeystoreFile(netConfig.Keystore, netConfig.KeyPass)
			if err != nil {
				return nil, fmt.Errorf("failed to parse keystore for %s: %w", netConfig.Name, err)
			}
		} else {
			return nil, fmt.Errorf("network %s requires either private_key or keystore", netConfig.Name)
		}

		if err := multiConfig.AddChainWithKey(chainInput, privateKey); err != nil {
			return nil, fmt.Errorf("failed to add network %s: %w", netConfig.Name, err)
		}
	}

	// Set default network
	if fileConfig.DefaultNetwork != "" {
		if _, exists := multiConfig.GetChain(fileConfig.DefaultNetwork); !exists {
			return nil, fmt.Errorf("default network %s is not configured", fileConfig.DefaultNetwork)
		}
		multiConfig.DefaultChain = fileConfig.DefaultNetwork
	}

	return multiConfig, nil
}

// GenerateMultiChainConfig creates a sample configuration file
func GenerateMultiChainConfig(outputPath string) error {
	sampleConfig := config.MultiChainConfigFile{
		HTTPPort:        8080,
		ProxyCount:      0,
		HcaptchaSiteKey: "",
		HcaptchaSecret:  "",
		DefaultNetwork:  "sepolia",
		Networks: []config.NetworkConfigFile{
			{
				Name:        "sepolia",
				DisplayName: "Ethereum Sepolia",
				ChainID:     11155111,
				Symbol:      "ETH",
				IsTestnet:   true,
				Provider:    "https://sepolia.infura.io/v3/<api-key>",
				ExplorerURL: "https://sepolia.etherscan.io/tx/",
				PrivateKey:  "0x1234567890abcdef...", // Replace with actual key
				Payout:      1.0,
				Interval:    1440,
			},
			{
				Name:        "polygon-amoy",
				DisplayName: "Polygon Amoy",
				ChainID:     80002,
				Symbol:      "POL",
				IsTestnet:   true,
				Provider:    "https://rpc-amoy.polygon.technology",
				ExplorerURL: "https://amoy.polygonscan.com/tx/",
				PrivateKey:  "0x1234567890abcdef...",
				Payout:      1.0,
				Interval:    1440,
			},
			{
				Name:        "bsc-testnet",
				DisplayName: "BSC Testnet",
				ChainID:     97,
				Symbol:      "BNB",
				IsTestnet:   true,
				Provider:    "https://data-seed-prebsc-1-s1.binance.org:8545",
				ExplorerURL: "https://testnet.bscscan.com/tx/",
				PrivateKey:  "0x1234567890abcdef...",
				Payout:      0.1,
				Interval:    1440,
			},
		},
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(sampleConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Sample multi-chain configuration written to %s\n", outputPath)
	fmt.Println("\nPlease update the private keys and other settings before using.")
	return nil
}

// Helper function to parse private key (moved here from config package)
func parsePrivateKeyHex(hexkey string) (*ecdsa.PrivateKey, error) {
	if strings.HasPrefix(hexkey, "0x") {
		hexkey = hexkey[2:]
	}
	return crypto.HexToECDSA(hexkey)
}

// Helper function to parse keystore
func parseKeystoreFile(keystorePath, passwordPath string) (*ecdsa.PrivateKey, error) {
	// Resolve keystore path
	keyfile, err := chain.ResolveKeyfilePath(keystorePath)
	if err != nil {
		return nil, err
	}

	// Read password
	password, err := os.ReadFile(passwordPath)
	if err != nil {
		return nil, err
	}

	// Decrypt keystore
	return chain.DecryptKeyfile(keyfile, strings.TrimRight(string(password), "\r\n"))
}
