package core

import (
	"blockchain-from-scratch/core/wallet"
	"bytes"
)

// TxInput represents an input in a transaction.
type TxInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

func (in *TxInput) UsesKey(publicHash []byte) bool {
	lockingHash := wallet.HashPubKey(publicHash)

	return bytes.Compare(lockingHash, publicHash) == 0
}
