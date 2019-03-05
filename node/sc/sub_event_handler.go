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
	"errors"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"math/big"
)

var (
	ErrGetServiceChainPHInCCEH = errors.New("ServiceChainPH isn't set in ChildChainEventHandler")
)

type ChildChainEventHandler struct {
	subbridge *SubBridge

	handler *SubBridgeHandler
}

func NewChildChainEventHandler(bridge *SubBridge, handler *SubBridgeHandler) (*ChildChainEventHandler, error) {
	return &ChildChainEventHandler{subbridge: bridge, handler: handler}, nil
}

func (cce *ChildChainEventHandler) HandleChainHeadEvent(block *types.Block) error {
	logger.Debug("bridgeNode block number", "number", block.Number())
	if cce.subbridge.GetAnchoringTx() {
		cce.handler.NewAnchoringTx(block)
	}
	return nil
}

func (cce *ChildChainEventHandler) HandleTxEvent(tx *types.Transaction) error {
	//TODO-Klaytn event handle
	return nil
}

func (cce *ChildChainEventHandler) HandleTxsEvent(txs []*types.Transaction) error {
	//TODO-Klaytn event handle
	return nil
}

func (cce *ChildChainEventHandler) HandleLogsEvent(logs []*types.Log) error {
	//TODO-Klaytn event handle
	return nil
}

func (cce *ChildChainEventHandler) HandleTokenReceivedEvent(token TokenReceivedEvent) error {
	//TODO-Klaytn event handle
	gatewayAddr := cce.subbridge.AddressManager().GetCounterPartGateway(token.ContractAddr)
	tokenAddr := cce.subbridge.AddressManager().GetCounterPartToken(token.TokenAddr)
	user := cce.subbridge.AddressManager().GetCounterPartUser(token.From)

	local, ok := cce.subbridge.gatewayMgr.IsLocal(gatewayAddr)
	if !ok {
		return errors.New("there is no gateway")
	}

	if local {
		auth := MakeTransactOpts(cce.handler.nodeKey, big.NewInt((int64)(cce.handler.getNodeAccountNonce())), cce.subbridge.getChainID(), big.NewInt(0))
		gateway := cce.subbridge.gatewayMgr.GetGateway(gatewayAddr)
		tx, err := gateway.WithdrawERC20(auth, token.Amount, user, tokenAddr)
		logger.Info("GateWay.WithdrawERC20", "tx", tx.Hash().Hex())
		return err
	} else {
		cce.handler.LockChainAccount()
		defer cce.handler.UnLockChainAccount()
		auth := MakeTransactOpts(cce.handler.chainKey, big.NewInt((int64)(cce.handler.getChainAccountNonce())), cce.handler.parentChainID, big.NewInt(0))
		gateway := cce.subbridge.gatewayMgr.GetGateway(gatewayAddr)
		tx, err := gateway.WithdrawERC20(auth, token.Amount, user, tokenAddr)
		if err == nil {
			cce.handler.addChainAccountNonce(1)
		}
		logger.Info("GateWay.WithdrawERC20", "tx", tx.Hash().Hex())
		return err
	}
}

func (cce *ChildChainEventHandler) HandleTokenTransferEvent(token TokenTransferEvent) error {
	//TODO-Klaytn event handle
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
	cce.subbridge.chainDB.WriteChildChainTxHash(ccBlockHash, ccTxHash)
}
