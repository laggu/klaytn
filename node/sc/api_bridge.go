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
	"github.com/ground-x/klaytn/contracts/bridge"
	"github.com/ground-x/klaytn/networks/p2p"
	"github.com/ground-x/klaytn/networks/p2p/discover"
	"github.com/ground-x/klaytn/node"
	"github.com/pkg/errors"
)

// MainBridgeAPI Implementation for main-bridge node
type MainBridgeAPI struct {
	sc *MainBridge
}

func (mbapi *MainBridgeAPI) GetChildChainIndexingEnabled() bool {
	return mbapi.sc.eventhandler.GetChildChainIndexingEnabled()
}

func (mbapi *MainBridgeAPI) ConvertChildChainBlockHashToParentChainTxHash(ccBlockHash common.Hash) common.Hash {
	return mbapi.sc.eventhandler.ConvertChildChainBlockHashToParentChainTxHash(ccBlockHash)
}

// Peers retrieves all the information we know about each individual peer at the
// protocol granularity.
func (mbapi *MainBridgeAPI) Peers() ([]*p2p.PeerInfo, error) {
	server := mbapi.sc.bridgeServer
	if server == nil {
		return nil, node.ErrNodeStopped
	}
	return server.PeersInfo(), nil
}

// NodeInfo retrieves all the information we know about the host node at the
// protocol granularity.
func (mbapi *MainBridgeAPI) NodeInfo() (*p2p.NodeInfo, error) {
	server := mbapi.sc.bridgeServer
	if server == nil {
		return nil, node.ErrNodeStopped
	}
	return server.NodeInfo(), nil
}

// SubBridgeAPI Implementation for sub-bridge node
type SubBridgeAPI struct {
	sc *SubBridge
}

func (sbapi *SubBridgeAPI) ConvertChildChainBlockHashToParentChainTxHash(ccBlockHash common.Hash) common.Hash {
	return sbapi.sc.eventhandler.ConvertChildChainBlockHashToParentChainTxHash(ccBlockHash)
}

func (sbapi *SubBridgeAPI) GetLatestAnchoredBlockNumber() uint64 {
	return sbapi.sc.handler.GetLatestAnchoredBlockNumber()
}

func (sbapi *SubBridgeAPI) GetReceiptFromParentChain(blockHash common.Hash) *types.Receipt {
	return sbapi.sc.handler.GetReceiptFromParentChain(blockHash)
}

func (sbapi *SubBridgeAPI) DeployBridge() ([]common.Address, error) {
	localAddr, err := sbapi.sc.bridgeManager.DeployBridge(sbapi.sc.localBackend, true)
	if err != nil {
		logger.Error("Failed to deploy scBridge.", "err", err)
		return nil, err
	}
	remoteAddr, err := sbapi.sc.bridgeManager.DeployBridge(sbapi.sc.remoteBackend, false)
	if err != nil {
		logger.Error("Failed to deploy mcBridge.", "err", err)
		return nil, err
	}

	err = sbapi.sc.bridgeManager.SetJournal(localAddr, remoteAddr)
	if err != nil {
		return nil, err
	}

	return []common.Address{localAddr, remoteAddr}, nil
}

// SubscribeEventBridge enables the given service/main chain bridges to subscribe the events.
func (sbapi *SubBridgeAPI) SubscribeEventBridge(cBridgeAddr, pBridgeAddr common.Address) error {
	err := sbapi.sc.AddressManager().AddBridge(cBridgeAddr, pBridgeAddr)
	if err != nil {
		return err
	}

	err = sbapi.sc.bridgeManager.SubscribeEvent(cBridgeAddr)
	if err != nil {
		logger.Error("Failed to SubscribeEventBridge Child Bridge", "addr", cBridgeAddr, "err", err)
		return err
	}

	err = sbapi.sc.bridgeManager.SubscribeEvent(pBridgeAddr)
	if err != nil {
		logger.Error("Failed to SubscribeEventBridge Parent Bridge", "addr", pBridgeAddr, "err", err)
		sbapi.sc.AddressManager().DeleteBridge(cBridgeAddr)
		sbapi.sc.bridgeManager.UnsubscribeEvent(cBridgeAddr)
		return err
	}

	// Update the journal's subscribed flag.
	sbapi.sc.bridgeManager.journal.rotate(sbapi.sc.bridgeManager.GetAllBridge())

	err = sbapi.sc.bridgeManager.AddRecovery(cBridgeAddr, pBridgeAddr)
	if err != nil {
		// TODO-Klaytn-ServiceChain: delete the journal
		sbapi.sc.AddressManager().DeleteBridge(cBridgeAddr)
		sbapi.sc.bridgeManager.UnsubscribeEvent(cBridgeAddr)
		sbapi.sc.bridgeManager.UnsubscribeEvent(pBridgeAddr)
		return err
	}
	return nil
}

