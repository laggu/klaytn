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
	"math/big"
	"github.com/ground-x/go-gxplatform/blockchain/types"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/log"
	"github.com/ground-x/go-gxplatform/params"
	"github.com/ground-x/go-gxplatform/ser/rlp"
	"github.com/ground-x/go-gxplatform/storage/rawdb"
)

type DBManager interface {
	// from accessors_chain.go
	ReadCanonicalHash(number uint64) common.Hash
	WriteCanonicalHash(hash common.Hash, number uint64)
	DeleteCanonicalHash(number uint64)

	ReadHeaderNumber(hash common.Hash) *uint64

	ReadHeadHeaderHash() common.Hash
	WriteHeadHeaderHash(hash common.Hash)

	ReadHeadBlockHash() common.Hash
	WriteHeadBlockHash(hash common.Hash)

	ReadHeadFastBlockHash() common.Hash
	WriteHeadFastBlockHash(hash common.Hash)

	ReadFastTrieProgress() uint64
	WriteFastTrieProgress(count uint64)

	HasHeader(hash common.Hash, number uint64) bool
	ReadHeader(hash common.Hash, number uint64) *types.Header
	ReadHeaderRLP(hash common.Hash, number uint64) rlp.RawValue
	WriteHeader(header *types.Header)
	DeleteHeader(hash common.Hash, number uint64)

	HasBody(hash common.Hash, number uint64) bool
	ReadBody(hash common.Hash, number uint64) *types.Body
	ReadBodyRLP(hash common.Hash, number uint64) rlp.RawValue
	WriteBody(hash common.Hash, number uint64, body *types.Body)
	WriteBodyRLP(hash common.Hash, number uint64, rlp rlp.RawValue)
	DeleteBody(hash common.Hash, number uint64)

	ReadTd(hash common.Hash, number uint64) *big.Int
	WriteTd(hash common.Hash, number uint64, td *big.Int)
	DeleteTd(hash common.Hash, number uint64)

	ReadReceipts(hash common.Hash, number uint64) types.Receipts
	WriteReceipts(hash common.Hash, number uint64, receipts types.Receipts)
	DeleteReceipts(hash common.Hash, number uint64)

	ReadBlock(hash common.Hash, number uint64) *types.Block
	WriteBlock(block *types.Block)
	DeleteBlock(hash common.Hash, number uint64)

	FindCommonAncestor(a, b *types.Header) *types.Header

	ReadIstanbulSnapshot(hash common.Hash) ([]byte, error)
	WriteIstanbulSnapshot(hash common.Hash, blob []byte) error

	WriteMerkleProof(key, value []byte) error

	ReadCachedTrieNode(hash common.Hash) ([]byte, error)
	ReadCachedTrieNodePreimage(secureKey []byte) ([]byte, error)


	// from accessors_indexes.go
	ReadTxLookupEntry(hash common.Hash) (common.Hash, uint64, uint64)
	WriteTxLookupEntries(block *types.Block)
	DeleteTxLookupEntry(hash common.Hash)

	ReadTransaction(hash common.Hash) (*types.Transaction, common.Hash, uint64, uint64)

	ReadReceipt(hash common.Hash) (*types.Receipt, common.Hash, uint64, uint64)

	ReadBloomBits(bit uint, section uint64, head common.Hash) ([]byte, error)
	WriteBloomBits(bit uint, section uint64, head common.Hash, bits []byte)

	ReadValidSections() ([]byte, error)
	WriteValidSections(encodedSections []byte)

	ReadSectionHead(encodedSection []byte) ([]byte, error)
	WriteSectionHead(encodedSection []byte, hash common.Hash)
	DeleteSectionHead(encodedSection []byte)


	// from accessors_metadata.go
	ReadDatabaseVersion() int
	WriteDatabaseVersion(version int)

	ReadChainConfig(hash common.Hash) *params.ChainConfig
	WriteChainConfig(hash common.Hash, cfg *params.ChainConfig)

	ReadPreimage(hash common.Hash) []byte
	WritePreimages(number uint64, preimages map[common.Hash][]byte)
}


// TODO-GX Some databases need to be refined or consolidated because some of them are actually the same one.
type DatabaseEntryType int
const (
	canonicalHashDB DatabaseEntryType = iota
	headerNumberDB
	headheaderHashDB
	headBlockHashDB
	headFastBlockHashDB
	fastTrieProgressDB

	headerDB
	bodyDB
	tdDB
	receiptsDB
	blockDB

	istanbulSnapshotDB
	merkleProofDB
	cachedTrieNodeDB
	cachedTrieNodePreimagesDB
	txLookUpEntryDB

	bloomBitsDB
	validSectionsDB
	sectionHeadDB

	databaseVersionDB
	chainConfigDB
	preimagesDB

	databaseEntryTypeSize
)

