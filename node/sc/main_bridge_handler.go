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
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/datasync/downloader"
	"github.com/ground-x/klaytn/networks/p2p"
	"github.com/ground-x/klaytn/ser/rlp"
)

type MainBridgeHandler struct {
	mainbridge *MainBridge

	protocolHandler ServiceChainProtocolHandler
}

func NewMainBridgeHandler(main *MainBridge) (*MainBridgeHandler, error) {

	handler := NewServiceChainProtocolHandler(main.config, main, main.eventhandler)

	return &MainBridgeHandler{mainbridge: main, protocolHandler: handler}, nil
}

func (mbh *MainBridgeHandler) HandleSubMsg(p BridgePeer, msg p2p.Msg) error {

	// Handle the message depending on its contents
	switch msg.Code {
	case StatusMsg:
		return nil
	case ServiceChainTxsMsg:
		scLogger.Debug("received ServiceChainTxsMsg")
		// Transactions arrived, make sure we have a valid and fresh chain to handle them
		//if atomic.LoadUint32(&pm.acceptTxs) == 0 {
		//	break
		//}
		//fmt.Println("========== received servicechainTX")
		if err := mbh.handleServiceChainTxDataMsg(p, msg); err != nil {
			return err
		}
	case ServiceChainParentChainInfoRequestMsg:
		scLogger.Debug("received ServiceChainParentChainInfoRequestMsg")
		if err := mbh.handleServiceChainParentChainInfoRequestMsg(p, msg); err != nil {
			return err
		}
	case ServiceChainParentChainInfoResponseMsg:
		scLogger.Debug("received ServiceChainParentChainInfoResponseMsg")
		if err := mbh.handleServiceChainParentChainInfoResponseMsg(p, msg); err != nil {
			return err
		}
	case ServiceChainReceiptResponseMsg:
		logger.Debug("received ServiceChainReceiptResponseMsg")
		if err := mbh.handleServiceChainReceiptResponseMsg(p, msg); err != nil {
			return err
		}
	case ServiceChainReceiptRequestMsg:
		scLogger.Debug("received ServiceChainReceiptRequestMsg")
		if err := mbh.handleServiceChainReceiptRequestMsg(p, msg); err != nil {
			return err
		}
	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}

// handleServiceChainTxDataMsg handles service chain transactions from child chain.
// It will return an error if given tx is not TxTypeChainDataAnchoring type.
func (mbh *MainBridgeHandler) handleServiceChainTxDataMsg(p BridgePeer, msg p2p.Msg) error {
	//pm.txMsgLock.Lock()
	// Transactions can be processed, parse all of them and deliver to the pool
	var txs []*types.Transaction
	if err := msg.Decode(&txs); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	// Only valid txs should be pushed into the pool.
	validTxs := make([]*types.Transaction, 0, len(txs))
	//validTxs := []*types.Transaction{}
	var err error
	for i, tx := range txs {
		if tx == nil {
			err = errResp(ErrDecode, "tx %d is nil", i)
			continue
		}
		if tx.Type() != types.TxTypeChainDataAnchoring {
			err = errResp(ErrUnexpectedTxType, "tx %d should be TxTypeChainDataAnchoring, but %s", i, tx.Type())
			continue
		}
		p.AddToKnownTxs(tx.Hash())
		validTxs = append(validTxs, tx)
	}
	mbh.mainbridge.txPool.AddRemotes(validTxs)
	return err
}

// handleServiceChainParentChainInfoRequestMsg handles parent chain info request message from child chain.
// It will send the nonce of the account and its gas price to the child chain peer who requested.
func (mbh *MainBridgeHandler) handleServiceChainParentChainInfoRequestMsg(p BridgePeer, msg p2p.Msg) error {
	var addr common.Address
	if err := msg.Decode(&addr); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	nonce := mbh.mainbridge.txPool.State().GetNonce(addr)
	pcInfo := parentChainInfo{nonce, mbh.mainbridge.blockchain.Config().UnitPrice}
	p.SendServiceChainInfoResponse(&pcInfo)
	scLogger.Debug("SendServiceChainInfoResponse", "addr", addr, "nonce", pcInfo.Nonce, "gasPrice", pcInfo.GasPrice)
	return nil
}

// handleServiceChainParentChainInfoResponseMsg handles parent chain info response message from parent chain.
// It will update the chainAccountNonce and remoteGasPrice of ServiceChainProtocolManager.
func (mbh *MainBridgeHandler) handleServiceChainParentChainInfoResponseMsg(p BridgePeer, msg p2p.Msg) error {
	var pcInfo parentChainInfo
	if err := msg.Decode(&pcInfo); err != nil {
		scLogger.Error("failed to decode", "err", err)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	if mbh.protocolHandler.getChainAccountNonce() > pcInfo.Nonce {
		// If received nonce is bigger than the current one, just leave a log and do nothing.
		scLogger.Warn("chain account nonce is bigger than the parent chain nonce.", "chainAccountNonce", mbh.protocolHandler.getChainAccountNonce(), "parentChainNonce", pcInfo.Nonce)
		return nil
	}
	mbh.protocolHandler.setChainAccountNonce(pcInfo.Nonce)
	mbh.protocolHandler.setChainAccountNonceSynced(true)
	mbh.protocolHandler.setRemoteGasPrice(pcInfo.GasPrice)
	scLogger.Debug("ServiceChainNonceResponse", "nonce", pcInfo.Nonce, "gasPrice", pcInfo.GasPrice)
	return nil
}

// handleServiceChainReceiptResponseMsg handles receipt response message from parent chain.
// It will store the received receipts and remove corresponding transaction in the resending list.
func (mbh *MainBridgeHandler) handleServiceChainReceiptResponseMsg(p BridgePeer, msg p2p.Msg) error {
	// TODO-Klaytn-ServiceChain Need to add an option, not to write receipts.
	// Decode the retrieval message
	var receipts []*types.ReceiptForStorage
	if err := msg.Decode(&receipts); err != nil && err != rlp.EOL {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	// Stores receipt and remove tx from sentServiceChainTxs only if the tx is successfully executed.
	mbh.protocolHandler.writeServiceChainTxReceipts(mbh.mainbridge.blockchain, receipts)
	return nil
}

// handleServiceChainReceiptRequestMsg handles receipt request message from child chain.
// It will find and send corresponding receipts with given transaction hashes.
func (mbh *MainBridgeHandler) handleServiceChainReceiptRequestMsg(p BridgePeer, msg p2p.Msg) error {
	// Decode the retrieval message
	msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
	if _, err := msgStream.List(); err != nil {
		return err
	}
	// Gather state data until the fetch or network limits is reached
	var (
		hash               common.Hash
		receiptsForStorage []*types.ReceiptForStorage
	)
	for len(receiptsForStorage) < downloader.MaxReceiptFetch {
		// Retrieve the hash of the next block
		if err := msgStream.Decode(&hash); err == rlp.EOL {
			break
		} else if err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Retrieve the receipt of requested service chain tx, skip if unknown.
		receipt := mbh.mainbridge.blockchain.GetReceiptByTxHash(hash)
		if receipt == nil {
			continue
		}

		receiptsForStorage = append(receiptsForStorage, (*types.ReceiptForStorage)(receipt))
	}
	return p.SendServiceChainReceiptResponse(receiptsForStorage)
}
