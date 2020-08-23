package sparkle

import (
	"encoding/hex"
	"math/big"
)

// Node ...
type Node struct {
	level  uint8
	prefix *big.Int
}

func (node *Node) isLeft() bool {
	return big.NewInt(0).Mod(node.prefix, big.NewInt(2)).Cmp(big.NewInt(0)) == 0
}

func (node *Node) getSibling() *Node {
	siblingNode := &Node{
		level: node.level,
	}

	if node.isLeft() {
		// sibling prefix is one greater than the current prefix
		siblingNode.prefix = big.NewInt(0).Add(node.prefix, big.NewInt(1))
	} else {
		// sibling prefix is one less than the current prefix
		siblingNode.prefix = big.NewInt(0).Add(node.prefix, big.NewInt(-1))
	}

	return siblingNode
}

func (node *Node) getParent() *Node {
	if node.level == 255 {
		return nil
	}

	ancestorNode := &Node{
		level: node.level + 1,
	}

	ancestorNode.prefix = big.NewInt(0).Rsh(node.prefix, 1)
	return ancestorNode
}

func (node *Node) key() string {
	levelByte := []byte{node.level}
	prefixBytes := node.prefix.Bytes()

	return hex.EncodeToString(levelByte) + hex.EncodeToString(prefixBytes)
}

// NewLeafNode ...
func NewLeafNode(index []byte) *Node {
	return &Node{
		level:  0,
		prefix: big.NewInt(0).SetBytes(index),
	}
}
