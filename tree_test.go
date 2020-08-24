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
	root, _ := tree.PublishRoot()

	steps, _ := tree.GenerateProof(data, root)

	if !VerifyProof(data, steps, root) {
		t.Error("proof verification failed")
	}

	// lets add some more data
	data2, _ := crypto.GenerateNonce()
	tree.AddLeaf(data2, data2)

	// publish new root
	root2, _ := tree.PublishRoot()

	// test previous data against previously published root
	steps, _ = tree.GenerateProof(data, root)
	if !VerifyProof(data, steps, root) {
		t.Error("proof verification failed")
	}

	// test new data against previously published root
	steps, _ = tree.GenerateProof(data2, root)
	if VerifyProof(data2, steps, root) {
		t.Error("expected proof verification to fail")
	}

	// test previous data against newly published root
	steps, _ = tree.GenerateProof(data, root2)
	if !VerifyProof(data, steps, root2) {
		t.Error("proof verification failed")
	}

	// test new data against newly published root
	steps, _ = tree.GenerateProof(data2, root2)
	if !VerifyProof(data2, steps, root2) {
		t.Error("proof verification failed")
	}
}
