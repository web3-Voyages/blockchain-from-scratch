package tests

import (
	"blockchain-from-scratch/core"
	"encoding/hex"
	"fmt"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestInitChain(t *testing.T) {
	chain := core.NewBlockChain()

	//chain.MineBlock("Send tx1")
	//chain.MineBlock("Send tx2")

	iterator := chain.Iterator()
	for {
		block := iterator.Next()
		if len(block.PrevBlockHash) == 0 {
			break
		}
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.HashTransactions())
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}

func Test(t *testing.T) {
	decode, _ := hex.DecodeString("/k31VYCoA7E6KfGR7f4uaMKwKlfSJrwvNE8F6I2pha8=")
	//hex.EncodeToString()
	logrus.Infof("%x", decode)
}
