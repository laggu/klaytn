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
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/accounts/abi/bind/backends"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/common"
	gatewaycontract "github.com/ground-x/klaytn/contracts/gateway"
	stoken "github.com/ground-x/klaytn/contracts/token"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/params"
	"log"
	"math/big"
	"os"
	"path"
	"sync"
	"testing"
)

// TestGateWayManager tests the following event/method of smart contract.
// Token Contract
// - DepositToGateway method
// Gateway Contract
// - DepositKLAY method
// - WithdrawKLAY method
// - WithdrawToken method
// - TokenWithdrawn event
// - TokenReceived event
func TestGateWayManager(t *testing.T) {

	defer func() {
		if err := os.Remove(path.Join(os.TempDir(), GatewayAddrJournal)); err != nil {
			t.Fatalf("fail to delete file %v", err)
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(5)

	// Generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	key2, _ := crypto.GenerateKey()
	auth2 := bind.NewKeyedTransactor(key2)

	key3, _ := crypto.GenerateKey()
	auth3 := bind.NewKeyedTransactor(key3)

	alloc := blockchain.GenesisAlloc{auth.From: {Balance: big.NewInt(params.KLAY)}, auth2.From: {Balance: big.NewInt(params.KLAY)}}
	sim := backends.NewSimulatedBackend(alloc)

	balance2, _ := sim.BalanceAt(context.Background(), auth2.From, big.NewInt(0))
	fmt.Println("after reward, balance :", balance2)

	config := &SCConfig{}
	config.nodekey = key
	config.chainkey = key2
	config.DataDir = os.TempDir()

	chainKeyAddr := crypto.PubkeyToAddress(config.chainkey.PublicKey)
	config.ChainAccountAddr = &chainKeyAddr

	sc := &SubBridge{
		config: config,
		peers:  newBridgePeerSet(),
	}
	var err error
	sc.handler, err = NewSubBridgeHandler(sc.config, sc)
	if err != nil {
		log.Fatalf("Failed to initialize bridgeHandler : %v", err)
		return
	}

	gatewayManager, err := NewGateWayManager(sc)

	testToken := big.NewInt(123)
	testKLAY := big.NewInt(321)

	// 1. Deploy Gateway Contract
	addr, err := gatewayManager.DeployGatewayTest(sim, false)
	if err != nil {
		log.Fatalf("Failed to deploy new gateway contract: %v", err)
	}
	gateway := gatewayManager.GetGateway(addr)
	fmt.Println("===== GatewayContract Addr ", addr.Hex())

	// 2. Deploy Token Contract
	gxtokenaddr, _, gxtoken, err := stoken.DeployGXToken(auth, sim, addr)
	if err != nil {
		log.Fatalf("Failed to DeployGXToken: %v", err)
	}

	balance, _ := sim.BalanceAt(context.Background(), auth.From, nil)
	fmt.Println("auth KLAY balance :", balance)

	balance, _ = sim.BalanceAt(context.Background(), auth2.From, nil)
	fmt.Println("auth2 KLAY balance :", balance)

	balance, _ = sim.BalanceAt(context.Background(), auth3.From, nil)
	fmt.Println("auth3 KLAY balance :", balance)

	// 3. Subscribe Gateway Contract
	gatewayManager.SubscribeEvent(addr)
	tokenCh := make(chan TokenReceivedEvent)
	tokenSendCh := make(chan TokenTransferEvent)
	gatewayManager.SubscribeTokenReceived(tokenCh)
	gatewayManager.SubscribeTokenWithDraw(tokenSendCh)

	go func() {
		for {
			select {
			case ev := <-tokenCh:
				fmt.Println("Deposit Event",
					"type", ev.TokenType,
					"amount", ev.Amount,
					"from", ev.From.String(),
					"to", ev.To.String(),
					"contract", ev.ContractAddr.String(),
					"token", ev.TokenAddr.String())

				switch ev.TokenType {
				case 0:
					// WithdrawKLAY by Event
					tx, err := gateway.WithdrawKLAY(&bind.TransactOpts{From: auth2.From, Signer: auth2.Signer}, ev.Amount, ev.To)
					if err != nil {
						log.Fatalf("Failed to WithdrawKLAY: %v", err)
					}
					fmt.Println("WithdrawKLAY Transaction by event ", tx.Hash().Hex())
					sim.Commit() // block

				case 1:
					// WithdrawToken by Event
					tx, err := gateway.WithdrawToken(&bind.TransactOpts{From: auth2.From, Signer: auth2.Signer}, ev.Amount, ev.To, gxtokenaddr)
					if err != nil {
						log.Fatalf("Failed to WithdrawToken: %v", err)
					}
					fmt.Println("WithdrawToken Transaction by event ", tx.Hash().Hex())
					sim.Commit() // block
				}

				wg.Done()

			case ev := <-tokenSendCh:
				fmt.Println("receive token withdraw event ", ev.ContractAddr.Hex())
				fmt.Println("Withdraw Event",
					"type", ev.TokenType,
					"amount", ev.Amount,
					"owner", ev.Owner.String(),
					"contract", ev.ContractAddr.String(),
					"token", ev.TokenAddr.String())
				wg.Done()
			}
		}
	}()

	// 4. WithdrawToken to Auth2 for charging
	tx, err := gateway.WithdrawToken(&bind.TransactOpts{From: auth2.From, Signer: auth2.Signer}, testToken, auth.From, gxtokenaddr)
	if err != nil {
		log.Fatalf("Failed to WithdrawToken: %v", err)
	}
	fmt.Println("WithdrawToken Transaction", tx.Hash().Hex())
	sim.Commit() // block

	balance, err = gxtoken.BalanceOfMine(&bind.CallOpts{From: auth.From})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("auth token balance", balance.String())
	balance, err = gxtoken.BalanceOfMine(&bind.CallOpts{From: auth2.From})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("auth2 token balance", balance.String())
	balance, err = gxtoken.BalanceOfMine(&bind.CallOpts{From: auth3.From})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("auth3 token balance", balance.String())

	// 5. DepositToGateway from auth to auth3
	tx, err = gxtoken.DepositToGateway(&bind.TransactOpts{From: auth.From, Signer: auth.Signer}, testToken, auth3.From)
	if err != nil {
		log.Fatalf("Failed to SafeTransferAndCall: %v", err)
	}
	fmt.Println("DepositToGateway Transaction", tx.Hash().Hex())
	sim.Commit() // block

	// 6. DepositKLAY from auth to auth3
	tx, err = gateway.DepositKLAY(&bind.TransactOpts{From: auth.From, Signer: auth.Signer, Value: testKLAY}, auth3.From)
	if err != nil {
		log.Fatalf("Failed to DepositKLAY: %v", err)
	}
	fmt.Println("DepositKLAY Transaction", tx.Hash().Hex())

	sim.Commit() // block

	wg.Wait()

	balance, err = gxtoken.BalanceOfMine(&bind.CallOpts{From: auth.From})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("auth token balance", balance.String())
	balance, err = gxtoken.BalanceOfMine(&bind.CallOpts{From: auth2.From})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("auth2 token balance", balance.String())
	balance, err = gxtoken.BalanceOfMine(&bind.CallOpts{From: auth3.From})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("auth3 token balance", balance.String())

	if balance.Cmp(testToken) != 0 {
		t.Fatal("testToken is mismatched,", "expected", testToken.String(), "result", balance.String())
	}

	balance, _ = sim.BalanceAt(context.Background(), auth.From, nil)
	fmt.Println("auth KLAY balance :", balance)

	balance, _ = sim.BalanceAt(context.Background(), auth2.From, nil)
	fmt.Println("auth2 KLAY balance :", balance)

	balance, _ = sim.BalanceAt(context.Background(), auth3.From, nil)
	fmt.Println("auth3 KLAY balance :", balance)

	if balance.Cmp(testKLAY) != 0 {
		t.Fatal("testKLAY is mismatched,", "expected", testKLAY.String(), "result", balance.String())
	}

	gatewayManager.Stop()
}

// for TestMethod
func (gwm *GateWayManager) DeployGatewayTest(backend bind.ContractBackend, local bool) (common.Address, error) {

	if local {
		addr, gateway, err := gwm.deployGatewayTest(big.NewInt(2019), big.NewInt((int64)(gwm.subBridge.handler.getNodeAccountNonce())), gwm.subBridge.handler.nodeKey, backend)
		gwm.localGateWays[addr] = gateway
		gwm.all[addr] = true
		return addr, err
	} else {
		addr, gateway, err := gwm.deployGatewayTest(gwm.subBridge.handler.parentChainID, big.NewInt((int64)(gwm.subBridge.handler.chainAccountNonce)), gwm.subBridge.handler.chainKey, backend)
		gwm.remoteGateWays[addr] = gateway
		gwm.all[addr] = false
		return addr, err
	}
}

func (gwm *GateWayManager) deployGatewayTest(chainID *big.Int, nonce *big.Int, accountKey *ecdsa.PrivateKey, backend bind.ContractBackend) (common.Address, *gatewaycontract.Gateway, error) {

	auth := bind.NewKeyedTransactor(accountKey)
	auth.Value = big.NewInt(10000)
	addr, tx, contract, err := gatewaycontract.DeployGateway(auth, backend, true)
	if err != nil {
		logger.Error("", "err", err)
		return common.Address{}, nil, err
	}
	logger.Info("Gateway is deploying on CurrentChain", "addr", addr, "txHash", tx.Hash().String())
	return addr, contract, nil
}
