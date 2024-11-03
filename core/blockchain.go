package core

import (
	"blockchain-from-scratch/utils"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain_test.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type Blockchain struct {
	//blocks []*Block
	tip []byte
	// TODO to support more db
	Db *bolt.DB
}

// MineBlock mines a new block with the provided transactions
func (chain *Blockchain) MineBlock(transactions []*Transaction) {
	for _, tx := range transactions {
		if !chain.VerifyTransaction(tx) {
			log.Panic("Error: Invalid transaction")
		}
	}

	// get last hash from db
	var lastHash []byte
	err := chain.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	// create new block and store
	newBlock := NewBlock(transactions, lastHash)
	err = chain.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err = b.Put(newBlock.Hash, utils.Serialize(newBlock))
		err = b.Put([]byte("l"), newBlock.Hash)
		chain.tip = newBlock.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// CreateBlockchain creates a new core DB
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	coinbaseTx := NewCoinbaseTx(address, genesisCoinbaseData)
	genesisBlock := NewGenesisBlock(coinbaseTx)

	var tip []byte
	// Open a DB file.
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}
		err = b.Put(genesisBlock.Hash, utils.Serialize(genesisBlock))
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"), genesisBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesisBlock.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return &Blockchain{tip, db}
}

func NewBlockChain() *Blockchain {
	if !dbExists() {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	// Open a DB file.
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return &Blockchain{tip, db}
}

func (chain *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	iterator := chain.Iterator()

	for {
		block := iterator.Next()

		for _, tx := range block.Transactions {
			txId := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.VOut {
				{
					// Check if the output was spent
					if spentTXOs[txId] != nil {
						for _, spentOut := range spentTXOs[txId] {
							if spentOut == outIdx {
								continue Outputs // Skip if the output was spent
							}
						}
					}

					outs := UTXO[txId]
					outs.Outputs = append(outs.Outputs, out)
					UTXO[txId] = outs
				}

				if !tx.IsCoinbase() {
					for _, in := range tx.Vin {
						inTxId := hex.EncodeToString(in.Txid)
						spentTXOs[inTxId] = append(spentTXOs[inTxId], in.Vout)
					}
				}
			}

		}
		// Break the loop if the genesis block is reached
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return UTXO
}

// Iterator returns a BlockchainIterator to iterate over the blocks of the core
func (chain *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{chain.tip, chain.Db}
}

func (chain *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	iterator := chain.Iterator()

	for {
		block := iterator.Next()
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("transaction is not found")
}

func (chain *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTx, err := chain.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(vin.Txid)] = prevTx
	}
	tx.Sign(privKey, prevTXs)
}

func (chain *Blockchain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		prevTx, err := chain.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(vin.Txid)] = prevTx
	}
	return tx.Verify(prevTXs)
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
