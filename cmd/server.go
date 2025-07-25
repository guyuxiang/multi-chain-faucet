package cmd

import (
	"crypto/ecdsa"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/chainflag/eth-faucet/internal/chain"
	"github.com/chainflag/eth-faucet/internal/config"
	"github.com/chainflag/eth-faucet/internal/server"
)

var (
	appVersion = "v1.2.0"

	// Legacy chainIDMap for backward compatibility
	chainIDMap = make(map[string]int)

	httpPortFlag       = flag.Int("httpport", 8080, "Listener port to serve HTTP connection")
	proxyCntFlag       = flag.Int("proxycount", 0, "Count of reverse proxies in front of the server")
	versionFlag        = flag.Bool("version", false, "Print version number")
	listNetworksFlag   = flag.Bool("list-networks", false, "List all supported networks and exit")
	multiChainFlag     = flag.String("multichain", "", "Path to multi-chain configuration file")
	generateConfigFlag = flag.Bool("generate-config", false, "Generate sample multi-chain configuration file")

	payoutFlag   = flag.Float64("faucet.amount", 1, "Number of Ethers to transfer per user request")
	intervalFlag = flag.Int("faucet.minutes", 1440, "Number of minutes to wait between funding rounds")
	netnameFlag  = flag.String("faucet.name", "testnet", "Network name to display on the frontend")
	symbolFlag   = flag.String("faucet.symbol", "ETH", "Token symbol to display on the frontend")

	keyJSONFlag  = flag.String("wallet.keyjson", os.Getenv("KEYSTORE"), "Keystore file to fund user requests with")
	keyPassFlag  = flag.String("wallet.keypass", "password.txt", "Passphrase text file to decrypt keystore")
	privKeyFlag  = flag.String("wallet.privkey", os.Getenv("PRIVATE_KEY"), "Private key hex to fund user requests with")
	providerFlag = flag.String("wallet.provider", os.Getenv("WEB3_PROVIDER"), "Endpoint for Ethereum JSON-RPC connection")

	hcaptchaSiteKeyFlag = flag.String("hcaptcha.sitekey", os.Getenv("HCAPTCHA_SITEKEY"), "hCaptcha sitekey")
	hcaptchaSecretFlag  = flag.String("hcaptcha.secret", os.Getenv("HCAPTCHA_SECRET"), "hCaptcha secret")
)

// ListSupportedNetworks prints all supported networks (useful for CLI help)
func ListSupportedNetworks() {
	fmt.Println("Supported networks:")
	fmt.Println("==================")

	categories := map[string][]string{
		"Ethereum":  {"mainnet", "sepolia", "holesky", "goerli"},
		"Polygon":   {"polygon", "polygon-mumbai", "polygon-amoy"},
		"BSC":       {"bsc", "bsc-testnet"},
		"Arbitrum":  {"arbitrum", "arbitrum-sepolia"},
		"Optimism":  {"optimism", "optimism-sepolia"},
		"Avalanche": {"avalanche", "avalanche-fuji"},
		"Base":      {"base", "base-sepolia"},
		"Fantom":    {"fantom", "fantom-testnet"},
		"Linea":     {"linea", "linea-sepolia"},
		"zkSync":    {"zksync", "zksync-sepolia"},
	}

	for category, networks := range categories {
		fmt.Printf("\n%s:\n", category)
		for _, networkName := range networks {
			if networkConfig, exists := config.GetNetworkByName(networkName); exists {
				testnetStr := ""
				if networkConfig.IsTestnet {
					testnetStr = " (Testnet)"
				}
				fmt.Printf("  %-20s - %s%s (Chain ID: %d, Symbol: %s)\n",
					networkName, networkConfig.Name, testnetStr, networkConfig.ChainID, networkConfig.Symbol)
			}
		}
	}
}

func init() {
	// Initialize legacy chainIDMap for backward compatibility
	chainIDMap = config.GetChainIDMap()

	flag.Parse()
	if *versionFlag {
		fmt.Println(appVersion)
		os.Exit(0)
	}
	if *listNetworksFlag {
		ListSupportedNetworks()
		os.Exit(0)
	}
	if *generateConfigFlag {
		if err := GenerateMultiChainConfig("multichain-config.json"); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating config: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if *multiChainFlag != "" {
		ExecuteMultiChain(*multiChainFlag)
		return
	}
}

func Execute() {
	privateKey, err := getPrivateKeyFromFlags()
	if err != nil {
		panic(fmt.Errorf("failed to read private key: %w", err))
	}

	// Get network configuration
	networkName := strings.ToLower(*netnameFlag)
	networkConfig, exists := config.GetNetworkByName(networkName)

	var chainID *big.Int
	var symbol string
	var displayName string

	if exists {
		chainID = big.NewInt(networkConfig.ChainID)
		// Use network default symbol if not explicitly set via flag
		if *symbolFlag == "ETH" { // Check if using default symbol
			symbol = networkConfig.Symbol
		} else {
			symbol = *symbolFlag
		}
		displayName = networkConfig.Name

		// Use default RPC if provider not set
		if *providerFlag == "" {
			*providerFlag = networkConfig.DefaultRPC
		}
	} else {
		// Fallback to legacy behavior
		if value, ok := chainIDMap[networkName]; ok {
			chainID = big.NewInt(int64(value))
		}
		symbol = *symbolFlag
		displayName = *netnameFlag
	}

	// Validate provider is set
	if *providerFlag == "" {
		panic(fmt.Errorf("web3 provider is required. Set via -wallet.provider flag or WEB3_PROVIDER environment variable"))
	}

	txBuilder, err := chain.NewTxBuilder(*providerFlag, privateKey, chainID)
	if err != nil {
		panic(fmt.Errorf("cannot connect to web3 provider: %w", err))
	}

	config := server.NewConfig(displayName, symbol, *httpPortFlag, *intervalFlag, *proxyCntFlag, *payoutFlag, *hcaptchaSiteKeyFlag, *hcaptchaSecretFlag)
	go server.NewServer(txBuilder, config).Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func getPrivateKeyFromFlags() (*ecdsa.PrivateKey, error) {
	if *privKeyFlag != "" {
		hexkey := *privKeyFlag
		if chain.Has0xPrefix(hexkey) {
			hexkey = hexkey[2:]
		}
		return crypto.HexToECDSA(hexkey)
	} else if *keyJSONFlag == "" {
		return nil, errors.New("missing private key or keystore")
	}

	keyfile, err := chain.ResolveKeyfilePath(*keyJSONFlag)
	if err != nil {
		return nil, err
	}
	password, err := os.ReadFile(*keyPassFlag)
	if err != nil {
		return nil, err
	}

	return chain.DecryptKeyfile(keyfile, strings.TrimRight(string(password), "\r\n"))
}
