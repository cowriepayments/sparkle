package sparkle

import "github.com/umran/crypto"

// ProofStep ...
type ProofStep struct {
	Left  bool
	Value crypto.Hash
}
