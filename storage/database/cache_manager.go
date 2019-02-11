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
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
	"reflect"
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

// NOTE-Klaytn-Cache BlockChain Caches
// Below is the list of the constants for cache size.
// TODO-Klaytn: Below should be handled by ini or other configurations.
const (
	maxBodyCache           = 256
	maxBlockCache          = 256
	maxRecentTransactions  = 30000
	maxRecentBlockReceipts = 30
	maxRecentTxReceipt     = 30000
)

const (
	numShardsBodyCache           = 4096
	numShardsBlockCache          = 4096
	numShardsRecentTransactions  = 4096
	numShardsRecentBlockReceipts = 4096
	numShardsRecentTxReceipt     = 4096
)

type blockChainCacheKey int

const (
	bodyCacheIndex blockChainCacheKey = iota
	bodyRLPCacheIndex
	blockCacheIndex
	recentTxAndLookupInfoIndex
	recentBlockReceiptsIndex
	recentTxReceiptIndex

	blockCacheKeySize
)

var blockLRUCacheConfig = [blockCacheKeySize]common.CacheConfiger{
	bodyCacheIndex:             common.LRUConfig{CacheSize: maxBodyCache},
	bodyRLPCacheIndex:          common.LRUConfig{CacheSize: maxBodyCache},
	blockCacheIndex:            common.LRUConfig{CacheSize: maxBlockCache},
	recentTxAndLookupInfoIndex: common.LRUConfig{CacheSize: maxRecentTransactions},
	recentBlockReceiptsIndex:   common.LRUConfig{CacheSize: maxRecentBlockReceipts},
	recentTxReceiptIndex:       common.LRUConfig{CacheSize: maxRecentTxReceipt},
}

var blockLRUShardCacheConfig = [blockCacheKeySize]common.CacheConfiger{
	bodyCacheIndex:             common.LRUShardConfig{CacheSize: maxBodyCache, NumShards: numShardsBodyCache},
	bodyRLPCacheIndex:          common.LRUShardConfig{CacheSize: maxBodyCache, NumShards: numShardsBodyCache},
	blockCacheIndex:            common.LRUShardConfig{CacheSize: maxBlockCache, NumShards: numShardsBlockCache},
	recentTxAndLookupInfoIndex: common.LRUShardConfig{CacheSize: maxRecentTransactions, NumShards: numShardsRecentTransactions},
	recentBlockReceiptsIndex:   common.LRUShardConfig{CacheSize: maxRecentBlockReceipts, NumShards: numShardsRecentBlockReceipts},
	recentTxReceiptIndex:       common.LRUShardConfig{CacheSize: maxRecentTxReceipt, NumShards: numShardsRecentTxReceipt},
}

func newBlockChainCache(cacheNameKey blockChainCacheKey, cacheType common.CacheType) common.Cache {
	var cache common.Cache

	switch cacheType {
	case common.LRUCacheType:
		cache, _ = common.NewCache(blockLRUCacheConfig[cacheNameKey])
	case common.LRUShardCacheType:
		cache, _ = common.NewCache(blockLRUShardCacheConfig[cacheNameKey])
	default:
		cache, _ = common.NewCache(blockLRUCacheConfig[cacheNameKey])
	}
	return cache
}

type TransactionLookup struct {
	Tx *types.Transaction
	*TxLookupEntry
}

// cacheManager handles caches of data structures stored in Database.
// Previously, most of them were handled by blockchain.HeaderChain or
// blockchain.BlockChain.
type cacheManager struct {
	// caches from blockchain.HeaderChain
	headerCache      common.Cache
	tdCache          common.Cache
	blockNumberCache common.Cache

	// caches from blockchain.BlockChain
	bodyCache             common.Cache // Cache for the most recent block bodies
	bodyRLPCache          common.Cache // Cache for the most recent block bodies in RLP encoded format
	blockCache            common.Cache // Cache for the most recent entire blocks
	recentTxAndLookupInfo common.Cache // recent TX and LookupInfo cache
	recentBlockReceipts   common.Cache // recent block receipts cache
	recentTxReceipt       common.Cache // recent TX receipt cache
}

