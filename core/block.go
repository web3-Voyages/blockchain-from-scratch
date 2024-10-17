package core

import (
	"time"
)

// Block represents a block in the blockchain
type Block struct {
	Timestamp int64
	// include tx...
	Data []byte
	// hash of the previous block header
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}

	// use consensus to generate hash and nonce
	// TODO can be extend more consensus mechanism
	pow := NewProofOfWork(block)
	nonce, hash, err := pow.Run()
	if err != nil {
		return nil
	}
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}
