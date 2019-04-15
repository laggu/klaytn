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

package contracts

import (
	"context"
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/accounts/abi/bind/backends"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/contracts/gateway"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/params"
	"github.com/stretchr/testify/assert"
	"log"
	"math/big"
	"testing"
	"time"
)

const (
	gasLimit uint64 = 100000          // gasLimit for contract transaction.
	timeOut         = 3 * time.Second // timeout of context and event loop for simulated backend.
)

// WaitMined waits the tx receipt until the timeout.
func WaitMined(tx *types.Transaction, backend bind.DeployBackend, t *testing.T) error {
	timeoutContext, cancelTimeout := context.WithTimeout(context.Background(), timeOut)
	defer cancelTimeout()

	receipt, err := bind.WaitMined(timeoutContext, backend, tx)
	if err != nil {
		t.Fatal("Failed to WaitMined.", "err", err, "txHash", tx.Hash().String(), "status", receipt.Status)
		return err
	}
	return nil
}

// TransferSignedTx sends the transaction to transfer KLAY from auth to `to` and waits the execution of the transaction.
func TransferSignedTx(auth *bind.TransactOpts, backend *backends.SimulatedBackend, to common.Address, value *big.Int, t *testing.T) (common.Hash, *big.Int, error) {
	ctx := context.Background()

	nonce, err := backend.NonceAt(ctx, auth.From, nil)
	assert.Equal(t, err, nil)

	chainID, err := backend.ChainID(ctx)
	assert.Equal(t, err, nil)

	gasPrice, err := backend.SuggestGasPrice(ctx)
	assert.Equal(t, err, nil)

	tx := types.NewTransaction(
		nonce,
		to,
		value,
		gasLimit,
		gasPrice,
		nil)

	signedTx, err := auth.Signer(types.NewEIP155Signer(chainID), auth.From, tx)
	assert.Equal(t, err, nil)

	fee := big.NewInt(0)

	err = backend.SendTransaction(ctx, signedTx)
	assert.Equal(t, err, nil)

	backend.Commit()

	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	receipt, err := bind.WaitMined(ctx, backend, signedTx)
	if err != nil {
		log.Fatalf("WaitMined time out %v", err)
	}

	fee.Mul(big.NewInt(int64(receipt.GasUsed)), gasPrice)

	return tx.Hash(), fee, nil
}

// RequestKLAYTransfer send a requestValueTransfer transaction to the bridge contract.
func RequestKLAYTransfer(bridge *gateway.Gateway, auth *bind.TransactOpts, to common.Address, value uint64, t *testing.T) {
	_, err := bridge.DepositKLAY(&bind.TransactOpts{From: auth.From, Signer: auth.Signer, GasLimit: gasLimit, Value: big.NewInt(1)}, to)
	if err != nil {
		t.Fatalf("fail to DepositKLAY %v", err)
	}
}

// SendHandleKLAYTransfer send a handleValueTransfer transaction to the bridge contract.
func SendHandleKLAYTransfer(bridge *gateway.Gateway, auth *bind.TransactOpts, to common.Address, value uint64, nonce uint64, t *testing.T) *types.Transaction {
	tx, err := bridge.WithdrawKLAY(&bind.TransactOpts{From: auth.From, Signer: auth.Signer, GasLimit: gasLimit}, big.NewInt(int64(value)), to, nonce)
	if err != nil {
		t.Fatalf("fail to WithdrawKLAY %v", err)
		return nil
	}
	return tx
}

// TestBridgeDeployWithKLAY checks to the state/contract balance of the bridge deployed.
func TestBridgeDeployWithKLAY(t *testing.T) {
	bridgeAccountKey, _ := crypto.GenerateKey()
	bridgeAccount := bind.NewKeyedTransactor(bridgeAccountKey)

	alloc := blockchain.GenesisAlloc{bridgeAccount.From: {Balance: big.NewInt(params.KLAY)}}
	backend := backends.NewSimulatedBackend(alloc)

	chargeAmount := big.NewInt(10000000)
	bridgeAccount.Value = chargeAmount
	bridgeAddress, tx, bridge, err := gateway.DeployGateway(bridgeAccount, backend, true)
	if err != nil {
		t.Fatalf("fail to DeployGateway %v", err)
	}
	backend.Commit()
	WaitMined(tx, backend, t)

	balanceContract, err := bridge.GetKLAY(nil)
	if err != nil {
		t.Fatalf("fail to GetKLAY %v", err)
	}

	balanceState, err := backend.BalanceAt(context.Background(), bridgeAddress, nil)
	if err != nil {
		t.Fatal("failed to BalanceAt")
	}

	assert.Equal(t, chargeAmount, balanceState)
	assert.Equal(t, chargeAmount, balanceContract)
}

