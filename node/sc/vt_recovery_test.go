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
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/accounts/abi/bind/backends"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/common/math"
	"github.com/ground-x/klaytn/contracts/bridge"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/params"
	"github.com/stretchr/testify/assert"
	"log"
	"math/big"
	"os"
	"path"
	"sync"
	"testing"
	"time"
)

type testInfo struct {
	sim        *backends.SimulatedBackend
	sc         *SubBridge
	bm         *BridgeManager
	localInfo  *BridgeInfo
	remoteInfo *BridgeInfo
	nodeAuth   *bind.TransactOpts
	chainAuth  *bind.TransactOpts
	aliceAuth  *bind.TransactOpts
	recoveryCh chan bool
}

const (
	testGasLimit     = 100000
	testAmount       = 321
	testTimeout      = 20 * time.Second
	testTxCount      = 7
	testBlockOffset  = 3 // +2 for genesis and bridge contract, +1 by a hardcoded hint
	testPendingCount = 3
	testNonceOffset  = testTxCount - testPendingCount
)

// TestBasicKLAYTransferRecovery tests each methods of the value transfer recovery.
func TestBasicKLAYTransferRecovery(t *testing.T) {
	defer func() {
		if err := os.Remove(path.Join(os.TempDir(), BridgeAddrJournal)); err != nil {
			t.Fatalf("fail to delete journal file %v", err)
		}
	}()

	// 1. Init dummy chain and do some value transfers.
	info := prepare(t, func(info *testInfo) {
		for i := 0; i < testTxCount; i++ {
			requestKLAYTransfer(info, info.localInfo.bridge, info.remoteInfo.bridge)
		}
	})
	vtr := NewValueTransferRecovery(&SCConfig{VTRecovery: true}, info.localInfo, info.remoteInfo, info.chainAuth)

	// 2. Get recovery hint.
	err := vtr.getRecoveryHint()
	if err != nil {
		t.Fatal("fail to get value transfer hint")
	}
	fmt.Println("value transfer hint", vtr.service2mainHint)
	assert.Equal(t, uint64(testTxCount-1), vtr.service2mainHint.requestNonce) // nonce begins at zero.
	assert.Equal(t, uint64(testTxCount-1-testPendingCount), vtr.service2mainHint.handleNonce)

	// 3. Request events by using the hint.
	err = vtr.getPendingEvents()
	if err != nil {
		t.Fatal("fail to get events from bridge contract")
	}

	// 4. Check pending events.
	fmt.Println("start to check pending tx", "len", len(vtr.serviceChainEvents))
	var count = 0
	for index, evt := range vtr.serviceChainEvents {
		assert.Equal(t, info.nodeAuth.From, evt.From)
		assert.Equal(t, info.aliceAuth.From, evt.To)
		assert.Equal(t, big.NewInt(testAmount), evt.Amount)
		assert.Equal(t, uint64(index+testNonceOffset), evt.RequestNonce)
		assert.Condition(t, func() bool {
			return uint64(testBlockOffset) <= evt.Raw.BlockNumber
		})
		count++
	}
	assert.Equal(t, testPendingCount, count)

	// 5. Recover pending transactions
	info.recoveryCh <- true
	assert.Equal(t, nil, vtr.recoverTransactions())
	dummyHandleRequestKLAYTransfer(&info, info.remoteInfo.bridge)
	info.sim.Commit()

	// 6. Check empty pending transactions.
	err = vtr.getRecoveryHint()
	if err != nil {
		t.Fatal("fail to get value transfer hint")
	}
	err = vtr.getPendingEvents()
	if err != nil {
		t.Fatal("fail to get events from bridge contract")
	}
	assert.Equal(t, 0, len(vtr.serviceChainEvents))

	assert.Equal(t, nil, vtr.Recover()) // nothing to recover
}

// TestBasicTokenTransferRecovery tests the token transfer recovery.
// TODO-Klaytn-ServiceChain: implement token transfer.
func TestBasicTokenTransferRecovery(t *testing.T) {
}

// TestBasicNFTTransferRecovery tests the NFT transfer recovery.
// TODO-Klaytn-ServiceChain: implement NFT transfer.
func TestBasicNFTTransferRecovery(t *testing.T) {
}

