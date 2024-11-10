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

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type Blockchain struct {
	//blocks []*Block
	tip []byte
	// TODO to support more db
	Db *bolt.DB
}

// MineBlock mines a new block with the provided transactions
func (chain *Blockchain) MineBlock(transactions []*Transaction) *Block {
	for _, tx := range transactions {
		if !chain.VerifyTransaction(tx) {
			log.Panic("Error: Invalid transaction")
		}
	}

	// get last hash from db
	var lastHash []byte
	var lastHeight int
	err := chain.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		var block Block
		utils.Deserialize(blockData, &block)
		lastHeight = block.Height
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	// create new block and store
	newBlock := NewBlock(transactions, lastHash, lastHeight+1)
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

	return newBlock
}

// AddBlock saves the block into the blockchain
func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		blockInDb := bucket.Get(block.Hash)
		if blockInDb != nil {
			return nil
		}

		blockData := utils.Serialize(block)
		err := bucket.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		// Get last block from chain
		var lastBlock Block
		lastHash := bucket.Get([]byte("l"))
		utils.Deserialize(bucket.Get(lastHash), &lastBlock)

		// if block height higher, update blockchain lastblock
		if block.Height > lastBlock.Height {
			err = bucket.Put([]byte("l"), block.Hash)
			bc.tip = block.Hash
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// CreateBlockchain creates a new core DB
func CreateBlockchain(address, nodeId string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeId)
	if dbExists(dbFile) {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	coinbaseTx := NewCoinbaseTx(address, genesisCoinbaseData)
	//utils.PrintJsonLog(coinbaseTx, "coinbaseTx")
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

func NewBlockChain(nodeId string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeId)
	if !dbExists(dbFile) {
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
	//utils.PrintJsonLog(tx, "SignTransaction")
}

func (chain *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	//utils.PrintJsonLog(tx, "VerifyTransaction")
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

// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block

	err := bc.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		lastHash := bucket.Get([]byte("l"))
		bucketData := bucket.Get(lastHash)
		utils.Deserialize(bucketData, &lastBlock)
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}

// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}

// GetBlock finds a block by its hash and returns it
func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found.")
		}

		utils.Deserialize(blockData, &block)
		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
