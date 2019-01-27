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

package cn

import (
	"crypto/ecdsa"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/log"
	"github.com/ground-x/klaytn/networks/p2p"
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
)

var scLogger = log.NewModuleLogger(log.ServiceChain)

type ServiceChainProtocolManager interface {
	// TODO-Klaytn-ServiceChain Please note that ServiceChainProtocolManager is made to separate its implementation
	// and message handling from pre-existing ProtocolManager. This can be done by pushing implementation into
	// a new peer type and handling message by a new peer type.
	// private functions
	getParentChainPeers() *peerSet // getParentChainPeers returns peers on the parent chain.
	getChildChainPeers() *peerSet  // getChildChainPeers returns peers on the child chain.
	removeParentPeer(id string)    // removeParentPeer removes a parent peer with given id.
	removeChildPeer(id string)     // removeChildPeer removes a child peer with given id.

	getRemoteNonce() uint64            // getRemoteNonce returns the nonce of address used for service chain tx.
	setRemoteNonce(newNonce uint64)    // setRemoteNonce sets the nonce of address used for service chain tx.
	getNonceSynced() bool              // getNonceSynced returns whether the nonce is synced or not.
	setNonceSynced(synced bool)        // setNonceSynced sets whether the nonce is synced or not.
	getRemoteGasPrice() uint64         // getRemoteGasPrice returns the gas price of parent chain.
	setRemoteGasPrice(gasPrice uint64) // setRemoteGasPrice sets the gas price of parent chain.

	getChainKey() *ecdsa.PrivateKey                                                             // getChainKey returns the private key used for signing service chain tx.
	addToSentServiceChainTxs(tx *types.Transaction)                                             // addToSentServiceChainTxs adds a transaction to SentServiceChainTxs.
	removeServiceChainTx(txHash common.Hash)                                                    // removeServiceChainTx removes a transaction from SentServiceChainTxs with the given transaction hash.
	writeServiceChainTxReceipts(bc *blockchain.BlockChain, receipts []*types.ReceiptForStorage) // writeServiceChainTxReceipt writes the received receipts of service chain transactions.

	// public functions
	GetChainAddr() *common.Address                               // GetChainAddr returns a pointer of a hex address of an account used for service chain.
	GetChainTxPeriod() uint64                                    // GetChainTxPeriod returns the period (in child chain blocks) of sending chain transactions.
	GetSentChainTxsLimit() uint64                                // GetSentChainTxsLimit returns the maximum number of stored chain transactions for resending.
	BroadcastServiceChainTxAndReceiptRequest(block *types.Block) // BroadcastServiceChainTxAndReceiptRequest broadcasts service chain transactions and request receipts to parent chain peers.
	BroadcastServiceChainTx(unsignedTx *types.Transaction)       // BroadcastServiceChainTx broadcasts service chain transactions to parent chain peers.
	SyncNonceAndGasPrice()                                       // SyncNonceAndGasPrice requests the nonce of address used for service chain tx to parent chain peers.
}

// serviceChainPM implements ServiceChainProtocolManager interface.
type serviceChainPM struct {
	parentChainPeers *peerSet
	childChainPeers  *peerSet

	// chainKey is a private key for account in parent chain, owned by service chain admin.
	// Used for signing transaction executed on the parent chain.
	chainKey *ecdsa.PrivateKey
	// ChainAddr is a hex account address used for chain identification from parent chain.
	ChainAddr *common.Address

	// remoteGasPrice means gas price of parent chain, used to make a service chain transaction.
	// Therefore, for now, it is only used by child chain side.
	remoteGasPrice      uint64
	remoteNonce         uint64
	nonceSynced         bool
	chainTxPeriod       uint64
	sentServiceChainTxs map[common.Hash]*types.Transaction

	// TODO-Klaytn-ServiceChain Need to limit the number independently? Or just managing the size of sentServiceChainTxs?
	sentServiceChainTxsLimit uint64
}

// ServiceChainConfig handles service chain configurations.
type ServiceChainConfig struct {
	ChainAddr   *common.Address
	ChainKey    *ecdsa.PrivateKey
	TxPeriod    uint64
	SentTxLimit uint64
}

