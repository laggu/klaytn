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
	"sync"
	"testing"
)

func TestGateWayManager(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(2)

	// Generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	key2, _ := crypto.GenerateKey()
	auth2 := bind.NewKeyedTransactor(key2)

	alloc := blockchain.GenesisAlloc{auth.From: {Balance: big.NewInt(params.KLAY)}, auth2.From: {Balance: big.NewInt(params.KLAY)}}
	sim := backends.NewSimulatedBackend(alloc)

	balance2, _ := sim.BalanceAt(context.Background(), auth2.From, big.NewInt(0))
	fmt.Println("after reward, balance :", balance2)

	config := &SCConfig{}
	config.nodekey = key
	config.chainkey = key2

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

	addr, err := gatewayManager.DeployGatewayTest(sim, false)
	if err != nil {
		log.Fatalf("Failed to deploy new gateway contract: %v", err)
	}
	fmt.Println("===== GatewayContract Addr ", addr.Hex())

	gatewayManager.SubscribeEvent(addr, false)

	tokenCh := make(chan TokenReceivedEvent)
	tokenSendCh := make(chan TokenTransferEvent)
	gatewayManager.SubscribeKRC20TokenReceived(tokenCh)
	gatewayManager.SubscribeKRC20WithDraw(tokenSendCh)

	go func() {
		for {
			select {
			case ev := <-tokenCh:
				fmt.Println(" receive token received event ", ev.ContractAddr.Hex())
				wg.Done()
			case ev := <-tokenSendCh:
				fmt.Println(" receive token withdraw event ", ev.ContractAddr.Hex())
				wg.Done()
			}
		}
	}()

	gxtokenaddr, _, gxtoken, err := stoken.DeployGXToken(auth, sim, addr)
	if err != nil {
		log.Fatalf("Failed to DeployGXToken: %v", err)
	}

	// Gateway transfer ERC20 to address
	gateway := gatewayManager.GetGateway(addr, false)
	tx, err := gateway.WithdrawERC20(&bind.TransactOpts{From: auth2.From, Signer: auth2.Signer}, big.NewInt(200), auth.From, gxtokenaddr)
	if err != nil {
		log.Fatalf("Failed to WithdrawERC20: %v", err)
	}
	fmt.Println(" Trasaction ", tx.Hash().Hex())
	sim.Commit() // block #1

	// address transfer ERC20 to Gateway
	tx, err = gxtoken.DepositToGateway(&bind.TransactOpts{From: auth.From, Signer: auth.Signer}, big.NewInt(100))
	if err != nil {
		log.Fatalf("Failed to SafeTransferAndCall: %v", err)
	}
	fmt.Println(" Trasaction ", tx.Hash().Hex())
	sim.Commit() // block #2

	balance, _ := sim.BalanceAt(context.Background(), auth.From, big.NewInt(2))
	fmt.Println("auth balance :", balance)

	//tx, err = gateway.WithdrawKLAY(&bind.TransactOpts{From: auth2.From, Signer: auth2.Signer},big.NewInt(200), auth.From)
	//if err != nil {
	//	log.Fatalf("Failed to WithdrawKLAY: %v", err)
	//}
	//fmt.Println(" Trasaction ", tx.Hash().Hex())

	sim.Commit() // block #3

	balance, _ = sim.BalanceAt(context.Background(), auth.From, big.NewInt(3))
	fmt.Println("auth balance :", balance)

	wg.Wait()

	gatewayManager.Stop()
	fmt.Println("GateWay Contract Addr ", addr.Hex())
}

// for TestMethod
func (gwm *GateWayManager) DeployGatewayTest(backend bind.ContractBackend, local bool) (common.Address, error) {

	if local {
		addr, gateway, err := gwm.deployGatewayTest(gwm.subBridge.getChainID(), big.NewInt((int64)(gwm.subBridge.handler.getNodeAccountNonce())), gwm.subBridge.handler.nodeKey, backend)
		gwm.localGateWays[addr] = gateway
		return addr, err
	} else {
		addr, gateway, err := gwm.deployGatewayTest(gwm.subBridge.handler.parentChainID, big.NewInt((int64)(gwm.subBridge.handler.chainAccountNonce)), gwm.subBridge.handler.chainKey, backend)
		gwm.remoteGateWays[addr] = gateway
		return addr, err
	}
}

func (gwm *GateWayManager) deployGatewayTest(chainID *big.Int, nonce *big.Int, accountKey *ecdsa.PrivateKey, backend bind.ContractBackend) (common.Address, *gatewaycontract.Gateway, error) {

	auth := bind.NewKeyedTransactor(accountKey)
	addr, tx, contract, err := gatewaycontract.DeployGateway(auth, backend, true)
	if err != nil {
		logger.Error("", "err", err)
		return common.Address{}, nil, err
	}
	logger.Info("Gateway is deploying on CurrentChain", "addr", addr, "txHash", tx.Hash().String())
	return addr, contract, nil
}
