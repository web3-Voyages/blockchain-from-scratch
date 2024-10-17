package model

import (
	"fmt"
	"testing"
)

func TestInitChain(t *testing.T) {
	chain := NewBlockChain()

	chain.AddBlock("Send tx1")
	chain.AddBlock("Send tx2")

	for _, block := range chain.blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}
