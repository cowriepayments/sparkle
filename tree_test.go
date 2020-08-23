package sparkle

import (
	"testing"

	"github.com/umran/crypto"
)

func TestProof(t *testing.T) {
	store := NewMemStore()
	tree := NewTree(store)

	data, _ := crypto.GenerateNonce()

	tree.AddLeaf(data, data)
	root := tree.PublishRoot()

	steps := tree.GenerateProof(data, root)

	if !VerifyProof(data, steps, root) {
		t.Error("proof verification failed")
	}

	// lets add some more data
	data2, _ := crypto.GenerateNonce()
	tree.AddLeaf(data2, data2)

	// publish new root
	root2 := tree.PublishRoot()

	// test previous data against previously published root
	if !VerifyProof(data, tree.GenerateProof(data, root), root) {
		t.Error("proof verification failed")
	}

	// test new data against previously published root
	if VerifyProof(data2, tree.GenerateProof(data2, root), root) {
		t.Error("expected proof verification to fail")
	}

	// test previous data against newly published root
	if !VerifyProof(data, tree.GenerateProof(data, root2), root2) {
		t.Error("proof verification failed")
	}

	// test new data against newly published root
	if !VerifyProof(data2, tree.GenerateProof(data2, root2), root2) {
		t.Error("proof verification failed")
	}
}