// newCacheManager returns a pointer of cacheManager with predefined configurations.
func newCacheManager() *cacheManager {
	cm := &cacheManager{
		headerCache:      newHeaderChainCache(hedearCacheIndex, common.DefaultCacheType),
		tdCache:          newHeaderChainCache(tdCacheIndex, common.DefaultCacheType),
		blockNumberCache: newHeaderChainCache(blockNumberCacheIndex, common.DefaultCacheType),

		bodyCache:    newBlockChainCache(bodyCacheIndex, common.DefaultCacheType),
		bodyRLPCache: newBlockChainCache(bodyRLPCacheIndex, common.DefaultCacheType),
		blockCache:   newBlockChainCache(blockCacheIndex, common.DefaultCacheType),

		recentTxAndLookupInfo: newBlockChainCache(recentTxAndLookupInfoIndex, common.DefaultCacheType),
		recentBlockReceipts:   newBlockChainCache(recentBlockReceiptsIndex, common.DefaultCacheType),
		recentTxReceipt:       newBlockChainCache(recentTxReceiptIndex, common.DefaultCacheType),
	}
	return cm
}

// clearHeaderChainCache flushes out 1) headerCache, 2) tdCache and 3) blockNumberCache.
func (cm *cacheManager) clearHeaderChainCache() {
	cm.headerCache.Purge()
	cm.tdCache.Purge()
	cm.blockNumberCache.Purge()
}

