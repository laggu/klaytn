// Copyright 2018 The go-klaytn Authors
//
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
	"errors"
	"github.com/patrickmn/go-cache"
	"sync"
	"time"
)

type CacheDatabase struct {
	cache *cache.Cache
	lock  sync.RWMutex
}

func NewCacheDatabase() *CacheDatabase {
	return &CacheDatabase{
		cache: cache.New(10*time.Second, 1*time.Minute),
	}
}

func NewCacheDatabaseWithCap(size int) *CacheDatabase {
	return &CacheDatabase{
		cache: cache.New(30*time.Second, 1*time.Minute),
	}
}

func (db *CacheDatabase) Type() string {
	return CACHEDB
}

func (db *CacheDatabase) Put(key []byte, value []byte) error {
	db.cache.Set(string(key), value, cache.DefaultExpiration)
	return nil
}

func (db *CacheDatabase) Has(key []byte) (bool, error) {
	_, found := db.cache.Get(string(key))
	return found, nil
}

func (db *CacheDatabase) Get(key []byte) ([]byte, error) {
	value, found := db.cache.Get(string(key))
	if !found {
		return nil, errors.New("not found")
	}
	return value.([]byte), nil
}

func (db *CacheDatabase) Keys() [][]byte {
	keys := [][]byte{}
	for key := range db.cache.Items() {
		keys = append(keys, []byte(key))
	}
	return keys
}

func (db *CacheDatabase) Delete(key []byte) error {
	db.cache.Delete(string(key))
	return nil
}

func (db *CacheDatabase) Close() {}
