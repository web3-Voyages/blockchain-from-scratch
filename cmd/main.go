package main

import (
	"blockchain-from-scratch/blockchain"
	"blockchain-from-scratch/cli"
)

func main() {
	chain := blockchain.NewBlockChain()
	defer chain.Db.Close()

	cli := cli.CLI{Chain: chain}
	cli.Run()
}
