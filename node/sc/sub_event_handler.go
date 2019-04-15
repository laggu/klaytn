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

// LogEventListener is listener to handle log event
type LogEventListener interface {
	Handle(logs []*types.Log) error
}

type ChildChainEventHandler struct {
	subbridge *SubBridge

	handler   *SubBridgeHandler
	listeners []LogEventListener
}

func NewChildChainEventHandler(bridge *SubBridge, handler *SubBridgeHandler) (*ChildChainEventHandler, error) {
	return &ChildChainEventHandler{subbridge: bridge, handler: handler}, nil
}

func (cce *ChildChainEventHandler) AddListener(listener LogEventListener) {
	// TODO-Klaytn improve listener management
	cce.listeners = append(cce.listeners, listener)
}

func (cce *ChildChainEventHandler) HandleChainHeadEvent(block *types.Block) error {
	logger.Debug("bridgeNode block number", "number", block.Number())
	cce.handler.LocalChainHeadEvent(block)
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
	for _, listener := range cce.listeners {
		if err := listener.Handle(logs); err != nil {
			logger.Error("fail to handle log", "err", err)
		}
	}
	return nil
}

func (cce *ChildChainEventHandler) HandleTokenReceivedEvent(token TokenReceivedEvent) error {
	//TODO-Klaytn event handle
	tokenType := token.TokenType
	bridgeAddr := cce.subbridge.AddressManager().GetCounterPartBridge(token.ContractAddr)
	tokenAddr := cce.subbridge.AddressManager().GetCounterPartToken(token.TokenAddr)
	to := token.To

	bridgeInfo, ok := cce.subbridge.bridgeManager.bridges[bridgeAddr]
	if !ok {
		return errors.New("there is no bridge")
	}
	switch tokenType {
	case KLAY:
		logger.Info("Got request KLAY transfer event")
		if bridgeInfo.onServiceChain {
			auth := MakeTransactOpts(cce.handler.nodeKey, big.NewInt((int64)(cce.handler.getNodeAccountNonce())), cce.subbridge.getChainID(), cce.subbridge.txPool.GasPrice())
			tx, err := bridgeInfo.bridge.HandleKLAYTransfer(auth, token.Amount, to, token.RequestNonce)
			if err != nil {
				logger.Error("Child Bridge failed to HandleKLAYTransfer", "err", err)
				return err
			}
			logger.Info("Child Bridge succeeded to HandleKLAYTransfer", "tx", tx.Hash().Hex())
			return nil
		} else {
			cce.handler.LockChainAccount()
			defer cce.handler.UnLockChainAccount()
			auth := MakeTransactOpts(cce.handler.chainKey, big.NewInt((int64)(cce.handler.getChainAccountNonce())), cce.handler.parentChainID, new(big.Int).SetUint64(cce.subbridge.handler.remoteGasPrice))
			tx, err := bridgeInfo.bridge.HandleKLAYTransfer(auth, token.Amount, to, token.RequestNonce)
			if err != nil {
				logger.Error("Parent Bridge failed to HandleKLAYTransfer", "err", err)
				return err
			}
			cce.handler.addChainAccountNonce(1)
			logger.Info("Parent Bridge succeeded to HandleKLAYTransfer", "tx", tx.Hash().Hex())
			return nil
		}
	case TOKEN:
		logger.Info("Got request token transfer event")
		if bridgeInfo.onServiceChain {
			auth := MakeTransactOpts(cce.handler.nodeKey, big.NewInt((int64)(cce.handler.getNodeAccountNonce())), cce.subbridge.getChainID(), cce.subbridge.txPool.GasPrice())
			tx, err := bridgeInfo.bridge.HandleTokenTransfer(auth, token.Amount, to, tokenAddr, token.RequestNonce)
			if err != nil {
				logger.Error("Child Bridge failed to HandleTokenTransfer", "err", err)
				return err
			}
			logger.Info("Child Bridge succeeded to HandleTokenTransfer", "tx", tx.Hash().Hex())
			return nil
		} else {
			cce.handler.LockChainAccount()
			defer cce.handler.UnLockChainAccount()
			auth := MakeTransactOpts(cce.handler.chainKey, big.NewInt((int64)(cce.handler.getChainAccountNonce())), cce.handler.parentChainID, new(big.Int).SetUint64(cce.subbridge.handler.remoteGasPrice))
			tx, err := bridgeInfo.bridge.HandleTokenTransfer(auth, token.Amount, to, tokenAddr, token.RequestNonce)
			if err != nil {
				logger.Error("Parent Bridge failed to HandleTokenTransfer", "err", err)
				return err
			}
			cce.handler.addChainAccountNonce(1)
			logger.Info("Parent Bridge succeeded to HandleTokenTransfer", "tx", tx.Hash().Hex())
			return nil
		}
	case NFT:
		logger.Info("Got request NFT transfer event")
		if bridgeInfo.onServiceChain {
			auth := MakeTransactOpts(cce.handler.nodeKey, big.NewInt((int64)(cce.handler.getNodeAccountNonce())), cce.subbridge.getChainID(), cce.subbridge.txPool.GasPrice())
			tx, err := bridgeInfo.bridge.HandleNFTTransfer(auth, token.Amount, to, tokenAddr, token.RequestNonce)
			if err != nil {
				logger.Error("Child Bridge failed to HandleNFTTransfer", "err", err)
				return err
			}
			logger.Info("Child Bridge succeeded to HandleNFTTransfer", "tx", tx.Hash().Hex())
			return nil
		} else {
			cce.handler.LockChainAccount()
			defer cce.handler.UnLockChainAccount()
			auth := MakeTransactOpts(cce.handler.chainKey, big.NewInt((int64)(cce.handler.getChainAccountNonce())), cce.handler.parentChainID, new(big.Int).SetUint64(cce.subbridge.handler.remoteGasPrice))
			tx, err := bridgeInfo.bridge.HandleNFTTransfer(auth, token.Amount, to, tokenAddr, token.RequestNonce)
			if err != nil {
				logger.Error("Parent Bridge failed to HandleNFTTransfer", "err", err)
				return err
			}
			cce.handler.addChainAccountNonce(1)
			logger.Info("Parent Bridge succeeded to HandleNFTTransfer", "tx", tx.Hash().Hex())
			return nil
		}
	default:
		logger.Error("Got Unknown Token Type ReceivedEvent")
	}

	return errors.New("unknown token type event")
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
