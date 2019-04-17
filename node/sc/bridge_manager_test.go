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
	"github.com/ground-x/klaytn/contracts/bridge"
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

// TestBridgeManager tests the event/method of Token/NFT/Bridge contracts.
// And It tests the nonce error case of bridge deploy (#2284)
// TODO-Klaytn-Servicechain needs to refine this test.
// - consider main/service chain simulated backend.
// - separate each test
func TestBridgeManager(t *testing.T) {
	defer func() {
		if err := os.Remove(path.Join(os.TempDir(), BridgeAddrJournal)); err != nil {
			t.Fatalf("fail to delete file %v", err)
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(6)

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
	config.MainChainAccountAddr = &chainKeyAddr

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

	bridgeManager, err := NewBridgeManager(sc)

	testToken := big.NewInt(123)
	testKLAY := big.NewInt(321)

	// 1. Deploy Bridge Contract
	addr, err := bridgeManager.DeployBridgeTest(sim, false)
	if err != nil {
		log.Fatalf("Failed to deploy new bridge contract: %v", err)
	}
	bridge := bridgeManager.bridges[addr].bridge
	fmt.Println("===== BridgeContract Addr ", addr.Hex())
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
	fmt.Printf("auth(%v) KLAY balance : %v\n", auth.From.String(), balance)

	balance, _ = sim.BalanceAt(context.Background(), auth2.From, nil)
	fmt.Printf("auth2(%v) KLAY balance : %v\n", auth2.From.String(), balance)

	balance, _ = sim.BalanceAt(context.Background(), auth3.From, nil)
	fmt.Printf("auth3(%v) KLAY balance : %v\n", auth3.From.String(), balance)

	balance, _ = sim.BalanceAt(context.Background(), auth4.From, nil)
	fmt.Printf("auth4(%v) KLAY balance : %v\n", auth4.From.String(), balance)

	// 4. Subscribe Bridge Contract
	bridgeManager.SubscribeEvent(addr)

	tokenCh := make(chan TokenReceivedEvent)
	tokenSendCh := make(chan TokenTransferEvent)
	bridgeManager.SubscribeTokenReceived(tokenCh)
	bridgeManager.SubscribeTokenWithDraw(tokenSendCh)

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
					"token", ev.TokenAddr.String(),
					"requestNonce", ev.RequestNonce)

				switch ev.TokenType {
				case 0:
					// WithdrawKLAY by Event
					tx, err := bridge.HandleKLAYTransfer(&bind.TransactOpts{From: auth2.From, Signer: auth2.Signer, GasLimit: 999999}, ev.Amount, ev.To, ev.RequestNonce)
					if err != nil {
						log.Fatalf("Failed to WithdrawKLAY: %v", err)
					}
					fmt.Println("WithdrawKLAY Transaction by event ", tx.Hash().Hex())
					sim.Commit() // block

				case 1:
					// WithdrawToken by Event
					tx, err := bridge.HandleTokenTransfer(&bind.TransactOpts{From: auth2.From, Signer: auth2.Signer, GasLimit: 999999}, ev.Amount, ev.To, tokenAddr, ev.RequestNonce)
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
					tx, err := bridge.HandleNFTTransfer(&bind.TransactOpts{From: auth2.From, Signer: auth2.Signer, GasLimit: 999999}, ev.Amount, ev.To, nftAddr, ev.RequestNonce)
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
					"token", ev.TokenAddr.String(),
					"handleNonce", ev.HandleNonce)
				wg.Done()
			}
		}
	}()

	// 5. transfer from auth to auth2 for charging and check balances
	{
		tx, err = token.Transfer(&bind.TransactOpts{From: auth.From, Signer: auth.Signer, GasLimit: 999999}, auth2.From, testToken)
		if err != nil {
			log.Fatalf("Failed to Transfer for charging: %v", err)
		}
		fmt.Println("Transfer Transaction", tx.Hash().Hex())
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
		tx, err = nft.Register(&bind.TransactOpts{From: auth.From, Signer: auth.Signer, GasLimit: 999999}, auth4.From, big.NewInt(int64(nftTokenID)))
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

	// 7. RequestValueTransfer from auth2 to auth3
	{
		tx, err = token.RequestValueTransfer(&bind.TransactOpts{From: auth2.From, Signer: auth2.Signer, GasLimit: 99999}, testToken, auth3.From)
		if err != nil {
			log.Fatalf("Failed to SafeTransferAndCall: %v", err)
		}
		fmt.Println("RequestValueTransfer Transaction", tx.Hash().Hex())
		sim.Commit() // block

		// TODO-Klaytn-Servicechain needs to support WaitMined
		//timeoutContext, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Second)
		//defer cancelTimeout()
		//
		//receipt, err := bind.WaitMined(timeoutContext, sim, tx)
		//if err != nil {
		//	log.Fatal("Failed to RequestValueTransfer.", "err", err, "txHash", tx.Hash().String(), "status", receipt.Status)
		//
		//}
		//fmt.Println("RequestValueTransfer is executed.", "addr", addr.String(), "txHash", tx.Hash().String())

	}

	// 8. DepositKLAY from auth to auth3
	{
		tx, err = bridge.RequestKLAYTransfer(&bind.TransactOpts{From: auth.From, Signer: auth.Signer, Value: testKLAY, GasLimit: 99999}, auth3.From)
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
			log.Fatalf("Failed to nft.RequestValueTransfer: %v", err)
		}
		fmt.Println("nft.RequestValueTransfer Transaction", tx.Hash().Hex())

		sim.Commit() // block

		timeoutContext, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelTimeout()

		receipt, err := bind.WaitMined(timeoutContext, sim, tx)
		if err != nil {
			log.Fatal("Failed to nft.RequestValueTransfer.", "err", err, "txHash", tx.Hash().String(), "status", receipt.Status)

		}
		fmt.Println("nft.RequestValueTransfer is executed.", "addr", addr.String(), "txHash", tx.Hash().String())
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
		addr2, err := bridgeManager.DeployBridgeNonceTest(sim)
		if err != nil {
			log.Fatalf("Failed to deploy new bridge contract: %v %v", err, addr2)
		}
	}

	bridgeManager.Stop()
}

