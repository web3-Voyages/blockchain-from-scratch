package core

// TXOutput represents a transaction output
type TxOutput struct {
	Value        int
	ScriptPubKey string
}

// NewTXOutput create a new TXOutput
