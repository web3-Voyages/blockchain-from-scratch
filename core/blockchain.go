package core

import (
	"github.com/boltdb/bolt"
	"log"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"

type Blockchain struct {
	//blocks []*Block
	tip []byte
	// TODO to support more db
	db *bolt.DB
}

func (chain *Blockchain) AddBlock(data string) {
	// get last hash from db
	var lastHash []byte
	err := chain.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	// create new block and store
	newBlock := NewBlock(data, lastHash)
	err = chain.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err = b.Put(newBlock.Hash, newBlock.Serialize())
		err = b.Put([]byte("l"), newBlock.Hash)
		chain.tip = newBlock.Hash
		return nil
	})
}

func NewBlockChain() *Blockchain {
	var tip []byte
	// Open a DB file.
	db, err := bolt.Open(dbFile, 0600, nil)
	err = db.Update(func(tx *bolt.Tx) error {
		// Check if there’s a blockchain stored in it.
		b := tx.Bucket([]byte(blocksBucket))
		// if no, Create a new Blockchain instance with its tip pointing at the genesis block
		if b == nil {
			genesisBlock := NewGenesisBlock()
			b, err = tx.CreateBucket([]byte(blocksBucket))
			err = b.Put(genesisBlock.Hash, genesisBlock.Serialize())
			err = b.Put([]byte("l"), genesisBlock.Hash)
			tip = genesisBlock.Hash
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return &Blockchain{tip, db}
}

// NewGenesisBlock when the chain created, init GenesisBlock
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}
