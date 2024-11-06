package node

import (
	"blockchain-from-scratch/core"
	"blockchain-from-scratch/utils"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
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
	case "addr":
		handleAddr(request)
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
	payload := decodeRequest(request)
	requestNodeBestHeight := payload.BestHeight

	if myBestHeight < requestNodeBestHeight {
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > requestNodeBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}
	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
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

func sendAddr(address string) {
}
func handleAddr(request []byte) {
}

func sendBlock(addr string, b *core.Block) {
}
func handleBlock(request []byte, bc *core.Blockchain) {
}

func sendInv(address, kind string, items [][]byte) {
}
func handleInv(request []byte, bc *core.Blockchain) {
}

func sendGetBlocks(address string) {
}
func handleGetBlocks(request []byte, bc *core.Blockchain) {
}

func sendGetData(address, kind string, id []byte) {
}
func handleGetData(request []byte, bc *core.Blockchain) {
}

func sendTx(addr string, tnx *core.Transaction) {
}
func handleTx(request []byte, bc *core.Blockchain) {
}

func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}

func decodeRequest(request []byte) *version {
	var buff bytes.Buffer
	var payload version
	buff.Write(request[commandLength:])
	err := gob.NewDecoder(&buff).Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	return &payload
}
