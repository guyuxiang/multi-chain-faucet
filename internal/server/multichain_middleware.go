package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jellydator/ttlcache/v2"
	log "github.com/sirupsen/logrus"
)

// MultiChainLimiter provides rate limiting per network
type MultiChainLimiter struct {
	limiters   map[string]*ttlcache.Cache
	proxyCount int
}

// NewMultiChainLimiter creates a new multi-chain rate limiter
func NewMultiChainLimiter(limiters map[string]*ttlcache.Cache, proxyCount int) *MultiChainLimiter {
	return &MultiChainLimiter{
		limiters:   limiters,
		proxyCount: proxyCount,
	}
}

// ServeHTTP implements the negroni middleware interface
func (ml *MultiChainLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Parse request to get network
	var req multiChainClaimRequest
	if err := decodeJSONBody(r, &req); err != nil {
		log.WithError(err).Error("Failed to decode request in rate limiter")
		renderJSON(w, claimResponse{Message: "invalid request format"}, http.StatusBadRequest)
		return
	}

	// Validate address
	if !isValidAddress(req.Address) {
		renderJSON(w, claimResponse{Message: "invalid address"}, http.StatusBadRequest)
		return
	}

	// Get network-specific limiter
	limiter, exists := ml.limiters[req.Network]
	if !exists {
		renderJSON(w, claimResponse{Message: "unsupported network"}, http.StatusBadRequest)
		return
	}

	// Check rate limits
	if err := ml.checkLimits(limiter, r, req.Address); err != nil {
		renderJSON(w, claimResponse{Message: err.Error()}, http.StatusTooManyRequests)
		return
	}

	// Continue to next handler
	next(w, r)
}

// checkLimits validates rate limiting rules
func (ml *MultiChainLimiter) checkLimits(limiter *ttlcache.Cache, r *http.Request, address string) error {
	// Get client IP
	ip := getClientIP(r, ml.proxyCount)

	// Check address-based limit
	if _, err := limiter.Get(address); err == nil {
		return fmt.Errorf("address %s is requesting too frequently", address)
	}

	// Check IP-based limit
	if _, err := limiter.Get(ip); err == nil {
		return fmt.Errorf("IP %s is requesting too frequently", ip)
	}

	// Set rate limit entries
	limiter.Set(address, true)
	limiter.Set(ip, true)

	log.WithFields(log.Fields{
		"address": address,
		"ip":      ip,
	}).Info("Rate limit passed")

	return nil
}

// getClientIP extracts client IP considering proxy configuration
func getClientIP(r *http.Request, proxyCount int) string {
	if proxyCount > 0 {
		// Handle X-Forwarded-For header with proxy count
		xff := r.Header.Get("X-Forwarded-For")
		if xff != "" {
			ips := strings.Split(xff, ",")
			if len(ips) >= proxyCount {
				return strings.TrimSpace(ips[len(ips)-proxyCount])
			}
		}
	}

	// Fallback to direct connection
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}

	return ip
}

// isValidAddress validates Ethereum address format
func isValidAddress(address string) bool {
	// Basic validation - you may want to use a more robust validation
	if len(address) != 42 {
		return false
	}
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	// Additional validation can be added here
	return true
}