// TestBridgeManagerJournal tests journal functionality.
func TestBridgeManagerJournal(t *testing.T) {
	defer func() {
		if err := os.Remove(path.Join(os.TempDir(), BridgeAddrJournal)); err != nil {
			t.Fatalf("fail to delete file %v", err)
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(2)

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

	config := &SCConfig{}
	config.nodekey = key
	config.chainkey = key2
	config.DataDir = os.TempDir()
	config.VTRecovery = true

	chainKeyAddr := crypto.PubkeyToAddress(config.chainkey.PublicKey)
	config.MainChainAccountAddr = &chainKeyAddr

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

	testKLAY := big.NewInt(321)

	// 1. Prepare manager and subscribe event
	bm, err := NewBridgeManager(sc)

	addr, err := bm.DeployBridgeTest(sim, false)
	bridge := bm.bridges[addr].bridge
	fmt.Println("===== BridgeContract Addr ", addr.Hex())
	sim.Commit() // block

	bm.bridges[addr] = &BridgeInfo{bridge, true, true}
	bm.journal.cache = []*BridgeJournal{}
	bm.journal.cache = append(bm.journal.cache, &BridgeJournal{addr, addr, true})

	bm.SubscribeEvent(addr)
	bm.unsubscribeEvent(addr)

	// 2. Reload bridge as if journal is loaded (it handles subscription internally)
	bm.loadBridge(addr, sc.remoteBackend, false, true)

	// 3. Run automatic subscription checker
	tokenCh := make(chan TokenReceivedEvent)
	tokenSendCh := make(chan TokenTransferEvent)
	bm.SubscribeTokenReceived(tokenCh)
	bm.SubscribeTokenWithDraw(tokenSendCh)

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
					tx, err := bridge.HandleKLAYTransfer(&bind.TransactOpts{From: auth2.From, Signer: auth2.Signer, GasLimit: 999999}, ev.Amount, ev.To, ev.RequestNonce)

					if err != nil {
						log.Fatalf("Failed to WithdrawKLAY: %v", err)
					}
					fmt.Println("WithdrawKLAY Transaction by event ", tx.Hash().Hex())
					sim.Commit() // block
				}
				wg.Done()

			case ev := <-tokenSendCh:
				fmt.Println("receive token withdraw event ", ev.ContractAddr.Hex())
				wg.Done()
			}
		}
	}()

	// 4. DepositKLAY from auth to auth3 to check subscription
	{
		tx, err := bridge.RequestKLAYTransfer(&bind.TransactOpts{From: auth.From, Signer: auth.Signer, Value: testKLAY, GasLimit: 99999}, auth3.From)
		if err != nil {
			log.Fatalf("Failed to DepositKLAY: %v", err)
		}
		fmt.Println("DepositKLAY Transaction", tx.Hash().Hex())

		sim.Commit() // block
	}

	wg.Wait()

	bm.Stop()
}

// for TestMethod
func (bm *BridgeManager) DeployBridgeTest(backend *backends.SimulatedBackend, local bool) (common.Address, error) {
	if local {
		addr, bridge, err := bm.deployBridgeTest(big.NewInt(2019), big.NewInt((int64)(bm.subBridge.handler.getServiceChainAccountNonce())), bm.subBridge.handler.nodeKey, backend)
		bm.SetBridge(addr, bridge, local, false)
		return addr, err
	} else {
		addr, bridge, err := bm.deployBridgeTest(bm.subBridge.handler.parentChainID, big.NewInt((int64)(bm.subBridge.handler.mainChainAccountNonce)), bm.subBridge.handler.chainKey, backend)
		bm.SetBridge(addr, bridge, local, false)
		return addr, err
	}
}

func (bm *BridgeManager) deployBridgeTest(chainID *big.Int, nonce *big.Int, accountKey *ecdsa.PrivateKey, backend *backends.SimulatedBackend) (common.Address, *bridge.Bridge, error) {
	auth := bind.NewKeyedTransactor(accountKey)
	auth.Value = big.NewInt(10000)
	addr, tx, contract, err := bridge.DeployBridge(auth, backend, true)
	if err != nil {
		logger.Error("", "err", err)
		return common.Address{}, nil, err
	}
	logger.Info("Bridge is deploying on CurrentChain", "addr", addr, "txHash", tx.Hash().String())

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
	//fmt.Println("deployBridge is executed.", "addr", addr.String(), "txHash", tx.Hash().String())

	return addr, contract, nil
}

// Nonce should not be increased when error occurs
func (bm *BridgeManager) DeployBridgeNonceTest(backend bind.ContractBackend) (common.Address, error) {
	key := bm.subBridge.handler.chainKey
	nonce := bm.subBridge.handler.getMainChainAccountNonce()
	bm.subBridge.handler.chainKey = nil
	addr, _ := bm.DeployBridge(backend, false)
	bm.subBridge.handler.chainKey = key

	if nonce != bm.subBridge.handler.getMainChainAccountNonce() {
		return addr, errors.New("nonce is accidentally increased")
	}

	return addr, nil
}
