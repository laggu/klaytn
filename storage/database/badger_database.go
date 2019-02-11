// Copyright 2018 The klaytn Authors
// This file is part of the klaytn library.
//
// The klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the klaytn library. If not, see <http://www.gnu.org/licenses/>.

package database

import (
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/ground-x/klaytn/log"
	"os"
)

type badgerDB struct {
	fn string // filename for reporting
	db *badger.DB

	logger log.Logger // Contextual logger tracking the database path
}

func NewBGDatabase(path string) (*badgerDB, error) {

	localLogger := logger.NewWith("path", path)

	if fi, err := os.Stat(path); err == nil {
		if !fi.IsDir() {
			return nil, fmt.Errorf("badger/database: open %s: not a directory", path)
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, err
		}
	}

	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	// optional options
	opts.NumMemtables = 5
	opts.SyncWrites = false
	opts.NumCompactors = 3
	opts.DoNotCompact = true
	opts.ReadOnly = false

	db, err := badger.Open(opts)
	if err != nil {
		logger.Error("fail to open badger", err)
	}

	// (Re)check for errors and abort if opening of the db failed
	if err != nil {
		return nil, err
	}

	return &badgerDB{
		fn:     path,
		db:     db,
		logger: localLogger,
	}, nil
}

func (db *badgerDB) Type() DBType {
	return BadgerDB
}

// Path returns the path to the database directory.
func (db *badgerDB) Path() string {
	return db.fn
}

// Put puts the given key / value to the queue
func (db *badgerDB) Put(key []byte, value []byte) error {
	// Generate the data to write to disk, update the meter and write
	//value = rle.Compress(value)

	txn := db.db.NewTransaction(true)
	defer txn.Discard()

	err := txn.Set(key, value)
	if err != nil {
		return err
	}
	return txn.Commit(nil)
}

func (db *badgerDB) Has(key []byte) (bool, error) {
	// Retrieve the key and increment the miss counter if not found
	txn := db.db.NewTransaction(false)
	defer txn.Discard()

	item, err := txn.Get(key)
	// badger.ErrKeyNotFound
	if err != nil {
		return false, err
	}
	value, err := item.Value()
	return value != nil, err
}

// Get returns the given key if it's present.
func (db *badgerDB) Get(key []byte) ([]byte, error) {
	// Retrieve the key and increment the miss counter if not found
	txn := db.db.NewTransaction(false)
	defer txn.Discard()

	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}
	return item.Value()
}

// Delete deletes the key from the queue and database
func (db *badgerDB) Delete(key []byte) error {
	// Execute the actual operation
	txn := db.db.NewTransaction(true)
	defer txn.Discard()

	err := txn.Delete(key)
	if err != nil {
		return err
	}
	return txn.Commit(nil)
}

func (db *badgerDB) NewIterator() *badger.Iterator {
	// Execute the actual operation
	txn := db.db.NewTransaction(false)
	return txn.NewIterator(badger.DefaultIteratorOptions)
}

func (db *badgerDB) Close() {
	err := db.db.Close()
	if err == nil {
		db.logger.Info("Database closed")
	} else {
		db.logger.Error("Failed to close database", "err", err)
	}
}

func (db *badgerDB) LDB() *badger.DB {
	return db.db
}

func (db *badgerDB) NewBatch() Batch {

	txn := db.db.NewTransaction(true)

	return &bdBatch{db: db.db, txn: txn}
}

func (db *badgerDB) Meter(prefix string) {
	logger.Warn("badgerDB does not support metrics!")
}

type bdBatch struct {
	db   *badger.DB
	txn  *badger.Txn
	size int
}

func (b *bdBatch) Put(key, value []byte) error {

	err := b.txn.Set(key, value)
	b.size += len(value)

	return err
}

func (b *bdBatch) Write() error {
	return b.txn.Commit(nil)
}

func (b *bdBatch) ValueSize() int {
	return b.size
}

func (b *bdBatch) Reset() {
	b.txn.Discard()
	b.size = 0
}

type bdtable struct {
	db     Database
	prefix string
}

func (dt *bdtable) Type() DBType {
	return dt.db.Type()
}

func (dt *bdtable) Put(key []byte, value []byte) error {
	return dt.db.Put(append([]byte(dt.prefix), key...), value)
}

func (dt *bdtable) Has(key []byte) (bool, error) {
	return dt.db.Has(append([]byte(dt.prefix), key...))
}

func (dt *bdtable) Get(key []byte) ([]byte, error) {
	return dt.db.Get(append([]byte(dt.prefix), key...))
}

func (dt *bdtable) Delete(key []byte) error {
	return dt.db.Delete(append([]byte(dt.prefix), key...))
}

func (dt *bdtable) Close() {
	// Do nothing; don't close the underlying DB.
}

func (dt *bdtable) Meter(prefix string) {
	dt.db.Meter(prefix)
}

type bdtableBatch struct {
	batch  Batch
	prefix string
}

func (dt *bdtable) NewBatch() Batch {
	return &bdtableBatch{dt.db.NewBatch(), dt.prefix}
}

func (tb *bdtableBatch) Put(key, value []byte) error {
	return tb.batch.Put(append([]byte(tb.prefix), key...), value)
}

func (tb *bdtableBatch) Write() error {
	return tb.batch.Write()
}

func (tb *bdtableBatch) ValueSize() int {
	return tb.batch.ValueSize()
}

func (tb *bdtableBatch) Reset() {
	tb.batch.Reset()
}
