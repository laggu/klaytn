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

package sc

import (
	"fmt"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/ser/rlp"
)

type EventManager interface {
	GetEventHadler() ChainEventHandler
}

type ChainEventHandler interface {
	GetChildChainIndexingEnabled() bool
	ConvertChildChainBlockHashToParentChainTxHash(ccBlockHash common.Hash) common.Hash
	WriteChildChainTxHash(ccBlockHash common.Hash, ccTxHash common.Hash)
	GetLatestAnchoredBlockNumber() uint64
	WriteAnchoredBlockNumber(blockNum uint64)
	WriteReceiptFromParentChain(blockHash common.Hash, receipt *types.Receipt)
	GetReceiptFromParentChain(blockHash common.Hash) *types.Receipt
	writeChildChainTxHashFromBlock(block *types.Block)
}

type MainChainEventHandler struct {
	mainbridge *MainBridge

	protocolHandler ServiceChainProtocolHandler
}

func NewMainChainEventHandler(bridge *MainBridge) (*MainChainEventHandler, error) {
	return &MainChainEventHandler{bridge, bridge.handler.protocolHandler}, nil
}

func (mce *MainChainEventHandler) HandleChainHeadEvent(block *types.Block) error {
	logger.Debug("bridgeNode block number", "number", block.Number())
	//mce.writeChildChainTxHashFromBlock(block)
	return nil
}

func (mce *MainChainEventHandler) HandleTxEvent(tx *types.Transaction) error {
	//@TODO-Klaytn event handle
	return nil
}

func (mce *MainChainEventHandler) HandleTxsEvent(txs []*types.Transaction) error {
	//@TODO-Klaytn event handle
	return nil
}

func (mce *MainChainEventHandler) HandleLogsEvent(logs []*types.Log) error {
	//@TODO-Klaytn event handle
	return nil
}

func (mce *MainChainEventHandler) RegisterNewPeer(p BridgePeer) error {
	if mce.protocolHandler.getParentChainID() == nil {
		mce.protocolHandler.setParentChainID(p.GetChainID())
		return nil
	}
	if mce.protocolHandler.getParentChainID().Cmp(p.GetChainID()) != 0 {
		return fmt.Errorf("attempt to add a peer with different chainID failed! existing chainID: %v, new chainID: %v", mce.protocolHandler.getParentChainID(), p.GetChainID())
	}
	return nil
}

// GetChildChainIndexingEnabled returns the current child chain indexing configuration.
func (mce *MainChainEventHandler) GetChildChainIndexingEnabled() bool {
	return mce.mainbridge.chainDB.ChildChainIndexingEnabled()
}

// ConvertChildChainBlockHashToParentChainTxHash returns a transaction hash of a transaction which contains
// ChainHashes, with the key made with given child chain block hash.
// Index is built when child chain indexing is enabled.
func (mce *MainChainEventHandler) ConvertChildChainBlockHashToParentChainTxHash(ccBlockHash common.Hash) common.Hash {
	return mce.mainbridge.chainDB.ConvertChildChainBlockHashToParentChainTxHash(ccBlockHash)
}

// WriteChildChainTxHash stores a transaction hash of a transaction which contains
// ChainHashes, with the key made with given child chain block hash.
// Index is built when child chain indexing is enabled.
func (mce *MainChainEventHandler) WriteChildChainTxHash(ccBlockHash common.Hash, ccTxHash common.Hash) {
	mce.mainbridge.chainDB.WriteChildChainTxHash(ccBlockHash, ccTxHash)
}

// GetLatestAnchoredBlockNumber returns the latest block number whose data has been anchored to the parent chain.
func (mce *MainChainEventHandler) GetLatestAnchoredBlockNumber() uint64 {
	return mce.mainbridge.chainDB.ReadAnchoredBlockNumber()
}

// WriteAnchoredBlockNumber writes the block number whose data has been anchored to the parent chain.
func (mce *MainChainEventHandler) WriteAnchoredBlockNumber(blockNum uint64) {
	mce.mainbridge.chainDB.WriteAnchoredBlockNumber(blockNum)
}

// WriteReceiptFromParentChain writes a receipt received from parent chain to child chain
// with corresponding block hash. It assumes that a child chain has only one parent chain.
func (mce *MainChainEventHandler) WriteReceiptFromParentChain(blockHash common.Hash, receipt *types.Receipt) {
	mce.mainbridge.chainDB.WriteReceiptFromParentChain(blockHash, receipt)
}

// GetReceiptFromParentChain returns a receipt received from parent chain to child chain
// with corresponding block hash. It assumes that a child chain has only one parent chain.
func (mce *MainChainEventHandler) GetReceiptFromParentChain(blockHash common.Hash) *types.Receipt {
	return mce.mainbridge.chainDB.ReadReceiptFromParentChain(blockHash)
}

