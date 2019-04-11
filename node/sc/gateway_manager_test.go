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
	"github.com/ground-x/klaytn/contracts/gateway"
	"github.com/ground-x/klaytn/contracts/servicechain_nft"
	"github.com/ground-x/klaytn/contracts/servicechain_token"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/params"
	"github.com/pkg/errors"
	"log"
	"math/big"
	"os"
	"path"
	"sync"
	"testing"
	"time"
)

// WaitGroupWithTimeOut waits the given wait group until the timout duration.
func WaitGroupWithTimeOut(wg *sync.WaitGroup, duration time.Duration, t *testing.T) {
	c := make(chan struct{})
	go func() {
		wg.Wait()
		c <- struct{}{}
	}()
	fmt.Println("start to wait group")
	select {
	case <-c:
		fmt.Println("waiting group is done")
	case <-time.After(duration):
		t.Fatal("timed out waiting group")
	}
}

// TestGateWayManager tests the event/method of Token/NFT/Gateway contracts.
// And It tests the nonce error case of gateway deploy (#2284)
// TODO-Klaytn-Servicechain needs to refine this test.
// - consider main/service chain simulated backend.
// - separate each test
func TestGateWayManager(t *testing.T) {
	defer func() {
		if err := os.Remove(path.Join(os.TempDir(), GatewayAddrJournal)); err != nil {
			t.Fatalf("fail to delete file %v", err)
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(7)

	// Generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	key2, _ := crypto.GenerateKey()
	auth2 := bind.NewKeyedTransactor(key2)

	key3, _ := crypto.GenerateKey()
	auth3 := bind.NewKeyedTransactor(key3)

	key4, _ := crypto.GenerateKey()
	auth4 := bind.NewKeyedTransactor(key4)

	alloc := blockchain.GenesisAlloc{auth.From: {Balance: big.NewInt(params.KLAY)}, auth2.From: {Balance: big.NewInt(params.KLAY)}, auth4.From: {Balance: big.NewInt(params.KLAY)}}
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
	sim.Commit() // block

	// 2. Deploy Token Contract
	tokenAddr, tx, token, err := sctoken.DeployServiceChainToken(auth, sim, addr)
	if err != nil {
		log.Fatalf("Failed to DeployGXToken: %v", err)
	}
	sim.Commit() // block

	// 3. Deploy NFT Contract
	nftTokenID := uint64(4438)
	nftAddr, tx, nft, err := scnft.DeployServiceChainNFT(auth, sim, addr)
	if err != nil {
		log.Fatalf("Failed to DeployServiceChainNFT: %v", err)
	}
	sim.Commit() // block

	// TODO-Klaytn-Servicechain needs to support WaitDeployed
	//timeoutContext, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancelTimeout()
	//
	//addr, err = bind.WaitDeployed(timeoutContext, sim, tx)
	//if err != nil {
	//	log.Fatal("Failed to DeployGXToken.", "err", err, "txHash", tx.Hash().String())
	//
	//}
	//fmt.Println("GXToken is deployed.", "addr", addr.String(), "txHash", tx.Hash().String())

	balance, _ := sim.BalanceAt(context.Background(), auth.From, nil)
	fmt.Println("auth KLAY balance :", balance)

	balance, _ = sim.BalanceAt(context.Background(), auth2.From, nil)
	fmt.Println("auth2 KLAY balance :", balance)

	balance, _ = sim.BalanceAt(context.Background(), auth3.From, nil)
	fmt.Println("auth3 KLAY balance :", balance)

	balance, _ = sim.BalanceAt(context.Background(), auth4.From, nil)
	fmt.Println("auth4 KLAY balance :", balance)

	// 4. Subscribe Gateway Contract
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
					tx, err := gateway.WithdrawToken(&bind.TransactOpts{From: auth2.From, Signer: auth2.Signer}, ev.Amount, ev.To, tokenAddr)
					if err != nil {
						log.Fatalf("Failed to WithdrawToken: %v", err)
					}
					fmt.Println("WithdrawToken Transaction by event ", tx.Hash().Hex())
					sim.Commit() // block

				case 2:
					owner, err := nft.OwnerOf(&bind.CallOpts{From: auth.From}, big.NewInt(int64(nftTokenID)))
					if err != nil {
						t.Fatal(err)
					}
					fmt.Println("NFT owner before WithdrawERC721: ", owner.String())

					// WithdrawToken by Event
					tx, err := gateway.WithdrawERC721(&bind.TransactOpts{From: auth2.From, Signer: auth2.Signer}, ev.Amount, nftAddr, ev.To)
					if err != nil {
						log.Fatalf("Failed to WithdrawERC721: %v", err)
					}
					fmt.Println("WithdrawERC721 Transaction by event ", tx.Hash().Hex())
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

	// 5. WithdrawToken to Auth for chargin and Check balances
	{
		tx, err = gateway.WithdrawToken(&bind.TransactOpts{From: auth2.From, Signer: auth2.Signer, GasLimit: 999999}, testToken, auth.From, tokenAddr)
		if err != nil {
			log.Fatalf("Failed to WithdrawToken: %v", err)
		}
		fmt.Println("WithdrawToken Transaction", tx.Hash().Hex())
		sim.Commit() // block

		// TODO-Klaytn-Servicechain needs to support WaitMined
		//timeoutContext, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Second)
		//defer cancelTimeout()
		//
		//receipt, err := bind.WaitMined(timeoutContext, sim, tx)
		//if err != nil {
		//	log.Fatal("Failed to WithdrawToken.", "err", err, "txHash", tx.Hash().String(),"status",receipt.Status)
		//}
		//fmt.Println("WithdrawToken is executed.", "addr", addr.String(), "txHash", tx.Hash().String())

		balance, err = token.BalanceOf(&bind.CallOpts{From: auth.From}, auth.From)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("auth token balance", balance.String())

		balance, err = token.BalanceOf(&bind.CallOpts{From: auth2.From}, auth2.From)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("auth2 token balance", balance.String())

		balance, err = token.BalanceOf(&bind.CallOpts{From: auth3.From}, auth3.From)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("auth3 token balance", balance.String())

		balance, err = token.BalanceOf(&bind.CallOpts{From: auth4.From}, auth4.From)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("auth4 token balance", balance.String())
	}

	// 6. Register (Mint) an NFT to Auth4
	{
		tx, err = nft.Register(&bind.TransactOpts{From: auth.From, Signer: auth.Signer, GasLimit: 999999}, auth4.From, nftTokenID)
		if err != nil {
			log.Fatalf("Failed to Register NFT: %v", err)
		}
		fmt.Println("Register NFT Transaction", tx.Hash().Hex())
		sim.Commit() // block

		balance, err = nft.BalanceOf(&bind.CallOpts{From: auth.From}, auth4.From)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("auth4 NFT balance", balance.String())
		fmt.Println("auth4 address", auth4.From.String())
		owner, err := nft.OwnerOf(&bind.CallOpts{From: auth.From}, big.NewInt(int64(nftTokenID)))
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("NFT owner after registering", owner.String())
	}

	// 7. DepositToGateway from auth to auth3
	{
		tx, err = token.DepositToGateway(&bind.TransactOpts{From: auth.From, Signer: auth.Signer, GasLimit: 99999}, testToken, auth3.From)
		if err != nil {
			log.Fatalf("Failed to SafeTransferAndCall: %v", err)
		}
		fmt.Println("DepositToGateway Transaction", tx.Hash().Hex())
		sim.Commit() // block

		// TODO-Klaytn-Servicechain needs to support WaitMined
		//timeoutContext, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Second)
		//defer cancelTimeout()
		//
		//receipt, err := bind.WaitMined(timeoutContext, sim, tx)
		//if err != nil {
		//	log.Fatal("Failed to DepositToGateway.", "err", err, "txHash", tx.Hash().String(), "status", receipt.Status)
		//
		//}
		//fmt.Println("DepositToGateway is executed.", "addr", addr.String(), "txHash", tx.Hash().String())

	}

	// 8. DepositKLAY from auth to auth3
	{
		tx, err = gateway.DepositKLAY(&bind.TransactOpts{From: auth.From, Signer: auth.Signer, Value: testKLAY, GasLimit: 99999}, auth3.From)
		if err != nil {
			log.Fatalf("Failed to DepositKLAY: %v", err)
		}
		fmt.Println("DepositKLAY Transaction", tx.Hash().Hex())

		sim.Commit() // block
	}

	// 9. Request NFT value transfer from auth4 to auth3
	{
		tx, err = nft.RequestValueTransfer(&bind.TransactOpts{From: auth4.From, Signer: auth4.Signer, GasLimit: 999999}, big.NewInt(int64(nftTokenID)), auth3.From)
		if err != nil {
			log.Fatalf("Failed to nft.DepositToGateway: %v", err)
		}
		fmt.Println("nft.DepositToGateway Transaction", tx.Hash().Hex())

		sim.Commit() // block

		timeoutContext, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelTimeout()

		receipt, err := bind.WaitMined(timeoutContext, sim, tx)
		if err != nil {
			log.Fatal("Failed to nft.DepositToGateway.", "err", err, "txHash", tx.Hash().String(), "status", receipt.Status)

		}
		fmt.Println("nft.DepositToGateway is executed.", "addr", addr.String(), "txHash", tx.Hash().String())
	}

	// Wait a few second for wait group
	WaitGroupWithTimeOut(&wg, 3*time.Second, t)

	// 10. Check Token balance
	{
		balance, err = token.BalanceOf(&bind.CallOpts{From: auth.From}, auth.From)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("auth token balance", balance.String())
		balance, err = token.BalanceOf(&bind.CallOpts{From: auth2.From}, auth2.From)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("auth2 token balance", balance.String())
		balance, err = token.BalanceOf(&bind.CallOpts{From: auth3.From}, auth3.From)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("auth3 token balance", balance.String())

		if balance.Cmp(testToken) != 0 {
			t.Fatal("testToken is mismatched,", "expected", testToken.String(), "result", balance.String())
		}
	}

	// 11. Check KLAY balance
	{
		balance, _ = sim.BalanceAt(context.Background(), auth.From, nil)
		fmt.Println("auth KLAY balance :", balance)

		balance, _ = sim.BalanceAt(context.Background(), auth2.From, nil)
		fmt.Println("auth2 KLAY balance :", balance)

		balance, _ = sim.BalanceAt(context.Background(), auth3.From, nil)
		fmt.Println("auth3 KLAY balance :", balance)

		if balance.Cmp(testKLAY) != 0 {
			t.Fatal("testKLAY is mismatched,", "expected", testKLAY.String(), "result", balance.String())
		}
	}

	// 12. Check NFT owner
	{
		owner, err := nft.OwnerOf(&bind.CallOpts{From: auth.From}, big.NewInt(int64(nftTokenID)))
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("NFT owner", owner.String())
		if owner != auth3.From {
			t.Fatal("NFT owner is mismatched", "expeted", auth3.From.String(), "result", owner.String())
		}
	}

	// 13. Nonce check on deploy error
	{
		addr2, err := gatewayManager.DeployGatewayNonceTest(sim)
		if err != nil {
			log.Fatalf("Failed to deploy new gateway contract: %v %v", err, addr2)
		}
	}

	gatewayManager.Stop()
}

// for TestMethod
func (gwm *GateWayManager) DeployGatewayTest(backend *backends.SimulatedBackend, local bool) (common.Address, error) {
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

func (gwm *GateWayManager) deployGatewayTest(chainID *big.Int, nonce *big.Int, accountKey *ecdsa.PrivateKey, backend *backends.SimulatedBackend) (common.Address, *gateway.Gateway, error) {
	auth := bind.NewKeyedTransactor(accountKey)
	auth.Value = big.NewInt(10000)
	addr, tx, contract, err := gateway.DeployGateway(auth, backend, true)
	if err != nil {
		logger.Error("", "err", err)
		return common.Address{}, nil, err
	}
	logger.Info("Gateway is deploying on CurrentChain", "addr", addr, "txHash", tx.Hash().String())

	// TODO-Klaytn-Servicechain needs to support WaitMined
	//backend.Commit()
	//
	//timeoutContext, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancelTimeout()
	//
	//receipt, err := bind.WaitMined(timeoutContext, backend, tx)
	//if err != nil {
	//	log.Fatal("Failed to deploy.", "err", err, "txHash", tx.Hash().String(), "status", receipt.Status)
	//	return common.Address{}, nil, err
	//}
	//fmt.Println("deployGateway is executed.", "addr", addr.String(), "txHash", tx.Hash().String())

	return addr, contract, nil
}

// Nonce should not be increased when error occurs
func (gwm *GateWayManager) DeployGatewayNonceTest(backend bind.ContractBackend) (common.Address, error) {
	key := gwm.subBridge.handler.chainKey
	nonce := gwm.subBridge.handler.getChainAccountNonce()
	gwm.subBridge.handler.chainKey = nil
	addr, _ := gwm.DeployGateway(backend, false)
	gwm.subBridge.handler.chainKey = key

	if nonce != gwm.subBridge.handler.getChainAccountNonce() {
		return addr, errors.New("nonce is accidentally increased")
	}

	return addr, nil
}
