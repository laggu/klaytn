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
	"encoding/binary"
	"bytes"
)

type DBManager interface {
	Close()
	NewBatch(dbType DatabaseEntryType) Batch
	GetMemDB() *MemDatabase

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
	WriteHeader(header *types.Header) error
	DeleteHeader(hash common.Hash, number uint64)

	HasBody(hash common.Hash, number uint64) bool
	ReadBody(hash common.Hash, number uint64) *types.Body
	ReadBodyRLP(hash common.Hash, number uint64) rlp.RawValue
	WriteBody(hash common.Hash, number uint64, body *types.Body) error
	PutBodyToBatch(batch Batch, hash common.Hash, number uint64, body *types.Body) error
	WriteBodyRLP(hash common.Hash, number uint64, rlp rlp.RawValue) error
	DeleteBody(hash common.Hash, number uint64)

	ReadTd(hash common.Hash, number uint64) *big.Int
	WriteTd(hash common.Hash, number uint64, td *big.Int)
	DeleteTd(hash common.Hash, number uint64)

	ReadReceipts(hash common.Hash, number uint64) types.Receipts
	WriteReceipts(hash common.Hash, number uint64, receipts types.Receipts) error
	DeleteReceipts(hash common.Hash, number uint64)

	ReadBlock(hash common.Hash, number uint64) *types.Block
	WriteBlock(block *types.Block) error
	DeleteBlock(hash common.Hash, number uint64)

	FindCommonAncestor(a, b *types.Header) *types.Header

	ReadIstanbulSnapshot(hash common.Hash) ([]byte, error)
	WriteIstanbulSnapshot(hash common.Hash, blob []byte) error

	WriteMerkleProof(key, value []byte) error

	ReadCachedTrieNode(hash common.Hash) ([]byte, error)
	ReadCachedTrieNodePreimage(secureKey []byte) ([]byte, error)

	ReadStateTrieNode(key []byte) ([]byte, error)
	HasStateTrieNode(key []byte) (bool, error)

	// from accessors_indexes.go
	ReadTxLookupEntry(hash common.Hash) (common.Hash, uint64, uint64)
	WriteTxLookupEntries(block *types.Block) error
	DeleteTxLookupEntry(hash common.Hash)

	ReadTransaction(hash common.Hash) (*types.Transaction, common.Hash, uint64, uint64)

	ReadReceipt(hash common.Hash) (*types.Receipt, common.Hash, uint64, uint64)

	ReadBloomBits(bloomBitsKey []byte) ([]byte, error)
	WriteBloomBits(bloomBitsKey []byte, bits []byte)

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
	WritePreimages(number uint64, preimages map[common.Hash][]byte) error
}


// TODO-GX Some databases need to be refined or consolidated because some of them are actually the same one.
type DatabaseEntryType int
const (
	canonicalHashDB DatabaseEntryType = iota

	headheaderHashDB
	headBlockHashDB
	headFastBlockHashDB
	fastTrieProgressDB

	headerDB
	BodyDB
	tdDB
	ReceiptsDB
	blockDB

	istanbulSnapshotDB
	merkleProofDB
	StateTrieDB
	PreimagesDB
	TxLookUpEntryDB

	BloomBitsDB
	indexSectionsDB

	databaseVersionDB
	chainConfigDB

	databaseEntryTypeSize
)

type databaseManager struct {
	dbs []Database
	isMemoryDB bool
}

func NewMemoryDBManager() (DBManager) {
	dbm := databaseManager{make([]Database, 1, 1), true}
	dbm.dbs[0] = NewMemDatabase()

	return &dbm
}

func NewDBManager(dir string, dbType string, cache, handles int) (DBManager, error) {
	dbm := databaseManager{make([]Database, databaseEntryTypeSize, databaseEntryTypeSize), false}

	// TODO-GX Should be replaced by initialization function with mapping information.
	var db Database
	var err error
	switch dbType {
	case LEVELDB:
		db, err = NewLDBDatabase(dir, cache, handles)
		db.Meter("klay/db/chaindata/")
	case BADGER:
		db, err = NewBGDatabase(dir)
	case MEMDB:
		db = NewMemDatabase()
	default:
		db, err = NewLDBDatabase(dir, cache, handles)
		log.Warn("database type is not set, fall back to default LevelDB")
	}

	if err != nil {
		return nil, err
	}

	for i:=0; i < int(databaseEntryTypeSize); i++ {
		if i == int(indexSectionsDB) {
			dbm.dbs[i] = NewTable(db, string(rawdb.BloomBitsIndexPrefix))
		} else {
			dbm.dbs[i] = db
		}
	}
	return &dbm, nil
}