// parentChainInfo handles the information of parent chain, which is needed from child chain.
type parentChainInfo struct {
	Nonce    uint64
	GasPrice uint64
}

// NewServiceChainProtocolManager generates a new ServiceChainProtocolManager with
// the given ServiceChainConfig.
func NewServiceChainProtocolManager(scc *ServiceChainConfig) ServiceChainProtocolManager {
	var chainAddr *common.Address
	if scc.ChainAddr != nil {
		chainAddr = scc.ChainAddr
	} else {
		chainKeyAddr := crypto.PubkeyToAddress(scc.ChainKey.PublicKey)
		chainAddr = &chainKeyAddr
	}
	return &serviceChainPM{
		parentChainPeers:         newPeerSet(),
		childChainPeers:          newPeerSet(),
		ChainAddr:                chainAddr,
		chainKey:                 scc.ChainKey,
		remoteGasPrice:           uint64(0),
		remoteNonce:              uint64(0),
		nonceSynced:              false,
		chainTxPeriod:            scc.TxPeriod,
		sentServiceChainTxs:      make(map[common.Hash]*types.Transaction),
		sentServiceChainTxsLimit: scc.SentTxLimit,
	}
}

// getParentChainPeers returns peers on the parent chain.
func (scpm *serviceChainPM) getParentChainPeers() *peerSet {
	return scpm.parentChainPeers
}

// getChildChainPeers returns peers on the child chain.
func (scpm *serviceChainPM) getChildChainPeers() *peerSet {
	return scpm.childChainPeers
}

// removeParentPeer removes a parent peer with given id.
func (scpm *serviceChainPM) removeParentPeer(id string) {
	// Short circuit if the peer was already removed
	peer := scpm.getParentChainPeers().Peer(id)
	if peer == nil {
		return
	}
	scLogger.Debug("Removing parent chain peer", "peer", id)
	// Unregister the peer from the downloader and parent chain peer set
	if err := scpm.getParentChainPeers().Unregister(id); err != nil {
		scLogger.Error("Parent chain peer removal failed", "peer", id, "err", err)
	}
	// Hard disconnect at the networking layer
	peer.GetP2PPeer().Disconnect(p2p.DiscUselessPeer)
}

// removeChildPeer removes a child peer with given id.
func (scpm *serviceChainPM) removeChildPeer(id string) {
	// Short circuit if the peer was already removed
	peer := scpm.getChildChainPeers().Peer(id)
	if peer == nil {
		return
	}
	scLogger.Debug("Removing child chain peer", "peer", id)
	// Unregister the peer from the downloader and child chain peer set
	if err := scpm.getChildChainPeers().Unregister(id); err != nil {
		scLogger.Error("Child chain peer removal failed", "peer", id, "err", err)
	}
	// Hard disconnect at the networking layer
	peer.GetP2PPeer().Disconnect(p2p.DiscUselessPeer)
}

// getRemoteNonce returns the nonce of address used for service chain tx.
func (scpm *serviceChainPM) getRemoteNonce() uint64 {
	return scpm.remoteNonce
}

// setRemoteNonce sets the nonce of address used for service chain tx.
func (scpm *serviceChainPM) setRemoteNonce(newNonce uint64) {
	scpm.remoteNonce = newNonce
}

// getNonceSynced returns whether the nonce is synced or not.
func (scpm *serviceChainPM) getNonceSynced() bool {
	return scpm.nonceSynced
}

// setNonceSynced sets whether the nonce is synced or not.
func (scpm *serviceChainPM) setNonceSynced(synced bool) {
	scpm.nonceSynced = synced
}

func (scpm *serviceChainPM) getRemoteGasPrice() uint64 {
	return scpm.remoteGasPrice
}

func (scpm *serviceChainPM) setRemoteGasPrice(gasPrice uint64) {
	scpm.remoteGasPrice = gasPrice
}

// GetChainAddr returns a pointer of a hex address of an account used for service chain.
// If given as a parameter, it will use it. If not given, it will use the address of the public key
// derived from chainKey.
func (scpm *serviceChainPM) GetChainAddr() *common.Address {
	return scpm.ChainAddr
}

