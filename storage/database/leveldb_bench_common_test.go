package database

import (
	"errors"
	"fmt"
	"github.com/ground-x/klaytn/common"
	"io/ioutil"
	"os"
	"testing"
)

const (
	// numDataInsertions is the amount of data to be pre-stored in the DB before it is read
	numDataInsertions = 1000 * 1000 * 2
	// readArea is the range of data to be attempted to read from the data stored in the db.
	readArea = numDataInsertions / 512
	// To read the data cached in the DB, the most recently input data of the DB should be read.
	// This is the offset to specify the starting point of the data to read.
	readCacheOffset = numDataInsertions - readArea
	// primeNumber is used as an interval for reading data.
	// It is a number used to keep the data from being read consecutively.
	// This interval is used because the prime number is small relative to any number except for its multiple,
	// so you can access the whole range of data by cycling the data by a small number of intervals.
	primeNumber = 71887
	// levelDBMemDBSize is the size of internal memory database of LevelDB. Data is first saved to memDB and then moved to persistent storage.
	levelDBMemDBSize = 64
)

// initTestDB creates the db and inputs the data in db for valSize
func initTestDB(valueSize int) (string, Database, [][]byte, error) {
	dir, err := ioutil.TempDir("", "bench-DB")
	if err != nil {
		return "", nil, nil, errors.New(fmt.Sprintf("can't create temporary directory: %v", err))
	}
	dbc := &DBConfig{Dir: dir, DBType: LEVELDB, LevelDBCacheSize: levelDBMemDBSize, LevelDBHandles: 0, ChildChainIndexing: false}
	db, err := newDatabase(dbc)
	if err != nil {
		return "", nil, nil, errors.New(fmt.Sprintf("can't create database: %v", err))
	}
	keys, values := genKeysAndValues(valueSize, numDataInsertions)

	for i, key := range keys {
		if err := db.Put(key, values[i]); err != nil {
			return "", nil, nil, errors.New(fmt.Sprintf("fail to put data to db: %v", err))
		}
	}

	return dir, db, keys, nil
}

// benchmarkReadDBFromFile is a benchmark function that reads the data stored in the ldb file.
// Reads the initially entered data to read the value stored in the file.
func benchmarkReadDBFromFile(b *testing.B, valueSize int) {
	dir, db, keys, err := initTestDB(valueSize)
	defer os.RemoveAll(dir)
	defer db.Close()
	if err != nil {
		b.Fatalf("database initialization error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Get(keys[(i*primeNumber)%readArea])
	}
}

// benchmarkReadDBFromMemDB is a benchmark function that reads data stored in memDB.
// Read the data entered later to read the value stored in memDB, not in the disk storage.
func benchmarkReadDBFromMemDB(b *testing.B, valueSize int) {
	dir, db, keys, err := initTestDB(valueSize)
	defer os.RemoveAll(dir)
	defer db.Close()
	if err != nil {
		b.Fatalf("database initialization error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Get(keys[(readCacheOffset)+(i*primeNumber)%readArea])
	}
}

// benchmarkReadCache is a benchmark function that reads the data stored in the cache.
func benchmarkReadCache(b *testing.B, valueSize int) {
	b.StopTimer()
	cache, _ := common.NewCache(common.LRUConfig{CacheSize: numDataInsertions})
	keys, values := genKeysAndValues(valueSize, numDataInsertions)
	hashKeys := make([]common.Hash, 0, numDataInsertions)

	for i, key := range keys {
		var hashKey common.Hash
		copy(hashKey[:], key[:32])
		hashKeys = append(hashKeys, hashKey)
		cache.Add(hashKey, values[i])
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(hashKeys[(i*primeNumber)%(numDataInsertions)])
	}
}

// getReadDataOptions is a test case for measuring read performance.
var getReadDataOptions = [...]struct {
	name        string
	valueLength int
	testFunc    func(b *testing.B, valueSize int)
}{
	{"DBFromFile1", 1, benchmarkReadDBFromFile},
	{"DBFromFile128", 128, benchmarkReadDBFromFile},
	{"DBFromFile256", 256, benchmarkReadDBFromFile},
	{"DBFromFile512", 512, benchmarkReadDBFromFile},

	{"DBFromMem1", 1, benchmarkReadDBFromMemDB},
	{"DBFromMem128", 128, benchmarkReadDBFromMemDB},
	{"DBFromMem256", 256, benchmarkReadDBFromMemDB},
	{"DBFromMem512", 512, benchmarkReadDBFromMemDB},

	{"Cache1", 1, benchmarkReadCache},
	{"Cache128", 128, benchmarkReadCache},
	{"Cache256", 256, benchmarkReadCache},
	{"Cache512", 512, benchmarkReadCache},
}

// Benchmark_read_data is a benchmark that measures data read performance in DB and cache.
func Benchmark_read_data(b *testing.B) {
	for _, bm := range getReadDataOptions {
		b.Run(bm.name, func(b *testing.B) {
			bm.testFunc(b, bm.valueLength)
		})
	}
}