func (dbm *databaseManager) NewBatch(dbEntryType DatabaseEntryType) Batch {
	return dbm.getDatabase(dbEntryType).NewBatch()
}

func (dbm *databaseManager) GetMemDB() *MemDatabase {
	if dbm.isMemoryDB {
		if memDB, ok := dbm.dbs[0].(*MemDatabase); ok {
			return memDB
		} else {
			log.Error("DBManager is set as memory DBManager, but actual value is not set as memory DBManager.")
			return nil
		}
	}
	log.Error("GetMemDB() call to non memory DBManager object.")
	return nil
}

func (dbm *databaseManager) getDatabase(dbEntryType DatabaseEntryType) Database {
	if dbm.isMemoryDB {
		return dbm.dbs[0]
	} else {
		return dbm.dbs[dbEntryType]
	}
}

func (dbm *databaseManager) Close() {
	dbm.dbs[0].Close()
	//TODO-GX should be enabled after individual databases are integrated.
	//for _, db := range dbm.dbs {
	//	db.Close()
	//}
}

// TODO-GX Some of below need to be invisible outside database package
// Canonical Hash operations.
// ReadCanonicalHash retrieves the hash assigned to a canonical block number.
func (dbm *databaseManager) ReadCanonicalHash(number uint64) common.Hash {
	db := dbm.getDatabase(canonicalHashDB)
	data, _ := db.Get(headerHashKey(number))
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteCanonicalHash stores the hash assigned to a canonical block number.
func (dbm *databaseManager) WriteCanonicalHash(hash common.Hash, number uint64) {
	db := dbm.getDatabase(canonicalHashDB)
	if err := db.Put(headerHashKey(number), hash.Bytes()); err != nil {
		log.Crit("Failed to store number to hash mapping", "err", err)
	}
}

// DeleteCanonicalHash removes the number to hash canonical mapping.
func (dbm *databaseManager) DeleteCanonicalHash(number uint64) {
	db := dbm.getDatabase(canonicalHashDB)
	if err := db.Delete(headerHashKey(number)); err != nil {
		log.Crit("Failed to delete number to hash mapping", "err", err)
	}
}

// Head Number operations.
// ReadHeaderNumber returns the header number assigned to a hash.
func (dbm *databaseManager) ReadHeaderNumber(hash common.Hash) *uint64 {
	db := dbm.getDatabase(headerDB)
	data, _ := db.Get(headerNumberKey(hash))
	if len(data) != 8 {
		return nil
	}
	number := binary.BigEndian.Uint64(data)
	return &number
}

// Head Header Hash operations.
// ReadHeadHeaderHash retrieves the hash of the current canonical head header.
func (dbm *databaseManager) ReadHeadHeaderHash() common.Hash {
	db := dbm.getDatabase(headheaderHashDB)
	data, _ := db.Get(headHeaderKey)
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteHeadHeaderHash stores the hash of the current canonical head header.
func (dbm *databaseManager) WriteHeadHeaderHash(hash common.Hash) {
	db := dbm.getDatabase(headheaderHashDB)
	if err := db.Put(headHeaderKey, hash.Bytes()); err != nil {
		log.Crit("Failed to store last header's hash", "err", err)
	}
}

// Block Hash operations.
func (dbm *databaseManager) ReadHeadBlockHash() common.Hash {
	db := dbm.getDatabase(headBlockHashDB)
	data, _ := db.Get(headBlockKey)
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteHeadBlockHash stores the head block's hash.
func (dbm *databaseManager) WriteHeadBlockHash(hash common.Hash) {
	db := dbm.getDatabase(headBlockHashDB)
	if err := db.Put(headBlockKey, hash.Bytes()); err != nil {
		log.Crit("Failed to store last block's hash", "err", err)
	}
}

// Head Fast Block Hash operations.
// ReadHeadFastBlockHash retrieves the hash of the current fast-sync head block.
func (dbm *databaseManager) ReadHeadFastBlockHash() common.Hash {
	db := dbm.getDatabase(headFastBlockHashDB)
	data, _ := db.Get(headFastBlockKey)
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteHeadFastBlockHash stores the hash of the current fast-sync head block.
func (dbm *databaseManager) WriteHeadFastBlockHash(hash common.Hash) {
	db := dbm.getDatabase(headFastBlockHashDB)
	if err := db.Put(headFastBlockKey, hash.Bytes()); err != nil {
		log.Crit("Failed to store last fast block's hash", "err", err)
	}
}

// Fast Trie Progress operations.
// ReadFastTrieProgress retrieves the number of tries nodes fast synced to allow
// reporting correct numbers across restarts.
func (dbm *databaseManager) ReadFastTrieProgress() uint64 {
	db := dbm.getDatabase(fastTrieProgressDB)
	data, _ := db.Get(fastTrieProgressKey)
	if len(data) == 0 {
		return 0
	}
	return new(big.Int).SetBytes(data).Uint64()
}

// WriteFastTrieProgress stores the fast sync trie process counter to support
// retrieving it across restarts.
func (dbm *databaseManager) WriteFastTrieProgress(count uint64) {
	db := dbm.getDatabase(fastTrieProgressDB)
	if err := db.Put(fastTrieProgressKey, new(big.Int).SetUint64(count).Bytes()); err != nil {
		log.Crit("Failed to store fast sync trie progress", "err", err)
	}
}

// (Block)Header operations.
// HasHeader verifies the existence of a block header corresponding to the hash.
func (dbm *databaseManager) HasHeader(hash common.Hash, number uint64) bool {
	db := dbm.getDatabase(headerDB)
	if has, err := db.Has(headerKey(number, hash)); !has || err != nil {
		return false
	}
	return true
}

// ReadHeader retrieves the block header corresponding to the hash.
func (dbm *databaseManager) ReadHeader(hash common.Hash, number uint64) *types.Header {
	data := dbm.ReadHeaderRLP(hash, number)
	if len(data) == 0 {
		return nil
	}
	header := new(types.Header)
	if err := rlp.Decode(bytes.NewReader(data), header); err != nil {
		log.Error("Invalid block header RLP", "hash", hash, "err", err)
		return nil
	}
	return header
}

// ReadHeaderRLP retrieves a block header in its raw RLP database encoding.
func (dbm *databaseManager) ReadHeaderRLP(hash common.Hash, number uint64) rlp.RawValue {
	db := dbm.getDatabase(headerDB)
	data, _ := db.Get(headerKey(number, hash))
	return data
}

// WriteHeader stores a block header into the database and also stores the hash-
// to-number mapping.
func (dbm *databaseManager) WriteHeader(header *types.Header) error {
	db := dbm.getDatabase(headerDB)
	// Write the hash -> number mapping
	var (
		hash    = header.Hash()
		number  = header.Number.Uint64()
		encoded = encodeBlockNumber(number)
	)
	key := headerNumberKey(hash)
	if err := db.Put(key, encoded); err != nil {
		log.Crit("Failed to store hash to number mapping", "err", err)
		return err
	}
	// Write the encoded header
	data, err := rlp.EncodeToBytes(header)
	if err != nil {
		log.Crit("Failed to RLP encode header", "err", err)
		return err
	}
	key = headerKey(number, hash)
	if err := db.Put(key, data); err != nil {
		log.Crit("Failed to store header", "err", err)
		return err
	}
	return nil
}

// DeleteHeader removes all block header data associated with a hash.
func (dbm *databaseManager) DeleteHeader(hash common.Hash, number uint64) {
	db := dbm.getDatabase(headerDB)
	if err := db.Delete(headerKey(number, hash)); err != nil {
		log.Crit("Failed to delete header", "err", err)
	}
	if err := db.Delete(headerNumberKey(hash)); err != nil {
		log.Crit("Failed to delete hash to number mapping", "err", err)
	}
}

// (Block)Body operations.
// HasBody verifies the existence of a block body corresponding to the hash.
func (dbm *databaseManager) HasBody(hash common.Hash, number uint64) bool {
	db := dbm.getDatabase(BodyDB)
	if has, err := db.Has(blockBodyKey(number, hash)); !has || err != nil {
		return false
	}
	return true
}

// ReadBody retrieves the block body corresponding to the hash.
func (dbm *databaseManager) ReadBody(hash common.Hash, number uint64) *types.Body {
	data := dbm.ReadBodyRLP(hash, number)
	if len(data) == 0 {
		return nil
	}
	body := new(types.Body)
	if err := rlp.Decode(bytes.NewReader(data), body); err != nil {
		log.Error("Invalid block body RLP", "hash", hash, "err", err)
		return nil
	}
	return body
}

// ReadBodyRLP retrieves the block body (transactions and uncles) in RLP encoding.
func (dbm *databaseManager) ReadBodyRLP(hash common.Hash, number uint64) rlp.RawValue {
	db := dbm.getDatabase(BodyDB)
	data, _ := db.Get(blockBodyKey(number, hash))
	return data
}

// WriteBody storea a block body into the database.
func (dbm *databaseManager) WriteBody(hash common.Hash, number uint64, body *types.Body) error {
	data, err := rlp.EncodeToBytes(body)
	if err != nil {
		log.Crit("Failed to RLP encode body", "err", err)
		return err
	}
	return dbm.WriteBodyRLP(hash, number, data)
}

func (dbm *databaseManager) PutBodyToBatch(batch Batch, hash common.Hash, number uint64, body *types.Body) error {
	data, err := rlp.EncodeToBytes(body)
	if err != nil {
		log.Crit("Failed to RLP encode body", "err", err)
		return err
	}

	if err := batch.Put(blockBodyKey(number, hash), data); err != nil {
		log.Crit("Failed to store block body", "err", err)
		return err
	}
	return nil
}

// WriteBodyRLP stores an RLP encoded block body into the database.
func (dbm *databaseManager) WriteBodyRLP(hash common.Hash, number uint64, rlp rlp.RawValue) error {
	db := dbm.getDatabase(BodyDB)
	if err := db.Put(blockBodyKey(number, hash), rlp); err != nil {
		log.Crit("Failed to store block body", "err", err)
		return err
	}
	return nil
}

// DeleteBody removes all block body data associated with a hash.
func (dbm *databaseManager) DeleteBody(hash common.Hash, number uint64) {
	db := dbm.getDatabase(BodyDB)
	if err := db.Delete(blockBodyKey(number, hash)); err != nil {
		log.Crit("Failed to delete block body", "err", err)
	}
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
	return rawdb.ReadReceipts(dbm.getDatabase(ReceiptsDB), hash, number)
}

func (dbm *databaseManager) WriteReceipts(hash common.Hash, number uint64, receipts types.Receipts) error {
	return rawdb.WriteReceipts(dbm.getDatabase(ReceiptsDB), hash, number, receipts)
}

func (dbm *databaseManager) DeleteReceipts(hash common.Hash, number uint64) {
	rawdb.DeleteReceipts(dbm.getDatabase(ReceiptsDB), hash, number)
}

// Block operations.
// ReadBlock retrieves an entire block corresponding to the hash, assembling it
// back from the stored header and body. If either the header or body could not
// be retrieved nil is returned.
//
// Note, due to concurrent download of header and block body the header and thus
// canonical hash can be stored in the database but the body data not (yet).
func (dbm *databaseManager) ReadBlock(hash common.Hash, number uint64) *types.Block {
	header := dbm.ReadHeader(hash, number)
	if header == nil {
		return nil
	}
	body := dbm.ReadBody(hash, number)
	if body == nil {
		return nil
	}
	return types.NewBlockWithHeader(header).WithBody(body.Transactions, body.Uncles)
}

func (dbm *databaseManager) WriteBlock(block *types.Block) error {
	if err := dbm.WriteBody(block.Hash(), block.NumberU64(), block.Body()); err != nil {
		return err
	}
	if err := dbm.WriteHeader(block.Header()); err != nil {
		return err
	}
	return nil
}

func (dbm *databaseManager) DeleteBlock(hash common.Hash, number uint64) {
	dbm.DeleteReceipts(hash, number)
	dbm.DeleteHeader(hash, number)
	dbm.DeleteBody(hash, number)
	dbm.DeleteTd(hash, number)
}

// Find Common Ancestor operation
// FindCommonAncestor returns the last common ancestor of two block headers
func (dbm *databaseManager) FindCommonAncestor(a, b *types.Header) *types.Header {
	for bn := b.Number.Uint64(); a.Number.Uint64() > bn; {
		a = dbm.ReadHeader(a.ParentHash, a.Number.Uint64()-1)
		if a == nil {
			return nil
		}
	}
	for an := a.Number.Uint64(); an < b.Number.Uint64(); {
		b = dbm.ReadHeader(b.ParentHash, b.Number.Uint64()-1)
		if b == nil {
			return nil
		}
	}
	for a.Hash() != b.Hash() {
		a = dbm.ReadHeader(a.ParentHash, a.Number.Uint64()-1)
		if a == nil {
			return nil
		}
		b = dbm.ReadHeader(b.ParentHash, b.Number.Uint64()-1)
		if b == nil {
			return nil
		}
	}
	return a
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
	return rawdb.ReadCachedTrieNode(dbm.getDatabase(StateTrieDB), hash)
}

// Cached Trie Node Preimage operation.
func (dbm *databaseManager) ReadCachedTrieNodePreimage(secureKey []byte) ([]byte, error) {
	return rawdb.ReadCachedTrieNodePreimage(dbm.getDatabase(PreimagesDB), secureKey)
}

// State Trie Related operations.
func (dbm *databaseManager) ReadStateTrieNode(key []byte) ([]byte, error) {
	return rawdb.ReadStateTrieNode(dbm.getDatabase(StateTrieDB), key)
}

func (dbm *databaseManager) HasStateTrieNode(key []byte) (bool, error) {
	val, err := rawdb.ReadStateTrieNode(dbm.getDatabase(StateTrieDB), key)
	if val == nil || err != nil {
		return false, err
	}
	return true, nil
}

// from accessors_indexes.go
func (dbm *databaseManager) ReadTxLookupEntry(hash common.Hash) (common.Hash, uint64, uint64) {
	return rawdb.ReadTxLookupEntry(dbm.getDatabase(TxLookUpEntryDB), hash)
}

func (dbm *databaseManager) WriteTxLookupEntries(block *types.Block) error {
	return rawdb.WriteTxLookupEntries(dbm.getDatabase(TxLookUpEntryDB), block)
}

func (dbm *databaseManager) DeleteTxLookupEntry(hash common.Hash) {
	rawdb.DeleteTxLookupEntry(dbm.getDatabase(TxLookUpEntryDB), hash)
}

// ReadTransaction retrieves a specific transaction from the database, along with
// its added positional metadata.
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
func (dbm *databaseManager) ReadBloomBits(bloomBitsKey []byte) ([]byte, error) {
	return rawdb.ReadBloomBits(dbm.getDatabase(BloomBitsDB), bloomBitsKey)
}

func (dbm *databaseManager) WriteBloomBits(bloomBitsKey, bits []byte) {
	rawdb.WriteBloomBits(dbm.getDatabase(BloomBitsDB), bloomBitsKey, bits)
}

// ValidSections operation.
func (dbm *databaseManager) ReadValidSections() ([]byte, error) {
	return rawdb.ReadValidSections(dbm.getDatabase(indexSectionsDB))
}

func (dbm *databaseManager) WriteValidSections(encodedSections []byte) {
	rawdb.WriteValidSections(dbm.getDatabase(indexSectionsDB), encodedSections)
}

// SectionHead operation.
func (dbm *databaseManager) ReadSectionHead(encodedSection []byte) ([]byte, error) {
	return rawdb.ReadSectionHead(dbm.getDatabase(indexSectionsDB), encodedSection)
}

func (dbm *databaseManager) WriteSectionHead(encodedSection []byte, hash common.Hash) {
	rawdb.WriteSectionHead(dbm.getDatabase(indexSectionsDB), encodedSection, hash)
}

func (dbm *databaseManager) DeleteSectionHead(encodedSection []byte) {
	rawdb.DeleteSectionHead(dbm.getDatabase(indexSectionsDB), encodedSection)
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
	return rawdb.ReadPreimage(dbm.getDatabase(PreimagesDB), hash)
}

func (dbm *databaseManager) WritePreimages(number uint64, preimages map[common.Hash][]byte) error {
	return rawdb.WritePreimages(dbm.getDatabase(PreimagesDB), number, preimages)
}

// bloomBitsPrefix + bit (uint16 big endian) + section (uint64 big endian) + hash -> bloom bits
var bloomBitsPrefix = []byte("B")

// bloomBitsKey = bloomBitsPrefix + bit (uint16 big endian) + section (uint64 big endian) + hash
func BloomBitsKey(bit uint, section uint64, hash common.Hash) []byte {
	key := append(append(bloomBitsPrefix, make([]byte, 10)...), hash.Bytes()...)

	binary.BigEndian.PutUint16(key[1:], uint16(bit))
	binary.BigEndian.PutUint64(key[3:], section)

	return key
}