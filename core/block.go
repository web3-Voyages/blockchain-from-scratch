package core

import (
	"bytes"
	"crypto/sha256"
	"github.com/sirupsen/logrus"
	"log"
	"time"
)

// Block represents a block in the core
type Block struct {
	Timestamp    int64
	Transactions []*Transaction
	// hash of the previous block header
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}

	// use consensus to generate hash and nonce
	// TODO can be extend more consensus mechanism
	pow := NewProofOfWork(block)
	nonce, hash, err := pow.Run()
	if err != nil {
		log.Panic(err)
	}
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// NewGenesisBlock when the chain created, init GenesisBlock
func NewGenesisBlock(coinbase *Transaction) *Block {
	logrus.Info("No existing blockchain found. Creating a new one...")
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

// HashTransactions TODO can use Merkle tree hash
// Bitcoin uses a more elaborate technique: it represents all transactions containing in a block as a Merkle tree
// and uses the root hash of the tree in the Proof-of-Work system.
// This approach allows to quickly check if a block contains certain transaction,
// having only just the root hash and without downloading all the transactions.
func (b *Block) HashTransactions() []byte {
	var txHashs [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashs = append(txHashs, tx.Serialize())
	}
	txHash = sha256.Sum256(bytes.Join(txHashs, []byte{}))
	return txHash[:]
}
