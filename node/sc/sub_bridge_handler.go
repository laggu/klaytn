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
	"crypto/ecdsa"
	"fmt"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/networks/p2p"
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
	"sync/atomic"
)

const (
	SyncRequestInterval = 10
)

// parentChainInfo handles the information of parent chain, which is needed from child chain.
type parentChainInfo struct {
	Nonce    uint64
	GasPrice uint64
}

type SubBridgeHandler struct {
	subbridge *SubBridge
	// parentChainID is the first received chainID from parent chain peer.
	// It will be reset to nil if there's no parent peer.
	parentChainID *big.Int
	// chainKey is a private key for account in parent chain, owned by service chain admin.
	// Used for signing transaction executed on the parent chain.
	chainKey *ecdsa.PrivateKey
	// ChainAccountAddr is a hex account address used for chain identification from parent chain.
	ChainAccountAddr *common.Address
	// remoteGasPrice means gas price of parent chain, used to make a service chain transaction.
	// Therefore, for now, it is only used by child chain side.
	remoteGasPrice    uint64
	chainAccountNonce uint64
	nonceSynced       bool
	chainTxPeriod     uint64

	// TODO-Klaytn-ServiceChain Need to limit the number independently? Or just managing the size of sentServiceChainTxs?
	sentServiceChainTxsLimit uint64

	skipSyncBlockCount int32
}

func NewSubBridgeHandler(scc *SCConfig, main *SubBridge) (*SubBridgeHandler, error) {
	var chainAccountAddr *common.Address
	if scc.ChainAccountAddr != nil {
		chainAccountAddr = scc.ChainAccountAddr
	} else {
		chainKeyAddr := crypto.PubkeyToAddress(scc.chainkey.PublicKey)
		chainAccountAddr = &chainKeyAddr
		scc.ChainAccountAddr = chainAccountAddr
	}
	return &SubBridgeHandler{
		subbridge:                main,
		ChainAccountAddr:         chainAccountAddr,
		chainKey:                 scc.chainkey,
		remoteGasPrice:           uint64(0),
		chainAccountNonce:        uint64(0),
		nonceSynced:              false,
		chainTxPeriod:            scc.AnchoringPeriod,
		sentServiceChainTxsLimit: scc.SentChainTxsLimit,
	}, nil
}

func (sbh *SubBridgeHandler) setParentChainID(chainId *big.Int) {
	sbh.parentChainID = chainId
}

func (sbh *SubBridgeHandler) getParentChainID() *big.Int {
	return sbh.parentChainID
}

// getChainAccountNonce returns the chain account nonce of chain account address.
func (sbh *SubBridgeHandler) getChainAccountNonce() uint64 {
	return atomic.LoadUint64(&sbh.chainAccountNonce)
}

// setChainAccountNonce sets the chain account nonce of chain account address.
func (sbh *SubBridgeHandler) setChainAccountNonce(newNonce uint64) {
	atomic.StoreUint64(&sbh.chainAccountNonce, newNonce)
}

// addChainAccountNonce increases nonce by number
func (sbh *SubBridgeHandler) addChainAccountNonce(number uint64) {
	atomic.AddUint64(&sbh.chainAccountNonce, number)
}

// getChainAccountNonceSynced returns whether the chain account nonce is synced or not.
func (sbh *SubBridgeHandler) getChainAccountNonceSynced() bool {
	return sbh.nonceSynced
}

// setChainAccountNonceSynced sets whether the chain account nonce is synced or not.
func (sbh *SubBridgeHandler) setChainAccountNonceSynced(synced bool) {
	sbh.nonceSynced = synced
}

func (sbh *SubBridgeHandler) getRemoteGasPrice() uint64 {
	return sbh.remoteGasPrice
}

func (sbh *SubBridgeHandler) setRemoteGasPrice(gasPrice uint64) {
	sbh.remoteGasPrice = gasPrice
}

// GetChainAccountAddr returns a pointer of a hex address of an account used for service chain.
// If given as a parameter, it will use it. If not given, it will use the address of the public key
// derived from chainKey.
func (sbh *SubBridgeHandler) GetChainAccountAddr() *common.Address {
	return sbh.ChainAccountAddr
}

// getChainKey returns the private key used for signing service chain tx.
func (sbh *SubBridgeHandler) getChainKey() *ecdsa.PrivateKey {
	return sbh.chainKey
}

// GetAnchoringPeriod returns the period to make and send a chain transaction to parent chain.
func (sbh *SubBridgeHandler) GetAnchoringPeriod() uint64 {
	return sbh.chainTxPeriod
}

