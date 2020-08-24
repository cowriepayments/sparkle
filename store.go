package sparkle

import (
	"github.com/umran/crypto"
)

// Store ...
type Store interface {
	ExecTx(func(Transaction) error) error
}

// Transaction ...
type Transaction interface {
	GetValue(key string, maxEpoch uint64) (crypto.Hash, error)
	SetValue(key string, value crypto.Hash) error
	CurrentEpoch() (uint64, error)
	CommitRoot(root crypto.Hash) error
	GetEpochByRoot(root crypto.Hash) (uint64, error)
	GetRootByEpoch(epoch uint64) (crypto.Hash, error)
}
