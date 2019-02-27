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

package tests

import (
	"context"
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/accounts/abi/bind/backends"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/contracts/reward/contract"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/params"
	"math/big"
	"reflect"
	"testing"
)

// TestLocalAddressBookInit test bootstrapping of AddressBook contract locally using simulated backend.
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
func TestLocalAddressBookInit(t *testing.T) {
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

	// Generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	key2, _ := crypto.GenerateKey()
	auth2 := bind.NewKeyedTransactor(key2)

	alloc := blockchain.GenesisAlloc{auth.From: {Balance: big.NewInt(params.KLAY)}, auth2.From: {Balance: big.NewInt(params.KLAY)}}
	sim := backends.NewSimulatedBackend(alloc)

	// Deploy a token contract on the simulated blockchain
	_, _, addressBook, err := contract.DeployAddressBook(auth, sim)
	if err != nil {
		t.Fatalf("Err %v", err)
	}

	// Call initTest
	transactOpt := bind.NewKeyedTransactor(key)
	tx, err := addressBook.InitTest(transactOpt)
	if err != nil {
		t.Fatalf("Err %v", err)
	}

	// Commit all pending transactions in the simulator and print the names again
	sim.Commit()

	ctx := context.Background()

	_, err = sim.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		t.Fatalf("Err %v", err)
	}

	// Call completeInitialization()
	transactOpt = bind.NewKeyedTransactor(key)
	_, err = addressBook.CompleteInitialization(transactOpt)
	if err != nil {
		t.Errorf("Err %v", err)
	} else {

		// Commit all pending transactions in the simulator and print the names again
		sim.Commit()

		_, err = sim.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			t.Errorf("Err %v", err)
		}
	}

	// Check data in AddressBook contract
	callOpt := &bind.CallOpts{
		Pending: false,
		From:    auth.From,
		Context: ctx,
	}
	_, stakingAddrs, rewardAddrs, _, _, err := addressBook.GetAllAddressInfo(callOpt)
	if err != nil {
		t.Fatalf("Err %v", err)
	}

	if reflect.DeepEqual(rewardAddrs, expectedRewardAddrs) != true {
		t.Errorf("rewardAddrs mismatched.")
	}
	if reflect.DeepEqual(stakingAddrs, expectedStakingAddrs) != true {
		t.Errorf("stakingAddrs mismatched.")
	}
}
