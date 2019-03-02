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
	"github.com/ground-x/klaytn/datasync/downloader"
	"github.com/ground-x/klaytn/networks/p2p"
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
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
	//protocolHandler     ServiceChainProtocolHandler
	sentServiceChainTxs map[common.Hash]*types.Transaction
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
		sentServiceChainTxs:      make(map[common.Hash]*types.Transaction),
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
	return sbh.chainAccountNonce
}

// setChainAccountNonce sets the chain account nonce of chain account address.
func (sbh *SubBridgeHandler) setChainAccountNonce(newNonce uint64) {
	sbh.chainAccountNonce = newNonce
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
	case ServiceChainTxsMsg:
		logger.Debug("received ServiceChainTxsMsg")
		// Transactions arrived, make sure we have a valid and fresh chain to handle them
		//if atomic.LoadUint32(&pm.acceptTxs) == 0 {
		//	break
		//}
		if err := sbh.handleServiceChainTxDataMsg(p, msg); err != nil {
			return err
		}
	case ServiceChainParentChainInfoRequestMsg:
		logger.Debug("received ServiceChainParentChainInfoRequestMsg")
		if err := sbh.handleServiceChainParentChainInfoRequestMsg(p, msg); err != nil {
			return err
		}
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
	case ServiceChainReceiptRequestMsg:
	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}

// handleServiceChainTxDataMsg handles service chain transactions from child chain.
// It will return an error if given tx is not TxTypeChainDataAnchoring type.
func (sbh *SubBridgeHandler) handleServiceChainTxDataMsg(p BridgePeer, msg p2p.Msg) error {
	//pm.txMsgLock.Lock()
	// Transactions can be processed, parse all of them and deliver to the pool
	var txs []*types.Transaction
	if err := msg.Decode(&txs); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	// Only valid txs should be pushed into the pool.
	validTxs := make([]*types.Transaction, 0, len(txs))
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
	sbh.subbridge.txPool.AddRemotes(validTxs)
	return err
}

