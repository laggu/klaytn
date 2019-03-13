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
	"crypto/ecdsa"
	"fmt"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/profile"
	"github.com/ground-x/klaytn/kerrors"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

// TestAccountCreationFailTxPool checks that inserting a TxTypeAccountCreation object
// into the txpool fails for an already existing account.
// It tests the following scenario:
// 1. Create an account decoupled using TxTypeAccountCreation.
// 2. Add the same creation tx into the txpool. In this case, it should fail.
func TestAccountCreationFailTxPool(t *testing.T) {
	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(6, 4)
	if err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer bcdata.Shutdown()

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// decoupled account
	decoupled, err := createDecoupledAccount("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
		common.HexToAddress("0x75c3098be5e4b63fbac05838daaee378dd48098d"))
	assert.Equal(t, nil, err)

	signer := types.MakeSigner(bcdata.bc.Config(), bcdata.bc.CurrentHeader().Number)
	gasPrice := big.NewInt(0)

	// 1. Create an account decoupled using TxTypeAccountCreation.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            decoupled.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    decoupled.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// make txpool
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// 2. Add the same creation tx into the txpool. In this case, it should fail.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            decoupled.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    decoupled.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		errs := txpool.AddRemotes(txs)

		assert.Equal(t, kerrors.ErrAccountAlreadyExists, errs[0])

		reservoir.Nonce += 1
	}
}

// TestSmartContractCreationFailTxPool checks that inserting a TxTypeSmartContractDeploy
// object into the txpool fails for an already existing account.
// It tests the following scenario:
// 1. Deploy smart contract (reservoir -> contract)
// 2. Check the the smart contract is deployed well.
// 3. Add the same deploy tx into the txpool. In this case, this should fail.
func TestSmartContractCreationFailTxPool(t *testing.T) {
	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(6, 4)
	if err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer bcdata.Shutdown()

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
	}

	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	gasPrice := new(big.Int).SetUint64(0)
	gasLimit := uint64(250000000)

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	var code string

	if isCompilerAvailable() {
		filename := string("../contracts/reward/contract/KlaytnReward.sol")
		codes, _ := compileSolidity(filename)
		code = codes[0]
	} else {
		// Falling back to use compiled code.
		code = "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"
	}

	// 1. Deploy smart contract (reservoir -> contract)
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(0)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            contract.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          common.FromHex(code),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 2. Check the the smart contract is deployed well.
	{
		statedb, err := bcdata.bc.State()
		if err != nil {
			t.Fatal(err)
		}
		code := statedb.GetCode(contract.Addr)
		assert.Equal(t, 478, len(code))
	}

	// make txpool
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// 3. Add the same deploy tx into the txpool. In this case, this should fail.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(0)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            contract.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          common.FromHex(code),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		errs := txpool.AddRemotes(txs)

		assert.Equal(t, kerrors.ErrAccountAlreadyExists, errs[0])

		reservoir.Nonce += 1
	}
}