// TestMethodRecover tests the valueTransferRecovery.Recover() method.
func TestMethodRecover(t *testing.T) {
	defer func() {
		if err := os.Remove(path.Join(os.TempDir(), BridgeAddrJournal)); err != nil {
			t.Fatalf("fail to delete journal file %v", err)
		}
	}()

	info := prepare(t, func(info *testInfo) {
		for i := 0; i < testTxCount; i++ {
			requestKLAYTransfer(info, info.localInfo.bridge, info.remoteInfo.bridge)
		}
	})
	vtr := NewValueTransferRecovery(&SCConfig{VTRecovery: true}, info.localInfo, info.remoteInfo, info.chainAuth)
	err := vtr.getRecoveryHint()
	if err != nil {
		t.Fatal("fail to get a value transfer hint")
	}
	assert.NotEqual(t, vtr.service2mainHint.requestNonce, vtr.service2mainHint.handleNonce)

	info.recoveryCh <- true
	err = vtr.Recover()
	if err != nil {
		t.Fatal("fail to recover the value transfer")
	}
	dummyHandleRequestKLAYTransfer(&info, info.remoteInfo.bridge)
	info.sim.Commit()

	err = vtr.getRecoveryHint()
	if err != nil {
		t.Fatal("fail to get a value transfer hint")
	}
	assert.Equal(t, vtr.service2mainHint.requestNonce, vtr.service2mainHint.handleNonce)
}

// TestMethodStop tests the Stop method for stop the internal goroutine.
func TestMethodStop(t *testing.T) {
	defer func() {
		if err := os.Remove(path.Join(os.TempDir(), BridgeAddrJournal)); err != nil {
			t.Fatalf("fail to delete journal file %v", err)
		}
	}()

	info := prepare(t, func(info *testInfo) {
		for i := 0; i < testTxCount; i++ {
			requestKLAYTransfer(info, info.localInfo.bridge, info.remoteInfo.bridge)
		}
	})
	vtr := NewValueTransferRecovery(&SCConfig{VTRecovery: true, VTRecoveryInterval: 1}, info.localInfo, info.remoteInfo, info.chainAuth)
	err := vtr.getRecoveryHint()
	if err != nil {
		t.Fatal("fail to get a value transfer hint")
	}
	assert.NotEqual(t, vtr.service2mainHint.requestNonce, vtr.service2mainHint.handleNonce)

	info.recoveryCh <- true
	err = vtr.Start()
	if err != nil {
		t.Fatal("fail to start the value transfer")
	}
	assert.Equal(t, true, vtr.status)
	err = vtr.Stop()
	if err != nil {
		t.Fatal("fail to stop the value transfer")
	}
	assert.Equal(t, false, vtr.status)
}

// TestFlagVTRecovery tests the disabled vtrecovery option.
func TestFlagVTRecovery(t *testing.T) {
	defer func() {
		if err := os.Remove(path.Join(os.TempDir(), BridgeAddrJournal)); err != nil {
			t.Fatalf("fail to delete the journal file %v", err)
		}
	}()

	info := prepare(t, func(info *testInfo) {
		for i := 0; i < testTxCount; i++ {
			requestKLAYTransfer(info, info.localInfo.bridge, info.remoteInfo.bridge)
		}
	})
	vtr := NewValueTransferRecovery(&SCConfig{VTRecovery: true}, info.localInfo, info.remoteInfo, info.chainAuth)
	err := vtr.getRecoveryHint()
	if err != nil {
		t.Fatal("fail to get a value transfer hint")
	}
	assert.NotEqual(t, vtr.service2mainHint.requestNonce, vtr.service2mainHint.handleNonce)

	info.recoveryCh <- true
	vtr.config.VTRecovery = false
	err = vtr.Recover()
	if err != nil {
		t.Fatal("fail to recover the value transfer")
	}
	dummyHandleRequestKLAYTransfer(&info, info.remoteInfo.bridge)
	info.sim.Commit()

	err = vtr.getRecoveryHint()
	if err != nil {
		t.Fatal("fail to get a value transfer hint")
	}
	assert.NotEqual(t, vtr.service2mainHint.requestNonce, vtr.service2mainHint.handleNonce)
}

