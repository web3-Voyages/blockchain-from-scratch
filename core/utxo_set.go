package core

// UTXOSet represents UTXO set
type UTXOSet struct {
	Blockchain *Blockchain
}

func (chain *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	//unspentTxs := chain.FindUnspentTransactions(address)
	accumulated := 0

	return accumulated, unspentOutputs
}

//func (chain *Blockchain) FindUnspentTransactions(address string) []Transaction {
//}
