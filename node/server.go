package node

import (
	"blockchain-from-scratch/core"
	"blockchain-from-scratch/utils"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"}
var blocksInTransit = [][]byte{}
var mempool = make(map[string]core.Transaction)

type version struct {
	Version   int
	BestHight int
	AddrFrom  string
}

func StartServer(nodeId, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeId)
	miningAddress = minerAddress
	nodeNet, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer nodeNet.Close()

	bc := core.NewBlockChain(nodeId)
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

func sendVersion(addr string, bc *core.Blockchain) {
	bestHeight := bc.GetBestHeight()
	payload := utils.Serialize(version{nodeVersion, bestHeight, nodeAddress})
	request := append(utils.CommandToBytes("version"), payload...)
	sendData(addr, request)
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

func handleConnection(conn net.Conn, bc *core.Blockchain) {
	request, err := io.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := utils.BytesToCommand(request[:commandLength])
	fmt.Printf("Received %s command\n", command)

	switch command {
	// case "addr":
	// 	handleAddr(request)
	// case "block":
	// 	handleBlock(request, bc)
	// case "inv":
	// 	handleInv(request, bc)
	// case "getblocks":
	// 	handleGetBlocks(request, bc)
	// case "getdata":
	// 	handleGetData(request, bc)
	// case "tx":
	// 	handleTx(request, bc)
	// case "version":
	// 	handleVersion(request, bc)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
}
