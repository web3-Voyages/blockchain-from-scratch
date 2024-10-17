package model

import (
	"bytes"
	"crypto/sha256"
	"strconv"
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
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}}
	block.SetHash()
	return &block
}

func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headeds := bytes.Join([][]byte{timestamp, b.PrevBlockHash, b.Data}, []byte{})
	hash := sha256.Sum256(headeds)
	b.Hash = hash[:]
}
