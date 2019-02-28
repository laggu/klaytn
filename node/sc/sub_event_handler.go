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
)

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
	//@TODO
	return nil
}

func (cce *ChildChainEventHandler) HandleTxsEvent(txs []*types.Transaction) error {
	//@TODO
	return nil
}

func (cce *ChildChainEventHandler) HandleLogsEvent(logs []*types.Log) error {
	//@TODO
	fmt.Println("call HandleLogsEvent ", len(logs))
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