type databaseManager struct {
	dbs [databaseEntryTypeSize]Database
}

func NewDBManager(dir string) (DBManager, error) {
	dbm := databaseManager{}

	// TODO-GX Should be replaced by initialization function with mapping information.
	db, err := NewLDBDatabase(dir, 128, 128)
	if err != nil {
		return nil, err
	}
	for i:=0; i < int(databaseEntryTypeSize); i++ {
		dbm.dbs[i] = db
	}

	return &dbm, nil
}

func (dbm *databaseManager) getDatabase(dbEntryType DatabaseEntryType) Database {
	return dbm.getDatabase(dbEntryType)
}

// TODO-GX Some of below need to be invisible outside database package
// Canonical Hash operations.
func (dbm *databaseManager) ReadCanonicalHash(number uint64) common.Hash {
	return rawdb.ReadCanonicalHash(dbm.getDatabase(canonicalHashDB), number)
}

func (dbm *databaseManager) WriteCanonicalHash(hash common.Hash, number uint64) {
	rawdb.WriteCanonicalHash(dbm.getDatabase(canonicalHashDB), hash, number)
}

func (dbm *databaseManager) DeleteCanonicalHash(number uint64) {
	rawdb.DeleteCanonicalHash(dbm.getDatabase(canonicalHashDB), number)
}

// Head Number operations.
func (dbm *databaseManager) ReadHeaderNumber(hash common.Hash) *uint64 {
	return rawdb.ReadHeaderNumber(dbm.getDatabase(headerNumberDB), hash)
}

// Head Header Hash operations.
func (dbm *databaseManager) ReadHeadHeaderHash() common.Hash {
	return rawdb.ReadHeadHeaderHash(dbm.getDatabase(headheaderHashDB))
}

func (dbm *databaseManager) WriteHeadHeaderHash(hash common.Hash) {
	rawdb.WriteHeadHeaderHash(dbm.getDatabase(headheaderHashDB), hash)
}

// Block Hash operations.
func (dbm *databaseManager) ReadHeadBlockHash() common.Hash {
	return rawdb.ReadHeadBlockHash(dbm.getDatabase(headBlockHashDB))
}

func (dbm *databaseManager) WriteHeadBlockHash(hash common.Hash) {
	rawdb.WriteHeadBlockHash(dbm.getDatabase(headBlockHashDB), hash)
}

// Head Fast Block Hash operations.
func (dbm *databaseManager) ReadHeadFastBlockHash() common.Hash {
	return rawdb.ReadHeadFastBlockHash(dbm.getDatabase(headFastBlockHashDB))
}

func (dbm *databaseManager) WriteHeadFastBlockHash(hash common.Hash) {
	rawdb.WriteHeadFastBlockHash(dbm.getDatabase(headFastBlockHashDB), hash)
}

// Fast Trie Progress operations.
func (dbm *databaseManager) ReadFastTrieProgress() uint64 {
	return rawdb.ReadFastTrieProgress(dbm.getDatabase(fastTrieProgressDB))
}

func (dbm *databaseManager) WriteFastTrieProgress(count uint64) {
	rawdb.WriteFastTrieProgress(dbm.getDatabase(fastTrieProgressDB), count)
}

// (Block)Header operations.
func (dbm *databaseManager) HasHeader(hash common.Hash, number uint64) bool {
	return rawdb.HasHeader(dbm.getDatabase(headerDB), hash, number)
}

func (dbm *databaseManager) ReadHeader(hash common.Hash, number uint64) *types.Header {
	return rawdb.ReadHeader(dbm.getDatabase(headerDB), hash, number)
}

func (dbm *databaseManager) ReadHeaderRLP(hash common.Hash, number uint64) rlp.RawValue {
	return rawdb.ReadHeaderRLP(dbm.getDatabase(headerDB), hash, number)
}

func (dbm *databaseManager) WriteHeader(header *types.Header) {
	rawdb.WriteHeader(dbm.getDatabase(headerDB), header)
}

func (dbm *databaseManager) DeleteHeader(hash common.Hash, number uint64) {
	rawdb.DeleteHeader(dbm.getDatabase(headerDB), hash, number)
}

// (Block)Body operations.
func (dbm *databaseManager) HasBody(hash common.Hash, number uint64) bool {
	return rawdb.HasBody(dbm.getDatabase(bodyDB), hash, number)
}

func (dbm *databaseManager) ReadBody(hash common.Hash, number uint64) *types.Body {
	return rawdb.ReadBody(dbm.getDatabase(bodyDB), hash, number)
}

