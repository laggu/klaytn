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
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"math/big"
)

const (
	headerCacheLimit      = 512
	tdCacheLimit          = 1024
	blockNumberCacheLimit = 2048
)

const (
	numShardsHeaderCache      = 4096
	numShardsTdCache          = 4096
	numShardsBlockNumberCache = 4096
)

type headerChainCacheKey int

const (
	hedearCacheIndex headerChainCacheKey = iota
	tdCacheIndex
	blockNumberCacheIndex

	headerChainCacheKeySize
)

var headerLRUCacheConfig = [headerChainCacheKeySize]common.CacheConfiger{
	hedearCacheIndex:      common.LRUConfig{CacheSize: headerCacheLimit},
	tdCacheIndex:          common.LRUConfig{CacheSize: tdCacheLimit},
	blockNumberCacheIndex: common.LRUConfig{CacheSize: blockNumberCacheLimit},
}

var headerLRUShardCacheConfig = [headerChainCacheKeySize]common.CacheConfiger{
	hedearCacheIndex:      common.LRUShardConfig{CacheSize: headerCacheLimit, NumShards: numShardsHeaderCache},
	tdCacheIndex:          common.LRUShardConfig{CacheSize: tdCacheLimit, NumShards: numShardsTdCache},
	blockNumberCacheIndex: common.LRUShardConfig{CacheSize: blockNumberCacheLimit, NumShards: numShardsBlockNumberCache},
}

func newHeaderChainCache(cacheNameKey headerChainCacheKey, cacheType common.CacheType) common.Cache {
	var cache common.Cache

	switch cacheType {
	case common.LRUCacheType:
		cache, _ = common.NewCache(headerLRUCacheConfig[cacheNameKey])
	case common.LRUShardCacheType:
		cache, _ = common.NewCache(headerLRUShardCacheConfig[cacheNameKey])
	default:
		cache, _ = common.NewCache(headerLRUCacheConfig[cacheNameKey])
	}
	return cache // NOTE-Klaytn-Cache HeaderChain Caches
}

// TODO-Klaytn-Storage Implementation for below caches will be handled in the next PR.
// NOTE-Klaytn-Cache BlockChain Caches
// Below is the list of the constants for cache size.
// TODO-Klaytn: Below should be handled by ini or other configurations.
//const (
//	maxBodyCache           = 256
//	maxBlockCache          = 256
//	maxFutureBlocks        = 256
//	maxTimeFutureBlocks    = 30
//	maxBadBlocks           = 10
//	maxRecentTransactions  = 30000
//	maxRecentBlockReceipts = 30
//	maxRecentTxReceipt     = 30000
//)
//
//const (
//	numShardsBodyCache           = 4096
//	numShardsBlockCache          = 4096
//	numShardsRecentTransactions  = 4096
//	numShardsRecentBlockReceipts = 4096
//	numShardsRecentTxReceipt     = 4096
//)
//
//const (
//	triesInMemory = 128
//	// BlockChainVersion ensures that an incompatible database forces a resync from scratch.
//	BlockChainVersion    = 3
//	DefaultBlockInterval = 128
//)
//
//type blockChainCacheKey int
//
//const (
//	bodyCacheIndex blockChainCacheKey = iota
//	bodyRLPCacheIndex
//	blockCacheIndex
//	recentTxAndLookupInfoIndex
//	recentBlockReceiptsIndex
//	recentTxReceiptIndex
//
//	blockCacheKeySize
//)
//
//var blockLRUCacheConfig = [blockCacheKeySize]common.CacheConfiger{
//	bodyCacheIndex:             common.LRUConfig{CacheSize: maxBodyCache},
//	bodyRLPCacheIndex:          common.LRUConfig{CacheSize: maxBodyCache},
//	blockCacheIndex:            common.LRUConfig{CacheSize: maxBlockCache},
//	recentTxAndLookupInfoIndex: common.LRUConfig{CacheSize: maxRecentTransactions},
//	recentBlockReceiptsIndex:   common.LRUConfig{CacheSize: maxRecentBlockReceipts},
//	recentTxReceiptIndex:       common.LRUConfig{CacheSize: maxRecentTxReceipt},
//}
//
//var blockLRUShardCacheConfig = [blockCacheKeySize]common.CacheConfiger{
//	bodyCacheIndex:             common.LRUShardConfig{CacheSize: maxBodyCache, NumShards: numShardsBodyCache},
//	bodyRLPCacheIndex:          common.LRUShardConfig{CacheSize: maxBodyCache, NumShards: numShardsBodyCache},
//	blockCacheIndex:            common.LRUShardConfig{CacheSize: maxBlockCache, NumShards: numShardsBlockCache},
//	recentTxAndLookupInfoIndex: common.LRUShardConfig{CacheSize: maxRecentTransactions, NumShards: numShardsRecentTransactions},
//	recentBlockReceiptsIndex:   common.LRUShardConfig{CacheSize: maxRecentBlockReceipts, NumShards: numShardsRecentBlockReceipts},
//	recentTxReceiptIndex:       common.LRUShardConfig{CacheSize: maxRecentTxReceipt, NumShards: numShardsRecentTxReceipt},
//}
//
//func newBlockChainCache(cacheNameKey blockChainCacheKey, cacheType common.CacheType) common.Cache {
//	var cache common.Cache
//
//	switch cacheType {
//	case common.LRUCacheType:
//		cache, _ = common.NewCache(blockLRUCacheConfig[cacheNameKey])
//	case common.LRUShardCacheType:
//		cache, _ = common.NewCache(blockLRUShardCacheConfig[cacheNameKey])
//	default:
//		cache, _ = common.NewCache(blockLRUCacheConfig[cacheNameKey])
//	}
//	return cache
//}

