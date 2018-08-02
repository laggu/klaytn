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

package common

import "github.com/hashicorp/golang-lru"

const (
	LRUCacheType = iota
	ARCCacheType
	//FREE_CACHE
	//BIG_CACHE
	//GO_CACHE
)

type Cache interface {
	Add(key, value interface{}) (evicted bool)
	Get(key interface{}) (value interface{}, ok bool)
	Contains(key interface{}) bool
	Purge()

	Keys() []interface{}
	Peek(key interface{}) (value interface{}, ok bool)
	Remove(key interface{})
	Len() int
}

type lruCache struct {
	lru *lru.Cache
}

func (cache *lruCache) Add(key, value interface{}) (evicted bool) {
	return cache.lru.Add(key, value)
}

func (cache *lruCache) Get(key interface{}) (value interface{}, ok bool) {
	value, ok = cache.lru.Get(key)
	cacheGetLRUTryMeter.Mark(1)
	if ok {
		cacheGetLRUHitMeter.Mark(1)
	}
	return
}

func (cache *lruCache) Contains(key interface{}) bool {
	return cache.lru.Contains(key)
}

func (cache *lruCache) Purge() {
	cache.lru.Purge()
}

func (cache *lruCache) Keys() []interface{} {
	return cache.lru.Keys()
}

func (cache *lruCache) Peek(key interface{}) (value interface{}, ok bool) {
	return cache.lru.Peek(key)
}

func (cache *lruCache) Remove(key interface{}) {
	cache.lru.Remove(key)
}

func (cache *lruCache) Len() int {
	return cache.lru.Len()
}

func NewLRUCache(size int) (*lruCache, error) {
	lru, err := lru.New(size)
	return &lruCache{lru}, err
}

type arcCache struct {
	arc *lru.ARCCache
}

func (cache *arcCache) Add(key, value interface{}) (evicted bool) {
	cache.arc.Add(key, value)
	//TODO-GX: need to be removed or should be added according to usage of evicted flag
	return true
}

func (cache *arcCache) Get(key interface{}) (value interface{}, ok bool) {
	return cache.arc.Get(key)
}

func (cache *arcCache) Contains(key interface{}) bool {
	return cache.arc.Contains(key)
}

func (cache *arcCache) Purge() {
	cache.arc.Purge()
}

func (cache *arcCache) Keys() []interface{} {
	return cache.arc.Keys()
}

func (cache *arcCache) Peek(key interface{}) (value interface{}, ok bool) {
	return cache.arc.Peek(key)
}

func (cache *arcCache) Remove(key interface{}) {
	cache.arc.Remove(key)
}

func (cache *arcCache) Len() int {
	return cache.arc.Len()
}

func NewARCCache(size int) (*arcCache, error) {
	arc, err := lru.NewARC(size)
	return &arcCache{arc}, err
}

func NewCache(cacheType, size int) (Cache, error) {
	var newCache Cache
	var err error

	switch cacheType {
	case LRUCacheType:
		newCache, err = NewLRUCache(size)
	case ARCCacheType:
		newCache, err = NewARCCache(size)
	default:
		// default caching policy = LRU
		newCache, err = NewLRUCache(size)
	}

	return newCache, err
}
