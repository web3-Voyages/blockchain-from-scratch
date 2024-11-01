package core

import (
	"blockchain-from-scratch/core/wallet"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

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
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TxInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{nil, []TxInput{txin}, []TxOutput{*txout}}
	tx.ID = tx.Hash()
	logrus.Infof("NewCoinbaseTx to '%s'", to)
	return &tx
}

func NewUTXOTransaction(from, to string, amount int, chain *Blockchain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	wallets, err := wallet.NewWallets()
	if err != nil {
		log.Panic(err)
	}

	fromWallet := wallets.GetWallet(from)
	pubKeyHash := wallet.HashPubKey(fromWallet.PublicKey)
	acc, validOutputs := chain.FindSpendableOutputs(pubKeyHash, amount)
	if acc < amount {
		log.Panic("Error: Not enough funds")
	}

	for txId, outs := range validOutputs {
		txID, err := hex.DecodeString(txId)
		if err != nil {
			log.Panic(err)
		}
		for _, out := range outs {
			input := TxInput{txID, out, nil, fromWallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	// TODO there's a bug when account transfer to themselves
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}
	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	// TODO need SignTransaction
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
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
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

		// Sign the transaction ID with the private key
		r, s, err := ecdsa.Sign(rand.Reader, &priKey, txCopy.ID)
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

	for inId, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		prepareForSigning(&txCopy, inId, &prevTx)

		r := big.Int{}
        s := big.Int{}
        sigLen := len(vin.Signature)
        r.SetBytes(vin.Signature[:(sigLen / 2)])
        s.SetBytes(vin.Signature[(sigLen / 2):])

        x := big.Int{}
        y := big.Int{}
        keyLen := len(vin.PubKey)
        x.SetBytes(vin.PubKey[:(keyLen / 2)])
        y.SetBytes(vin.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, tx.Hash(), &r, &s) {
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