// writeChildChainTxHashFromBlock writes transaction hashes of transactions which contain
// ChainHashes.
func (mce *MainChainEventHandler) writeChildChainTxHashFromBlock(block *types.Block) {
	txs := block.Transactions()
	signer := types.MakeSigner(mce.mainbridge.blockchain.Config(), block.Number())
	for _, tx := range txs {
		// TODO-Klaytn-ServiceChain GetChildChainAccountAddr will be removed once new transaction type is introduced.
		if ccAddr := tx.GetChildChainAccountAddr(signer); ccAddr == nil {
			continue
		}
		chainHashes := new(types.ChainHashes)
		data, err := tx.AnchoredData()
		if err != nil {
			logger.Error("writeChildChainTxHashFromBlock : failed to get anchoring data from the tx", "txHash", tx.Hash())
			continue
		}
		if err := rlp.DecodeBytes(data, chainHashes); err != nil {
			logger.Error("writeChildChainTxHashFromBlock : failed to decode anchoring data")
			continue
		}
		mce.WriteChildChainTxHash(chainHashes.BlockHash, tx.Hash())
	}
}

// event handler for childchain
type ChildChainEventHandler struct {
	subbridge *SubBridge

	protocolHandler ServiceChainProtocolHandler
}

func NewChildChainEventHandler(bridge *SubBridge) (*ChildChainEventHandler, error) {
	return &ChildChainEventHandler{bridge, bridge.handler.protocolHandler}, nil
}

func (cce *ChildChainEventHandler) HandleChainHeadEvent(block *types.Block) error {
	logger.Debug("bridgeNode block number", "number", block.Number())
	cce.protocolHandler.BroadcastServiceChainTxAndReceiptRequest(block)
	return nil
}

func (cce *ChildChainEventHandler) HandleTxEvent(tx *types.Transaction) error {
	//@TODO-Klaytn event handle
	return nil
}

func (cce *ChildChainEventHandler) HandleTxsEvent(txs []*types.Transaction) error {
	//@TODO-Klaytn event handle
	return nil
}

func (cce *ChildChainEventHandler) HandleLogsEvent(logs []*types.Log) error {
	//@TODO-Klaytn event handle
	return nil
}

// GetChildChainIndexingEnabled returns the current child chain indexing configuration.
func (cce *ChildChainEventHandler) GetChildChainIndexingEnabled() bool {
	return cce.subbridge.chainDB.ChildChainIndexingEnabled()
}

// ConvertChildChainBlockHashToParentChainTxHash returns a transaction hash of a transaction which contains
// ChainHashes, with the key made with given child chain block hash.
// Index is built when child chain indexing is enabled.
func (cce *ChildChainEventHandler) ConvertChildChainBlockHashToParentChainTxHash(ccBlockHash common.Hash) common.Hash {
	return cce.subbridge.chainDB.ConvertChildChainBlockHashToParentChainTxHash(ccBlockHash)
}

// WriteChildChainTxHash stores a transaction hash of a transaction which contains
// ChainHashes, with the key made with given child chain block hash.
// Index is built when child chain indexing is enabled.
func (cce *ChildChainEventHandler) WriteChildChainTxHash(ccBlockHash common.Hash, ccTxHash common.Hash) {
}

// GetLatestAnchoredBlockNumber returns the latest block number whose data has been anchored to the parent chain.
func (cce *ChildChainEventHandler) GetLatestAnchoredBlockNumber() uint64 {
	return 0
}

// WriteAnchoredBlockNumber writes the block number whose data has been anchored to the parent chain.
func (cce *ChildChainEventHandler) WriteAnchoredBlockNumber(blockNum uint64) {
}

// WriteReceiptFromParentChain writes a receipt received from parent chain to child chain
// with corresponding block hash. It assumes that a child chain has only one parent chain.
func (cce *ChildChainEventHandler) WriteReceiptFromParentChain(blockHash common.Hash, receipt *types.Receipt) {
}

// GetReceiptFromParentChain returns a receipt received from parent chain to child chain
// with corresponding block hash. It assumes that a child chain has only one parent chain.
func (cce *ChildChainEventHandler) GetReceiptFromParentChain(blockHash common.Hash) *types.Receipt {
	return nil
}

// writeChildChainTxHashFromBlock writes transaction hashes of transactions which contain
// ChainHashes.
func (cce *ChildChainEventHandler) writeChildChainTxHashFromBlock(block *types.Block) {
}

func (cce *ChildChainEventHandler) RegisterNewPeer(p BridgePeer) error {
	if cce.protocolHandler.getParentChainID() == nil {
		cce.protocolHandler.setParentChainID(p.GetChainID())
		return nil
	}
	if cce.protocolHandler.getParentChainID().Cmp(p.GetChainID()) != 0 {
		return fmt.Errorf("attempt to add a peer with different chainID failed! existing chainID: %v, new chainID: %v", cce.protocolHandler.getParentChainID(), p.GetChainID())
	}
	return nil
}
