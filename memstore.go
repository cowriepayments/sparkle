package sparkle

import (
	"sync"

	"github.com/umran/crypto"
)

// MemStore ...
type MemStore struct {
	mux          *sync.Mutex
	currentEpoch uint64
	values       map[string]map[uint64]crypto.Hash
	epochsByRoot map[string]uint64
	rootsByEpoch map[uint64]crypto.Hash
}

// ExecTx ...
func (ms *MemStore) ExecTx(handler func(Transaction) error) error {
	ms.mux.Lock()
	defer ms.mux.Unlock()

	return handler(&MemStoreTransaction{
		store: ms,
	})
}

// MemStoreTransaction ...
type MemStoreTransaction struct {
	store *MemStore
}

// GetValue ...
func (tx *MemStoreTransaction) GetValue(key string, maxEpoch uint64) (crypto.Hash, error) {
	versions := tx.store.values[key]
	if versions == nil {
		return nil, nil
	}

	var (
		latestEpoch   uint64
		latestVersion crypto.Hash
	)

	for epoch, version := range versions {
		if epoch >= latestEpoch && epoch <= maxEpoch {
			latestEpoch = epoch
			latestVersion = version
		}
	}

	return latestVersion, nil
}

// SetValue ...
func (tx *MemStoreTransaction) SetValue(key string, value crypto.Hash) error {
	_, ok := tx.store.values[key]
	if !ok {
		tx.store.values[key] = make(map[uint64]crypto.Hash)
	}

	tx.store.values[key][tx.store.currentEpoch] = value
	return nil
}

// CurrentEpoch ...
func (tx *MemStoreTransaction) CurrentEpoch() (uint64, error) {
	return tx.store.currentEpoch, nil
}

// CommitRoot ...
func (tx *MemStoreTransaction) CommitRoot(root crypto.Hash) error {
	tx.store.rootsByEpoch[tx.store.currentEpoch] = root
	tx.store.epochsByRoot[root.HexString()] = tx.store.currentEpoch
	tx.store.currentEpoch++

	return nil
}

// GetEpochByRoot ...
func (tx *MemStoreTransaction) GetEpochByRoot(root crypto.Hash) (uint64, error) {
	return tx.store.epochsByRoot[root.HexString()], nil
}

// GetRootByEpoch ...
func (tx *MemStoreTransaction) GetRootByEpoch(epoch uint64) (crypto.Hash, error) {
	return tx.store.rootsByEpoch[epoch], nil
}

// NewMemStore ...
func NewMemStore() *MemStore {
	return &MemStore{
		mux:          new(sync.Mutex),
		currentEpoch: 0,
		values:       make(map[string]map[uint64]crypto.Hash),
		epochsByRoot: make(map[string]uint64),
		rootsByEpoch: make(map[uint64]crypto.Hash),
	}
}
