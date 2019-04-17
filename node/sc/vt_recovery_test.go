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
	"github.com/ground-x/klaytn/common"
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
	sim       *backends.SimulatedBackend
	sc        *SubBridge
	bm        *BridgeManager
	bridge    *bridge.Bridge
	nodeAuth  *bind.TransactOpts
	aliceAuth *bind.TransactOpts
	contract  common.Address
}

const testGasLimit = 100000
const testTimeout = 5 * time.Second
const testAmount = 321
const testTxCount = 7
const testBlockOffset = 3 // +2 for genesis and deploy, +1 by hardcoded hint
const testNonceOffset = 1 // +1 by hardcoded hint

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
	hint := vtr.getRecoveryHint(info.contract)

	// 3. Request logs from the hint
	events, err := vtr.getPendingEvents(info.bridge, hint)
	if err != nil {
		t.Fatal("fail to get logs from bridge contract")
	}

	// 4. Confirm pending transactions
	fmt.Println("start to check pending tx")
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
	assert.Equal(t, testTxCount-1, count)

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
	config.ChainAccountAddr = &chainKeyAddr

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
	addr, err := bm.DeployBridgeTest(sim, false)
	br := bm.bridges[addr].bridge
	fmt.Println("BridgeContract", addr.Hex())
	sim.Commit() // block
	err = bm.SubscribeEvent(addr)
	if err != nil {
		t.Fatalf("bridge manager event subscription failed")
	}
	info := testInfo{sim, sc, bm, br, nodeAuth, aliceAuth, addr}

	// Prepare channel for event handling
	tokenCh := make(chan TokenReceivedEvent)
	tokenSendCh := make(chan TokenTransferEvent)
	bm.SubscribeTokenReceived(tokenCh)
	bm.SubscribeTokenWithDraw(tokenSendCh)

	wg := sync.WaitGroup{}
	wg.Add(2 * testTxCount)

	// Start a event handling loop
	go func() {
		for {
			select {
			case ev := <-tokenCh:
				fmt.Println("\treceive TokenReceivedEvent", ev.ContractAddr.String())

				switch ev.TokenType {
				case 0:
					assert.Equal(t, addr, ev.ContractAddr)
					opts := &bind.TransactOpts{From: chainAuth.From, Signer: chainAuth.Signer, GasLimit: testGasLimit}
					tx, err := br.HandleKLAYTransfer(opts, ev.Amount, ev.To, ev.RequestNonce)
					if err != nil {
						log.Fatalf("\tFailed to HandleKLAYTransfer: %v", err)
					}
					fmt.Println("\tHandleKLAYTransfer Tx", tx.Hash().Hex())
					sim.Commit() // block
				}
				wg.Done()

			case ev := <-tokenSendCh:
				assert.Equal(t, addr, ev.ContractAddr)
				fmt.Println("\treceive TokenTransferEvent", ev.ContractAddr.Hex())
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
	tx, err := info.bridge.RequestKLAYTransfer(opts, info.aliceAuth.From)
	if err != nil {
		log.Fatalf("Failed to RequestKLAYTransfer: %v", err)
	}
	fmt.Println("\tRequestKLAYTransfer Tx", tx.Hash().Hex())
	info.sim.Commit() // block
}
