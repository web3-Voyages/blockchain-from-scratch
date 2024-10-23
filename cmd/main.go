package main

import (
	"blockchain-from-scratch/cli"
	"blockchain-from-scratch/core"
)

func main() {
	chain := core.NewBlockChain()
	defer chain.Db.Close()

	cli := cli.CLI{Chain: chain}
	cli.Run()
}
