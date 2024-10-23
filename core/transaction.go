package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
)

// use UTXO transcation
// see link: https://trustwallet.com/blog/what-is-a-utxo-unspent-transaction-output
type Transaction struct {
	ID   []byte
	Vin  []TxInput
	VOut []TxOutput
}

// subsidy is the amount of reward.
// In Bitcoin, this number is not stored anywhere and calculated based only on the total number of blocks: the number of blocks is divided by 210000.
// Mining the genesis block produced 50 BTC, and every 210000 blocks the reward is halved.
// In our implementation, we’ll store the reward as a constant
const subsidy = 10

func NewCoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TxInput{[]byte{}, -1, data}
	txout := TxOutput{subsidy, to}
	tx := Transaction{nil, []TxInput{txin}, []TxOutput{txout}}
	tx.ID = tx.Hash()
	return &tx
}

func NewUTXOTransaction(from, to string, amount int, chain *Blockchain) *Transaction {
	return NewCoinbaseTx(from, to)
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}