// TestScenarioMainChainRecovery tests the value transfer recovery of the main chain to service chain value transfers.
func TestScenarioMainChainRecovery(t *testing.T) {
	defer func() {
		if err := os.Remove(path.Join(os.TempDir(), BridgeAddrJournal)); err != nil {
			t.Fatalf("fail to delete journal file %v", err)
		}
	}()

	info := prepare(t, func(info *testInfo) {
		for i := 0; i < testTxCount; i++ {
			requestKLAYTransfer(info, info.remoteInfo.bridge, info.localInfo.bridge)
		}
	})
	vtr := NewValueTransferRecovery(&SCConfig{VTRecovery: true}, info.localInfo, info.remoteInfo, info.chainAuth)
	err := vtr.getRecoveryHint()
	if err != nil {
		t.Fatal("fail to get a value transfer hint")
	}
	assert.NotEqual(t, vtr.main2serviceHint.requestNonce, vtr.main2serviceHint.handleNonce)

	info.recoveryCh <- true
	err = vtr.Recover()
	if err != nil {
		t.Fatal("fail to recover the value transfer")
	}
	dummyHandleRequestKLAYTransfer(&info, info.localInfo.bridge)
	info.sim.Commit()

	err = vtr.getRecoveryHint()
	if err != nil {
		t.Fatal("fail to get a value transfer hint")
	}
	assert.Equal(t, vtr.main2serviceHint.requestNonce, vtr.main2serviceHint.handleNonce)
}

// TestScenarioAutomaticRecovery tests the recovery of the internal goroutine.
func TestScenarioAutomaticRecovery(t *testing.T) {
	defer func() {
		if err := os.Remove(path.Join(os.TempDir(), BridgeAddrJournal)); err != nil {
			t.Fatalf("fail to delete journal file %v", err)
		}
	}()

	info := prepare(t, func(info *testInfo) {
		for i := 0; i < testTxCount; i++ {
			requestKLAYTransfer(info, info.localInfo.bridge, info.remoteInfo.bridge)
		}
	})
	vtr := NewValueTransferRecovery(&SCConfig{VTRecovery: true, VTRecoveryInterval: 1}, info.localInfo, info.remoteInfo, info.chainAuth)
	err := vtr.getRecoveryHint()
	if err != nil {
		t.Fatal("fail to get a value transfer hint")
	}
	assert.NotEqual(t, vtr.service2mainHint.requestNonce, vtr.service2mainHint.handleNonce)

	info.recoveryCh <- true
	err = vtr.Start()
	if err != nil {
		t.Fatal("fail to start the value transfer")
	}
	dummyHandleRequestKLAYTransfer(&info, info.remoteInfo.bridge)
	info.sim.Commit()

	err = vtr.getRecoveryHint()
	if err != nil {
		t.Fatal("fail to get a value transfer hint")
	}
	assert.Equal(t, vtr.service2mainHint.requestNonce, vtr.service2mainHint.handleNonce)
}

