package server

import (
	"net/http"

	"github.com/guyuxiang/multi-chain-faucet/web"
)

const frontendBasePath = "/multi-chain-faucet"

func mountStatic(router *http.ServeMux) {
	fileServer := http.FileServer(web.Dist())
	router.Handle("/", fileServer)
	router.Handle(frontendBasePath, http.RedirectHandler(frontendBasePath+"/", http.StatusMovedPermanently))
	router.Handle(frontendBasePath+"/", http.StripPrefix(frontendBasePath, fileServer))
}
