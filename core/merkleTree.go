package core

import "crypto/sha256"

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	rNode := MerkleNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		rNode.Data = hash[:]
	} else {
		prevHashs := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashs)
		rNode.Data = hash[:]
	}

	rNode.Left = left
	rNode.Right = right
	return &rNode
}

func NewMerkleTree(datas [][]byte) *MerkleTree {
	var nodes []MerkleNode

	if len(datas)%2 != 0 {
		//  In case there is an odd number of transactions, the last transaction is duplicated
		// (in the Merkle tree, not in the block!)
		datas = append(datas, datas[len(datas)-1])
	}

	// build leaf nodes
	for _, data := range datas {
		node := NewMerkleNode(nil, nil, data)
		nodes = append(nodes, *node)
	}

	// build tree level nodes
	for i := 0; i < len(datas)/2; i++ {
		var newLevel []MerkleNode
		for j := 0; j < len(nodes); j += 2 {
			levelNode := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, *levelNode)
		}
		nodes = newLevel
	}
	return &MerkleTree{&nodes[0]}
}
