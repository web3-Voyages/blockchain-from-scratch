package core

import (
	"blockchain-from-scratch/utils"
	"encoding/hex"
	"github.com/boltdb/bolt"
	"log"
)

const utxoBucket = "chainstate"

// UTXOSet represents UTXO set
type UTXOSet struct {
	Blockchain *Blockchain
}

func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.Db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			var outs TXOutputs
			utils.Deserialize(v, &outs)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return accumulated, unspentOutputs
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput
	db := u.Blockchain.Db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var outs TXOutputs
			utils.Deserialize(v, &outs)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}

func (utxo UTXOSet) Reindex() {
	db := utxo.Blockchain.Db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		_, err = tx.CreateBucket(bucketName)
		return err
	})

	if err != nil {
		log.Panic(err)
	}

	UTXO := utxo.Blockchain.FindUTXO()
	err = db.Update(func(tx *bolt.Tx) error {
		block := tx.Bucket(bucketName)
		for txId, outs := range UTXO {
			key, err := hex.DecodeString(txId)
			if err != nil {
				log.Panic(err)
			}
			err = block.Put(key, utils.Serialize(outs))
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

}

func (u UTXOSet) Update(block *Block) {
	db := u.Blockchain.Db

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					updatedOuts := TXOutputs{}
					outsBytes := b.Get(in.Txid)
					var outs TXOutputs
					utils.Deserialize(outsBytes, &outs)

					for outIndex, out := range outs.Outputs {
						if outIndex != in.Vout {
							outs.Outputs = append(outs.Outputs, out)
						}
					}

					var err error
					if len(updatedOuts.Outputs) == 0 {
						err = b.Delete(in.Txid)
					} else {
						err = b.Put(in.Txid, utils.Serialize(updatedOuts))
					}
					if err != nil {
						log.Panic(err)
					}
				}
			}

			newOutputs := TXOutputs{}
			for _, out := range tx.VOut {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			err := b.Put(tx.ID, utils.Serialize(newOutputs))
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