// handleServiceChainParentChainInfoRequestMsg handles parent chain info request message from child chain.
// It will send the nonce of the account and its gas price to the child chain peer who requested.
func (sbh *SubBridgeHandler) handleServiceChainParentChainInfoRequestMsg(p BridgePeer, msg p2p.Msg) error {
	var addr common.Address
	if err := msg.Decode(&addr); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	stateDB, err := sbh.subbridge.blockchain.State()
	if err != nil {
		// TODO-Klaytn-ServiceChain This error and inefficient balance error should be transferred.
		return errResp(ErrFailedToGetStateDB, "failed to get stateDB, err: %v", err)
	} else {
		pcInfo := parentChainInfo{stateDB.GetNonce(addr), sbh.subbridge.blockchain.Config().UnitPrice}
		p.SendServiceChainInfoResponse(&pcInfo)
		logger.Debug("SendServiceChainInfoResponse", "addr", addr, "nonce", pcInfo.Nonce, "gasPrice", pcInfo.GasPrice)
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
	if sbh.getChainAccountNonce() > pcInfo.Nonce {
		// If received nonce is bigger than the current one, just leave a log and do nothing.
		logger.Warn("chain account nonce is bigger than the parent chain nonce.", "chainAccountNonce", sbh.getChainAccountNonce(), "parentChainNonce", pcInfo.Nonce)
		return nil
	}
	sbh.setChainAccountNonce(pcInfo.Nonce)
	sbh.setChainAccountNonceSynced(true)
	sbh.setRemoteGasPrice(pcInfo.GasPrice)
	logger.Debug("ServiceChainNonceResponse", "nonce", pcInfo.Nonce, "gasPrice", pcInfo.GasPrice)
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

// handleServiceChainReceiptRequestMsg handles receipt request message from child chain.
// It will find and send corresponding receipts with given transaction hashes.
func (sbh *SubBridgeHandler) handleServiceChainReceiptRequestMsg(p BridgePeer, msg p2p.Msg) error {
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
		receipt := sbh.subbridge.blockchain.GetReceiptByTxHash(hash)
		if receipt == nil {
			continue
		}

		receiptsForStorage = append(receiptsForStorage, (*types.ReceiptForStorage)(receipt))
	}
	return p.SendServiceChainReceiptResponse(receiptsForStorage)
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

// BroadcastServiceChainTxAndReceiptRequest broadcasts service chain transactions and
// request receipts to parent chain peers.
func (sbh *SubBridgeHandler) BroadcastServiceChainTxAndReceiptRequest(block *types.Block) {
	// Before broadcasting service chain transactions and receipt requests,
	// check connection and nonceSynced.
	if sbh.subbridge.BridgePeerSet().Len() == 0 {
		sbh.setChainAccountNonceSynced(false)
		sbh.parentChainID = nil
		return
	}
	if !sbh.getChainAccountNonceSynced() {
		sbh.SyncNonceAndGasPrice()
		// If nonce is not synced, clear sent service chain txs.
		sbh.sentServiceChainTxs = make(map[common.Hash]*types.Transaction)
		return
	}
	if block.NumberU64()%sbh.chainTxPeriod != 0 {
		return
	}
	tx, err := sbh.genUnsignedServiceChainTx(block)
	if err != nil {
		logger.Error("Failed to generate service chain transaction", "blockNum", block.NumberU64(), "err", err)
		return
	}
	sbh.BroadcastServiceChainTx(tx)
	sbh.broadcastServiceChainReceiptRequest()
}

// BroadcastServiceChainTx broadcasts service chain transactions to parent chain peers.
// It signs the given unsigned transaction with parent chain ID and then send it to its
// parent chain peers.
func (sbh *SubBridgeHandler) BroadcastServiceChainTx(unsignedTx *types.Transaction) {
	parentChainID := sbh.parentChainID
	if parentChainID == nil {
		logger.Error("unexpected nil parentChainID while BroadcastServiceChainTx")
		return
	}
	// TODO-Klaytn-ServiceChain Change types.NewEIP155Signer to types.MakeSigner using parent chain's chain config and block number
	signedTx, err := types.SignTx(unsignedTx, types.NewEIP155Signer(sbh.parentChainID), sbh.getChainKey())
	if err != nil {
		logger.Error("failed signing tx", "err", err)
		return
	}
	sbh.chainAccountNonce++
	sbh.addToSentServiceChainTxs(signedTx)
	txs := sbh.getSentServiceChainTxsSlice()

	for _, peer := range sbh.subbridge.BridgePeerSet().peers {
		if peer.GetChainID() != parentChainID {
			logger.Debug("parent peer with different parent chainID", "peerID", peer.GetID(), "peer chainID", peer.GetChainID(), "parent chainID", parentChainID)
			continue
		}
		peer.SendServiceChainTxs(txs)
		logger.Debug("sent ServiceChainTxData", "peerID", peer.GetID())
	}
}

// writeServiceChainTxReceipt writes the received receipts of service chain transactions.
func (sbh *SubBridgeHandler) writeServiceChainTxReceipts(bc *blockchain.BlockChain, receipts []*types.ReceiptForStorage) {
	sentServiceChainTxs := sbh.sentServiceChainTxs
	for _, receipt := range receipts {
		txHash := receipt.TxHash
		if tx, ok := sentServiceChainTxs[txHash]; ok {
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
			sbh.removeServiceChainTx(txHash)
		} else {
			logger.Error("received service chain transaction receipt does not exist in sentServiceChainTxs", "txHash", txHash.String())
		}

		logger.Info("received service chain transaction receipt", "txHash", txHash.String())
	}
}

// addToSentServiceChainTxs adds a transaction to SentServiceChainTxs.
func (sbh *SubBridgeHandler) addToSentServiceChainTxs(tx *types.Transaction) {
	if uint64(len(sbh.sentServiceChainTxs)) > sbh.sentServiceChainTxsLimit {
		logger.Warn("Number of txs in sentServiceChainTxs already exceeds the limit", "sentServiceChainTxsLimit", sbh.sentServiceChainTxsLimit)
		return
	}
	if _, ok := sbh.sentServiceChainTxs[tx.Hash()]; ok {
		logger.Error("ServiceChainTx already exists in sentServiceChainTxs", "txHash", tx.Hash())
		return
	}
	sbh.sentServiceChainTxs[tx.Hash()] = tx
}

// removeServiceChainTx removes a transaction from SentServiceChainTxs with the given
// transaction hash.
func (sbh *SubBridgeHandler) removeServiceChainTx(txHash common.Hash) {
	if _, ok := sbh.sentServiceChainTxs[txHash]; !ok {
		logger.Error("ServiceChainTx does not exists in sentServiceChainTxs", "txHash", txHash)
		return
	}
	delete(sbh.sentServiceChainTxs, txHash)
}

// getSentServiceChainTxsHashes returns only the hashes of SentServiceChainTxs.
func (sbh *SubBridgeHandler) getSentServiceChainTxsHashes() []common.Hash {
	var hashes []common.Hash
	for k := range sbh.sentServiceChainTxs {
		hashes = append(hashes, k)
	}
	return hashes
}

// getSentServiceChainTxsSlice returns SentServiceChainTxs in types.Transactions.
func (sbh *SubBridgeHandler) getSentServiceChainTxsSlice() types.Transactions {
	var txs types.Transactions
	for _, v := range sbh.sentServiceChainTxs {
		txs = append(txs, v)
	}
	return txs
}

// WriteReceiptFromParentChain writes a receipt received from parent chain to child chain
// with corresponding block hash. It assumes that a child chain has only one parent chain.
func (sbh *SubBridgeHandler) WriteReceiptFromParentChain(blockHash common.Hash, receipt *types.Receipt) {
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
func (scpm *SubBridgeHandler) broadcastServiceChainReceiptRequest() {
	hashes := scpm.getSentServiceChainTxsHashes()
	for _, peer := range scpm.subbridge.BridgePeerSet().peers {
		peer.SendServiceChainReceiptRequest(hashes)
		logger.Debug("sent ServiceChainReceiptRequest", "peerID", peer.GetID(), "numReceiptsRequested", len(hashes))
	}
}

// SyncNonceAndGasPrice requests the nonce of address used for service chain tx to parent chain peers.
func (scpm *SubBridgeHandler) SyncNonceAndGasPrice() {
	addr := scpm.GetChainAccountAddr()
	for _, peer := range scpm.subbridge.BridgePeerSet().peers {
		peer.SendServiceChainInfoRequest(addr)
	}
}