// getChainKey returns the private key used for signing service chain tx.
func (scpm *serviceChainPM) getChainKey() *ecdsa.PrivateKey {
	return scpm.chainKey
}

// GetChainTxPeriod returns the period to make and send a chain transaction to parent chain.
func (scpm *serviceChainPM) GetChainTxPeriod() uint64 {
	return scpm.chainTxPeriod
}

// GetSentChainTxsLimit returns the maximum number of stored chain transactions for resending.
func (scpm *serviceChainPM) GetSentChainTxsLimit() uint64 {
	return scpm.sentServiceChainTxsLimit
}

// addToSentServiceChainTxs adds a transaction to SentServiceChainTxs.
func (scpm *serviceChainPM) addToSentServiceChainTxs(tx *types.Transaction) {
	if uint64(len(scpm.sentServiceChainTxs)) > scpm.sentServiceChainTxsLimit {
		scLogger.Warn("Number of txs in sentServiceChainTxs already exceeds the limit", "sentServiceChainTxsLimit", scpm.sentServiceChainTxsLimit)
		return
	}
	if _, ok := scpm.sentServiceChainTxs[tx.Hash()]; ok {
		scLogger.Error("ServiceChainTx already exists in sentServiceChainTxs", "txHash", tx.Hash())
		return
	}
	scpm.sentServiceChainTxs[tx.Hash()] = tx
}

// removeServiceChainTx removes a transaction from SentServiceChainTxs with the given
// transaction hash.
func (scpm *serviceChainPM) removeServiceChainTx(txHash common.Hash) {
	if _, ok := scpm.sentServiceChainTxs[txHash]; !ok {
		scLogger.Error("ServiceChainTx does not exists in sentServiceChainTxs", "txHash", txHash)
		return
	}
	delete(scpm.sentServiceChainTxs, txHash)
}

// getSentServiceChainTxsHashes returns only the hashes of SentServiceChainTxs.
func (scpm *serviceChainPM) getSentServiceChainTxsHashes() []common.Hash {
	var hashes []common.Hash
	for k := range scpm.sentServiceChainTxs {
		hashes = append(hashes, k)
	}
	return hashes
}

// getSentServiceChainTxsSlice returns SentServiceChainTxs in types.Transactions.
func (scpm *serviceChainPM) getSentServiceChainTxsSlice() types.Transactions {
	var txs types.Transactions
	for _, v := range scpm.sentServiceChainTxs {
		txs = append(txs, v)
	}
	return txs
}

// writeServiceChainTxReceipt writes the received receipts of service chain transactions.
func (scpm *serviceChainPM) writeServiceChainTxReceipts(bc *blockchain.BlockChain, receipts []*types.ReceiptForStorage) {
	sentServiceChainTxs := scpm.sentServiceChainTxs
	for _, receipt := range receipts {
		txHash := receipt.TxHash
		if tx, ok := sentServiceChainTxs[txHash]; ok {
			ccTxData := new(types.ChildChainTxData)
			data, err := tx.PeggedData()
			if err != nil {
				scLogger.Error("failed to get pegging tx type from the tx", "txHash", txHash.String())
				return
			}
			if err := rlp.DecodeBytes(data, ccTxData); err != nil {
				scLogger.Error("failed to RLP decode ChildChainTxData", "txHash", txHash.String())
				return
			}
			bc.WriteReceiptFromParentChain(ccTxData.BlockHash, (*types.Receipt)(receipt))
			scpm.removeServiceChainTx(txHash)
		} else {
			scLogger.Error("received service chain transaction receipt does not exist in sentServiceChainTxs", "txHash", txHash.String())
		}

		scLogger.Info("received service chain transaction receipt", "txHash", txHash.String())
	}

}

