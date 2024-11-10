package core

import (
	"blockchain-from-scratch/utils"
	"encoding/hex"
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
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
					//if out.IsLockedWithKey(pubKeyHash) {
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
			//logrus.Infof("get '%x' vin: ", k)
			//utils.PrintJsonLog(&outs, "FindUTXO")

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

func (u UTXOSet) GetUTXODetails() {
	db := u.Blockchain.Db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var outs TXOutputs
			utils.Deserialize(v, &outs)
			logrus.Infof("get vin '%x' : ", k)
			utils.PrintJsonLog(&outs, "GetUTXODetails")
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (utxo UTXOSet) Reindex() {
	db := utxo.Blockchain.Db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		_, err = tx.CreateBucket(bucketName)
		return err
	})
	//utils.PrintJsonLog(utxo, "Reindex")
	if err != nil {
		log.Panic(err)
	}

	UTXO := utxo.Blockchain.FindUTXO()
	//utils.PrintJsonLog(UTXO, "Reindex")
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
	logrus.Info("======= Reindex UTXO ======")
}

func (u UTXOSet) Update(block *Block) {
	db := u.Blockchain.Db

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		for _, blockTx := range block.Transactions {
			if !blockTx.IsCoinbase() {
				for _, in := range blockTx.Vin {
					updatedOuts := TXOutputs{}
					outsBytes := b.Get(in.Txid)
					var outs TXOutputs
					utils.Deserialize(outsBytes, &outs)
					//utils.PrintJsonLog(outs, "outs")

					for outIndex, out := range outs.Outputs {
						if outIndex != in.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
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
			for _, out := range blockTx.VOut {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			err := b.Put(blockTx.ID, utils.Serialize(newOutputs))
			//utils.PrintJsonLog(newOutputs, fmt.Sprintf("newOutputs: %x", blockTx.ID))
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
