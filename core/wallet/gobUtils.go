package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"math/big"
)

// Struct for wrapping private key for gob serialization
type privateKeyGob struct {
	D *big.Int
	X *big.Int
	Y *big.Int
}

func init() {
	gob.Register(elliptic.P256())
	gob.Register(privateKeyGob{})
}

// Convert PrivateKey to serializable format
func (w *Wallet) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	pkGob := privateKeyGob{
		D: w.PrivateKey.D,
		X: w.PrivateKey.X,
		Y: w.PrivateKey.Y,
	}

	err := enc.Encode(pkGob)
	if err != nil {
		return nil, err
	}

	err = enc.Encode(w.PublicKey)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Restore Wallet from serialized data
func (w *Wallet) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	var pkGob privateKeyGob
	err := dec.Decode(&pkGob)
	if err != nil {
		return err
	}

	w.PrivateKey = ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     pkGob.X,
			Y:     pkGob.Y,
		},
		D: pkGob.D,
	}

	return dec.Decode(&w.PublicKey)
}
