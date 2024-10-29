package core

import (
	"blockchain-from-scratch/core/wallet"
	"bytes"
)

// TXOutput represents a transaction output
type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet.HashPubKey(address)
	// Slicing [1:len(pubKeyHash) - 4] here is to remove the version byte and checksum.
	// The version byte is usually the first byte, and the checksum is usually the last four bytes. See field version/addressChecksumLen.
	// The version byte is 0x00, which is one byte. In hexadecimal, each pair (e.g., 00) represents 1 byte.
	out.PubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTXOutput create a new TXOutput
func NewTXOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}
