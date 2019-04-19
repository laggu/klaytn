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
	sim          *backends.SimulatedBackend
	sc           *SubBridge
	bm           *BridgeManager
	localBridge  *bridge.Bridge
	remoteBridge *bridge.Bridge
	nodeAuth     *bind.TransactOpts
	aliceAuth    *bind.TransactOpts
}

const testGasLimit = 100000
const testAmount = 321
const testTimeout = 5 * time.Second
const testTxCount = 7
const testBlockOffset = 3 // +2 for genesis and bridge contract, +1 by a hardcoded hint
const testPendingCount = 2
const testNonceOffset = testTxCount - testPendingCount

// TestValueTransferRecovery tests value transfer recovery.
func TestValueTransferRecovery(t *testing.T) {
	defer func() {
		if err := os.Remove(path.Join(os.TempDir(), BridgeAddrJournal)); err != nil {
			t.Fatalf("fail to delete journal file %v", err)
		}
	}()

	// 1. Init dummy chain and value transfers
	info := prepare(t)
	vtr := NewValueTransferRecovery(&SCConfig{VTRecovery: true})

	// 2. Get recovery hint (currently hardcoded to block number 3)
	hint, err := vtr.getValueTransferHint(info.localBridge, info.remoteBridge)
	if err != nil {
		t.Fatal("fail to get request value transfer hint")
	}
	fmt.Println("value transfer hint", hint)
	assert.Equal(t, uint64(testTxCount-1), hint.requestNonce) // nonce begins at zero.
	assert.Equal(t, uint64(testTxCount-1-testPendingCount), hint.handleNonce)

	// 3. Request logs from the hint
	events, err := vtr.getPendingEvents(info.localBridge, hint)
	if err != nil {
		t.Fatal("fail to get logs from bridge contract")
	}

	// 4. Check pending transactions
	fmt.Println("start to check pending tx", "len", len(events))
	var count = 0
	for index, evt := range events {
		fmt.Println("\tPending Tx:", evt.Raw.TxHash.String())
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
	vtr.recoverTransactions()
	assert.Equal(t, true, vtr.done)
}

// prepare generates dummy blocks for testing pending transactions.
func prepare(t *testing.T) testInfo {
	// Setup configuration
	config := &SCConfig{}
	config.nodekey, _ = crypto.GenerateKey()
	config.chainkey, _ = crypto.GenerateKey()
	config.DataDir = os.TempDir()
	config.VTRecovery = true

	// Generate a new random account and a funded simulator
	nodeAuth := bind.NewKeyedTransactor(config.nodekey)
	chainAuth := bind.NewKeyedTransactor(config.chainkey)
	aliceKey, _ := crypto.GenerateKey()
	aliceAuth := bind.NewKeyedTransactor(aliceKey)
	chainKeyAddr := crypto.PubkeyToAddress(config.chainkey.PublicKey)
	config.ServiceChainAccountAddr = &chainKeyAddr

	// Alloc genesis and create a simulator
	alloc := blockchain.GenesisAlloc{
		nodeAuth.From:  {Balance: big.NewInt(params.KLAY)},
		chainAuth.From: {Balance: big.NewInt(params.KLAY)},
		aliceAuth.From: {Balance: big.NewInt(params.KLAY)},
	}
	sim := backends.NewSimulatedBackend(alloc)

	sc := &SubBridge{config: config, peers: newBridgePeerSet()}
	handler, err := NewSubBridgeHandler(sc.config, sc)
	if err != nil {
		log.Fatalf("Failed to initialize bridgeHandler : %v", err)
		return testInfo{}
	}
	sc.handler = handler

	// Prepare manager and deploy bridge contract
	bm, err := NewBridgeManager(sc)
	localAddr, err := bm.DeployBridgeTest(sim, false)
	if err != nil {
		t.Fatal("deploy bridge test failed", localAddr)
	}
	remoteAddr, err := bm.DeployBridgeTest(sim, false) // mimic remote by local
	if err != nil {
		t.Fatal("deploy bridge test failed", remoteAddr)
	}
	localBridge := bm.bridges[localAddr].bridge
	remoteBridge := bm.bridges[remoteAddr].bridge
	fmt.Println("BridgeContract", localAddr.Hex())
	fmt.Println("BridgeContract", remoteAddr.Hex())
	sim.Commit() // block

	// Subscribe events
	err = bm.SubscribeEvent(localAddr)
	if err != nil {
		t.Fatalf("local bridge manager event subscription failed")
	}
	err = bm.SubscribeEvent(remoteAddr)
	if err != nil {
		t.Fatalf("remote bridge manager event subscription failed")
	}
	info := testInfo{sim, sc, bm, localBridge, remoteBridge, nodeAuth, aliceAuth}

	// Prepare channel for event handling
	tokenCh := make(chan TokenReceivedEvent)
	tokenSendCh := make(chan TokenTransferEvent)
	bm.SubscribeTokenReceived(tokenCh)
	bm.SubscribeTokenWithDraw(tokenSendCh)

	wg := sync.WaitGroup{}
	wg.Add((2 * testTxCount) - testPendingCount)

	// Start a event handling loop
	go func() {
		for {
			select {
			case ev := <-tokenCh:
				fmt.Println("\tTokenReceivedEvent", ev.ContractAddr.String())

				switch ev.TokenType {
				case 0:
					// Intentionally lost a single handle value transfer.
					// Since the increase of monotony in nonce is checked in the contract,
					// all subsequent handle transfer will be failed.
					if ev.RequestNonce == (testTxCount - testPendingCount) {
						fmt.Println("missing handle value transfer", "nonce", ev.RequestNonce)
						break
					}
					assert.Equal(t, localAddr, ev.ContractAddr)
					opts := &bind.TransactOpts{From: chainAuth.From, Signer: chainAuth.Signer, GasLimit: testGasLimit}
					_, err := remoteBridge.HandleKLAYTransfer(opts, ev.Amount, ev.To, ev.RequestNonce, ev.BlockNumber)
					if err != nil {
						log.Fatalf("\tFailed to HandleKLAYTransfer: %v", err)
					}
					sim.Commit() // block
				}
				wg.Done()

			case ev := <-tokenSendCh:
				assert.Equal(t, remoteAddr, ev.ContractAddr)
				fmt.Println("\tTokenTransferEvent", ev.ContractAddr.Hex())
				wg.Done()
			}
		}
	}()

	// Request value transfer
	for i := 0; i < testTxCount; i++ {
		valueTransfer(info)
	}

	WaitGroupWithTimeOut(&wg, testTimeout, t)

	return info
}

// valueTransfer requests a klay transfer transaction.
func valueTransfer(info testInfo) {
	opts := &bind.TransactOpts{
		From:     info.nodeAuth.From,
		Signer:   info.nodeAuth.Signer,
		Value:    big.NewInt(testAmount),
		GasLimit: testGasLimit,
	}
	_, err := info.localBridge.RequestKLAYTransfer(opts, info.aliceAuth.From)
	if err != nil {
		log.Fatalf("Failed to RequestKLAYTransfer: %v", err)
	}
	info.sim.Commit() // block
}
