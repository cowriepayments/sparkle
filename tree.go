package sparkle

import (
	"math/big"

	"github.com/umran/crypto"
)

// Tree ...
type Tree struct {
	store         Store
	levelDefaults map[uint8]crypto.Hash
}

// AddLeaf ...
func (tree *Tree) AddLeaf(index []byte, data []byte) {
	node := &Node{
		level:  0,
		prefix: big.NewInt(0).SetBytes(index),
	}
	nodeValue := crypto.GenerateHash(data)
	ancestorNode := node.getAncestor()

	tree.store.ExecTx(func(tx Transaction) {
		tx.SetValue(node.key(), nodeValue)
		for ancestorNode != nil {
			siblingNode := node.getSibling()
			siblingValue := tx.GetValue(siblingNode.key(), tx.CurrentEpoch())
			if siblingValue == nil {
				siblingValue = tree.levelDefaults[siblingNode.level]
			}

			var ancestorValue crypto.Hash
			if node.isLeft() {
				ancestorValue = crypto.GenerateHash(nodeValue.Merge(siblingValue))
			} else {
				ancestorValue = crypto.GenerateHash(siblingValue.Merge(nodeValue))
			}

			tx.SetValue(ancestorNode.key(), ancestorValue)

			node = ancestorNode
			nodeValue = ancestorValue
			ancestorNode = node.getAncestor()
		}
	})
}

// PublishRoot ...
func (tree *Tree) PublishRoot() crypto.Hash {
	var root crypto.Hash
	tree.store.ExecTx(func(tx Transaction) {
		left := tx.GetValue("ff", tx.CurrentEpoch())
		right := tx.GetValue("ff01", tx.CurrentEpoch())

		if left == nil {
			left = tree.levelDefaults[255]
		}

		if right == nil {
			right = tree.levelDefaults[255]
		}

		root = crypto.GenerateHash(left.Merge(right))

		tx.CommitRoot(root)
	})

	return root
}

// GetRoot returns the most recently published root
func (tree *Tree) GetRoot() crypto.Hash {
	var root crypto.Hash
	tree.store.ExecTx(func(tx Transaction) {
		root = tx.GetRootByEpoch(tx.CurrentEpoch() - 1)
	})

	return root
}

// GenerateProof ...
func (tree *Tree) GenerateProof(index []byte, root crypto.Hash) []*ProofStep {
	proof := make([]*ProofStep, 256)

	tree.store.ExecTx(func(tx Transaction) {
		// note that if the root doesn't really exist we will
		// get back an epoch of 0, which will still be valid
		// against the first root committed. a proof will still
		// be generated in such a case
		epoch := tx.GetEpochByRoot(root)

		node := &Node{
			level:  0,
			prefix: big.NewInt(0).SetBytes(index),
		}

		// get the sibling value of the leaf node
		leafSibling := node.getSibling()
		leafSiblingValue := tx.GetValue(leafSibling.key(), epoch)
		if leafSiblingValue == nil {
			leafSiblingValue = tree.levelDefaults[leafSibling.level]
		}
		proof[0] = &ProofStep{
			Left:  leafSibling.isLeft(),
			Value: leafSiblingValue,
		}

		// get all ancestor siblings
		for proofIndex, ancestor := 1, node.getAncestor(); ancestor != nil; proofIndex, ancestor = proofIndex+1, ancestor.getAncestor() {
			sibling := ancestor.getSibling()
			siblingValue := tx.GetValue(sibling.key(), epoch)
			if siblingValue == nil {
				siblingValue = tree.levelDefaults[sibling.level]
			}
			proof[proofIndex] = &ProofStep{
				Left:  sibling.isLeft(),
				Value: siblingValue,
			}
		}
	})

	return proof
}

// VerifyProof ...
func VerifyProof(data []byte, steps []*ProofStep, root crypto.Hash) bool {
	rootCandidate := crypto.GenerateHash(data)
	for _, step := range steps {
		if step.Left {
			rootCandidate = crypto.GenerateHash(step.Value.Merge(rootCandidate))
		} else {
			rootCandidate = crypto.GenerateHash(rootCandidate.Merge(step.Value))
		}
	}

	return root.Equal(rootCandidate)
}

// NewTree ...
func NewTree(store Store) *Tree {
	return &Tree{
		store:         store,
		levelDefaults: generateLevelDefaults(),
	}
}

func generateLevelDefaults() map[uint8]crypto.Hash {
	defaults := make(map[uint8]crypto.Hash)

	for i := 0; i < 256; i++ {
		if i == 0 {
			defaults[uint8(i)] = crypto.GenerateHash(make([]byte, 0))
		}

		defaults[uint8(i)] = crypto.GenerateHash(defaults[uint8(i)-1].Merge(defaults[uint8(i)-1]))
	}

	return defaults
}