// prepare generates dummy blocks for testing pending transactions.
func prepare(t *testing.T, vtcallback func(*testInfo)) testInfo {
	// Setup configuration.
	config := &SCConfig{}
	config.nodekey, _ = crypto.GenerateKey()
	config.chainkey, _ = crypto.GenerateKey()
	config.DataDir = os.TempDir()
	config.VTRecovery = true

	// Generate a new random account and a funded simulator.
	nodeAuth := bind.NewKeyedTransactor(config.nodekey)
	chainAuth := bind.NewKeyedTransactor(config.chainkey)
	aliceKey, _ := crypto.GenerateKey()
	aliceAuth := bind.NewKeyedTransactor(aliceKey)
	chainKeyAddr := crypto.PubkeyToAddress(config.chainkey.PublicKey)
	config.ServiceChainAccountAddr = &chainKeyAddr

	// Alloc genesis and create a simulator.
	alloc := blockchain.GenesisAlloc{
		nodeAuth.From:  {Balance: big.NewInt(params.KLAY)},
		chainAuth.From: {Balance: big.NewInt(params.KLAY)},
		aliceAuth.From: {Balance: big.NewInt(params.KLAY)},
	}
	sim := backends.NewSimulatedBackend(alloc)

	sc := &SubBridge{config: config, peers: newBridgePeerSet()}
	handler, err := NewSubBridgeHandler(sc.config, sc)
	if err != nil {
		log.Fatalf("Failed to initialize the bridgeHandler : %v", err)
		return testInfo{}
	}
	sc.handler = handler

	// Prepare manager and deploy bridge contract.
	bm, err := NewBridgeManager(sc)
	localAddr, err := bm.DeployBridgeTest(sim, false)
	if err != nil {
		t.Fatal("deploy bridge test failed", localAddr)
	}
	remoteAddr, err := bm.DeployBridgeTest(sim, false) // mimic remote by local
	if err != nil {
		t.Fatal("deploy bridge test failed", remoteAddr)
	}
	localInfo := bm.bridges[localAddr]
	remoteInfo := bm.bridges[remoteAddr]
	sim.Commit() // block

	// Subscribe events.
	err = bm.SubscribeEvent(localAddr)
	if err != nil {
		t.Fatalf("local bridge manager event subscription failed")
	}
	err = bm.SubscribeEvent(remoteAddr)
	if err != nil {
		t.Fatalf("remote bridge manager event subscription failed")
	}

	// Prepare channel for event handling.
	requestVTCh := make(chan TokenReceivedEvent)
	handleVTCh := make(chan TokenTransferEvent)
	recoveryCh := make(chan bool)
	bm.SubscribeTokenReceived(requestVTCh)
	bm.SubscribeTokenWithDraw(handleVTCh)

	info := testInfo{
		sim, sc, bm, localInfo, remoteInfo,
		nodeAuth, chainAuth, aliceAuth, recoveryCh,
	}

	// Start a event handling loop.
	wg := sync.WaitGroup{}
	wg.Add((2 * testTxCount) - testPendingCount)
	var isRecovery = false
	go func() {
		for {
			select {
			case ev := <-recoveryCh:
				isRecovery = ev
			case ev := <-requestVTCh:

				switch ev.TokenType {
				case 0:
					// Intentionally lost a single handle value transfer.
					// Since the increase of monotony in nonce is checked in the contract,
					// all subsequent handle transfer will be failed.
					if ev.RequestNonce == (testTxCount - testPendingCount) {
						fmt.Println("missing handle value transfer", "nonce", ev.RequestNonce)
						break
					}
					opts := &bind.TransactOpts{From: chainAuth.From, Signer: chainAuth.Signer, GasLimit: testGasLimit}
					_, err := info.remoteInfo.bridge.HandleKLAYTransfer(opts, ev.Amount, ev.To, ev.RequestNonce, ev.BlockNumber)
					if err != nil {
						log.Fatalf("\tFailed to HandleKLAYTransfer: %v", err)
					}
					sim.Commit()
				}
				if !isRecovery {
					wg.Done()
				}

			case _ = <-handleVTCh:
				if !isRecovery {
					wg.Done()
				}
			}
		}
	}()

	// Request value transfer.
	vtcallback(&info)

	WaitGroupWithTimeOut(&wg, testTimeout, t)

	return info
}

// dummyHandleValueTransfer emulates the HandleValueTransfer instead of the handleRequestValueTransferEvent().
// TODO-Klaytn-ServiceChain: use ChildChainEventHandler
func dummyHandleRequestKLAYTransfer(info *testInfo, bridge *bridge.Bridge) {
	info.localInfo.nextHandleNonce = math.MaxUint64 // set a large value to get all events
	events := info.localInfo.GetReadyRequestValueTransferEvents()
	for _, ev := range events {
		opts := &bind.TransactOpts{From: info.chainAuth.From, Signer: info.chainAuth.Signer, GasLimit: testGasLimit}
		_, err := bridge.HandleKLAYTransfer(opts, ev.Amount, ev.To, ev.RequestNonce, ev.BlockNumber)
		if err != nil {
			fmt.Println("dummy HandleValueTransfer failed", err, ev)
		}
	}
}

// valueTransfer requests the klay transfer transactions where 'from' and 'to' bridges are specified.
func requestKLAYTransfer(info *testInfo, from *bridge.Bridge, to *bridge.Bridge) {
	opts := &bind.TransactOpts{
		From:     info.nodeAuth.From,
		Signer:   info.nodeAuth.Signer,
		Value:    big.NewInt(testAmount),
		GasLimit: testGasLimit,
	}
	oldLocal, oldRemote := info.localInfo.bridge, info.remoteInfo.bridge
	info.localInfo.bridge, info.remoteInfo.bridge = from, to
	_, err := from.RequestKLAYTransfer(opts, info.aliceAuth.From)
	if err != nil {
		log.Fatalf("Failed to RequestKLAYTransfer: %v", err)
	}
	info.sim.Commit() // block
	info.localInfo.bridge, info.remoteInfo.bridge = oldLocal, oldRemote
}
