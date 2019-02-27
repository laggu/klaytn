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

// +build Reward

package tests

import (
	"context"
	"fmt"
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/client"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/contracts/reward/contract"
	"github.com/ground-x/klaytn/crypto"
	"reflect"
	"testing"
	"time"
)

// TestRemoteAddressBookInit start bootstrapping of Klaytn network with AddressBook contract deployed in genesis block.
// Prerequisites:
//   1. Launch Klaytn network with AddressBook deployed in genesis block.
//   2. Update endPoint and faucet infomration in TestAddressBookInit for launched Klaytn network
//   3. go test -run TestAddressBookInit -tags Reward
//
// Staking addresses are prepared as below and also defined in AddressBook contract.
// > personal.importRawKey("5696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
// "0xac5e047d39692be8c81d0724543d5de721d0dd54"
// > personal.importRawKey("6696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
// "0x382ef85439dc874a0f55ab4d9801a5056e371b37"
// > personal.importRawKey("7696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
// "0xb821a659c21cb39745144931e71d0e9d09c8647f"
// > personal.importRawKey("8696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
// "0xc1094c666657937ab7ed23040207a8ee68781350"
//
// Reward addresses are prepared as below and also defined in AddressBook contract.
//> personal.importRawKey("1696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
// "0x9b0bf94f5ab62c73e454dfc55adb2d2fa6cd3af5"
// > personal.importRawKey("2696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
// "0x4d83f1795ecdd684e94f1c5893ae6904ebeaeb94"
//> personal.importRawKey("3696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
// "0x9b58fe24f7a7cb9d102e21b3376bd80eefdc320b"
// > personal.importRawKey("4696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
// "0xe60bf7b625e54e9f67767fad0e564f6aec297652"
func TestRemoteAddressBookInit(t *testing.T) {
	expectedStakingAddrs := []common.Address{
		common.HexToAddress("0xac5e047d39692be8c81d0724543d5de721d0dd54"),
		common.HexToAddress("0x382ef85439dc874a0f55ab4d9801a5056e371b37"),
		common.HexToAddress("0xb821a659c21cb39745144931e71d0e9d09c8647f"),
		common.HexToAddress("0xc1094c666657937ab7ed23040207a8ee68781350"),
	}
	expectedRewardAddrs := []common.Address{
		common.HexToAddress("0x9b0bf94f5ab62c73e454dfc55adb2d2fa6cd3af5"),
		common.HexToAddress("0x4d83f1795ecdd684e94f1c5893ae6904ebeaeb94"),
		common.HexToAddress("0x9b58fe24f7a7cb9d102e21b3376bd80eefdc320b"),
		common.HexToAddress("0xe60bf7b625e54e9f67767fad0e564f6aec297652"),
	}

	// Update endpoint and faucet information

	// endpoint
	endPoint := "http://localhost:8545"
	// faucet
	faucetAddress := common.HexToAddress("da4755d9af7acc24a7e579f5abe48225df800a0e")
	faucetKey, err := crypto.HexToECDSA("f8e1c91dccf9c8f3b73d1bae4f21c90391884d635898aa884793c63c774d4c06")
	if err != nil {
		t.Fatalf("Err %v", err)
	}

	//Get RPC client of Klaytn node
	client, err := client.Dial(endPoint)
	if err != nil {
		t.Fatalf("Dial Err : %v", err)
	}

	ctx := context.Background()

	// Check faucet

	blockNum, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("BlockNumber Err %v", err)
	}
	fmt.Printf("block number = %v\n", blockNum)

	balance, err := client.BalanceAt(ctx, faucetAddress, blockNum)
	if err != nil {
		t.Fatalf("Err %v", err)
	}
	fmt.Printf("balanceOf(%v)=[%v]\n", faucetAddress.String(), balance)

	addressBook, err := contract.NewAddressBook(common.HexToAddress(contract.AddressBookAddress), client)
	if err != nil {
		t.Fatalf("Err %v", err)
	}

	// Call initTest()
	transactOpt := bind.NewKeyedTransactor(faucetKey)
	fmt.Printf("transactOpt=%v, %v\n", transactOpt, transactOpt.From.String())
	tx, err := addressBook.InitTest(transactOpt)
	if err != nil {
		t.Fatalf("Err %v", err)
	}
	fmt.Printf("Tx=%v\n", tx)

	time.Sleep(2 * time.Second)

	receipt, err := client.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		t.Fatalf("Err %v", err)
	}
	fmt.Printf("Receipt=%v\n", receipt)

	// Call completeInitialization()
	transactOpt = bind.NewKeyedTransactor(faucetKey)
	fmt.Printf("transactOpt=%v, %v\n", transactOpt, transactOpt.From.String())
	tx, err = addressBook.CompleteInitialization(transactOpt)
	if err != nil {
		t.Errorf("Err %v", err)
	} else {
		fmt.Printf("Tx=%v\n", tx)

		time.Sleep(2 * time.Second)

		receipt, err = client.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			t.Errorf("Err %v", err)
		}
		fmt.Printf("Receipt=%v\n", receipt)
	}

	// Check data in AddressBook contract
	callOpt := &bind.CallOpts{
		Pending: false,
		From:    faucetAddress,
		Context: ctx,
	}
	nodeIds, stakingAddrs, rewardAddrs, pocAddr, kirAddr, err := addressBook.GetAllAddressInfo(callOpt)
	if err != nil {
		t.Fatalf("Err %v", err)
	}
	fmt.Printf("nodeIds=\n")
	for _, nodeId := range nodeIds {
		fmt.Printf("  %v\n", nodeId.String())
	}
	fmt.Printf("stakingAddrs=\n")
	for _, stakingAddr := range stakingAddrs {
		fmt.Printf("  %v\n", stakingAddr.String())
	}
	fmt.Printf("rewardAddrs=\n")
	for _, rewardAddr := range rewardAddrs {
		fmt.Printf("  %v\n", rewardAddr.String())
	}
	fmt.Printf("pocAddr=%v, kirAddr=%v\n", pocAddr.String(), kirAddr.String())

	if reflect.DeepEqual(rewardAddrs, expectedRewardAddrs) != true {
		t.Errorf("rewardAddrs mismatched.")
	}
	if reflect.DeepEqual(stakingAddrs, expectedStakingAddrs) != true {
		t.Errorf("stakingAddrs mismatched.")
	}

	// Get the block
	block, err := client.BlockByNumber(ctx, blockNum)
	if err != nil {
		t.Fatalf("Err %v", err)
	}
	fmt.Printf("Block hash = %v\n", block.Hash().String())
}
