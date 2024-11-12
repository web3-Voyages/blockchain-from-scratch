package tests

import (
	"blockchain-from-scratch/core"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"testing"
)

func TestInitChain(t *testing.T) {
	chain := core.NewBlockChain("1")

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
	var tx core.Transaction
	jsonStr := `{"ID":"u+KpSWUNoowD2vpoEbivs4IbjL1j7wqDXNFsrXRPJb0=","Vin":[{"Txid":"ccXSWEcXrjDSYoX9wttqnMkbuTYJHeb0ts6SVrb7ZbI=","Vout":0,"Signature":"cz7z23kFFXRLcN93p5B7At7OJCfA+9woFztZhp8cEs5k+ulvfaPmkR1r2VfFfqvTuOJu2OWq/G7N86Yei6CUZg==","PubKey":"MsZKElx+u7z35ANBrZ6Pv9C+uIY3/IkiiDDgRowhCLCiHDfr4MvpztFDieCTNa0YDUh90Nsb93+9AzXAsJDkvg=="}],"VOut":[{"Value":1,"PubKeyHash":"1BemKb104Gzwjx2tjNa7xLA4hPo="},{"Value":9,"PubKeyHash":"1T2TO5pErUyq97jpcetEDenmZEQ="}]}`
	err := json.Unmarshal([]byte(jsonStr), &tx)
	if err != nil {
		log.Panic(err)
	}
	logrus.Infof("%x", tx.Hash())
	logrus.Infof("%x", tx.Hash())
}
