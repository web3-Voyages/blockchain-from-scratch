package node

import (
	"blockchain-from-scratch/core"
	"blockchain-from-scratch/utils"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/sirupsen/logrus"
)

func StartServer(nodeId, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeId)
	miningAddress = minerAddress
	nodeNet, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}

	defer nodeNet.Close()

	bc := core.NewBlockChain(nodeId)

	// If the current node is not the central node,
	// it must send a version message to the central node to query whether its blockchain is outdated.
	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}

	for {
		conn, err := nodeNet.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc)
	}
}

func handleConnection(conn net.Conn, bc *core.Blockchain) {
	request, err := io.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request, bc)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		handleVersion(request, bc)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
}

func sendVersion(addr string, bc *core.Blockchain) {
	bestHeight := bc.GetBestHeight()
	payload := utils.Serialize(version{nodeVersion, bestHeight, nodeAddress})
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}
func handleVersion(request []byte, bc *core.Blockchain) {
	myBestHeight := bc.GetBestHeight()
	var payload version
	decodeRequest(request, &payload)
	requestNodeBestHeight := payload.BestHeight

	if myBestHeight < requestNodeBestHeight {
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > requestNodeBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}
	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
		for _, node := range knownNodes {
			sendGetBlocks(node)
		}
	}
}

func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		// add new  node
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range knownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, addr)
			}
		}
		knownNodes = updatedNodes
		return
	}
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

// func sendAddr(address string) {
// }
// func handleAddr(request []byte) {
// }

func sendBlock(address string, b *core.Block) {
	data := nodeBlock{nodeAddress, utils.Serialize(b)}
	payload := utils.Serialize(data)
	request := append(commandToBytes("block"), payload...)
	sendData(address, request)
}
func handleBlock(request []byte, bc *core.Blockchain) {
	var payload nodeBlock
	decodeRequest(request, &payload)

	blockData := payload.Block
	var block core.Block
	utils.Deserialize(blockData, block)
	bc.AddBlock(&block)
	logrus.Info("Added block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)
		blocksInTransit = blocksInTransit[1:]
	} else {
		utxoSet := core.UTXOSet{Blockchain: bc}
		utxoSet.Reindex()
	}
}

func sendInv(address, kind string, items [][]byte) {
	inventory := inv{nodeAddress, kind, items}
	payload := utils.Serialize(inventory)
	request := append(commandToBytes("inv"), payload...)
	sendData(address, request)
}
func handleInv(request []byte, bc *core.Blockchain) {
	var payload inv
	decodeRequest(request, &payload)
	logrus.Infof("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items
		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, payload.Type, blockHash)

		// TODO process blocksInTransit
		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txId := payload.Items[0]
		// TODO mempool process
		if mempool[hex.EncodeToString(txId)].ID == nil {
			sendGetData(payload.AddrFrom, payload.Type, txId)
		}
	}
}

func sendGetBlocks(address string) {
	payload := utils.Serialize(getblocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)
	sendData(address, request)
}
func handleGetBlocks(request []byte, bc *core.Blockchain) {
	var payload getblocks
	decodeRequest(request, &payload)
	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)
}

func sendGetData(address, kind string, id []byte) {
	payload := utils.Serialize(getdata{address, kind, id})
	request := append(commandToBytes("getdata"), payload...)
	sendData(address, request)
}
func handleGetData(request []byte, bc *core.Blockchain) {
	var payload getdata
	decodeRequest(request, &payload)

	// TODO should check the block or tx is exist
	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]
		sendTx(payload.AddrFrom, &tx)
	}
}

func sendTx(address string, tnx *core.Transaction) {
	payload := utils.Serialize(tx{address, utils.Serialize(tnx)})
	request := append(commandToBytes("tx"), payload...)
	sendData(address, request)
}
func handleTx(request []byte, bc *core.Blockchain) {
	var payload tx
	decodeRequest(request, payload)
	txData := payload.Transaction
	var tx core.Transaction
	utils.Deserialize(txData, &tx)

	// add tx into mempool
	mempool[hex.EncodeToString(tx.ID)] = tx

	// if node is master node, just send inv 
	// TODO should be decentralized
	if nodeAddress != knownNodes[0] {
		for _, node := range knownNodes {
			if nodeAddress != node && node != payload.AddFrom {
				sendInv(payload.AddFrom, "tx", [][]byte{tx.ID})
			}
		}
	} else if len(mempool) >= 2 && len(miningAddress) > 0 {
		// if mempool is not empty , try to mine block 
	MineTransactions:
		var txs []*core.Transaction
		// verfiy tx
		for id := range mempool {
			tx := mempool[id]
			if bc.VerifyTransaction(&tx) {
				txs = append(txs, &tx)
			}
		}
		if len(txs) == 0 {
			logrus.Info("All transactions are invalid! Waiting for new ones...")
			return
		}

		// mine new block
		cbtx := core.NewCoinbaseTx(miningAddress, "")
		txs = append(txs, cbtx)
		newBlock := bc.MineBlock(txs)
		utxoSet := core.UTXOSet{Blockchain: bc}
		utxoSet.Reindex()

		for _, tx := range txs {
			txId := hex.EncodeToString(tx.ID)
			delete(mempool, txId)
		}

		for _, node := range knownNodes {
			if node != nodeAddress {
				sendInv(node, "block", [][]byte{newBlock.Hash})
			}
		}

		if len(mempool) > 0 {
			goto MineTransactions
		}
	}

}

func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}

func decodeRequest(request []byte, v interface{}) {
	var buff bytes.Buffer

	buff.Write(request[commandLength:])
	err := gob.NewDecoder(&buff).Decode(&v)
	if err != nil {
		log.Panic(err)
	}
}
