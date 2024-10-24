package core

// TxInput represents an input in a transaction.
type TxInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}

func (in *TxInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}