// cacheManager handles caches of data structures stored in Database.
// Previously, most of them were handled by blockchain.HeaderChain or
// blockchain.BlockChain.
type cacheManager struct {
	// caches from blockchain.HeaderChain
	headerCache      common.Cache
	tdCache          common.Cache
	blockNumberCache common.Cache

	//TODO-Klaytn-Storage Implementation for below caches will be handled in the next PR.
	// caches from blockchain.BlockChain
	//bodyCache    common.Cache   // Cache for the most recent block bodies
	//bodyRLPCache common.Cache   // Cache for the most recent block bodies in RLP encoded format
	//blockCache   common.Cache   // Cache for the most recent entire blocks
	//recentTxAndLookupInfo common.Cache // recent TX and LookupInfo cache
	//recentBlockReceipts   common.Cache // recent block receipts cache
	//recentTxReceipt       common.Cache // recent TX receipt cache
}

// newCacheManager returns a pointer of cacheManager with predefined configurations.
func newCacheManager() *cacheManager {
	cm := &cacheManager{
		headerCache:      newHeaderChainCache(hedearCacheIndex, common.DefaultCacheType),
		tdCache:          newHeaderChainCache(tdCacheIndex, common.DefaultCacheType),
		blockNumberCache: newHeaderChainCache(blockNumberCacheIndex, common.DefaultCacheType),

		//TODO-Klaytn-Storage Implementation for below caches will be handled in the next PR.
		//bodyCache:    newBlockChainCache(bodyCacheIndex, common.DefaultCacheType),
		//bodyRLPCache: newBlockChainCache(bodyRLPCacheIndex, common.DefaultCacheType),
		//blockCache:   newBlockChainCache(blockCacheIndex, common.DefaultCacheType),
		//
		//recentTxAndLookupInfo: newBlockChainCache(recentTxAndLookupInfoIndex, common.DefaultCacheType),
		//recentBlockReceipts:   newBlockChainCache(recentBlockReceiptsIndex, common.DefaultCacheType),
		//recentTxReceipt:       newBlockChainCache(recentTxReceiptIndex, common.DefaultCacheType),
	}
	return cm
}

// clearHeaderCache flushes out 1) headerCache, 2) tdCache and 3) blockNumberCache.
func (cm *cacheManager) clearHeaderCache() {
	cm.headerCache.Purge()
	cm.tdCache.Purge()
	cm.blockNumberCache.Purge()
}

// readHeaderCache looks for cached header in headerCache.
// It returns nil if not found.
func (cm *cacheManager) readHeaderCache(hash common.Hash) *types.Header {
	if header, ok := cm.headerCache.Get(hash); ok && header != nil {
		cacheGetHeaderHitMeter.Mark(1)
		return header.(*types.Header)
	}
	cacheGetHeaderMissMeter.Mark(1)
	return nil
}

// writeHeaderCache writes header as a value, headerHash as a key.
func (cm *cacheManager) writeHeaderCache(hash common.Hash, header *types.Header) {
	if header == nil {
		return
	}
	cm.headerCache.Add(hash, header)
}

// deleteHeaderCache writes nil as a value, headerHash as a key, to indicate given
// headerHash is deleted in headerCache.
func (cm *cacheManager) deleteHeaderCache(hash common.Hash) {
	cm.headerCache.Add(hash, nil)
}

// hasHeaderInCache returns if a cachedHeader exists with given headerHash.
func (cm *cacheManager) hasHeaderInCache(hash common.Hash) bool {
	if cm.blockNumberCache.Contains(hash) || cm.headerCache.Contains(hash) {
		return true
	}
	return false
}

// readTdCache looks for cached total difficulty in tdCache.
// It returns nil if not found.
func (cm *cacheManager) readTdCache(hash common.Hash) *big.Int {
	if cached, ok := cm.tdCache.Get(hash); ok && cached != nil {
		cacheGetTDHitMeter.Mark(1)
		return cached.(*big.Int)
	}
	cacheGetTDMissMeter.Mark(1)
	return nil
}

// writeHeaderCache writes total difficulty as a value, headerHash as a key.
func (cm *cacheManager) writeTdCache(hash common.Hash, td *big.Int) {
	if td == nil {
		return
	}
	cm.tdCache.Add(hash, td)
}

// deleteTdCache writes nil as a value, headerHash as a key, to indicate given
// headerHash is deleted in TdCache.
func (cm *cacheManager) deleteTdCache(hash common.Hash) {
	cm.tdCache.Add(hash, nil)
}

// readBlockNumberCache looks for cached headerNumber in blockNumberCache.
// It returns nil if not found.
func (cm *cacheManager) readBlockNumberCache(hash common.Hash) *uint64 {
	if cached, ok := cm.blockNumberCache.Get(hash); ok {
		cacheGetBlockNumberHitMeter.Mark(1)
		blockNumber := cached.(uint64)
		return &blockNumber
	}
	cacheGetBlockNumberMissMeter.Mark(1)
	return nil
}

// writeHeaderCache writes headerNumber as a value, headerHash as a key.
func (cm *cacheManager) writeBlockNumberCache(hash common.Hash, number uint64) {
	cm.blockNumberCache.Add(hash, number)
}
