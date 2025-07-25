package main

import (
	"github.com/guyuxiang/multi-chain-faucet/cmd"
)

//go:generate npm run build
func main() {
	cmd.Execute()
}