func (dbm *databaseManager) ReadBodyRLP(hash common.Hash, number uint64) rlp.RawValue {
	return rawdb.ReadBodyRLP(dbm.getDatabase(bodyDB), hash, number)
}

func (dbm *databaseManager) WriteBody(hash common.Hash, number uint64, body *types.Body) {
	rawdb.WriteBody(dbm.getDatabase(bodyDB), hash, number, body)
}

func (dbm *databaseManager) WriteBodyRLP(hash common.Hash, number uint64, rlp rlp.RawValue) {
	rawdb.WriteBodyRLP(dbm.getDatabase(bodyDB), hash, number, rlp)
}

func (dbm *databaseManager) DeleteBody(hash common.Hash, number uint64) {
	rawdb.DeleteBody(dbm.getDatabase(bodyDB), hash, number)
}

// TotalDifficulty operations.
func (dbm *databaseManager) ReadTd(hash common.Hash, number uint64) *big.Int {
	return rawdb.ReadTd(dbm.getDatabase(tdDB), hash, number)
}

func (dbm *databaseManager) WriteTd(hash common.Hash, number uint64, td *big.Int) {
	rawdb.WriteTd(dbm.getDatabase(tdDB), hash, number, td)
}

func (dbm *databaseManager) DeleteTd(hash common.Hash, number uint64) {
	rawdb.DeleteTd(dbm.getDatabase(tdDB), hash, number)
}

// Receipts operations.
func (dbm *databaseManager) ReadReceipts(hash common.Hash, number uint64) types.Receipts {
	return rawdb.ReadReceipts(dbm.getDatabase(receiptsDB), hash, number)
}

func (dbm *databaseManager) WriteReceipts(hash common.Hash, number uint64, receipts types.Receipts) {
	rawdb.WriteReceipts(dbm.getDatabase(receiptsDB), hash, number, receipts)
}

func (dbm *databaseManager) DeleteReceipts(hash common.Hash, number uint64) {
	rawdb.DeleteReceipts(dbm.getDatabase(receiptsDB), hash, number)
}

// Block operations.
func (dbm *databaseManager) ReadBlock(hash common.Hash, number uint64) *types.Block {
	return rawdb.ReadBlock(dbm.getDatabase(blockDB), hash, number)
}

func (dbm *databaseManager) WriteBlock(block *types.Block) {
	rawdb.WriteBlock(dbm.getDatabase(blockDB), block)
}

func (dbm *databaseManager) DeleteBlock(hash common.Hash, number uint64) {
	rawdb.DeleteBlock(dbm.getDatabase(blockDB), hash, number)
}

// Find Common Ancestor operation
// rawdb.FindCommonAncestor uses ReadHeader to find common ancestor.
// Therefore, pass headerDB as a database parameter.
func (dbm *databaseManager) FindCommonAncestor(a, b *types.Header) *types.Header {
	return rawdb.FindCommonAncestor(dbm.getDatabase(headerDB), a, b)
}

// Istanbul Snapshot operations.
func (dbm *databaseManager) ReadIstanbulSnapshot(hash common.Hash) ([]byte, error) {
	return rawdb.ReadIstanbulSnapshot(dbm.getDatabase(istanbulSnapshotDB), hash)
}

func (dbm *databaseManager) WriteIstanbulSnapshot(hash common.Hash, blob []byte) error {
	return rawdb.WriteIstanbulSnapshot(dbm.getDatabase(istanbulSnapshotDB), hash, blob)
}

// Merkle Proof operation.
func (dbm *databaseManager) WriteMerkleProof(key, value []byte) error {
	return rawdb.WriteMerkleProof(dbm.getDatabase(merkleProofDB), key, value)
}

// Cached Trie Node operation.
func (dbm *databaseManager) ReadCachedTrieNode(hash common.Hash) ([]byte, error) {
	return rawdb.ReadCachedTrieNode(dbm.getDatabase(cachedTrieNodeDB), hash)
}

// Cached Trie Node Preimage operation.
func (dbm *databaseManager) ReadCachedTrieNodePreimage(secureKey []byte) ([]byte, error) {
	return rawdb.ReadCachedTrieNodePreimage(dbm.getDatabase(cachedTrieNodePreimagesDB), secureKey)
}

// from accessors_indexes.go
func (dbm *databaseManager) ReadTxLookupEntry(hash common.Hash) (common.Hash, uint64, uint64) {
	return rawdb.ReadTxLookupEntry(dbm.getDatabase(txLookUpEntryDB), hash)
}