// clearBlockChainCache flushes out 1) bodyCache, 2) bodyRLPCache, 3) blockCache,
// 4) recentTxAndLookupInfo, 5) recentBlockReceipts and 6) recentTxReceipt.
func (cm *cacheManager) clearBlockChainCache() {
	cm.bodyCache.Purge()
	cm.bodyRLPCache.Purge()
	cm.blockCache.Purge()
	cm.recentTxAndLookupInfo.Purge()
	cm.recentBlockReceipts.Purge()
	cm.recentTxReceipt.Purge()
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

// readBodyCache looks for cached blockBody in bodyCache.
// It returns nil if not found.
func (cm *cacheManager) readBodyCache(hash common.Hash) *types.Body {
	if cachedBody, ok := cm.bodyCache.Get(hash); ok && cachedBody != nil {
		cacheGetBlockBodyHitMeter.Mark(1)
		return cachedBody.(*types.Body)
	}
	cacheGetBlockBodyMissMeter.Mark(1)
	return nil
}

// writeBodyCache writes blockBody as a value, blockHash as a key.
func (cm *cacheManager) writeBodyCache(hash common.Hash, body *types.Body) {
	if body == nil {
		return
	}
	cm.bodyCache.Add(hash, body)
}

// deleteBodyCache writes nil as a value, blockHash as a key, to indicate given
// txHash is deleted in bodyCache and bodyRLPCache.
func (cm *cacheManager) deleteBodyCache(hash common.Hash) {
	cm.bodyCache.Add(hash, nil)
	cm.bodyRLPCache.Add(hash, nil)
}

// readBodyRLPCache looks for cached RLP-encoded blockBody in bodyRLPCache.
// It returns nil if not found.
func (cm *cacheManager) readBodyRLPCache(hash common.Hash) rlp.RawValue {
	if cachedBodyRLP, ok := cm.bodyRLPCache.Get(hash); ok && cachedBodyRLP != nil {
		cacheGetBlockBodyRLPHitMeter.Mark(1)
		return cachedBodyRLP.(rlp.RawValue)
	}
	cacheGetBlockBodyRLPMissMeter.Mark(1)
	return nil
}

// writeBodyRLPCache writes RLP-encoded blockBody as a value, blockHash as a key.
func (cm *cacheManager) writeBodyRLPCache(hash common.Hash, bodyRLP rlp.RawValue) {
	if bodyRLP == nil {
		return
	}
	cm.bodyRLPCache.Add(hash, bodyRLP)
}

// readBlockCache looks for cached block in blockCache.
// It returns nil if not found.
func (cm *cacheManager) readBlockCache(hash common.Hash) *types.Block {
	if cachedBlock, ok := cm.blockCache.Get(hash); ok && cachedBlock != nil {
		cacheGetBlockHitMeter.Mark(1)
		return cachedBlock.(*types.Block)
	}
	cacheGetBlockMissMeter.Mark(1)
	return nil
}

// hasBlockInCache returns if given hash exists in blockCache.
func (cm *cacheManager) hasBlockInCache(hash common.Hash) bool {
	return cm.blockCache.Contains(hash)
}

// writeBlockCache writes block as a value, blockHash as a key.
func (cm *cacheManager) writeBlockCache(hash common.Hash, block *types.Block) {
	if block == nil {
		return
	}
	cm.blockCache.Add(hash, block)
}

// deleteBlockCache writes nil as a value, blockHash as a key,
// to indicate given blockHash is deleted in recentBlockReceipts.
func (cm *cacheManager) deleteBlockCache(hash common.Hash) {
	cm.blockCache.Add(hash, nil)
}

// readTxAndLookupInfoInCache looks for cached tx and its look up information in recentTxAndLookupInfo.
// It returns nil and empty values if not found.
func (cm *cacheManager) readTxAndLookupInfoInCache(txHash common.Hash) (*types.Transaction, common.Hash, uint64, uint64) {
	if value, ok := cm.recentTxAndLookupInfo.Get(txHash); ok && value != nil {
		cacheGetRecentTransactionsHitMeter.Mark(1)
		txLookup, ok := value.(*TransactionLookup)
		if !ok {
			logger.Error("invalid type in recentTxAndLookupInfo. expected=*TransactionLookup", "actual=", reflect.TypeOf(value))
			return nil, common.Hash{}, 0, 0
		}
		return txLookup.Tx, txLookup.BlockHash, txLookup.BlockIndex, txLookup.Index
	}
	cacheGetRecentTransactionsMissMeter.Mark(1)
	return nil, common.Hash{}, 0, 0
}

// writeTxAndLookupInfoCache writes a tx and its lookup information as a value, txHash as a key.
func (cm *cacheManager) writeTxAndLookupInfoCache(txHash common.Hash, txLookup *TransactionLookup) {
	if txLookup == nil {
		return
	}
	cm.recentTxAndLookupInfo.Add(txHash, txLookup)
}

// readBlockReceiptsInCache looks for cached blockReceipts in recentBlockReceipts.
// It returns nil if not found.
func (cm *cacheManager) readBlockReceiptsInCache(blockHash common.Hash) types.Receipts {
	if cachedBlockReceipts, ok := cm.recentBlockReceipts.Get(blockHash); ok && cachedBlockReceipts != nil {
		cacheGetRecentBlockReceiptsHitMeter.Mark(1)
		return cachedBlockReceipts.(types.Receipts)
	}
	cacheGetRecentBlockReceiptsMissMeter.Mark(1)
	return nil
}

// writeBlockReceiptsCache writes blockReceipts as a value, blockHash as a key.
func (cm *cacheManager) writeBlockReceiptsCache(blockHash common.Hash, receipts types.Receipts) {
	if receipts == nil {
		return
	}
	cm.recentBlockReceipts.Add(blockHash, receipts)
}

// deleteBlockReceiptsCache writes nil as a value, blockHash as a key, to indicate given
// blockHash is deleted in recentBlockReceipts.
func (cm *cacheManager) deleteBlockReceiptsCache(blockHash common.Hash) {
	cm.recentBlockReceipts.Add(blockHash, nil)
}

// readTxReceiptInCache looks for cached txReceipt in recentTxReceipt.
// It returns nil if not found.
func (cm *cacheManager) readTxReceiptInCache(txHash common.Hash) *types.Receipt {
	if cachedReceipt, ok := cm.recentTxReceipt.Get(txHash); ok && cachedReceipt != nil {
		cacheGetRecentTxReceiptHitMeter.Mark(1)
		return cachedReceipt.(*types.Receipt)
	}
	cacheGetRecentTxReceiptMissMeter.Mark(1)
	return nil
}

// writeTxReceiptCache writes txReceipt as a value, txHash as a key.
func (cm *cacheManager) writeTxReceiptCache(txHash common.Hash, receipt *types.Receipt) {
	if receipt == nil {
		return
	}
	cm.recentTxReceipt.Add(txHash, receipt)
}

// deleteTxReceiptCache writes  writes nil as a value, blockHash as a key, to indicate given
// txHash is deleted in recentTxReceipt.
func (cm *cacheManager) deleteTxReceiptCache(txHash common.Hash) {
	cm.recentTxReceipt.Add(txHash, nil)
}
