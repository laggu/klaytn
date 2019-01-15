// Copyright 2018 The go-klaytn Authors
// This file is part of the go-klaytn library.
//
// The go-klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-klaytn library. If not, see <http://www.gnu.org/licenses/>.

package database

import (
	"github.com/ground-x/klaytn/common"
	"github.com/syndtr/goleveldb/leveldb"
	"strconv"
	"strings"
)

const (
	SHARDS = 4
)

type ShardDatabase struct {
	dbs []*levelDB
}

func NewShardDatabase(file string, cache int, handles int) (*ShardDatabase, error) {

	dbs := make([]*levelDB, SHARDS)
	for i := 0; i < SHARDS; i++ {
		shardName := file + "-" + string(i)
		db, err := NewLDBDatabase(shardName, cache, handles)
		if err != nil {
			return nil, err
		}
		dbs[i] = db
	}

	return &ShardDatabase{
		dbs: dbs,
	}, nil

}

func (db *ShardDatabase) Type() string {
	return SHARDDB
}

func getPartition(key []byte) int64 {
	hashstring := strings.TrimPrefix(common.Bytes2Hex(key), "0x")
	if len(hashstring) > 15 {
		hashstring = hashstring[:15]
	}
	seed, _ := strconv.ParseInt(hashstring, 16, 64)
	partition := seed % int64(SHARDS)

	return partition
}

// Put puts the given key / value to the queue
func (db *ShardDatabase) Put(key []byte, value []byte) error {
	return db.dbs[getPartition(key)].Put(key, value)
}

func (db *ShardDatabase) Has(key []byte) (bool, error) {
	return db.dbs[getPartition(key)].Has(key)
}

// Get returns the given key if it's present.
func (db *ShardDatabase) Get(key []byte) ([]byte, error) {
	return db.dbs[getPartition(key)].Get(key)
}

// Delete deletes the key from the queue and database
func (db *ShardDatabase) Delete(key []byte) error {
	return db.dbs[getPartition(key)].Delete(key)
}

func (db *ShardDatabase) Close() {
	for _, db := range db.dbs {
		db.Close()
	}
}

func (db *ShardDatabase) NewBatch() Batch {
	return &shardBatch{dbs: db.dbs}
}

type shardBatch struct {
	dbs  []*levelDB
	b    []*leveldb.Batch
	size int
}

func (b *shardBatch) Put(key, value []byte) error {
	b.dbs[getPartition(key)].Put(key, value)
	b.size += len(value)
	return nil
}

func (b *shardBatch) Write() error {
	return nil
}

func (b *shardBatch) ValueSize() int {
	return b.size
}

func (b *shardBatch) Reset() {
	b.size = 0
}

type shardTable struct {
	db     Database
	prefix string
}

func (dt *shardTable) Type() string {
	return dt.db.Type()
}

func (dt *shardTable) Put(key []byte, value []byte) error {
	return dt.db.Put(append([]byte(dt.prefix), key...), value)
}

func (dt *shardTable) Has(key []byte) (bool, error) {
	return dt.db.Has(append([]byte(dt.prefix), key...))
}

func (dt *shardTable) Get(key []byte) ([]byte, error) {
	return dt.db.Get(append([]byte(dt.prefix), key...))
}

func (dt *shardTable) Delete(key []byte) error {
	return dt.db.Delete(append([]byte(dt.prefix), key...))
}

func (dt *shardTable) Close() {
	// Do nothing; don't close the underlying DB.
}

type shardTableBatch struct {
	batch  Batch
	prefix string
}

func (dt *shardTable) NewBatch() Batch {
	return &shardTableBatch{dt.db.NewBatch(), dt.prefix}
}

func (tb *shardTableBatch) Put(key, value []byte) error {
	return tb.batch.Put(append([]byte(tb.prefix), key...), value)
}

func (tb *shardTableBatch) Write() error {
	return tb.batch.Write()
}

func (tb *shardTableBatch) ValueSize() int {
	return tb.batch.ValueSize()
}

func (tb *shardTableBatch) Reset() {
	tb.batch.Reset()
}
