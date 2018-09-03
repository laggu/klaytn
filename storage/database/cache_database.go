package database

import (
	"sync"
	"errors"
	"github.com/patrickmn/go-cache"
	"time"
)


type CacheDatabase struct {
	cache  *cache.Cache
	lock sync.RWMutex
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
	db.cache.Set(string(key),value, cache.DefaultExpiration)
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
	for key,_ := range db.cache.Items() {
		keys = append(keys, []byte(key))
	}
	return keys
}

func (db *CacheDatabase) Delete(key []byte) error {
	db.cache.Delete(string(key))
	return nil
}

func (db *CacheDatabase) Close() {}