func (dbm *databaseManager) WriteTxLookupEntries(block *types.Block) {
	rawdb.WriteTxLookupEntries(dbm.getDatabase(txLookUpEntryDB), block)
}

func (dbm *databaseManager) DeleteTxLookupEntry(hash common.Hash) {
	rawdb.DeleteTxLookupEntry(dbm.getDatabase(txLookUpEntryDB), hash)
}

// Transaction read operation.
// Directly copied rawdb operation because it uses two different databases.
func (dbm *databaseManager) ReadTransaction(hash common.Hash) (*types.Transaction, common.Hash, uint64, uint64) {
	blockHash, blockNumber, txIndex := dbm.ReadTxLookupEntry(hash)
	if blockHash == (common.Hash{}) {
		return nil, common.Hash{}, 0, 0
	}
	body := dbm.ReadBody(blockHash, blockNumber)
	if body == nil || len(body.Transactions) <= int(txIndex) {
		log.Error("Transaction referenced missing", "number", blockNumber, "hash", blockHash, "index", txIndex)
		return nil, common.Hash{}, 0, 0
	}
	return body.Transactions[txIndex], blockHash, blockNumber, txIndex
}

// Receipt read operation.
// Directly copied rawdb operation because it uses two different databases.
func (dbm *databaseManager) ReadReceipt(hash common.Hash) (*types.Receipt, common.Hash, uint64, uint64) {
	blockHash, blockNumber, receiptIndex := dbm.ReadTxLookupEntry(hash)
	if blockHash == (common.Hash{}) {
		return nil, common.Hash{}, 0, 0
	}
	receipts := dbm.ReadReceipts(blockHash, blockNumber)
	if len(receipts) <= int(receiptIndex) {
		log.Error("Receipt refereced missing", "number", blockNumber, "hash", blockHash, "index", receiptIndex)
		return nil, common.Hash{}, 0, 0
	}
	return receipts[receiptIndex], blockHash, blockNumber, receiptIndex
}

// BloomBits operations.
func (dbm *databaseManager) ReadBloomBits(bit uint, section uint64, head common.Hash) ([]byte, error) {
	return rawdb.ReadBloomBits(dbm.getDatabase(bloomBitsDB), bit, section, head)
}

func (dbm *databaseManager) WriteBloomBits(bit uint, section uint64, head common.Hash, bits []byte) {
	rawdb.WriteBloomBits(dbm.getDatabase(bloomBitsDB), bit, section, head, bits)
}

// ValidSections operation.
func (dbm *databaseManager) ReadValidSections() ([]byte, error) {
	return rawdb.ReadValidSections(dbm.getDatabase(validSectionsDB))
}

func (dbm *databaseManager) WriteValidSections(encodedSections []byte) {
	rawdb.WriteValidSections(dbm.getDatabase(validSectionsDB), encodedSections)
}

// SectionHead operation.
func (dbm *databaseManager) ReadSectionHead(encodedSection []byte) ([]byte, error) {
	return rawdb.ReadSectionHead(dbm.getDatabase(sectionHeadDB), encodedSection)
}

func (dbm *databaseManager) WriteSectionHead(encodedSection []byte, hash common.Hash) {
	rawdb.WriteSectionHead(dbm.getDatabase(sectionHeadDB), encodedSection, hash)
}

func (dbm *databaseManager) DeleteSectionHead(encodedSection []byte) {
	rawdb.DeleteSectionHead(dbm.getDatabase(sectionHeadDB), encodedSection)
}

// from accessors_metadata.go
func (dbm *databaseManager) ReadDatabaseVersion() int {
	return rawdb.ReadDatabaseVersion(dbm.getDatabase(databaseVersionDB))
}

func (dbm *databaseManager) WriteDatabaseVersion(version int) {
	rawdb.WriteDatabaseVersion(dbm.getDatabase(databaseVersionDB), version)
}

// ChainConfig operations.
func (dbm *databaseManager) ReadChainConfig(hash common.Hash) *params.ChainConfig {
	return rawdb.ReadChainConfig(dbm.getDatabase(chainConfigDB), hash)
}

func (dbm *databaseManager) WriteChainConfig(hash common.Hash, cfg *params.ChainConfig) {
	rawdb.WriteChainConfig(dbm.getDatabase(chainConfigDB), hash, cfg)
}

// Preimages operations.
func (dbm *databaseManager) ReadPreimage(hash common.Hash) []byte {
	return rawdb.ReadPreimage(dbm.getDatabase(preimagesDB), hash)
}

func (dbm *databaseManager) WritePreimages(number uint64, preimages map[common.Hash][]byte) {
	rawdb.WritePreimages(dbm.getDatabase(preimagesDB), number, preimages)
}
