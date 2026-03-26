package server

import "net/http"

// mountAPI registers an API handler on both root and frontend base paths so the
// frontend can live under /multi-chain-faucet/ while calling ./api/*.
func mountAPI(router *http.ServeMux, path string, handler http.Handler) {
	router.Handle(path, handler)
	router.Handle(frontendBasePath+path, handler)
}
