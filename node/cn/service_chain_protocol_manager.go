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
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/networks/p2p"
)

type ServiceChainProtocolManager interface {
	// TODO-GX-ServiceChain Please note that ServiceChainProtocolManager is made to separate its implementation
	// and message handling from pre-existing ProtocolManager. This can be done by pushing implementation into
	// a new peer type and handling message by a new peer type.
	// private functions
	getParentChainPeers() *peerSet // getParentChainPeers returns peers on the parent chain.
	getChildChainPeers() *peerSet  // getChildChainPeers returns peers on the child chain.
	removeParentPeer(id string)    // removeParentPeer removes a parent peer with given id.
	removeChildPeer(id string)     // removeChildPeer removes a child peer with given id.

	getRemoteNonce() uint64         // getRemoteNonce returns the nonce of address used for service chain tx.
	setRemoteNonce(newNonce uint64) // setRemoteNonce sets the nonce of address used for service chain tx.
	getNonceSynced() bool           // getNonceSynced returns whether the nonce is synced or not.
	setNonceSynced(synced bool)     // setNonceSynced sets whether the nonce is synced or not.

	getChainKey() *ecdsa.PrivateKey                 // getChainKey returns the private key used for signing service chain tx.
	addToSentServiceChainTxs(tx *types.Transaction) // addToSentServiceChainTxs adds a transaction to SentServiceChainTxs.
	removeServiceChainTx(txHash common.Hash)        // removeServiceChainTx removes a transaction from SentServiceChainTxs with the given transaction hash.

	// public functions
	GetChainAddr() *common.Address // GetChainAddr returns a pointer of a hex address of an account used for service chain.
	GetChainTxPeriod() uint64      // GetChainTxPeriod returns the period (in child chain blocks) of sending chain transactions.
	GetSentChainTxsLimit() uint64  // GetSentChainTxsLimit returns the maximum number of stored chain transactions for resending.
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

	remoteNonce         uint64
	nonceSynced         bool
	chainTxPeriod       uint64
	sentServiceChainTxs map[common.Hash]*types.Transaction

	// TODO-GX-ServiceChain Need to limit the number independently? Or just managing the size of sentServiceChainTxs?
	sentServiceChainTxsLimit uint64
}

// ServiceChainConfig handles service chain configurations.
type ServiceChainConfig struct {
	ChainAddr   *common.Address
	ChainKey    *ecdsa.PrivateKey
	TxPeriod    uint64
	SentTxLimit uint64
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
	logger.Debug("Removing parent chain peer", "peer", id)
	// Unregister the peer from the downloader and parent chain peer set
	if err := scpm.getParentChainPeers().Unregister(id); err != nil {
		logger.Error("Parent chain peer removal failed", "peer", id, "err", err)
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
	logger.Debug("Removing child chain peer", "peer", id)
	// Unregister the peer from the downloader and child chain peer set
	if err := scpm.getChildChainPeers().Unregister(id); err != nil {
		logger.Error("Child chain peer removal failed", "peer", id, "err", err)
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
		logger.Warn("Number of txs in sentServiceChainTxs already exceeds the limit", "sentServiceChainTxsLimit", scpm.sentServiceChainTxsLimit)
		return
	}
	if _, ok := scpm.sentServiceChainTxs[tx.Hash()]; ok {
		logger.Error("ServiceChainTx already exists in sentServiceChainTxs", "txHash", tx.Hash())
		return
	}
	scpm.sentServiceChainTxs[tx.Hash()] = tx
}

// removeServiceChainTx removes a transaction from SentServiceChainTxs with the given
// transaction hash.
func (scpm *serviceChainPM) removeServiceChainTx(txHash common.Hash) {
	if _, ok := scpm.sentServiceChainTxs[txHash]; !ok {
		logger.Error("ServiceChainTx does not exists in sentServiceChainTxs", "txHash", txHash)
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
