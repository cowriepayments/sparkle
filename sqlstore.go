package sparkle

import (
	"database/sql"

	"github.com/umran/crypto"
	"github.com/umran/db"
)

// SQLStore ...
type SQLStore struct {
	connection db.Connection
}

// ExecTx ...
func (s *SQLStore) ExecTx(handler func(Transaction) error) error {
	return s.connection.ExecTx(func(tx db.Transaction) error {
		return handler(&SQLStoreTransaction{
			tx: tx,
		})
	})
}

// SQLStoreTransaction ...
type SQLStoreTransaction struct {
	tx db.Transaction
}

// GetValue ...
func (st *SQLStoreTransaction) GetValue(key string, maxEpoch uint64) (crypto.Hash, error) {
	var valueHex string

	statement := `
		SELECT value AS valueHex
		FROM nodes
		WHERE key = $1 AND epoch <= $2
		ORDER BY epoch DESC
		LIMIT 1
	`

	err := st.tx.QueryRow(statement, key, maxEpoch).Scan(&valueHex)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return crypto.HashFromHexString(valueHex)
}

// SetValue ...
func (st *SQLStoreTransaction) SetValue(key string, value crypto.Hash) error {
	currentEpoch, err := st.CurrentEpoch()
	if err != nil {
		return err
	}

	statement := `
		INSERT INTO nodes (key, epoch, value)
		VALUES ($1, $2, $3)
		ON CONFLICT ON CONSTRAINT nodes_primary_key
		DO UPDATE SET value = $3
	`

	_, err = st.tx.Exec(statement, key, currentEpoch, value.HexString())
	return err
}

// CurrentEpoch ...
func (st *SQLStoreTransaction) CurrentEpoch() (uint64, error) {
	var epoch uint64

	statement := `
		SELECT epoch
		FROM state
		ORDER BY epoch DESC
		LIMIT 1
	`

	err := st.tx.QueryRow(statement).Scan(&epoch)
	if err != nil {
		if err == sql.ErrNoRows {
			return epoch, nil
		}
		return epoch, err
	}

	return epoch + 1, nil
}

// CommitRoot ...
func (st *SQLStoreTransaction) CommitRoot(root crypto.Hash) error {
	currentEpoch, err := st.CurrentEpoch()
	if err != nil {
		return err
	}

	statement := `
		INSERT INTO state (epoch, root)
		VALUES ($1, $2)
	`

	_, err = st.tx.Exec(statement, currentEpoch, root.HexString())
	return err
}

// GetEpochByRoot ...
func (st *SQLStoreTransaction) GetEpochByRoot(root crypto.Hash) (uint64, error) {
	var epoch uint64

	statement := `
		SELECT epoch
		FROM state
		WHERE root = $1
		ORDER BY epoch DESC
		LIMIT 1
	`

	err := st.tx.QueryRow(statement, root.HexString()).Scan(&epoch)
	if err != nil {
		return epoch, err
	}

	return epoch, nil
}

// GetRootByEpoch ...
func (st *SQLStoreTransaction) GetRootByEpoch(epoch uint64) (crypto.Hash, error) {
	var rootHex string

	statement := `
		SELECT root
		FROM state
		WHERE epoch = $1
	`

	err := st.tx.QueryRow(statement, epoch).Scan(&rootHex)
	if err != nil {
		return nil, err
	}

	return crypto.HashFromHexString(rootHex)
}

// NewSQLStore ...
func NewSQLStore(connection db.Connection) *SQLStore {
	return &SQLStore{
		connection: connection,
	}
}