// TestBridgeRequestValueTransferNonce checks the bridge emit events with serialized nonce.
func TestBridgeRequestValueTransferNonce(t *testing.T) {
	bridgeAccountKey, _ := crypto.GenerateKey()
	bridgeAccount := bind.NewKeyedTransactor(bridgeAccountKey)

	testAccKey, _ := crypto.GenerateKey()
	testAcc := bind.NewKeyedTransactor(testAccKey)

	alloc := blockchain.GenesisAlloc{bridgeAccount.From: {Balance: big.NewInt(params.KLAY)}}
	backend := backends.NewSimulatedBackend(alloc)

	chargeAmount := big.NewInt(10000000)
	bridgeAccount.Value = chargeAmount
	addr, tx, bridge, err := gateway.DeployGateway(bridgeAccount, backend, true)
	if err != nil {
		t.Fatalf("fail to DeployGateway %v", err)
	}
	backend.Commit()
	WaitMined(tx, backend, t)
	t.Log("1. Bridge is deployed.", "addr=", addr.String(), "txHash=", tx.Hash().String())

	requestValueTransferEventCh := make(chan *gateway.GatewayTokenReceived, 100)
	requestSub, err := bridge.WatchTokenReceived(nil, requestValueTransferEventCh)
	defer requestSub.Unsubscribe()
	if err != nil {
		t.Fatalf("fail to DepositKLAY %v", err)
	}
	t.Log("2. Bridge is subscribed.")

	RequestKLAYTransfer(bridge, bridgeAccount, testAcc.From, 1, t)
	backend.Commit()

	expectedNonce := uint64(0)

loop:
	for {
		select {
		case ev := <-requestValueTransferEventCh:
			assert.Equal(t, expectedNonce, ev.ReqeustNonce)

			if expectedNonce == 1000 {
				return
			}
			expectedNonce++

			// TODO-Klaytn added more request token/NFT transfer cases,
			RequestKLAYTransfer(bridge, bridgeAccount, testAcc.From, 1, t)
			backend.Commit()

		case err := <-requestSub.Err():
			t.Log("Contract Event Loop Running Stop by requestSub.Err()", "err", err)
			break loop

		case <-time.After(timeOut):
			t.Log("Contract Event Loop Running Stop by timeout")
			break loop
		}
	}

	t.Fatal("fail to check monotone increasing nonce", "lastNonce", expectedNonce)
}

// TestBridgeHandleValueTransferNonce checks the bridge allow the handle value transfer tx with only serialized nonce.
func TestBridgeHandleValueTransferNonce(t *testing.T) {
	bridgeAccountKey, _ := crypto.GenerateKey()
	bridgeAccount := bind.NewKeyedTransactor(bridgeAccountKey)

	testAccKey, _ := crypto.GenerateKey()
	testAcc := bind.NewKeyedTransactor(testAccKey)

	alloc := blockchain.GenesisAlloc{bridgeAccount.From: {Balance: big.NewInt(params.KLAY)}}
	backend := backends.NewSimulatedBackend(alloc)

	chargeAmount := big.NewInt(10000000)
	bridgeAccount.Value = chargeAmount
	bridgeAddress, tx, bridge, err := gateway.DeployGateway(bridgeAccount, backend, true)
	if err != nil {
		t.Fatalf("fail to DeployGateway %v", err)
	}
	backend.Commit()
	WaitMined(tx, backend, t)
	t.Log("1. Bridge is deployed.", "bridgeAddress=", bridgeAddress.String(), "txHash=", tx.Hash().String())

	// TODO-Klaytn This routine should be removed. It is temporary code for the bug of bridge contract.
	TransferSignedTx(bridgeAccount, backend, bridgeAddress, chargeAmount, t)

	handleValueTransferEventCh := make(chan *gateway.GatewayTokenWithdrawn, 100)
	handleSub, err := bridge.WatchTokenWithdrawn(nil, handleValueTransferEventCh)
	defer handleSub.Unsubscribe()
	if err != nil {
		t.Fatalf("fail to DepositKLAY %v", err)
	}
	t.Log("2. Bridge is subscribed.")

	sentNonce := uint64(0)
	testCount := uint64(1000)
	transferAmount := uint64(100)
	tx = SendHandleKLAYTransfer(bridge, bridgeAccount, testAcc.From, transferAmount, sentNonce, t)
	backend.Commit()

	timeoutContext, cancelTimeout := context.WithTimeout(context.Background(), timeOut)
	defer cancelTimeout()

	receipt, err := bind.WaitMined(timeoutContext, backend, tx)

	if err != nil {
		t.Fatal("Failed to WaitMined.", "err", err, "txHash", tx.Hash().String(), "status", receipt.Status)
	}

loop:
	for {
		select {
		case ev := <-handleValueTransferEventCh:
			assert.Equal(t, sentNonce, ev.HandleNonce)

			if sentNonce == testCount {
				bal, err := backend.BalanceAt(context.Background(), testAcc.From, nil)
				assert.Equal(t, err, nil)

				assert.Equal(t, bal, big.NewInt(int64(transferAmount*(testCount+1))))
				return
			}
			sentNonce++

			// fail case : smaller nonce
			SendHandleKLAYTransfer(bridge, bridgeAccount, testAcc.From, transferAmount, sentNonce+1, t)

			// fail case : bigger nonce
			SendHandleKLAYTransfer(bridge, bridgeAccount, testAcc.From, transferAmount, sentNonce-1, t)

			// success case : right nonce
			SendHandleKLAYTransfer(bridge, bridgeAccount, testAcc.From, transferAmount, sentNonce, t)
			backend.Commit()

		case err := <-handleSub.Err():
			t.Log("Contract Event Loop Running Stop by handleSub.Err()", "err", err)
			break loop

		case <-time.After(timeOut):
			t.Log("Contract Event Loop Running Stop by timeout")
			break loop
		}
	}

	t.Fatal("fail to check monotone increasing nonce", "lastNonce", sentNonce)
}
