# Blockchain from Scratch

This project aims to implement a simple blockchain from scratch using Golang. The goal is to understand the core components of blockchain technology, such as block creation, chain linking, and mining algorithms.

## Features
- Block creation with basic data structure
- Hashing mechanism for linking blocks
- Consensus mechanism, maybe POW、POS...
- Peer-to-peer communication 
- Transaction management 
- .....more

## Getting Started
To run the project, clone the repository and follow the instructions below.

### Prerequisites
- Golang 1.23 or later

### Installation
```bash
git clone https://github.com/yourusername/blockchain-from-scratch.git
cd blockchain-from-scratch
```
### Run Command

1. createwallet
```bash
go run cmd/main.go createwallet
go run cmd/main.go createwallet
```

2. createblockchain
```bash
go run cmd/main.go createblockchain -address 15NNQXN8JyMrtPK3qzS1DE2hgGLABmkXWT
INFO[0000] NewCoinbaseTx to '15NNQXN8JyMrtPK3qzS1DE2hgGLABmkXWT'
INFO[0000] No existing blockchain found. Creating a new one...
INFO[0000] Mining a new block
009122593169229cd3b11ef4c6564ea3ab82c43540a28cade87b6b8c8563b308
INFO[0000] ======= Reindex UTXO ======
Done!
```

3. transfer
```bash
go run cmd/main.go send -from WALLET_1 -to WALLET_2 -amount 10 -mine
go run cmd/main.go send -from WALLET_3 -to WALLET_4 -amount 10
```
-mine 标志指的是块会立刻被同一节点挖出来, 不指定的话交易将由矿工打包出块

4. getBalance 
```bash
go run cmd/main.go getbalance  -address 15NNQXN8JyMrtPK3qzS1DE2hgGLABmkXWT
INFO[0000] Balance of '15NNQXN8JyMrtPK3qzS1DE2hgGLABmkXWT': 14
```

5. printChain
```bash
go run cmd/main.go printchain
```
  
7. // TODO....

## Release & Deliverable
- [docs](./docs)

## References
- [How does Blockchain work? - Simply Explained](https://www.youtube.com/watch?v=SSo_EIwHSd4)  - A YouTube video explaining blockchain basics in simple terms.
- [Build your own Blockchain / Cryptocurrency](https://github.com/EgoSay/build-your-own-x?tab=readme-ov-file#build-your-own-blockchain--cryptocurrency) - A list of how to design you own chain
- [Yu](https://github.com/yu-org/yu) - Yu is a highly customizable blockchain framework.
- [Building Blockchain in Go](https://jeiwan.net/posts/building-blockchain-in-go-part-1/) - A blog that guide you Building Blockchain in Go step by step