// UnsubscribeEventBridge disables the event subscription of the given service/main chain bridges.
func (sbapi *SubBridgeAPI) UnsubscribeEventBridge(cBridgeAddr, pBridgeAddr common.Address) error {
	sbapi.sc.bridgeManager.UnsubscribeEvent(cBridgeAddr)
	sbapi.sc.bridgeManager.UnsubscribeEvent(pBridgeAddr)

	if _, _, err := sbapi.sc.AddressManager().DeleteBridge(cBridgeAddr); err != nil {
		return err
	}

	sbapi.sc.bridgeManager.journal.rotate(sbapi.sc.bridgeManager.GetAllBridge())
	return nil
}

func (sbapi *SubBridgeAPI) TxPendingCount() int {
	return len(sbapi.sc.GetBridgeTxPool().Pending())
}

func (sbapi *SubBridgeAPI) ListDeployedBridge() []*BridgeJournal {
	return sbapi.sc.bridgeManager.GetAllBridge()
}

func (sbapi *SubBridgeAPI) Anchoring(flag bool) bool {
	return sbapi.sc.SetAnchoringTx(flag)
}

func (sbapi *SubBridgeAPI) GetAnchoring() bool {
	return sbapi.sc.GetAnchoringTx()
}

func (sbapi *SubBridgeAPI) RegisterBridge(cBridgeAddr common.Address, pBridgeAddr common.Address) error {
	cBridge, err := bridge.NewBridge(cBridgeAddr, sbapi.sc.localBackend)
	if err != nil {
		return err
	}
	pBridge, err := bridge.NewBridge(pBridgeAddr, sbapi.sc.remoteBackend)
	if err != nil {
		return err
	}

	bm := sbapi.sc.bridgeManager
	err = bm.SetBridgeInfo(cBridgeAddr, cBridge, sbapi.sc.bridgeAccountManager.scAccount, true, false)
	if err != nil {
		return err
	}
	err = bm.SetBridgeInfo(pBridgeAddr, pBridge, sbapi.sc.bridgeAccountManager.mcAccount, false, false)
	if err != nil {
		bm.DeleteBridgeInfo(cBridgeAddr)
		return err
	}

	err = bm.SetJournal(cBridgeAddr, pBridgeAddr)
	if err != nil {
		return err
	}

	return nil
}

func (sbapi *SubBridgeAPI) DeregisterBridge(cBridgeAddr common.Address, pBridgeAddr common.Address) error {
	if sbapi.sc.AddressManager().GetCounterPartBridge(cBridgeAddr) != pBridgeAddr {
		return errors.New("invalid bridge pair")
	}

	_, _, err := sbapi.sc.AddressManager().DeleteBridge(cBridgeAddr)
	if err != nil {
		return err
	}

	// TODO-Klaytn Add removing journal and pair information
	return nil
}

func (sbapi *SubBridgeAPI) RegisterToken(cBridgeAddr, pBridgeAddr, cTokenAddr, pTokenAddr common.Address) error {
	cBi, cExist := sbapi.sc.bridgeManager.GetBridgeInfo(cBridgeAddr)
	pBi, pExist := sbapi.sc.bridgeManager.GetBridgeInfo(pBridgeAddr)

	if !cExist || !pExist {
		return errors.New("bridge does not exist")
	}

	if sbapi.sc.AddressManager().GetCounterPartBridge(cBridgeAddr) != pBridgeAddr {
		return errors.New("invalid bridge pair")
	}

	cBi.account.Lock()
	tx, err := cBi.bridge.RegisterToken(cBi.account.GetTransactOpts(), cTokenAddr, pTokenAddr)
	if err != nil {
		cBi.account.UnLock()
		return err
	}
	cBi.account.IncNonce()
	cBi.account.UnLock()
	logger.Debug("scBridge registered token", "txHash", tx.Hash().String(), "scToken", cTokenAddr.String(), "mcToken", pTokenAddr.String())

	pBi.account.Lock()
	tx, err = pBi.bridge.RegisterToken(pBi.account.GetTransactOpts(), pTokenAddr, cTokenAddr)
	if err != nil {
		pBi.account.UnLock()
		return err
	}
	pBi.account.IncNonce()
	pBi.account.UnLock()

	logger.Debug("mcBridge registered token", "txHash", tx.Hash().String(), "scToken", cTokenAddr.String(), "mcToken", pTokenAddr.String())

	if err := sbapi.sc.AddressManager().AddToken(cTokenAddr, pTokenAddr); err != nil {
		return err
	}
	logger.Info("Register token", "scToken", cTokenAddr.String(), "mcToken", pTokenAddr.String())
	return nil
}

