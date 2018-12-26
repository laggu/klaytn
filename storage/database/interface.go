// Copyright 2018 the go-klaytn Authors
// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from ethdb/interface.go (2018/06/04).
// Modified and improved for the go-klaytn development.

package database

// Code using batches should try to add this much data to the batch.
// The value was determined empirically.
const (
	LEVELDB = "leveldb"
	BADGER  = "badger"
	MEMDB   = "memdb"
	CACHEDB = "cachedb"
	SHARDDB = "sharddb"

	IdealBatchSize = 100 * 1024
)

// Putter wraps the database write operation supported by both batches and regular databases.
type Putter interface {
	Put(key []byte, value []byte) error
}

// Database wraps all database operations. All methods are safe for concurrent use.
type Database interface {
	Putter
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	Delete(key []byte) error
	Close()
	NewBatch() Batch
	Type() string
	Meter(prefix string)
}

// Batch is a write-only database that commits changes to its host database
// when Write is called. Batch cannot be used concurrently.
type Batch interface {
	Putter
	ValueSize() int // amount of data in the batch
	Write() error
	// Reset resets the batch for reuse
	Reset()
}

// NewTable returns a Database object that prefixes all keys with a given
// string.
func NewTable(db Database, prefix string) Database {

	switch db.Type() {
	case LEVELDB:
		return &table{
			db:     db,
			prefix: prefix,
		}
	case BADGER:
		return &bdtable{
			db:     db,
			prefix: prefix,
		}
	default:
		return nil
	}
}

// NewTableBatch returns a Batch object which prefixes all keys with a given string.
func NewTableBatch(db Database, prefix string) Batch {

	switch db.Type() {
	case LEVELDB:
		return &tableBatch{db.NewBatch(), prefix}
	case BADGER:
		return &bdtableBatch{db.NewBatch(), prefix}
	default:
		return nil
	}
}
