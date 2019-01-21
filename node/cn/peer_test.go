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
	"fmt"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/networks/p2p"
	"github.com/ground-x/klaytn/networks/p2p/discover"
	"gopkg.in/fatih/set.v0"
	"math/big"
	"math/rand"
	"strings"
	"testing"
)

func randomID() (id discover.NodeID) {
	for i := range id {
		id[i] = byte(rand.Intn(255))
	}
	return id
}

type testPeer struct {
	basePeer
}

func (p *testPeer) isOnTheSameChain(status *statusData, network uint64, genesis common.Hash, chainID *big.Int) (bool, error) {
	result, err := p.basePeer.isOnTheSameChain(status, network, genesis, chainID)
	return result, err
}

func newTestPeer(onParentChain bool) *testPeer {
	id := randomID()
	p := p2p.NewPeer(id, "", nil)

	peer := &testPeer{
		basePeer: basePeer{
			Peer:        p,
			rw:          nil,
			version:     klay62,
			id:          fmt.Sprintf("%x", id[:8]),
			knownTxs:    set.New(),
			knownBlocks: set.New(),
			queuedTxs:   make(chan []*types.Transaction, maxQueuedTxs),
			queuedProps: make(chan *propEvent, maxQueuedProps),
			queuedAnns:  make(chan *types.Block, maxQueuedAnns),
			term:        make(chan struct{}),
		},
	}

	peer.SetOnParentChain(onParentChain)
	return peer
}

type testCase struct {
	//Input
	peer   *testPeer
	status statusData

	networkID uint64
	genesis   common.Hash
	chainID   *big.Int

	//Output
	expectedOnSameChain bool
	expectedErr         string
}

func sameError(errType string, err error) bool {
	if errType == "" && err == nil {
		return true
	}
	return strings.HasPrefix(err.Error(), errType)
}

func createTestCase(sameChain bool, statusOnChildPeer bool, onParentChain bool, invalidNetworkID bool, invalidProtocolVersion bool, wantOnSameChain bool, wantErr string) *testCase {
	var myNetworkID, peerNetworkID uint64
	myNetworkID = (uint64)(1000)
	peerNetworkID = myNetworkID
	if invalidNetworkID {
		peerNetworkID = (uint64)(2000)
	}

	TD := big.NewInt(100)
	parentPeer := newTestPeer(true)
	notParentPeer := newTestPeer(false)
	dummyHash := common.HexToHash("1")

	GenesisA := common.HexToHash("1111")
	GenesisB := common.HexToHash("2222")
	ChainIDA := big.NewInt(1111)
	ChainIDB := big.NewInt(2222)

	var peerGenesis, myGenesis common.Hash
	var peerChainID, myChainID *big.Int

	peerGenesis = GenesisA
	peerChainID = ChainIDA

	if sameChain {
		myGenesis = peerGenesis
		myChainID = peerChainID
	} else {
		myGenesis = GenesisB
		myChainID = ChainIDB
	}

	var peer *testPeer

	if onParentChain {
		peer = parentPeer
	} else {
		peer = notParentPeer
	}

	peerProtocolVersion := peer.version
	if invalidProtocolVersion {
		peerProtocolVersion = peer.version + 1
	}

	status := statusData{
		(uint32)(peerProtocolVersion),
		peerNetworkID,
		TD,
		dummyHash,
		peerGenesis,
		peerChainID,
		false,
	}

	status.OnChildChain = statusOnChildPeer

	return &testCase{
		// peer are on the same chain. And the peer is added by admin.addPeer() => No Error
		peer:                peer,
		status:              status,
		networkID:           myNetworkID,
		genesis:             myGenesis,
		chainID:             myChainID,
		expectedOnSameChain: wantOnSameChain,
		expectedErr:         wantErr,
	}
}

func TestHandShakePeers(t *testing.T) {
	// Check all combinations of handshake depends on chain genesis/ID, hierarchy which the node know, hierarchy received
	tests := []*testCase{
		createTestCase(true, true, true, false, false, true, errorToString[ErrInvalidPeerHierarchy]),
		createTestCase(true, true, false, false, false, true, errorToString[ErrInvalidPeerHierarchy]),
		createTestCase(true, false, true, false, false, true, errorToString[ErrInvalidPeerHierarchy]),
		createTestCase(true, false, false, false, false, true, ""),

		createTestCase(false, true, true, false, false, false, errorToString[ErrInvalidPeerHierarchy]),
		createTestCase(false, true, false, false, false, false, ""),
		createTestCase(false, false, true, false, false, false, ""),
		createTestCase(false, false, false, false, false, false, errorToString[ErrInvalidPeerHierarchy]),

		createTestCase(true, false, false, true, false, true, errorToString[ErrNetworkIdMismatch]),
		createTestCase(true, false, false, false, true, true, errorToString[ErrProtocolVersionMismatch]),
	}

	for idx, test := range tests {
		onTheSameChain, err := test.peer.isOnTheSameChain(&test.status, test.networkID, test.genesis, test.chainID)

		if onTheSameChain != test.expectedOnSameChain || !sameError(test.expectedErr, err) {
			t.Fatalf("TC#%v : wrong result => Result(OnSameChain:%v, err:%v)",
				idx, onTheSameChain, err)
		}

		t.Logf("TC #%v is passed. err: %v", idx, err)
	}
}
