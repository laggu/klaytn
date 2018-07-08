// Copyright 2014 The go-ethereum Authors
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

package gxdb_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/ground-x/go-gxplatform/gxdb"
)

func newTestLDB() (*gxdb.LDBDatabase, func()) {
	dirName, err := ioutil.TempDir(os.TempDir(), "gxdb_leveldb_test_")
	if err != nil {
		panic("failed to create test file: " + err.Error())
	}
	db, err := gxdb.NewLDBDatabase(dirName, 0, 0)
	if err != nil {
		panic("failed to create test database: " + err.Error())
	}

	return db, func() {
		db.Close()
		os.RemoveAll(dirName)
	}
}

func newTestBadgerDB() (*gxdb.BGDatabase, func()) {
	dirName, err := ioutil.TempDir(os.TempDir(), "gxdb_badgerdb_test_")
	if err != nil {
		panic("failed to create test file: " + err.Error())
	}
	db, err := gxdb.NewBGDatabase(dirName)
	if err != nil {
		panic("failed to create test database: " + err.Error())
	}

	return db, func() {
		db.Close()
		os.RemoveAll(dirName)
	}
}

var test_values = []string{"a", "1251", "\x00123\x00"}

//var test_values = []string{"", "a", "1251", "\x00123\x00"} original test_values; modified since badgerDB can't store empty key

func TestLDB_PutGet(t *testing.T) {
	db, remove := newTestLDB()
	defer remove()
	testPutGet(db, t)
}

func TestBadgerDB_PutGet(t *testing.T) {
	db, remove := newTestBadgerDB()
	defer remove()
	testPutGet(db, t)
}

func TestMemoryDB_PutGet(t *testing.T) {
	testPutGet(gxdb.NewMemDatabase(), t)
}

func testPutGet(db gxdb.Database, t *testing.T) {
	t.Parallel()

	// put
	for _, v := range test_values {
		err := db.Put([]byte(v), []byte(v))
		if err != nil {
			t.Fatalf("put failed: %v", err)
		}
	}

	// get
	for _, v := range test_values {
		data, err := db.Get([]byte(v))
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		if !bytes.Equal(data, []byte(v)) {
			t.Fatalf("get returned wrong result, got %q expected %q", string(data), v)
		}
	}

	// override with "?"
	for _, v := range test_values {
		err := db.Put([]byte(v), []byte("?"))
		if err != nil {
			t.Fatalf("put override failed: %v", err)
		}
	}

	// get "?" by key
	for _, v := range test_values {
		data, err := db.Get([]byte(v))
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		if !bytes.Equal(data, []byte("?")) {
			t.Fatalf("get returned wrong result, got %q expected ?", string(data))
		}
	}

	// override returned value
	for _, v := range test_values {
		orig, err := db.Get([]byte(v))
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		orig[0] = byte(0xff)
		data, err := db.Get([]byte(v))
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		if !bytes.Equal(data, []byte("?")) {
			t.Fatalf("get returned wrong result, got %q expected ?", string(data))
		}
	}

	// delete
	for _, v := range test_values {
		err := db.Delete([]byte(v))
		if err != nil {
			t.Fatalf("delete %q failed: %v", v, err)
		}
	}

	// try to get deleted values
	for _, v := range test_values {
		_, err := db.Get([]byte(v))
		if err == nil {
			t.Fatalf("got deleted value %q", v)
		}
	}
}

func TestLDB_ParallelPutGet(t *testing.T) {
	db, remove := newTestLDB()
	defer remove()
	testParallelPutGet(db, t)
}

func TestBadgerDB_ParallelPutGet(t *testing.T) {
	db, remove := newTestBadgerDB()
	defer remove()
	testParallelPutGet(db, t)
}

func TestMemoryDB_ParallelPutGet(t *testing.T) {
	testParallelPutGet(gxdb.NewMemDatabase(), t)
}

func testParallelPutGet(db gxdb.Database, t *testing.T) {
	const n = 8
	var pending sync.WaitGroup

	pending.Add(n)
	for i := 0; i < n; i++ {
		go func(key string) {
			defer pending.Done()
			err := db.Put([]byte(key), []byte("v"+key))
			if err != nil {
				panic("put failed: " + err.Error())
			}
		}(strconv.Itoa(i))
	}
	pending.Wait()

	pending.Add(n)
	for i := 0; i < n; i++ {
		go func(key string) {
			defer pending.Done()
			data, err := db.Get([]byte(key))
			if err != nil {
				panic("get failed: " + err.Error())
			}
			if !bytes.Equal(data, []byte("v"+key)) {
				panic(fmt.Sprintf("get failed, got %q expected %q", []byte(data), []byte("v"+key)))
			}
		}(strconv.Itoa(i))
	}
	pending.Wait()

	pending.Add(n)
	for i := 0; i < n; i++ {
		go func(key string) {
			defer pending.Done()
			err := db.Delete([]byte(key))
			if err != nil {
				panic("delete failed: " + err.Error())
			}
		}(strconv.Itoa(i))
	}
	pending.Wait()

	pending.Add(n)
	for i := 0; i < n; i++ {
		go func(key string) {
			defer pending.Done()
			_, err := db.Get([]byte(key))
			if err == nil {
				panic("got deleted value")
			}
		}(strconv.Itoa(i))
	}
	pending.Wait()
}
