package tests

import (
	"blockchain-from-scratch/core/blockchain"
	"fmt"
	"testing"
)

func TestInitChain(t *testing.T) {
	chain := blockchain.NewBlockChain()

	chain.AddBlock("Send tx1")
	chain.AddBlock("Send tx2")

	iterator := chain.Iterator()
	for {
		block := iterator.Next()
		if len(block.PrevBlockHash) == 0 {
			break
		}
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}
