package node

import (
	"blockchain-from-scratch/core"
)

const commandLength = 12
const protocol = "tcp"
const nodeVersion = 1

var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"}
var minerNodes = []string{}
var blocksInTransit = [][]byte{}
var mempool = make(map[string]core.Transaction)

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return string(command)
}

func extractCommand(request []byte) []byte {
	return request[:commandLength]
}
