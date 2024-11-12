package core

import (
	"blockchain-from-scratch/core/wallet"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"log"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// use UTXO transcation, just like the Bitcoin protocol does not track user balances directly;
// instead, it tracks UTXOs and which addresses they belong to.
// see link:
//  1. https://trustwallet.com/blog/what-is-a-utxo-unspent-transaction-output
//  2. https://mp.weixin.qq.com/s/LsHf2jhy9YdQcAyM8b9bdg
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
		// using random number generation to ensure uniqueness of data and transaction ID uniqueness
		timestamp := time.Now().Unix()
		randomUUID := uuid.New().String()
		data = fmt.Sprintf("Reward to '%s' at %d with UUID %s", to, timestamp, randomUUID)
	}

	txin := TxInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{nil, []TxInput{txin}, []TxOutput{*txout}}
	tx.ID = tx.Hash()
	logrus.Infof("NewCoinbaseTx to '%s'", to)
	return &tx
}

func NewUTXOTransaction(nodeWallet *wallet.Wallet, to string, amount int, UTXOSet *UTXOSet) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	pubKeyHash := wallet.HashPubKey(nodeWallet.PublicKey)
	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)
	if acc < amount {
		log.Panic("Error: Not enough funds")
	}

	for txId, outs := range validOutputs {
		txID, err := hex.DecodeString(txId)
		if err != nil {
			log.Panic(err)
		}
		for _, out := range outs {
			input := TxInput{txID, out, nil, nodeWallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	from := string(nodeWallet.GetAddress())
	logrus.Infof("NewUTXOTransaction from '%s' to '%s' amount %d", from, to, acc)
	// Ensure correct balance when transferring to self
	if from == to {
		outputs = append(outputs, *NewTXOutput(acc, from))
	} else {
		outputs = append(outputs, *NewTXOutput(amount, to))
		if acc > amount {
			outputs = append(outputs, *NewTXOutput(acc-amount, from))
		}
	}
	tx := Transaction{nil, inputs, outputs}
	//utils.PrintJsonLog(tx, "NewUTXOTransaction")
	tx.ID = tx.Hash()
	// need SignTransaction
	UTXOSet.Blockchain.SignTransaction(&tx, nodeWallet.PrivateKey)
	return &tx
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
	result, err := msgpack.Marshal(tx)
	if err != nil {
		log.Panic(err)
	}
	return result
}

func (tx *Transaction) Sign(priKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	// Create a trimmed copy of the transaction for signing
	// Using txCopy is to isolate the signing data and prevent transaction malleability issues
	txCopy := tx.TrimmedCopy()

	for inId, vin := range txCopy.Vin {
		// Retrieve the previous transaction corresponding to the current input
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		prepareForSigning(&txCopy, inId, &prevTx)
		dataToSign := txCopy.ID
		txCopy.Vin[inId].Signature = nil
		txCopy.Vin[inId].PubKey = prevTx.VOut[vin.Vout].PubKeyHash
		//dataToSign := []byte(fmt.Sprintf("%x\n", txCopy))
		// Sign the transaction ID with the private key
		r, s, err := ecdsa.Sign(rand.Reader, &priKey, dataToSign)
		if err != nil {
			log.Panic(err)
		}
		// Append the signature to the original transaction's input
		tx.Vin[inId].Signature = append(r.Bytes(), s.Bytes()...)
	}

}

func prepareForSigning(txCopy *Transaction, inId int, prevTx *Transaction) {
	txCopy.Vin[inId].Signature = nil
	txCopy.Vin[inId].PubKey = prevTx.Hash()
	txCopy.ID = txCopy.Hash()
	txCopy.Vin[inId].PubKey = nil
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inId, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		prepareForSigning(&txCopy, inId, &prevTx)
		dataToVerify := txCopy.ID
		//utils.PrintJsonLog(&prevTx, "Verify")
		//txCopy.Vin[inId].Signature = nil
		//txCopy.Vin[inId].PubKey = prevTx.VOut[vin.Vout].PubKeyHash
		//dataToVerify := []byte(fmt.Sprintf("%x\n", txCopy))

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])
		// logrus.Infof("r '%x': , s '%x', id: '%x'", r.Bytes(), s.Bytes(), txCopy.ID)

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, dataToVerify, &r, &s) {
			return false
		}
	}

	return true
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TxInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.VOut {
		outputs = append(outputs, TxOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}