// BroadcastServiceChainTxAndReceiptRequest broadcasts service chain transactions and
// request receipts to parent chain peers.
func (scpm *serviceChainPM) BroadcastServiceChainTxAndReceiptRequest(block *types.Block) {
	// Before broadcasting service chain transactions and receipt requests,
	// check connection and nonceSynced.
	if scpm.getParentChainPeers().Len() == 0 {
		scpm.setNonceSynced(false)
		return
	}
	if !scpm.getNonceSynced() {
		scpm.SyncNonceAndGasPrice()
		// If nonce is not synced, clear sent service chain txs.
		scpm.sentServiceChainTxs = make(map[common.Hash]*types.Transaction)
		return
	}
	if block.NumberU64()%scpm.chainTxPeriod != 0 {
		return
	}
	tx, err := scpm.genUnsignedServiceChainTx(block)
	if err != nil {
		scLogger.Error("Failed to generate service chain transaction", "blockNum", block.NumberU64(), "err", err)
		return
	}
	scpm.BroadcastServiceChainTx(tx)
	scpm.broadcastServiceChainReceiptRequest()
}

// genUnsignedServiceChainTx generates an unsigned transaction, which type is TxTypeChainDataPegging.
// Nonce of account used for service chain transaction will be increased after the signing.
func (scpm *serviceChainPM) genUnsignedServiceChainTx(block *types.Block) (*types.Transaction, error) {
	ccTxData := types.NewChildChainTxData(block)
	encodedCCTxData, err := rlp.EncodeToBytes(ccTxData)
	if err != nil {
		return nil, err
	}

	values := map[types.TxValueKeyType]interface{}{
		types.TxValueKeyNonce:      scpm.getRemoteNonce(), // nonce will be increased after the signing.
		types.TxValueKeyFrom:       *scpm.GetChainAddr(),
		types.TxValueKeyTo:         *scpm.GetChainAddr(),
		types.TxValueKeyAmount:     new(big.Int).SetUint64(0),
		types.TxValueKeyGasLimit:   uint64(999999999998), // TODO-Klaytn-ServiceChain should define proper gas limit
		types.TxValueKeyGasPrice:   new(big.Int).SetUint64(scpm.remoteGasPrice),
		types.TxValueKeyPeggedData: encodedCCTxData,
	}

	if tx, err := types.NewTransactionWithMap(types.TxTypeChainDataPegging, values); err != nil {
		return nil, err
	} else {
		return tx, nil
	}
}

// BroadcastServiceChainTx broadcasts service chain transactions to parent chain peers.
// It signs the given unsigned transaction with parent chain ID and then send it to its
// parent chain peers.
func (scpm *serviceChainPM) BroadcastServiceChainTx(unsignedTx *types.Transaction) {
	var parentChainID *big.Int
	var txs types.Transactions
	for _, peer := range scpm.getParentChainPeers().peers {
		if parentChainID == nil {
			parentChainID = peer.GetChainID()
			// TODO-Klaytn-ServiceChain Change types.NewEIP155Signer to types.MakeSigner using parent chain's chain config and block number
			signedTx, err := types.SignTx(unsignedTx, types.NewEIP155Signer(parentChainID), scpm.getChainKey())
			if err != nil {
				scLogger.Error("failed signing tx", "err", err)
				return
			}
			scpm.addToSentServiceChainTxs(signedTx)
			txs = scpm.getSentServiceChainTxsSlice()
			scpm.remoteNonce++
		}
		if peer.GetChainID() != parentChainID {
			scLogger.Debug("parent peer with different parent chainID", "peerID", peer.GetID(), "peer chainID", peer.GetChainID(), "parent chainID", parentChainID)
			continue
		}
		peer.SendServiceChainTxs(txs)
		scLogger.Debug("sent ServiceChainTxData", "peerID", peer.GetID())
	}
}

// broadcastServiceChainReceiptRequest broadcasts receipt requests for service chain transactions.
func (scpm *serviceChainPM) broadcastServiceChainReceiptRequest() {
	hashes := scpm.getSentServiceChainTxsHashes()
	for _, peer := range scpm.getParentChainPeers().peers {
		peer.SendServiceChainReceiptRequest(hashes)
		scLogger.Debug("sent ServiceChainReceiptRequest", "peerID", peer.GetID(), "numReceiptsRequested", len(hashes))
	}
}

// SyncNonceAndGasPrice requests the nonce of address used for service chain tx to parent chain peers.
func (scpm *serviceChainPM) SyncNonceAndGasPrice() {
	addr := scpm.GetChainAddr()
	for _, peer := range scpm.getParentChainPeers().peers {
		peer.SendServiceChainInfoRequest(addr)
	}
}
