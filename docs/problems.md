
## 遇到的问题汇总
- 需要注意自己给自己转账时, UTXO 是不会变化的, 即交易输出依旧是自己
```go
// Ensure correct balance when transferring to self
if from == to {
    outputs = append(outputs, *NewTXOutput(acc, from))
} else {
    outputs = append(outputs, *NewTXOutput(amount, to))
    if acc > amount {
        outputs = append(outputs, *NewTXOutput(acc-amount, from))
    }
}
```
  

- 初始块 `NewCoinbaseTx data` 数据应该随机生成，确保不会生成同样的交易 `hash id`, 保证交易 id 唯一性


- `transaction.Serialize` 使用 gob 序列化是不稳定的，存在数据结构以及内容都完全一致的对象序列化出来的结果不一致情况
```go
import "github.com/vmihailenco/msgpack/v5"
func (tx Transaction) Serialize() []byte {
  result, err := msgpack.Marshal(tx)
  if err != nil {
    log.Panic(err)
  }
  return result
}
```

- 确保每次新块生成并确认加入到链之后, 更新链上最新区块信息
```go
err = chain.Db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte(blocksBucket))
    err = b.Put(newBlock.Hash, utils.Serialize(newBlock))
    err = b.Put([]byte("l"), newBlock.Hash)
    chain.tip = newBlock.Hash
}
```



