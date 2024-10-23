package core

// TxInput represents an input in a transaction.
type TxInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}