func (sbapi *SubBridgeAPI) DeregisterToken(cBridgeAddr, pBridgeAddr, cTokenAddr, pTokenAddr common.Address) error {
	if sbapi.sc.AddressManager().GetCounterPartToken(cTokenAddr) != pTokenAddr {
		return errors.New("invalid toke pair")
	}

	cBi, cExist := sbapi.sc.bridgeManager.bridges[cBridgeAddr]
	pBi, pExist := sbapi.sc.bridgeManager.bridges[pBridgeAddr]

	if !cExist || !pExist {
		return errors.New("bridge does not exist")
	}

	cBi.account.Lock()
	defer cBi.account.UnLock()
	tx, err := cBi.bridge.DeregisterToken(cBi.account.GetTransactOpts(), cTokenAddr)
	if err != nil {
		return err
	}
	cBi.account.IncNonce()
	logger.Debug("scBridge deregistered token", "txHash", tx.Hash().String(), "scToken", cTokenAddr.String(), "mcToken", pTokenAddr.String())

	pBi.account.Lock()
	defer pBi.account.UnLock()
	tx, err = pBi.bridge.DeregisterToken(pBi.account.GetTransactOpts(), pTokenAddr)
	if err != nil {
		return err
	}
	pBi.account.IncNonce()
	logger.Debug("mcBridge deregistered token", "txHash", tx.Hash().String(), "scToken", cTokenAddr.String(), "mcToken", pTokenAddr.String())

	_, _, err = sbapi.sc.AddressManager().DeleteToken(cTokenAddr)
	return err
}

// AddPeer requests connecting to a remote node, and also maintaining the new
// connection at all times, even reconnecting if it is lost.
func (sbapi *SubBridgeAPI) AddPeer(url string) (bool, error) {
	// Make sure the server is running, fail otherwise
	server := sbapi.sc.bridgeServer
	if server == nil {
		return false, node.ErrNodeStopped
	}
	// TODO-Klaytn Refactoring this to check whether the url is valid or not by dialing and return it.
	if _, err := addPeerInternal(server, url); err != nil {
		return false, err
	} else {
		return true, nil
	}
}

// addPeerInternal does common part for AddPeer.
func addPeerInternal(server p2p.Server, url string) (*discover.Node, error) {
	// Try to add the url as a static peer and return
	node, err := discover.ParseNode(url)
	if err != nil {
		return nil, fmt.Errorf("invalid kni: %v", err)
	}
	server.AddPeer(node, false)
	return node, nil
}

// RemovePeer disconnects from a a remote node if the connection exists
func (sbapi *SubBridgeAPI) RemovePeer(url string) (bool, error) {
	// Make sure the server is running, fail otherwise
	server := sbapi.sc.bridgeServer
	if server == nil {
		return false, node.ErrNodeStopped
	}
	// Try to remove the url as a static peer and return
	node, err := discover.ParseNode(url)
	if err != nil {
		return false, fmt.Errorf("invalid kni: %v", err)
	}
	server.RemovePeer(node)
	return true, nil
}

// Peers retrieves all the information we know about each individual peer at the
// protocol granularity.
func (sbapi *SubBridgeAPI) Peers() ([]*p2p.PeerInfo, error) {
	server := sbapi.sc.bridgeServer
	if server == nil {
		return nil, node.ErrNodeStopped
	}
	return server.PeersInfo(), nil
}

// NodeInfo retrieves all the information we know about the host node at the
// protocol granularity.
func (sbapi *SubBridgeAPI) NodeInfo() (*p2p.NodeInfo, error) {
	server := sbapi.sc.bridgeServer
	if server == nil {
		return nil, node.ErrNodeStopped
	}
	return server.NodeInfo(), nil
}

func (sbapi *SubBridgeAPI) GetMainChainAccountAddr() string {
	return sbapi.sc.config.MainChainAccountAddr.Hex()
}

func (sbapi *SubBridgeAPI) GetMainChainAccountNonce() uint64 {
	return sbapi.sc.handler.getMainChainAccountNonce()
}

func (sbapi *SubBridgeAPI) GetServiceChainAccountAddr() string {
	return sbapi.sc.config.ServiceChainAccountAddr.Hex()
}

func (sbapi *SubBridgeAPI) GetAnchoringPeriod() uint64 {
	return sbapi.sc.config.AnchoringPeriod
}

func (sbapi *SubBridgeAPI) GetSentChainTxsLimit() uint64 {
	return sbapi.sc.config.SentChainTxsLimit
}
