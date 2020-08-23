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
func (ms *MemStore) ExecTx(task func(Transaction)) {
	ms.mux.Lock()
	defer ms.mux.Unlock()

	task(&MemStoreTransaction{
		store: ms,
	})
}

// MemStoreTransaction ...
type MemStoreTransaction struct {
	store *MemStore
}

// GetValue ...
func (tx *MemStoreTransaction) GetValue(key string, maxEpoch uint64) crypto.Hash {
	versions := tx.store.values[key]
	if versions == nil {
		return nil
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

	return latestVersion
}

// SetValue ...
func (tx *MemStoreTransaction) SetValue(key string, value crypto.Hash) {
	_, ok := tx.store.values[key]
	if !ok {
		tx.store.values[key] = make(map[uint64]crypto.Hash)
	}

	tx.store.values[key][tx.store.currentEpoch] = value
}

// CurrentEpoch ...
func (tx *MemStoreTransaction) CurrentEpoch() uint64 {
	return tx.store.currentEpoch
}

// CommitRoot ...
func (tx *MemStoreTransaction) CommitRoot(root crypto.Hash) {
	tx.store.rootsByEpoch[tx.store.currentEpoch] = root
	tx.store.epochsByRoot[root.HexString()] = tx.store.currentEpoch
	tx.store.currentEpoch++
}

// GetEpochByRoot ...
func (tx *MemStoreTransaction) GetEpochByRoot(root crypto.Hash) uint64 {
	return tx.store.epochsByRoot[root.HexString()]
}

// GetRootByEpoch ...
func (tx *MemStoreTransaction) GetRootByEpoch(epoch uint64) crypto.Hash {
	return tx.store.rootsByEpoch[epoch]
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
