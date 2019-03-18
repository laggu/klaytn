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
	"github.com/ground-x/klaytn/contracts/gateway"
	"github.com/ground-x/klaytn/networks/p2p"
	"github.com/ground-x/klaytn/networks/p2p/discover"
	"github.com/ground-x/klaytn/node"
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

func (mbapi *MainBridgeAPI) GetLatestAnchoredBlockNumber() uint64 {
	return mbapi.sc.eventhandler.GetLatestAnchoredBlockNumber()
}

func (mbapi *MainBridgeAPI) GetReceiptFromParentChain(blockHash common.Hash) *types.Receipt {
	return mbapi.sc.eventhandler.GetReceiptFromParentChain(blockHash)
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

func (mbapi *MainBridgeAPI) GetAnchoringPeriod() uint64 {
	return mbapi.sc.config.AnchoringPeriod
}

func (mbapi *MainBridgeAPI) GetSentChainTxsLimit() uint64 {
	return mbapi.sc.config.SentChainTxsLimit
}

// SubBridgeAPI Implementation for sub-bridge node
type SubBridgeAPI struct {
	sc *SubBridge
}

func (sbapi *SubBridgeAPI) GetChildChainIndexingEnabled() bool {
	return sbapi.sc.eventhandler.GetChildChainIndexingEnabled()
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

func (sbapi *SubBridgeAPI) DeployGateway() ([]common.Address, error) {
	localAddr, err := sbapi.sc.gatewayMgr.DeployGateway(sbapi.sc.localBackend, true)
	if err != nil {
		return nil, err
	}
	remoteAddr, err := sbapi.sc.gatewayMgr.DeployGateway(sbapi.sc.remoteBackend, false)
	if err != nil {
		return nil, err
	}
	return []common.Address{localAddr, remoteAddr}, nil
}

func (sbapi *SubBridgeAPI) DeployGatewayOnLocalChain() (common.Address, error) {
	return sbapi.sc.gatewayMgr.DeployGateway(sbapi.sc.localBackend, true)
}

func (sbapi *SubBridgeAPI) DeployGatewayOnParentChain() (common.Address, error) {
	return sbapi.sc.gatewayMgr.DeployGateway(sbapi.sc.remoteBackend, false)
}

// TODO-Klaytn needs to make unSubscribe() method and enable user can unSubscribeEvent.
func (sbapi *SubBridgeAPI) SubscribeEventGateway(cGatewayAddr common.Address, pGatewayAddr common.Address) error {
	cErr := sbapi.sc.gatewayMgr.SubscribeEvent(cGatewayAddr)
	if cErr != nil {
		return cErr
	}

	pErr := sbapi.sc.gatewayMgr.SubscribeEvent(pGatewayAddr)
	if pErr != nil {
		return pErr
	}
	// TODO-Klaytn needs to make unSubscribe() method and deal with the exception case.

	sbapi.sc.AddressManager().AddGateway(cGatewayAddr, pGatewayAddr)
	return nil
}

func (sbapi *SubBridgeAPI) TxPendingCount() int {
	return len(sbapi.sc.GetBridgeTxPool().Pending())
}

func (sbapi *SubBridgeAPI) ListDeployedGateway() []*GateWayJournal {
	return sbapi.sc.gatewayMgr.GetAllGateway()
}

func (sbapi *SubBridgeAPI) Anchoring(flag bool) bool {
	return sbapi.sc.SetAnchoringTx(flag)
}

func (sbapi *SubBridgeAPI) GetAnchoring() bool {
	return sbapi.sc.GetAnchoringTx()
}

func (sbapi *SubBridgeAPI) RegisterGateway(cGatewayAddr common.Address, pGatewayAddr common.Address) bool {
	cGateway, cErr := gateway.NewGateway(cGatewayAddr, sbapi.sc.localBackend)
	pGateway, pErr := gateway.NewGateway(cGatewayAddr, sbapi.sc.remoteBackend)

	if cErr != nil || pErr != nil {
		return false
	}

	sbapi.sc.gatewayMgr.localGateWays[cGatewayAddr] = cGateway
	sbapi.sc.gatewayMgr.all[cGatewayAddr] = true

	sbapi.sc.gatewayMgr.remoteGateWays[pGatewayAddr] = pGateway
	sbapi.sc.gatewayMgr.all[pGatewayAddr] = false

	return true
}

func (sbapi *SubBridgeAPI) UnRegisterGateway(gateway common.Address) {
	sbapi.sc.AddressManager().DeleteGateway(gateway)
}

func (sbapi *SubBridgeAPI) RegisterToken(token1 common.Address, token2 common.Address) {
	sbapi.sc.AddressManager().AddToken(token1, token2)
}

func (sbapi *SubBridgeAPI) UnRegisterToken(token common.Address) {
	sbapi.sc.AddressManager().DeleteToken(token)
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

func (mbapi *SubBridgeAPI) GetChainAccountAddr() string {
	return mbapi.sc.config.ChainAccountAddr.Hex()
}

func (mbapi *SubBridgeAPI) GetChainAccountNonce() uint64 {
	return mbapi.sc.handler.getChainAccountNonce()
}

func (mbapi *SubBridgeAPI) GetNodeAccountAddr() string {
	return mbapi.sc.config.NodeAccountAddr.Hex()
}

func (mbapi *SubBridgeAPI) GetAnchoringPeriod() uint64 {
	return mbapi.sc.config.AnchoringPeriod
}

func (mbapi *SubBridgeAPI) GetSentChainTxsLimit() uint64 {
	return mbapi.sc.config.SentChainTxsLimit
}
