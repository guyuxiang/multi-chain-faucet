package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jellydator/ttlcache/v2"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni/v3"

	"github.com/chainflag/eth-faucet/internal/chain"
	"github.com/chainflag/eth-faucet/internal/config"
	"github.com/chainflag/eth-faucet/web"
)

// MultiChainServer manages multiple blockchain networks
type MultiChainServer struct {
	multiConfig *config.MultiChainConfig
	builders    map[string]chain.TxBuilder
	limiters    map[string]*ttlcache.Cache // Per-network rate limiters
}

// NewMultiChainServer creates a new multi-chain faucet server
func NewMultiChainServer(multiConfig *config.MultiChainConfig) (*MultiChainServer, error) {
	server := &MultiChainServer{
		multiConfig: multiConfig,
		builders:    make(map[string]chain.TxBuilder),
		limiters:    make(map[string]*ttlcache.Cache),
	}

	// Initialize TxBuilders for each chain
	for network, chainInstance := range multiConfig.GetActiveChains() {
		// Create TxBuilder
		builder, err := chain.NewTxBuilder(
			chainInstance.Provider,
			chainInstance.PrivateKey,
			nil, // Chain ID will be auto-detected or set from config
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create TxBuilder for %s: %w", network, err)
		}

		server.builders[network] = builder

		// Create rate limiter for this network
		limiter := ttlcache.NewCache()
		limiter.SetTTL(time.Duration(chainInstance.Interval) * time.Minute)
		server.limiters[network] = limiter

		log.Infof("Initialized %s network (Chain ID: %d, Symbol: %s)",
			chainInstance.Config.Name, chainInstance.Config.ChainID, chainInstance.Config.Symbol)
	}

	return server, nil
}

// setupRouter creates HTTP routes for the multi-chain server
func (s *MultiChainServer) setupRouter() *http.ServeMux {
	router := http.NewServeMux()

	// Serve static files
	router.Handle("/", http.FileServer(web.Dist()))

	// API routes
	router.Handle("/api/claim", negroni.New(
		NewMultiChainLimiter(s.limiters, s.multiConfig.ProxyCount),
		NewCaptcha(s.multiConfig.HcaptchaSiteKey, s.multiConfig.HcaptchaSecret),
		negroni.Wrap(s.handleMultiChainClaim()),
	))
	router.Handle("/api/info", s.handleMultiChainInfo())
	router.Handle("/api/networks", s.handleNetworkList())

	return router
}

// handleMultiChainClaim processes faucet claims for multiple networks
func (s *MultiChainServer) handleMultiChainClaim() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}

		// Parse request
		var req multiChainClaimRequest
		if err := decodeJSONBody(r, &req); err != nil {
			log.WithError(err).Error("Failed to decode request")
			renderJSON(w, claimResponse{Message: err.Error()}, http.StatusBadRequest)
			return
		}

		// Validate network
		chainInstance, exists := s.multiConfig.GetChain(req.Network)
		if !exists {
			renderJSON(w, claimResponse{Message: "unsupported network"}, http.StatusBadRequest)
			return
		}

		// Get TxBuilder for the network
		builder, exists := s.builders[req.Network]
		if !exists {
			renderJSON(w, claimResponse{Message: "network not available"}, http.StatusInternalServerError)
			return
		}

		// Process transaction
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		txHash, err := builder.Transfer(ctx, req.Address, chain.EtherToWei(chainInstance.Payout))
		if err != nil {
			log.WithError(err).WithField("network", req.Network).Error("Failed to send transaction")
			renderJSON(w, claimResponse{Message: err.Error()}, http.StatusInternalServerError)
			return
		}

		log.WithFields(log.Fields{
			"txHash":  txHash,
			"address": req.Address,
			"network": req.Network,
			"amount":  chainInstance.Payout,
			"symbol":  chainInstance.Config.Symbol,
		}).Info("Transaction sent successfully")

		resp := claimResponse{
			Message: fmt.Sprintf("Txhash: %s", txHash),
		}
		renderJSON(w, resp, http.StatusOK)
	}
}

// handleMultiChainInfo returns information about all active networks
func (s *MultiChainServer) handleMultiChainInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}

		// Build network info for active chains
		activeNetworks := make(map[string]ActiveNetworkInfo)
		for network, chainInstance := range s.multiConfig.GetActiveChains() {
			builder := s.builders[network]
			activeNetworks[network] = ActiveNetworkInfo{
				Name:      chainInstance.Config.Name,
				Symbol:    chainInstance.Config.Symbol,
				ChainID:   chainInstance.Config.ChainID,
				IsTestnet: chainInstance.Config.IsTestnet,
				Account:   builder.Sender().String(),
				Payout:    strconv.FormatFloat(chainInstance.Payout, 'f', -1, 64),
			}
		}

		// Convert all supported networks to DTO format
		supportedNetworks := make(map[string]NetworkInfo)
		for name, netConfig := range config.GetSupportedNetworks() {
			supportedNetworks[name] = NetworkInfo{
				Name:      netConfig.Name,
				Symbol:    netConfig.Symbol,
				ChainID:   netConfig.ChainID,
				IsTestnet: netConfig.IsTestnet,
			}
		}

		resp := multiChainInfoResponse{
			DefaultNetwork:    s.multiConfig.DefaultChain,
			ActiveNetworks:    activeNetworks,
			SupportedNetworks: supportedNetworks,
			HcaptchaSiteKey:   s.multiConfig.HcaptchaSiteKey,
		}

		renderJSON(w, resp, http.StatusOK)
	}
}

// handleNetworkList returns list of active networks
func (s *MultiChainServer) handleNetworkList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}

		networks := s.multiConfig.GetChainNetworks()
		renderJSON(w, map[string]interface{}{
			"networks": networks,
			"default":  s.multiConfig.DefaultChain,
		}, http.StatusOK)
	}
}

// Run starts the multi-chain server
func (s *MultiChainServer) Run() {
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(s.setupRouter())

	log.Infof("Starting multi-chain faucet server on port %d", s.multiConfig.HTTPPort)
	log.Infof("Active networks: %v", s.multiConfig.GetChainNetworks())
	log.Infof("Default network: %s", s.multiConfig.DefaultChain)

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(s.multiConfig.HTTPPort), n))
}
