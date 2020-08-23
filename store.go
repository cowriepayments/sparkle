package sparkle

import (
	"github.com/umran/crypto"
)

// Store ...
type Store interface {
	ExecTx(func(Transaction))
}

// Transaction ...
type Transaction interface {
	GetValue(key string, maxEpoch uint64) crypto.Hash
	SetValue(key string, value crypto.Hash)
	CurrentEpoch() uint64
	CommitRoot(root crypto.Hash)
	GetEpochByRoot(root crypto.Hash) uint64
	GetRootByEpoch(epoch uint64) crypto.Hash
}
