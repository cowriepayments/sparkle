package sparkle

import (
	"github.com/umran/crypto"
)

// Tree ...
type Tree struct {
	store         Store
	levelDefaults map[uint8]crypto.Hash
}

// AddLeaf ...
func (tree *Tree) AddLeaf(index []byte, data []byte) error {
	node := NewLeafNode(index)

	nodeValue := crypto.GenerateHash(data)
	parentNode := node.getParent()

	return tree.store.ExecTx(func(tx Transaction) error {
		currentEpoch, err := tx.CurrentEpoch()
		if err != nil {
			return err
		}

		if err := tx.SetValue(node.key(), nodeValue); err != nil {
			return err
		}

		for parentNode != nil {
			siblingNode := node.getSibling()
			siblingValue, err := tx.GetValue(siblingNode.key(), currentEpoch)
			if err != nil {
				return err
			}

			if siblingValue == nil {
				siblingValue = tree.levelDefaults[siblingNode.level]
			}

			var parentValue crypto.Hash
			if node.isLeft() {
				parentValue = crypto.GenerateHash(nodeValue.Merge(siblingValue))
			} else {
				parentValue = crypto.GenerateHash(siblingValue.Merge(nodeValue))
			}

			if err := tx.SetValue(parentNode.key(), parentValue); err != nil {
				return err
			}

			node = parentNode
			nodeValue = parentValue
			parentNode = node.getParent()
		}

		return nil
	})
}

// PublishRoot ...
func (tree *Tree) PublishRoot() (crypto.Hash, error) {
	var root crypto.Hash

	err := tree.store.ExecTx(func(tx Transaction) error {
		currentEpoch, err := tx.CurrentEpoch()
		if err != nil {
			return err
		}

		left, err := tx.GetValue("ff", currentEpoch)
		if err != nil {
			return err
		}

		right, err := tx.GetValue("ff01", currentEpoch)
		if err != nil {
			return err
		}

		if left == nil {
			left = tree.levelDefaults[255]
		}

		if right == nil {
			right = tree.levelDefaults[255]
		}

		root = crypto.GenerateHash(left.Merge(right))
		return tx.CommitRoot(root)
	})

	return root, err
}

// GetRoot returns the most recently published root
func (tree *Tree) GetRoot() (crypto.Hash, error) {
	var root crypto.Hash

	err := tree.store.ExecTx(func(tx Transaction) error {
		currentEpoch, err := tx.CurrentEpoch()
		if err != nil {
			return err
		}

		root, err = tx.GetRootByEpoch(currentEpoch - 1)
		return err
	})

	return root, err
}

// GenerateProof ...
func (tree *Tree) GenerateProof(index []byte, root crypto.Hash) ([]*ProofStep, error) {
	proof := make([]*ProofStep, 256)

	err := tree.store.ExecTx(func(tx Transaction) error {
		// note that if the root doesn't really exist we will
		// get back an epoch of 0, which will still be valid
		// against the first root committed. a proof will still
		// be generated in such a case
		epoch, err := tx.GetEpochByRoot(root)
		if err != nil {
			return err
		}

		node := NewLeafNode(index)
		leafSibling := node.getSibling()
		leafSiblingValue, err := tx.GetValue(leafSibling.key(), epoch)
		if err != nil {
			return err
		}

		if leafSiblingValue == nil {
			leafSiblingValue = tree.levelDefaults[leafSibling.level]
		}

		proof[0] = &ProofStep{
			Left:  leafSibling.isLeft(),
			Value: leafSiblingValue,
		}

		// get all parent siblings
		for proofIndex, parent := 1, node.getParent(); parent != nil; proofIndex, parent = proofIndex+1, parent.getParent() {
			sibling := parent.getSibling()
			siblingValue, err := tx.GetValue(sibling.key(), epoch)
			if err != nil {
				return err
			}

			if siblingValue == nil {
				siblingValue = tree.levelDefaults[sibling.level]
			}

			proof[proofIndex] = &ProofStep{
				Left:  sibling.isLeft(),
				Value: siblingValue,
			}
		}

		return nil
	})

	return proof, err
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
			continue
		}

		defaults[uint8(i)] = crypto.GenerateHash(defaults[uint8(i)-1].Merge(defaults[uint8(i)-1]))
	}

	return defaults
}
