package database

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/ground-x/go-gxplatform/common"
	"strings"
	"strconv"
)

const (
	SHARDS = 4
)

type ShardDatabase struct {
    dbs  []*LDBDatabase
}

func NewShardDatabase(file string, cache int, handles int) (*ShardDatabase, error) {

	dbs := make([]*LDBDatabase,SHARDS)
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
	hashstring := strings.TrimPrefix(common.Bytes2Hex(key),"0x")
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
	return &sahrdBatch{dbs:db.dbs}
}

type sahrdBatch struct {
	dbs   []*LDBDatabase
	b     []*leveldb.Batch
	size int
}

func (b *sahrdBatch) Put(key, value []byte) error {
	b.dbs[getPartition(key)].Put(key,value)
	b.size += len(value)
	return nil
}

func (b *sahrdBatch) Write() error {
	return nil
}

func (b *sahrdBatch) ValueSize() int {
	return b.size
}

func (b *sahrdBatch) Reset() {
	b.size = 0
}

type shardtable struct {
	db     Database
	prefix string
}

func (dt *shardtable) Type() string {
	return dt.db.Type()
}

func (dt *shardtable) Put(key []byte, value []byte) error {
	return dt.db.Put(append([]byte(dt.prefix), key...), value)
}

func (dt *shardtable) Has(key []byte) (bool, error) {
	return dt.db.Has(append([]byte(dt.prefix), key...))
}

func (dt *shardtable) Get(key []byte) ([]byte, error) {
	return dt.db.Get(append([]byte(dt.prefix), key...))
}

func (dt *shardtable) Delete(key []byte) error {
	return dt.db.Delete(append([]byte(dt.prefix), key...))
}

func (dt *shardtable) Close() {
	// Do nothing; don't close the underlying DB.
}

type shardtableBatch struct {
	batch  Batch
	prefix string
}

func (dt *shardtable) NewBatch() Batch {
	return &shardtableBatch{dt.db.NewBatch(), dt.prefix}
}

func (tb *shardtableBatch) Put(key, value []byte) error {
	return tb.batch.Put(append([]byte(tb.prefix), key...), value)
}

func (tb *shardtableBatch) Write() error {
	return tb.batch.Write()
}

func (tb *shardtableBatch) ValueSize() int {
	return tb.batch.ValueSize()
}

func (tb *shardtableBatch) Reset() {
	tb.batch.Reset()
}