// GetSentChainTxsLimit returns the maximum number of stored chain transactions for resending.
func (sbh *SubBridgeHandler) GetSentChainTxsLimit() uint64 {
	return sbh.sentServiceChainTxsLimit
}

func (sbh *SubBridgeHandler) HandleMainMsg(p BridgePeer, msg p2p.Msg) error {
	// Handle the message depending on its contents
	switch msg.Code {
	case StatusMsg:
		return nil
	case ServiceChainParentChainInfoResponseMsg:
		logger.Debug("received ServiceChainParentChainInfoResponseMsg")
		if err := sbh.handleServiceChainParentChainInfoResponseMsg(p, msg); err != nil {
			return err
		}
	case ServiceChainReceiptResponseMsg:
		logger.Debug("received ServiceChainReceiptResponseMsg")
		if err := sbh.handleServiceChainReceiptResponseMsg(p, msg); err != nil {
			return err
		}
	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}

// handleServiceChainParentChainInfoResponseMsg handles parent chain info response message from parent chain.
// It will update the chainAccountNonce and remoteGasPrice of ServiceChainProtocolManager.
func (sbh *SubBridgeHandler) handleServiceChainParentChainInfoResponseMsg(p BridgePeer, msg p2p.Msg) error {
	var pcInfo parentChainInfo
	if err := msg.Decode(&pcInfo); err != nil {
		logger.Error("failed to decode", "err", err)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	poolNonce := sbh.subbridge.bridgeTxPool.GetMaxTxNonce(sbh.GetChainAccountAddr())
	if poolNonce > 0 {
		poolNonce += 1
		// just check
		if sbh.getChainAccountNonce() > poolNonce {
			logger.Error("chain account nonce is bigger than the chain pool nonce.", "chainPoolNonce", poolNonce, "chainAccountNonce", sbh.getChainAccountNonce())
		}
		if poolNonce < pcInfo.Nonce {
			// bridgeTxPool journal miss txs which already sent to parent-chain
			logger.Error("chain pool nonce is less than the parent chain nonce.", "chainPoolNonce", poolNonce, "parentChainNonce", pcInfo.Nonce)
			sbh.setChainAccountNonce(pcInfo.Nonce)
		} else {
			// bridgeTxPool journal has txs which don't receive receipt from parent-chain
			sbh.setChainAccountNonce(poolNonce)
		}
	} else if sbh.getChainAccountNonce() > pcInfo.Nonce {
		logger.Error("chain account nonce is bigger than the parent chain nonce.", "chainAccountNonce", sbh.getChainAccountNonce(), "parentChainNonce", pcInfo.Nonce)
		sbh.setChainAccountNonce(pcInfo.Nonce)
	} else {
		// there is no tx in bridgetTxPool, so parent-chain's nonce is used
		sbh.setChainAccountNonce(pcInfo.Nonce)
	}
	sbh.setChainAccountNonceSynced(true)
	sbh.setRemoteGasPrice(pcInfo.GasPrice)
	logger.Info("ServiceChainNonceResponse", "nonce", pcInfo.Nonce, "gasPrice", pcInfo.GasPrice, "chainAccountNonce", sbh.getChainAccountNonce())
	return nil
}

// handleServiceChainReceiptResponseMsg handles receipt response message from parent chain.
// It will store the received receipts and remove corresponding transaction in the resending list.
func (sbh *SubBridgeHandler) handleServiceChainReceiptResponseMsg(p BridgePeer, msg p2p.Msg) error {
	// TODO-Klaytn-ServiceChain Need to add an option, not to write receipts.
	// Decode the retrieval message
	var receipts []*types.ReceiptForStorage
	if err := msg.Decode(&receipts); err != nil && err != rlp.EOL {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	// Stores receipt and remove tx from sentServiceChainTxs only if the tx is successfully executed.
	sbh.writeServiceChainTxReceipts(sbh.subbridge.blockchain, receipts)
	return nil
}

// genUnsignedServiceChainTx generates an unsigned transaction, which type is TxTypeChainDataAnchoring.
// Nonce of account used for service chain transaction will be increased after the signing.
func (sbh *SubBridgeHandler) genUnsignedServiceChainTx(block *types.Block) (*types.Transaction, error) {
	chainHashes := types.NewChainHashes(block)
	encodedCCTxData, err := rlp.EncodeToBytes(chainHashes)
	if err != nil {
		return nil, err
	}

	values := map[types.TxValueKeyType]interface{}{
		types.TxValueKeyNonce:        sbh.getChainAccountNonce(), // chain account nonce will be increased after signing a transaction.
		types.TxValueKeyFrom:         *sbh.GetChainAccountAddr(),
		types.TxValueKeyTo:           *sbh.GetChainAccountAddr(),
		types.TxValueKeyAmount:       new(big.Int).SetUint64(0),
		types.TxValueKeyGasLimit:     uint64(999999999998), // TODO-Klaytn-ServiceChain should define proper gas limit
		types.TxValueKeyGasPrice:     new(big.Int).SetUint64(sbh.remoteGasPrice),
		types.TxValueKeyAnchoredData: encodedCCTxData,
	}

	if tx, err := types.NewTransactionWithMap(types.TxTypeChainDataAnchoring, values); err != nil {
		return nil, err
	} else {
		return tx, nil
	}
}

// NewAnchoringTx broadcasts service chain transactions and
func (sbh *SubBridgeHandler) NewAnchoringTx(block *types.Block) {
	if sbh.getChainAccountNonceSynced() {
		sbh.blockAnchoringManager(block)
		sbh.broadcastServiceChainTx()
		sbh.broadcastServiceChainReceiptRequest()
		sbh.skipSyncBlockCount = 0
	} else {
		if sbh.skipSyncBlockCount%SyncRequestInterval == 0 {
			// TODO-Klaytn too many request while sync main-net
			sbh.SyncNonceAndGasPrice()
			// check tx's receipts which parent-chain already executed in bridgeTxPool
			go sbh.broadcastServiceChainReceiptRequest()
		}
		sbh.skipSyncBlockCount++
	}
}

// broadcastServiceChainTx broadcasts service chain transactions to parent chain peers.
// It signs the given unsigned transaction with parent chain ID and then send it to its
// parent chain peers.
func (sbh *SubBridgeHandler) broadcastServiceChainTx() {
	parentChainID := sbh.parentChainID
	if parentChainID == nil {
		logger.Error("unexpected nil parentChainID while broadcastServiceChainTx")
	}
	txs := sbh.subbridge.GetBridgeTxPool().PendingTxsByAddress(sbh.ChainAccountAddr, (int)(sbh.GetSentChainTxsLimit()))
	peers := sbh.subbridge.BridgePeerSet().peers
	if len(peers) == 0 {
		sbh.setChainAccountNonceSynced(false)
	}
	for _, peer := range peers {
		if peer.GetChainID().Cmp(parentChainID) != 0 {
			logger.Error("parent peer with different parent chainID", "peerID", peer.GetID(), "peer chainID", peer.GetChainID(), "parent chainID", parentChainID)
			continue
		}
		peer.SendServiceChainTxs(txs)
		logger.Debug("sent ServiceChainTxData", "peerID", peer.GetID())
	}
	logger.Info("broadcastServiceChainTx ServiceChainTxData", "len(txs)", len(txs), "len(peers)", len(peers))
}

// writeServiceChainTxReceipt writes the received receipts of service chain transactions.
func (sbh *SubBridgeHandler) writeServiceChainTxReceipts(bc *blockchain.BlockChain, receipts []*types.ReceiptForStorage) {
	for _, receipt := range receipts {
		txHash := receipt.TxHash
		if tx := sbh.subbridge.GetBridgeTxPool().Get(txHash); tx != nil {
			chainHashes := new(types.ChainHashes)
			data, err := tx.AnchoredData()
			if err != nil {
				logger.Error("failed to get anchoring tx type from the tx", "txHash", txHash.String())
				return
			}
			if err := rlp.DecodeBytes(data, chainHashes); err != nil {
				logger.Error("failed to RLP decode ChainHashes", "txHash", txHash.String())
				return
			}
			sbh.WriteReceiptFromParentChain(chainHashes.BlockHash, (*types.Receipt)(receipt))
			sbh.subbridge.GetBridgeTxPool().RemoveTx(tx)
		} else {
			logger.Error("received service chain transaction receipt does not exist in sentServiceChainTxs", "txHash", txHash.String())
		}

		logger.Info("received service chain transaction receipt", "txHash", txHash.String())
	}
}

func (sbh *SubBridgeHandler) RegisterNewPeer(p BridgePeer) error {
	if sbh.getParentChainID() == nil {
		sbh.setParentChainID(p.GetChainID())
		return nil
	}
	if sbh.getParentChainID().Cmp(p.GetChainID()) != 0 {
		return fmt.Errorf("attempt to add a peer with different chainID failed! existing chainID: %v, new chainID: %v", sbh.getParentChainID(), p.GetChainID())
	}
	return nil
}

// broadcastServiceChainReceiptRequest broadcasts receipt requests for service chain transactions.
func (sbh *SubBridgeHandler) broadcastServiceChainReceiptRequest() {
	hashes := sbh.subbridge.GetBridgeTxPool().PendingTxHashsByAddress(sbh.GetChainAccountAddr(), (int)(sbh.GetSentChainTxsLimit()))
	for _, peer := range sbh.subbridge.BridgePeerSet().peers {
		peer.SendServiceChainReceiptRequest(hashes)
		logger.Debug("sent ServiceChainReceiptRequest", "peerID", peer.GetID(), "numReceiptsRequested", len(hashes))
	}
}

func (sbh *SubBridgeHandler) blockAnchoringManager(block *types.Block) {
	latestAnchoredBlockNumber := sbh.GetLatestAnchoredBlockNumber()
	var successCnt, cnt, blkNum uint64
	for cnt, blkNum = 0, latestAnchoredBlockNumber+1; cnt <= sbh.sentServiceChainTxsLimit && blkNum <= block.Number().Uint64(); cnt, blkNum = cnt+1, blkNum+1 {
		if err := sbh.generateAndAddAnchoringTxIntoTxPool(sbh.subbridge.blockchain.GetBlockByNumber(blkNum)); err == nil {
			sbh.WriteAnchoredBlockNumber(blkNum)
			successCnt++
		} else {
			logger.Error("blockAnchoringManager: break to generateAndAddAnchoringTxIntoTxPool", "cnt", cnt, "startBlockNumber", latestAnchoredBlockNumber+1, "FaildBlockNumber", blkNum, "latestBlockNum", block.NumberU64())
			break
		}
	}
	logger.Info("blockAnchoringManager: Success to generate anchoring txs", "successCnt", successCnt, "startBlockNumber", latestAnchoredBlockNumber+1, "latestBlockNum", block.NumberU64())
}

func (sbh *SubBridgeHandler) generateAndAddAnchoringTxIntoTxPool(block *types.Block) error {
	// Generating Anchoring Tx
	if block.NumberU64()%sbh.chainTxPeriod != 0 {
		return nil
	}
	//scpm.muChainAccount.Lock()
	//defer scpm.muChainAccount.Unlock()
	unsignedTx, err := sbh.genUnsignedServiceChainTx(block)
	if err != nil {
		logger.Error("Failed to generate service chain transaction", "blockNum", block.NumberU64(), "err", err)
		return err
	}
	// TODO-Klaytn-ServiceChain Change types.NewEIP155Signer to types.MakeSigner using parent chain's chain config and block number
	signedTx, err := types.SignTx(unsignedTx, types.NewEIP155Signer(sbh.parentChainID), sbh.getChainKey())
	if err != nil {
		logger.Error("failed signing tx", "err", err)
		return err
	}
	if err := sbh.subbridge.GetBridgeTxPool().AddLocal(signedTx); err == nil {
		sbh.addChainAccountNonce(1)
	} else {
		logger.Error("failed to add signed tx into txpool", "err", err)
		return err
	}
	return nil
}

// SyncNonceAndGasPrice requests the nonce of address used for service chain tx to parent chain peers.
func (scpm *SubBridgeHandler) SyncNonceAndGasPrice() {
	addr := scpm.GetChainAccountAddr()
	for _, peer := range scpm.subbridge.BridgePeerSet().peers {
		peer.SendServiceChainInfoRequest(addr)
	}
}

// GetLatestAnchoredBlockNumber returns the latest block number whose data has been anchored to the parent chain.
func (sbh *SubBridgeHandler) GetLatestAnchoredBlockNumber() uint64 {
	return sbh.subbridge.ChainDB().ReadAnchoredBlockNumber()
}

// WriteAnchoredBlockNumber writes the block number whose data has been anchored to the parent chain.
func (sbh *SubBridgeHandler) WriteAnchoredBlockNumber(blockNum uint64) {
	sbh.subbridge.chainDB.WriteAnchoredBlockNumber(blockNum)
}

// WriteReceiptFromParentChain writes a receipt received from parent chain to child chain
// with corresponding block hash. It assumes that a child chain has only one parent chain.
func (sbh *SubBridgeHandler) WriteReceiptFromParentChain(blockHash common.Hash, receipt *types.Receipt) {
	sbh.subbridge.chainDB.WriteReceiptFromParentChain(blockHash, receipt)
}

// GetReceiptFromParentChain returns a receipt received from parent chain to child chain
// with corresponding block hash. It assumes that a child chain has only one parent chain.
func (sbh *SubBridgeHandler) GetReceiptFromParentChain(blockHash common.Hash) *types.Receipt {
	return sbh.subbridge.chainDB.ReadReceiptFromParentChain(blockHash)
}
