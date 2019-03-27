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
	"github.com/ground-x/klaytn/accounts/abi"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/profile"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/params"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
	"time"
)

// TestTransactionGas checks gas calculation of base transaction types such as:
// 1.  LegacyTransaction
// 2.  TxTypeValueTransfer
// 3.  TxTypeValueTransferMemo with non-zero values.
// 4.  TxTypeValueTransferMemo with zero values.
// 5.  TxTypeAccountCreation
// 6.  TxTypeAccountUpdate
// 7.  TxTypeSmartContractDeploy
// 8.  TxTypeSmartContractExecution
// 9.  TxTypeCancel
// 10. TxTypeChainDataAnchoring
func TestTransactionGas(t *testing.T) {
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

	// anonymous account
	anon, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	// decoupled account
	decoupled, err := createDecoupledAccount("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
		common.HexToAddress("0x75c3098be5e4b63fbac05838daaee378dd48098d"))
	assert.Equal(t, nil, err)

	colin, err := createHumanReadableAccount("ed580f5bd71a2ee4dae5cb43e331b7d0318596e561e6add7844271ed94156b20", "colin")
	assert.Equal(t, nil, err)

	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
		fmt.Println("decoupledAddr = ", decoupled.Addr.String())
		fmt.Println("colinAddr = ", colin.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// Preparing step. Create an account decoupled.
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

	// Preparing step. Deploy a smart contract.
	var abiStr string
	var code string

	if isCompilerAvailable() {
		filename := string("../contracts/reward/contract/KlaytnReward.sol")
		codes, abistrings := compileSolidity(filename)
		code = codes[0]
		abiStr = abistrings[0]
	} else {
		// Falling back to use compiled code.
		code = "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"
		abiStr = `[{"constant":true,"inputs":[],"name":"totalAmount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"receiver","type":"address"}],"name":"reward","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"safeWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"}]`
	}

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

	// 1. LegacyTransaction
	{
		amount := new(big.Int).SetUint64(100000000000)
		tx := types.NewTransaction(reservoir.Nonce,
			anon.Addr, amount, gasLimit, gasPrice, []byte{})

		err := tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		assert.Equal(t, params.TxGas, gas)
	}

	// 2. TxTypeValueTransfer
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000), // smaller than total amount
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		assert.Equal(t, params.TxGas, gas)
	}

	// 3. TxTypeValueTransferMemo with non-zero values.
	{
		data := []byte{1, 2, 3, 4}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000), // smaller than total amount
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		assert.Equal(t, params.TxGas+uint64(len(data))*params.TxDataNonZeroGas, gas)
	}

	// 4. TxTypeValueTransferMemo with zero values.
	{
		data := []byte{0, 0, 0, 0}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000), // smaller than total amount
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		assert.Equal(t, params.TxGas+uint64(len(data))*params.TxDataZeroGas, gas)
	}

	// 5. TxTypeAccountCreation
	{
		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            colin.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    colin.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		assert.Equal(t, params.TxGasAccountCreation+params.TxAccountCreationGasDefault+uint64(len(colin.Keys))*params.TxAccountCreationGasPerKey, gas)
	}

	// 6. TxTypeAccountUpdate
	// Note that we have to use decoupled as a sender since an AccountUpdate transaction cannot be done with reservoir.
	{
		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      decoupled.Nonce,
			types.TxValueKeyFrom:       decoupled.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		numKeys := uint64(len(decoupled.Keys))
		validationGas := params.TxValidationGasDefault + numKeys*params.TxValidationGasPerKey

		assert.Equal(t, params.TxGasAccountUpdate+validationGas+params.TxAccountCreationGasDefault+numKeys*params.TxAccountCreationGasPerKey, gas)
	}

	// 7. TxTypeSmartContractDeploy
	{
		amount := new(big.Int).SetUint64(0)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            common.HexToAddress("12345678"),
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          common.FromHex(code),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(common.FromHex(code), true, true)
		assert.Equal(t, nil, err)

		executionGas := uint64(0x175fd)
		assert.Equal(t, intrinsicGas+executionGas, gas)
	}

	// 8. TxTypeSmartContractExecution
	{
		amountToSend := new(big.Int).SetUint64(10)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", reservoir.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       contract.Addr,
			types.TxValueKeyAmount:   amountToSend,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(data, false, true)
		assert.Equal(t, nil, err)

		executionGas := uint64(0x9ec4)
		assert.Equal(t, intrinsicGas+executionGas, gas)
	}

	// 9. TxTypeCancel
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeCancel, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		assert.Equal(t, params.TxGasCancel, gas)
	}

	// 10. TxTypeChainDataAnchoring
	{
		anchoredData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:        reservoir.Nonce,
			types.TxValueKeyFrom:         reservoir.Addr,
			types.TxValueKeyTo:           reservoir.Addr,
			types.TxValueKeyAmount:       big.NewInt(100000),
			types.TxValueKeyGasLimit:     gasLimit,
			types.TxValueKeyGasPrice:     gasPrice,
			types.TxValueKeyAnchoredData: anchoredData,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeChainDataAnchoring, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		assert.Equal(t, params.TxChainDataAnchoringGas+params.ChainDataAnchoringGas*(uint64)(len(anchoredData)), gas)
	}
}

// TestFeeDelegatedTransactionGas checks gas calculation of fee delegated transaction types such as:
// 1. TxTypeFeeDelegatedValueTransfer
// 2. TxTypeFeeDelegatedValueTransferMemo with non-zero values.
// 3. TxTypeFeeDelegatedValueTransferMemo with zero values.
// 4. TxTypeFeeDelegatedAccountUpdate
// 5. TxTypeFeeDelegatedSmartContractDeploy
// 6. TxTypeFeeDelegatedSmartContractExecution
// 7. TxTypeFeeDelegatedCancel
func TestFeeDelegatedTransactionGas(t *testing.T) {
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

	// anonymous account
	anon, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	// decoupled account
	decoupled, err := createDecoupledAccount("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
		common.HexToAddress("0x75c3098be5e4b63fbac05838daaee378dd48098d"))
	assert.Equal(t, nil, err)

	colin, err := createHumanReadableAccount("ed580f5bd71a2ee4dae5cb43e331b7d0318596e561e6add7844271ed94156b20", "colin")
	assert.Equal(t, nil, err)

	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
		fmt.Println("decoupledAddr = ", decoupled.Addr.String())
		fmt.Println("colinAddr = ", colin.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// Preparing step. Create an account decoupled.
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

	// Preparing step. Deploy a smart contract.
	var abiStr string
	var code string

	if isCompilerAvailable() {
		filename := string("../contracts/reward/contract/KlaytnReward.sol")
		codes, abistrings := compileSolidity(filename)
		code = codes[0]
		abiStr = abistrings[0]
	} else {
		// Falling back to use compiled code.
		code = "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"
		abiStr = `[{"constant":true,"inputs":[],"name":"totalAmount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"receiver","type":"address"}],"name":"reward","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"safeWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"}]`
	}

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

	// 1. TxTypeFeeDelegatedValueTransfer
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyFeePayer: decoupled.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGas + params.TxGasFeeDelegated
		gasFeePayer := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 2. TxTypeFeeDelegatedValueTransferMemo with non-zero values.
	{
		data := []byte{1, 2, 3, 4}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: decoupled.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataNonZeroGas
		intrinsicGas := params.TxGas + gasPayload + params.TxGasFeeDelegated
		gasFeePayer := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 3. TxTypeFeeDelegatedValueTransferMemo with zero values.
	{
		data := []byte{0, 0, 0, 0}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: decoupled.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataZeroGas
		intrinsicGas := params.TxGas + gasPayload + params.TxGasFeeDelegated
		gasFeePayer := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 4. TxTypeFeeDelegatedAccountUpdate
	// Note that we have to use decoupled as a sender since an AccountUpdate transaction cannot be done with reservoir.
	{
		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      decoupled.Nonce,
			types.TxValueKeyFrom:       decoupled.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey),
			types.TxValueKeyFeePayer:   reservoir.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		numKeys := uint64(len(decoupled.Keys))
		gasKey := params.TxAccountCreationGasDefault + numKeys*params.TxAccountCreationGasPerKey
		intrinsicGas := params.TxGasAccountUpdate + gasKey + params.TxGasFeeDelegated
		gasFrom := params.TxValidationGasDefault + numKeys*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 5. TxTypeFeeDelegatedSmartContractDeploy
	{
		amount := new(big.Int).SetUint64(0)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            common.HexToAddress("12345678"),
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyFeePayer:      decoupled.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(common.FromHex(code), true, true)
		assert.Equal(t, nil, err)

		intrinsicGas = intrinsicGas + params.TxGasFeeDelegated
		executionGas := uint64(0x175fd)
		gasFeePayer := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+executionGas+gasFeePayer, gas)
	}

	// 6. TxTypeFeeDelegatedSmartContractExecution
	{
		amount := new(big.Int).SetUint64(10)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", reservoir.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       contract.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: decoupled.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(data, false, true)
		assert.Equal(t, nil, err)

		intrinsicGas = intrinsicGas + params.TxGasFeeDelegated
		executionGas := uint64(0x9ec4)
		gasFeePayer := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+executionGas+gasFeePayer, gas)
	}

	// 7. TxTypeFeeDelegatedCancel
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyFeePayer: decoupled.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedCancel, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGasCancel + params.TxGasFeeDelegated
		gasFeePayer := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}
}

// TestFeeDelegatedWithRatioTransactionGas checks gas calculation of fee delegated with ratio transaction types such as:
// 1. TxTypeFeeDelegatedValueTransferWithRatio
// 2. TxTypeFeeDelegatedValueTransferMemoWithRatio with non-zero values.
// 3. TxTypeFeeDelegatedValueTransferMemoWithRatio with zero values.
// 4. TxTypeFeeDelegatedAccountUpdateWithRatio
// 5. TxTypeFeeDelegatedSmartContractDeployWithRatio
// 6. TxTypeFeeDelegatedSmartContractExecutionWithRatio
// 7. TxTypeFeeDelegatedCancelWithRatio
func TestFeeDelegatedWithRatioTransactionGas(t *testing.T) {
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

	// anonymous account
	anon, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	// decoupled account
	decoupled, err := createDecoupledAccount("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
		common.HexToAddress("0x75c3098be5e4b63fbac05838daaee378dd48098d"))
	assert.Equal(t, nil, err)

	colin, err := createHumanReadableAccount("ed580f5bd71a2ee4dae5cb43e331b7d0318596e561e6add7844271ed94156b20", "colin")
	assert.Equal(t, nil, err)

	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
		fmt.Println("decoupledAddr = ", decoupled.Addr.String())
		fmt.Println("colinAddr = ", colin.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// Preparing step. Create an account decoupled.
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

	// Preparing step. Deploy a smart contract.
	var abiStr string
	var code string

	if isCompilerAvailable() {
		filename := string("../contracts/reward/contract/KlaytnReward.sol")
		codes, abistrings := compileSolidity(filename)
		code = codes[0]
		abiStr = abistrings[0]
	} else {
		// Falling back to use compiled code.
		code = "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"
		abiStr = `[{"constant":true,"inputs":[],"name":"totalAmount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"receiver","type":"address"}],"name":"reward","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"safeWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"}]`
	}

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

	// 1. TxTypeFeeDelegatedValueTransferWithRatio
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 reservoir.Addr,
			types.TxValueKeyAmount:             big.NewInt(100000),
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyFeePayer:           decoupled.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGas + params.TxGasFeeDelegatedWithRatio
		gasFeePayer := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 2. TxTypeFeeDelegatedValueTransferMemoWithRatio with non-zero values.
	{
		data := []byte{1, 2, 3, 4}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 reservoir.Addr,
			types.TxValueKeyAmount:             big.NewInt(100000),
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyData:               data,
			types.TxValueKeyFeePayer:           decoupled.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemoWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataNonZeroGas
		intrinsicGas := params.TxGas + gasPayload + params.TxGasFeeDelegatedWithRatio
		gasFeePayer := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 3. TxTypeFeeDelegatedValueTransferMemoWithRatio with zero values.
	{
		data := []byte{0, 0, 0, 0}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 reservoir.Addr,
			types.TxValueKeyAmount:             big.NewInt(100000),
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyData:               data,
			types.TxValueKeyFeePayer:           decoupled.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemoWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataZeroGas
		intrinsicGas := params.TxGas + gasPayload + params.TxGasFeeDelegatedWithRatio
		gasFeePayer := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 4. TxTypeFeeDelegatedAccountUpdateWithRatio
	// Note that we have to use decoupled as a sender since an AccountUpdate transaction cannot be done with reservoir.
	{
		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              decoupled.Nonce,
			types.TxValueKeyFrom:               decoupled.Addr,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyAccountKey:         accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey),
			types.TxValueKeyFeePayer:           reservoir.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdateWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		numKeys := uint64(len(decoupled.Keys))
		gasKey := params.TxAccountCreationGasDefault + numKeys*params.TxAccountCreationGasPerKey
		intrinsicGas := params.TxGasAccountUpdate + gasKey + params.TxGasFeeDelegatedWithRatio
		gasFrom := params.TxValidationGasDefault + numKeys*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 5. TxTypeFeeDelegatedSmartContractDeployWithRatio
	{
		amount := new(big.Int).SetUint64(0)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 common.HexToAddress("12345678"),
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyHumanReadable:      false,
			types.TxValueKeyData:               common.FromHex(code),
			types.TxValueKeyFeePayer:           decoupled.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeployWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(common.FromHex(code), true, true)
		assert.Equal(t, nil, err)

		intrinsicGas = intrinsicGas + params.TxGasFeeDelegatedWithRatio
		executionGas := uint64(0x175fd)
		gasFeePayer := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+executionGas+gasFeePayer, gas)
	}

	// 6. TxTypeFeeDelegatedSmartContractExecutionWithRatio
	{
		amount := new(big.Int).SetUint64(10)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", reservoir.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 contract.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyData:               data,
			types.TxValueKeyFeePayer:           decoupled.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractExecutionWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(data, false, true)
		assert.Equal(t, nil, err)

		intrinsicGas = intrinsicGas + params.TxGasFeeDelegatedWithRatio
		executionGas := uint64(0x9ec4)
		gasFeePayer := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+executionGas+gasFeePayer, gas)
	}

	// 7. TxTypeFeeDelegatedCancelWithRatio
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyFeePayer:           decoupled.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedCancelWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGasCancel + params.TxGasFeeDelegatedWithRatio
		gasFeePayer := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}
}

// TestTransactionGasWithAccountKeyPublic checks gas calculations using AccountKeyPublic for base transaction types such as:
// 1. TxTypeValueTransfer
// 2. TxTypeValueTransferMemo with non-zero values.
// 3. TxTypeValueTransferMemo with zero values.
// 4. TxTypeAccountCreation
// 5. TxTypeSmartContractDeploy
// 6. TxTypeSmartContractExecution
// 7. TxTypeCancel
// 8. TxTypeChainDataAnchoring
func TestTransactionGasWithAccountKeyPublic(t *testing.T) {
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

	colin, err := createHumanReadableAccount("ed580f5bd71a2ee4dae5cb43e331b7d0318596e561e6add7844271ed94156b20", "colin")
	assert.Equal(t, nil, err)

	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("decoupledAddr = ", decoupled.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// Preparing step. Create an account decoupled.
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

	// Preparing step. Deploy a smart contract.
	var abiStr string
	var code string

	if isCompilerAvailable() {
		filename := string("../contracts/reward/contract/KlaytnReward.sol")
		codes, abistrings := compileSolidity(filename)
		code = codes[0]
		abiStr = abistrings[0]
	} else {
		// Falling back to use compiled code.
		code = "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"
		abiStr = `[{"constant":true,"inputs":[],"name":"totalAmount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"receiver","type":"address"}],"name":"reward","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"safeWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"}]`
	}

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

	// 1. TxTypeValueTransfer
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    decoupled.Nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGas
		gasFrom := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 2. TxTypeValueTransferMemo with non-zero values.
	{
		data := []byte{1, 2, 3, 4}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    decoupled.Nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataNonZeroGas
		intrinsicGas := params.TxGas + gasPayload
		gasFrom := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 3. TxTypeValueTransferMemo with zero values.
	{
		data := []byte{0, 0, 0, 0}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    decoupled.Nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataZeroGas
		intrinsicGas := params.TxGas + gasPayload
		gasFrom := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 4. TxTypeAccountCreation
	{
		amount := new(big.Int).SetUint64(100000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         decoupled.Nonce,
			types.TxValueKeyFrom:          decoupled.Addr,
			types.TxValueKeyTo:            colin.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    colin.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasKey := params.TxAccountCreationGasDefault + uint64(len(colin.Keys))*params.TxAccountCreationGasPerKey
		intrinsicGas := params.TxGasAccountCreation + gasKey
		gasFrom := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 5. TxTypeSmartContractDeploy
	{
		amount := new(big.Int).SetUint64(0)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         decoupled.Nonce,
			types.TxValueKeyFrom:          decoupled.Addr,
			types.TxValueKeyTo:            common.HexToAddress("12345678"),
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          common.FromHex(code),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(common.FromHex(code), true, true)
		assert.Equal(t, nil, err)

		gasFrom := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		executionGas := uint64(0x175fd)

		assert.Equal(t, intrinsicGas+executionGas+gasFrom, gas)
	}

	// 6. TxTypeSmartContractExecution
	{
		amount := new(big.Int).SetUint64(10)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", decoupled.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    decoupled.Nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyTo:       contract.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(data, false, true)
		assert.Equal(t, nil, err)

		gasFrom := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		executionGas := uint64(0x9ec4)

		assert.Equal(t, intrinsicGas+executionGas+gasFrom, gas)
	}

	// 7. TxTypeCancel
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    decoupled.Nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeCancel, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGasCancel
		gasFrom := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 8. TxTypeChainDataAnchoring
	{
		anchoredData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:        decoupled.Nonce,
			types.TxValueKeyFrom:         decoupled.Addr,
			types.TxValueKeyTo:           reservoir.Addr,
			types.TxValueKeyAmount:       big.NewInt(100000),
			types.TxValueKeyGasLimit:     gasLimit,
			types.TxValueKeyGasPrice:     gasPrice,
			types.TxValueKeyAnchoredData: anchoredData,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeChainDataAnchoring, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasAnchoring := params.ChainDataAnchoringGas * (uint64)(len(anchoredData))
		intrinsicGas := params.TxChainDataAnchoringGas + gasAnchoring
		gasFrom := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}
}

// TestTransactionGasWithAccountKeyWeightedMultiSig checks gas calculations using AccountKeyWeightedMultisig for base transaction types such as:
// 1. TxTypeValueTransfer
// 2. TxTypeValueTransferMemo with non-zero values.
// 3. TxTypeValueTransferMemo with zero values.
// 4. TxTypeAccountCreation
// 5. TxTypeAccountUpdate
// 6. TxTypeSmartContractDeploy
// 7. TxTypeSmartContractExecution
// 8. TxTypeCancel
// 9. TxTypeChainDataAnchoring
func TestTransactionGasWithAccountKeyWeightedMultiSig(t *testing.T) {
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

	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	multisig, err := createMultisigAccount(uint(2),
		[]uint{1, 1, 1},
		[]string{"bb113e82881499a7a361e8354a5b68f6c6885c7bcba09ea2b0891480396c322e",
			"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae59438e989",
			"c32c471b732e2f56103e2f8e8cfd52792ef548f05f326e546a7d1fbf9d0419ec"},
		common.HexToAddress("0xbbfa38050bf3167c887c086758f448ce067ea8ea"))

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("multisigAddr = ", multisig.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// Preparing step. Create an account multisig.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            multisig.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    multisig.AccKey,
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

	// Preparing step. Deploy a smart contract.
	var abiStr string
	var code string

	if isCompilerAvailable() {
		filename := string("../contracts/reward/contract/KlaytnReward.sol")
		codes, abistrings := compileSolidity(filename)
		code = codes[0]
		abiStr = abistrings[0]
	} else {
		// Falling back to use compiled code.
		code = "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"
		abiStr = `[{"constant":true,"inputs":[],"name":"totalAmount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"receiver","type":"address"}],"name":"reward","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"safeWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"}]`
	}

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

	// 1. TxTypeValueTransfer
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    multisig.Nonce,
			types.TxValueKeyFrom:     multisig.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGas
		gasFrom := params.TxValidationGasDefault + uint64(len(multisig.Keys))*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 2. TxTypeValueTransferMemo with non-zero values.
	{
		data := []byte{1, 2, 3, 4}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    multisig.Nonce,
			types.TxValueKeyFrom:     multisig.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataNonZeroGas
		intrinsicGas := params.TxGas + gasPayload
		gasFrom := params.TxValidationGasDefault + uint64(len(multisig.Keys))*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 3. TxTypeValueTransferMemo with zero values.
	{
		data := []byte{0, 0, 0, 0}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    multisig.Nonce,
			types.TxValueKeyFrom:     multisig.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataZeroGas
		intrinsicGas := params.TxGas + gasPayload
		gasFrom := params.TxValidationGasDefault + uint64(len(multisig.Keys))*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 4. TxTypeAccountCreation
	{
		newAddr, err := common.FromHumanReadableAddress("newMultiSig")
		assert.Equal(t, nil, err)

		newMultisig, err := createMultisigAccount(uint(2),
			[]uint{1, 1},
			[]string{"98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab",
				"c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c"}, newAddr)

		amount := new(big.Int).SetUint64(100000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         multisig.Nonce,
			types.TxValueKeyFrom:          multisig.Addr,
			types.TxValueKeyTo:            newMultisig.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    newMultisig.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasKey := params.TxAccountCreationGasDefault + uint64(len(newMultisig.Keys))*params.TxAccountCreationGasPerKey
		intrinsicGas := params.TxGasAccountCreation + gasKey
		gasFrom := params.TxValidationGasDefault + uint64(len(multisig.Keys))*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 5. TxTypeAccountUpdate
	{
		k1, err := crypto.HexToECDSA("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
		assert.Equal(t, nil, err)

		k2, err := crypto.HexToECDSA("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c")
		assert.Equal(t, nil, err)

		threshold := uint(2)
		keys := accountkey.WeightedPublicKeys{
			accountkey.NewWeightedPublicKey(1, (*accountkey.PublicKeySerializable)(&k1.PublicKey)),
			accountkey.NewWeightedPublicKey(1, (*accountkey.PublicKeySerializable)(&k2.PublicKey)),
		}
		updateKey := accountkey.NewAccountKeyWeightedMultiSigWithValues(threshold, keys)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      multisig.Nonce,
			types.TxValueKeyFrom:       multisig.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: updateKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasKey := params.TxAccountCreationGasDefault + uint64(len(updateKey.Keys))*params.TxAccountCreationGasPerKey
		intrinsicGas := params.TxGasAccountUpdate + gasKey
		gasFrom := params.TxValidationGasDefault + uint64(len(multisig.Keys))*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 6. TxTypeSmartContractDeploy
	{
		amount := new(big.Int).SetUint64(0)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         multisig.Nonce,
			types.TxValueKeyFrom:          multisig.Addr,
			types.TxValueKeyTo:            common.HexToAddress("12345678"),
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          common.FromHex(code),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(common.FromHex(code), true, true)
		assert.Equal(t, nil, err)

		gasFrom := params.TxValidationGasDefault + uint64(len(multisig.Keys))*params.TxValidationGasPerKey
		executionGas := uint64(0x175fd)

		assert.Equal(t, intrinsicGas+executionGas+gasFrom, gas)
	}

	// 7. TxTypeSmartContractExecution
	{
		amount := new(big.Int).SetUint64(10)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", multisig.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    multisig.Nonce,
			types.TxValueKeyFrom:     multisig.Addr,
			types.TxValueKeyTo:       contract.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(data, false, true)
		assert.Equal(t, nil, err)

		gasFrom := params.TxValidationGasDefault + uint64(len(multisig.Keys))*params.TxValidationGasPerKey
		executionGas := uint64(0x9ec4)

		assert.Equal(t, intrinsicGas+executionGas+gasFrom, gas)
	}

	// 8. TxTypeCancel
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    multisig.Nonce,
			types.TxValueKeyFrom:     multisig.Addr,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeCancel, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGasCancel
		gasFrom := params.TxValidationGasDefault + uint64(len(multisig.Keys))*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 9. TxTypeChainDataAnchoring
	{
		anchoredData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:        multisig.Nonce,
			types.TxValueKeyFrom:         multisig.Addr,
			types.TxValueKeyTo:           reservoir.Addr,
			types.TxValueKeyAmount:       big.NewInt(100000),
			types.TxValueKeyGasLimit:     gasLimit,
			types.TxValueKeyGasPrice:     gasPrice,
			types.TxValueKeyAnchoredData: anchoredData,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeChainDataAnchoring, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasAnchoring := params.ChainDataAnchoringGas * (uint64)(len(anchoredData))
		intrinsicGas := params.TxChainDataAnchoringGas + gasAnchoring
		gasFrom := params.TxValidationGasDefault + uint64(len(multisig.Keys))*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}
}

// TestTransactionGasWithAccountKeyRoleBasedWithAccountKeyPublic checks gas calculations using AccountKeyRoleBased with AccountKeyPublic for base transaction types such as:
// 1.  TxTypeValueTransfer
// 2.  TxTypeValueTransferMemo with non-zero values.
// 3.  TxTypeValueTransferMemo with zero values.
// 4.  TxTypeAccountCreation
// 5.  TxTypeAccountUpdate
// 6.  TxTypeAccountUpdate with AccountNil.
// 7.  TxTypeSmartContractDeploy
// 8.  TxTypeSmartContractExecution
// 9.  TxTypeCancel
// 10. TxTypeChainDataAnchoring
func TestTransactionGasWithAccountKeyRoleBasedWithAccountKeyPublic(t *testing.T) {
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

	colin, err := createHumanReadableAccount("ed580f5bd71a2ee4dae5cb43e331b7d0318596e561e6add7844271ed94156b20", "colin")
	assert.Equal(t, nil, err)

	// get humanreadable address for role based account
	roleBasedAddr, err := common.FromHumanReadableAddress("roleBasedAddr")
	assert.Equal(t, nil, err)

	// role based account
	roleBased, err := createRoleBasedAccountWithAccountKeyPublic(
		[]string{"98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab",
			"c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
			"41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867"}, roleBasedAddr)
	assert.Equal(t, nil, err)

	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("colinAddr = ", colin.Addr.String())
		fmt.Println("roleBasedKeyAddr = ", roleBased.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// Preparing step. Create an role based account.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            roleBased.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleBased.AccKey,
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

	// Preparing step. Deploy a smart contract.
	var abiStr string
	var code string

	if isCompilerAvailable() {
		filename := string("../contracts/reward/contract/KlaytnReward.sol")
		codes, abistrings := compileSolidity(filename)
		code = codes[0]
		abiStr = abistrings[0]
	} else {
		// Falling back to use compiled code.
		code = "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"
		abiStr = `[{"constant":true,"inputs":[],"name":"totalAmount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"receiver","type":"address"}],"name":"reward","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"safeWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"}]`
	}

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

	// 1. TxTypeValueTransfer
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGas
		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFrom := 3 * gasAccountKeyPublicForSigValidation

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 2. TxTypeValueTransferMemo with non-zero values.
	{
		data := []byte{1, 2, 3, 4}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataNonZeroGas
		intrinsicGas := params.TxGas + gasPayload
		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFrom := 3 * gasAccountKeyPublicForSigValidation

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 3. TxTypeValueTransferMemo with zero values.
	{
		data := []byte{0, 0, 0, 0}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataZeroGas
		intrinsicGas := params.TxGas + gasPayload
		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFrom := 3 * gasAccountKeyPublicForSigValidation

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 4. TxTypeAccountCreation
	{
		newAddr, err := common.FromHumanReadableAddress("newAddr")
		assert.Equal(t, nil, err)

		newRoleBased, err := createRoleBasedAccountWithAccountKeyPublic(
			[]string{
				"41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867",
				"98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab",
				"c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c"}, newAddr)
		assert.Equal(t, nil, err)

		amount := new(big.Int).SetUint64(100000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         roleBased.Nonce,
			types.TxValueKeyFrom:          roleBased.Addr,
			types.TxValueKeyTo:            newRoleBased.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    newRoleBased.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasKeyPerAcocuntKeyPublic := params.TxAccountCreationGasDefault + 1*params.TxAccountCreationGasPerKey
		gasKey := 3 * gasKeyPerAcocuntKeyPublic
		intrinsicGas := params.TxGasAccountCreation + gasKey
		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFrom := 3 * gasAccountKeyPublicForSigValidation

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 5. TxTypeAccountUpdate
	{
		k1, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		assert.Equal(t, nil, err)

		k2, err := crypto.HexToECDSA("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
		assert.Equal(t, nil, err)

		k3, err := crypto.HexToECDSA("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c")
		assert.Equal(t, nil, err)

		updateKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&k1.PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&k2.PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&k3.PublicKey),
		})

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      roleBased.Nonce,
			types.TxValueKeyFrom:       roleBased.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: updateKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.UpdateKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasKey := params.TxAccountCreationGasDefault + 3*params.TxAccountCreationGasPerKey
		intrinsicGas := params.TxGasAccountUpdate + gasKey
		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFrom := 3 * gasAccountKeyPublicForSigValidation

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 6. TxTypeAccountUpdate with AccountNil.
	{
		key, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		assert.Equal(t, nil, err)

		updateKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyNil(),
			accountkey.NewAccountKeyPublicWithValue(&key.PublicKey),
			accountkey.NewAccountKeyNil(),
		})

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      roleBased.Nonce,
			types.TxValueKeyFrom:       roleBased.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: updateKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.UpdateKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasKey := params.TxAccountCreationGasDefault + 1*params.TxAccountCreationGasPerKey
		intrinsicGas := params.TxGasAccountUpdate + gasKey
		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFrom := 3 * gasAccountKeyPublicForSigValidation

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 7. TxTypeSmartContractDeploy
	{
		amount := new(big.Int).SetUint64(0)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         roleBased.Nonce,
			types.TxValueKeyFrom:          roleBased.Addr,
			types.TxValueKeyTo:            common.HexToAddress("12345678"),
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          common.FromHex(code),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(common.FromHex(code), true, true)
		assert.Equal(t, nil, err)

		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFrom := 3 * gasAccountKeyPublicForSigValidation
		executionGas := uint64(0x175fd)

		assert.Equal(t, intrinsicGas+executionGas+gasFrom, gas)
	}

	// 8. TxTypeSmartContractExecution
	{
		amount := new(big.Int).SetUint64(10)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", roleBased.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyTo:       contract.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(data, false, true)
		assert.Equal(t, nil, err)

		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFrom := 3 * gasAccountKeyPublicForSigValidation
		executionGas := uint64(0x9ec4)

		assert.Equal(t, intrinsicGas+executionGas+gasFrom, gas)
	}

	// 9. TxTypeCancel
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeCancel, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGasCancel
		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFrom := 3 * gasAccountKeyPublicForSigValidation

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 10. TxTypeChainDataAnchoring
	{
		anchoredData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:        roleBased.Nonce,
			types.TxValueKeyFrom:         roleBased.Addr,
			types.TxValueKeyTo:           reservoir.Addr,
			types.TxValueKeyAmount:       big.NewInt(100000),
			types.TxValueKeyGasLimit:     gasLimit,
			types.TxValueKeyGasPrice:     gasPrice,
			types.TxValueKeyAnchoredData: anchoredData,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeChainDataAnchoring, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasAnchoring := params.ChainDataAnchoringGas * (uint64)(len(anchoredData))
		intrinsicGas := params.TxChainDataAnchoringGas + gasAnchoring
		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFrom := 3 * gasAccountKeyPublicForSigValidation

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}
}

// TestTransactionGasWithAccountKeyRoleBasedWithAccountKeyWeightedMultiSig checks gas calculations using AccountKeyRoleBased with AccountKeyWeightedMultiSig for base transaction types such as:
// 1.  TxTypeValueTransfer
// 2.  TxTypeValueTransferMemo with non-zero values.
// 3.  TxTypeValueTransferMemo with zero values.
// 4.  TxTypeAccountCreation
// 5.  TxTypeAccountUpdate
// 6.  TxTypeAccountUpdate with AccountKeyNil.
// 7.  TxTypeSmartContractDeploy
// 8.  TxTypeSmartContractExecution
// 9.  TxTypeCancel
// 10. TxTypeChainDataAnchoring
func TestTransactionGasWithAccountKeyRoleBasedWithAccountKeyWeightedMultiSig(t *testing.T) {
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

	colin, err := createHumanReadableAccount("ed580f5bd71a2ee4dae5cb43e331b7d0318596e561e6add7844271ed94156b20", "colin")
	assert.Equal(t, nil, err)

	param1 := TestCreateMultisigAccountParam{
		Threshold: uint(2),
		Weights:   []uint{1, 1, 1},
		PrvKeys: []string{"bb113e82881499a7a361e8354a5b68f6c6885c7bcba09ea2b0891480396c322e",
			"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae59438e989",
			"c32c471b732e2f56103e2f8e8cfd52792ef548f05f326e546a7d1fbf9d0419ec"},
	}

	param2 := TestCreateMultisigAccountParam{
		Threshold: uint(2),
		Weights:   []uint{1, 1, 1},
		PrvKeys: []string{"98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab",
			"c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
			"41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867"},
	}

	param3 := TestCreateMultisigAccountParam{
		Threshold: uint(2),
		Weights:   []uint{1, 1, 1},
		PrvKeys: []string{"ed580f5bd71a2ee4dae5cb43e331b7d0318596e561e6add7844271ed94156b20",
			"98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab",
			"c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c"},
	}

	// get humanreadable address for role based account
	roleBasedAddr, err := common.FromHumanReadableAddress("roleBasedAddr")
	assert.Equal(t, nil, err)

	// role based account
	roleBased, err := createRoleBasedAccountWithAccountKeyWeightedMultiSig(
		[]TestCreateMultisigAccountParam{param1, param2, param3}, roleBasedAddr)
	assert.Equal(t, nil, err)

	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("colinAddr = ", colin.Addr.String())
		fmt.Println("roleBasedAddr = ", roleBased.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// Preparing step. Create an role based account.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            roleBased.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleBased.AccKey,
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

	// Preparing step. Deploy a smart contract.
	var abiStr string
	var code string

	if isCompilerAvailable() {
		filename := string("../contracts/reward/contract/KlaytnReward.sol")
		codes, abistrings := compileSolidity(filename)
		code = codes[0]
		abiStr = abistrings[0]
	} else {
		// Falling back to use compiled code.
		code = "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"
		abiStr = `[{"constant":true,"inputs":[],"name":"totalAmount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"receiver","type":"address"}],"name":"reward","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"safeWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"}]`
	}

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

	// 1. TxTypeValueTransfer
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGas

		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFrom := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 2. TxTypeValueTransferMemo with non-zero values.
	{
		data := []byte{1, 2, 3, 4}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataNonZeroGas
		intrinsicGas := params.TxGas + gasPayload

		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFrom := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 3. TxTypeValueTransferMemo with zero values.
	{
		data := []byte{0, 0, 0, 0}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataZeroGas
		intrinsicGas := params.TxGas + gasPayload

		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFrom := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 4. TxTypeAccountCreation
	{
		newRoleBasedAddr, err := common.FromHumanReadableAddress("newRoleBasedAddr")
		assert.Equal(t, nil, err)

		newRoleBased, err := createRoleBasedAccountWithAccountKeyWeightedMultiSig(
			[]TestCreateMultisigAccountParam{param3, param1, param2}, newRoleBasedAddr)
		assert.Equal(t, nil, err)

		amount := new(big.Int).SetUint64(100000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         roleBased.Nonce,
			types.TxValueKeyFrom:          roleBased.Addr,
			types.TxValueKeyTo:            newRoleBased.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    newRoleBased.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasTxKeysForCreation := params.TxAccountCreationGasDefault + uint64(len(newRoleBased.TxKeys))*params.TxAccountCreationGasPerKey
		gasUpdateKeysForCreation := params.TxAccountCreationGasDefault + uint64(len(newRoleBased.UpdateKeys))*params.TxAccountCreationGasPerKey
		gasFeeKeysForCreation := params.TxAccountCreationGasDefault + uint64(len(newRoleBased.FeeKeys))*params.TxAccountCreationGasPerKey
		gasKey := gasTxKeysForCreation + gasUpdateKeysForCreation + gasFeeKeysForCreation
		intrinsicGas := params.TxGasAccountCreation + gasKey

		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFrom := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 5. TxTypeAccountUpdate
	{
		newRoleBasedAddr, err := common.FromHumanReadableAddress("newRoleBasedAddr")
		assert.Equal(t, nil, err)

		newRoleBased, err := createRoleBasedAccountWithAccountKeyWeightedMultiSig(
			[]TestCreateMultisigAccountParam{param3, param1, param2}, newRoleBasedAddr)
		assert.Equal(t, nil, err)

		updateKey := newRoleBased.AccKey

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      roleBased.Nonce,
			types.TxValueKeyFrom:       roleBased.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: updateKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.UpdateKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasTxKeysForCreation := params.TxAccountCreationGasDefault + uint64(len(newRoleBased.TxKeys))*params.TxAccountCreationGasPerKey
		gasUpdateKeysForCreation := params.TxAccountCreationGasDefault + uint64(len(newRoleBased.UpdateKeys))*params.TxAccountCreationGasPerKey
		gasFeeKeysForCreation := params.TxAccountCreationGasDefault + uint64(len(newRoleBased.FeeKeys))*params.TxAccountCreationGasPerKey
		intrinsicGas := params.TxGasAccountUpdate + gasTxKeysForCreation + gasUpdateKeysForCreation + gasFeeKeysForCreation

		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFrom := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 6. TxTypeAccountUpdate with AccountKeyNil.
	{
		multisigAddr, err := common.FromHumanReadableAddress("multisigAddr")
		assert.Equal(t, nil, err)

		multisig, err := createMultisigAccount(uint(2),
			[]uint{1, 1},
			[]string{"98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab",
				"c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c"}, multisigAddr)

		updateKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyNil(),
			multisig.AccKey,
			multisig.AccKey,
		})

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      roleBased.Nonce,
			types.TxValueKeyFrom:       roleBased.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: updateKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.UpdateKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasTxKeysForCreation := params.TxAccountCreationGasDefault + 0*params.TxAccountCreationGasPerKey
		gasUpdateKeysForCreation := params.TxAccountCreationGasDefault + 2*params.TxAccountCreationGasPerKey
		gasFeeKeysForCreation := params.TxAccountCreationGasDefault + 2*params.TxAccountCreationGasPerKey
		intrinsicGas := params.TxGasAccountUpdate + gasTxKeysForCreation + gasUpdateKeysForCreation + gasFeeKeysForCreation

		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFrom := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 7. TxTypeSmartContractDeploy
	{
		amount := new(big.Int).SetUint64(0)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         roleBased.Nonce,
			types.TxValueKeyFrom:          roleBased.Addr,
			types.TxValueKeyTo:            common.HexToAddress("12345678"),
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          common.FromHex(code),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(common.FromHex(code), true, true)
		assert.Equal(t, nil, err)

		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFrom := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys
		executionGas := uint64(0x175fd)

		assert.Equal(t, intrinsicGas+executionGas+gasFrom, gas)
	}

	// 8. TxTypeSmartContractExecution
	{
		amount := new(big.Int).SetUint64(10)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", roleBased.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyTo:       contract.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(data, false, true)
		assert.Equal(t, nil, err)

		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFrom := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys
		executionGas := uint64(0x9ec4)

		assert.Equal(t, intrinsicGas+executionGas+gasFrom, gas)
	}

	// 9. TxTypeCancel
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeCancel, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGasCancel
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFrom := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}

	// 10. TxTypeChainDataAnchoring
	{
		anchoredData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:        roleBased.Nonce,
			types.TxValueKeyFrom:         roleBased.Addr,
			types.TxValueKeyTo:           reservoir.Addr,
			types.TxValueKeyAmount:       big.NewInt(100000),
			types.TxValueKeyGasLimit:     gasLimit,
			types.TxValueKeyGasPrice:     gasPrice,
			types.TxValueKeyAnchoredData: anchoredData,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeChainDataAnchoring, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, roleBased.TxKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasAnchoring := params.ChainDataAnchoringGas * (uint64)(len(anchoredData))
		intrinsicGas := params.TxChainDataAnchoringGas + gasAnchoring

		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFrom := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFrom, gas)
	}
}

// TestFeeDelegatedTransactionGasWithLegacyAndLegacyPayer checks gas calculations
// using AccountKeyLegacy sender and AccountKeyLegacy fee payer for fee delegated transaction types such as:
// 1. TxTypeFeeDelegatedValueTransfer
// 2. TxTypeFeeDelegatedValueTransferMemo with non-zero values.
// 3. TxTypeFeeDelegatedValueTransferMemo with zero values.
// 4. TxTypeFeeDelegatedSmartContractDeploy
// 5. TxTypeFeeDelegatedSmartContractExecution
// 6. TxTypeFeeDelegatedCancel
func TestFeeDelegatedTransactionGasWithLegacyAndLegacyPayer(t *testing.T) {
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

	// anonymous account
	anon, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	// decoupled account
	decoupled, err := createDecoupledAccount("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
		common.HexToAddress("0x75c3098be5e4b63fbac05838daaee378dd48098d"))
	assert.Equal(t, nil, err)

	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// Preparing step. Send Klay to anon account.
	{
		var txs types.Transactions

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       anon.Addr,
			types.TxValueKeyAmount:   big.NewInt(1000000000000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// Preparing step. Create an account decoupled.
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

	// Preparing step. Deploy a smart contract.
	var abiStr string
	var code string

	if isCompilerAvailable() {
		filename := string("../contracts/reward/contract/KlaytnReward.sol")
		codes, abistrings := compileSolidity(filename)
		code = codes[0]
		abiStr = abistrings[0]
	} else {
		// Falling back to use compiled code.
		code = "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"
		abiStr = `[{"constant":true,"inputs":[],"name":"totalAmount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"receiver","type":"address"}],"name":"reward","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"safeWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"}]`
	}

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

	// 1. TxTypeFeeDelegatedValueTransfer
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       anon.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyFeePayer: anon.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGas + params.TxGasFeeDelegated
		gasFeePayer := params.TxValidationGasDefault

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 2. TxTypeFeeDelegatedValueTransferMemo with non-zero values.
	{
		data := []byte{1, 2, 3, 4}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       anon.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: anon.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataNonZeroGas
		intrinsicGas := params.TxGas + gasPayload + params.TxGasFeeDelegated
		gasFeePayer := params.TxValidationGasDefault

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 3. TxTypeFeeDelegatedValueTransferMemo with zero values.
	{
		data := []byte{0, 0, 0, 0}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       anon.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: anon.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataZeroGas
		intrinsicGas := params.TxGas + gasPayload + params.TxGasFeeDelegated
		gasFeePayer := params.TxValidationGasDefault

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 4. TxTypeFeeDelegatedSmartContractDeploy
	{
		amount := new(big.Int).SetUint64(0)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            common.HexToAddress("12345678"),
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyFeePayer:      anon.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(common.FromHex(code), true, true)
		assert.Equal(t, nil, err)

		intrinsicGas = intrinsicGas + params.TxGasFeeDelegated
		executionGas := uint64(0x175fd)
		gasFeePayer := params.TxValidationGasDefault

		assert.Equal(t, intrinsicGas+executionGas+gasFeePayer, gas)
	}

	// 5. TxTypeFeeDelegatedSmartContractExecution
	{
		amount := new(big.Int).SetUint64(10)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", reservoir.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       contract.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: anon.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(data, false, true)
		assert.Equal(t, nil, err)

		intrinsicGas = intrinsicGas + params.TxGasFeeDelegated
		executionGas := uint64(0x9ec4)
		gasFeePayer := params.TxValidationGasDefault

		assert.Equal(t, intrinsicGas+executionGas+gasFeePayer, gas)
	}

	// 6. TxTypeFeeDelegatedCancel
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyFeePayer: anon.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedCancel, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGasCancel + params.TxGasFeeDelegated
		gasFeePayer := params.TxValidationGasDefault

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}
}

// TestFeeDelegatedTransactionGasWithLegacyAndMultiSigPayer checks gas calculations
// using AccountKeyLegacy sender and AccountKeyWeightedMultiSig fee payer for fee delegated transaction types such as:
// 1. TxTypeFeeDelegatedValueTransfer
// 2. TxTypeFeeDelegatedValueTransferMemo with non-zero values.
// 3. TxTypeFeeDelegatedValueTransferMemo with zero values.
// 4. TxTypeFeeDelegatedSmartContractDeploy
// 5. TxTypeFeeDelegatedSmartContractExecution
// 6. TxTypeFeeDelegatedCancel
func TestFeeDelegatedTransactionGasWithLegacyAndMultiSigPayer(t *testing.T) {
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

	multisig, err := createMultisigAccount(uint(2),
		[]uint{1, 1, 1},
		[]string{"bb113e82881499a7a361e8354a5b68f6c6885c7bcba09ea2b0891480396c322e",
			"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae59438e989",
			"c32c471b732e2f56103e2f8e8cfd52792ef548f05f326e546a7d1fbf9d0419ec"},
		common.HexToAddress("0xbbfa38050bf3167c887c086758f448ce067ea8ea"))

	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// Preparing step. Create a multiSig account.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            multisig.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    multisig.AccKey,
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

	// Preparing step. Deploy a smart contract.
	var abiStr string
	var code string

	if isCompilerAvailable() {
		filename := string("../contracts/reward/contract/KlaytnReward.sol")
		codes, abistrings := compileSolidity(filename)
		code = codes[0]
		abiStr = abistrings[0]
	} else {
		// Falling back to use compiled code.
		code = "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"
		abiStr = `[{"constant":true,"inputs":[],"name":"totalAmount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"receiver","type":"address"}],"name":"reward","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"safeWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"}]`
	}

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

	// 1. TxTypeFeeDelegatedValueTransfer
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       multisig.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyFeePayer: multisig.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGas + params.TxGasFeeDelegated
		gasFeePayer := params.TxValidationGasDefault + uint64(len(multisig.Keys))*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 2. TxTypeFeeDelegatedValueTransferMemo with non-zero values.
	{
		data := []byte{1, 2, 3, 4}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       multisig.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: multisig.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataNonZeroGas
		intrinsicGas := params.TxGas + gasPayload + params.TxGasFeeDelegated
		gasFeePayer := params.TxValidationGasDefault + uint64(len(multisig.Keys))*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 3. TxTypeFeeDelegatedValueTransferMemo with zero values.
	{
		data := []byte{0, 0, 0, 0}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       multisig.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: multisig.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataZeroGas
		intrinsicGas := params.TxGas + gasPayload + params.TxGasFeeDelegated
		gasFeePayer := params.TxValidationGasDefault + uint64(len(multisig.Keys))*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 4. TxTypeFeeDelegatedSmartContractDeploy
	{
		amount := new(big.Int).SetUint64(0)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            common.HexToAddress("12345678"),
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyFeePayer:      multisig.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(common.FromHex(code), true, true)
		assert.Equal(t, nil, err)

		intrinsicGas = intrinsicGas + params.TxGasFeeDelegated
		executionGas := uint64(0x175fd)
		gasFeePayer := params.TxValidationGasDefault + uint64(len(multisig.Keys))*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+executionGas+gasFeePayer, gas)
	}

	// 5. TxTypeFeeDelegatedSmartContractExecution
	{
		amount := new(big.Int).SetUint64(10)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", reservoir.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       contract.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: multisig.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(data, false, true)
		assert.Equal(t, nil, err)

		intrinsicGas = intrinsicGas + params.TxGasFeeDelegated
		executionGas := uint64(0x9ec4)
		gasFeePayer := params.TxValidationGasDefault + uint64(len(multisig.Keys))*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+executionGas+gasFeePayer, gas)
	}

	// 6. TxTypeFeeDelegatedCancel
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyFeePayer: multisig.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedCancel, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGasCancel + params.TxGasFeeDelegated
		gasFeePayer := params.TxValidationGasDefault + uint64(len(multisig.Keys))*params.TxValidationGasPerKey

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}
}

// TestFeeDelegatedTransactionGasWithLegacyAndRoleBasedWithPublicPayer checks gas calculations
// using AccountKeyLegacy sender and AccountKeyRoleBased with AccountKeyPublic fee payer for fee delegated transaction types such as:
// 1. TxTypeFeeDelegatedValueTransfer
// 2. TxTypeFeeDelegatedValueTransferMemo with non-zero values.
// 3. TxTypeFeeDelegatedValueTransferMemo with zero values.
// 4. TxTypeFeeDelegatedSmartContractDeploy
// 5. TxTypeFeeDelegatedSmartContractExecution
// 6. TxTypeFeeDelegatedCancel
func TestFeeDelegatedTransactionGasWithLegacyAndRoleBasedWithPublicPayer(t *testing.T) {
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

	// get humanreadable address for role based account
	roleBasedAddr, err := common.FromHumanReadableAddress("roleBasedAddr")
	assert.Equal(t, nil, err)

	// role based account
	roleBased, err := createRoleBasedAccountWithAccountKeyPublic(
		[]string{"98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab",
			"c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
			"41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867"}, roleBasedAddr)
	assert.Equal(t, nil, err)

	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// Preparing step. Create a role based account with accountKeyPublic.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            roleBased.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleBased.AccKey,
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

	// Preparing step. Deploy a smart contract.
	var abiStr string
	var code string

	if isCompilerAvailable() {
		filename := string("../contracts/reward/contract/KlaytnReward.sol")
		codes, abistrings := compileSolidity(filename)
		code = codes[0]
		abiStr = abistrings[0]
	} else {
		// Falling back to use compiled code.
		code = "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"
		abiStr = `[{"constant":true,"inputs":[],"name":"totalAmount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"receiver","type":"address"}],"name":"reward","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"safeWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"}]`
	}

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

	// 1. TxTypeFeeDelegatedValueTransfer
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       roleBased.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGas + params.TxGasFeeDelegated
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFeePayer := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 2. TxTypeFeeDelegatedValueTransferMemo with non-zero values.
	{
		data := []byte{1, 2, 3, 4}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       roleBased.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataNonZeroGas
		intrinsicGas := params.TxGas + gasPayload + params.TxGasFeeDelegated
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFeePayer := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 3. TxTypeFeeDelegatedValueTransferMemo with zero values.
	{
		data := []byte{0, 0, 0, 0}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       roleBased.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataZeroGas
		intrinsicGas := params.TxGas + gasPayload + params.TxGasFeeDelegated
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFeePayer := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 4. TxTypeFeeDelegatedSmartContractDeploy
	{
		amount := new(big.Int).SetUint64(0)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            common.HexToAddress("12345678"),
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyFeePayer:      roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(common.FromHex(code), true, true)
		assert.Equal(t, nil, err)

		intrinsicGas = intrinsicGas + params.TxGasFeeDelegated
		executionGas := uint64(0x175fd)
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFeePayer := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+executionGas+gasFeePayer, gas)
	}

	// 5. TxTypeFeeDelegatedSmartContractExecution
	{
		amount := new(big.Int).SetUint64(10)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", reservoir.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       contract.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(data, false, true)
		assert.Equal(t, nil, err)

		intrinsicGas = intrinsicGas + params.TxGasFeeDelegated
		executionGas := uint64(0x9ec4)
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFeePayer := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+executionGas+gasFeePayer, gas)
	}

	// 6. TxTypeFeeDelegatedCancel
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedCancel, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGasCancel + params.TxGasFeeDelegated
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFeePayer := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}
}

// TestFeeDelegatedTransactionGasWithLegacyAndRoleBasedWithdMultiSigPayer checks gas calculations
// using AccountKeyLegacy sender and AccountKeyRoleBased with AccountKeyWeightedMultiSig fee payer for fee delegated transaction types such as:
// 1. TxTypeFeeDelegatedValueTransfer
// 2. TxTypeFeeDelegatedValueTransferMemo with non-zero values.
// 3. TxTypeFeeDelegatedValueTransferMemo with zero values.
// 4. TxTypeFeeDelegatedSmartContractDeploy
// 5. TxTypeFeeDelegatedSmartContractExecution
// 6. TxTypeFeeDelegatedCancel
func TestFeeDelegatedTransactionGasWithLegacyAndRoleBasedWithdMultiSigPayer(t *testing.T) {
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

	param1 := TestCreateMultisigAccountParam{
		Threshold: uint(2),
		Weights:   []uint{1, 1, 1},
		PrvKeys: []string{"bb113e82881499a7a361e8354a5b68f6c6885c7bcba09ea2b0891480396c322e",
			"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae59438e989",
			"c32c471b732e2f56103e2f8e8cfd52792ef548f05f326e546a7d1fbf9d0419ec"},
	}

	param2 := TestCreateMultisigAccountParam{
		Threshold: uint(2),
		Weights:   []uint{1, 1, 1},
		PrvKeys: []string{"98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab",
			"c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
			"41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867"},
	}

	param3 := TestCreateMultisigAccountParam{
		Threshold: uint(2),
		Weights:   []uint{1, 1, 1},
		PrvKeys: []string{"ed580f5bd71a2ee4dae5cb43e331b7d0318596e561e6add7844271ed94156b20",
			"98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab",
			"c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c"},
	}

	// get humanreadable address for role based account
	roleBasedAddr, err := common.FromHumanReadableAddress("roleBasedAddr")
	assert.Equal(t, nil, err)

	// role based account
	roleBased, err := createRoleBasedAccountWithAccountKeyWeightedMultiSig(
		[]TestCreateMultisigAccountParam{param1, param2, param3}, roleBasedAddr)
	assert.Equal(t, nil, err)

	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// Preparing step. Create a role based account with accountKeyPublic.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            roleBased.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleBased.AccKey,
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

	// Preparing step. Deploy a smart contract.
	var abiStr string
	var code string

	if isCompilerAvailable() {
		filename := string("../contracts/reward/contract/KlaytnReward.sol")
		codes, abistrings := compileSolidity(filename)
		code = codes[0]
		abiStr = abistrings[0]
	} else {
		// Falling back to use compiled code.
		code = "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"
		abiStr = `[{"constant":true,"inputs":[],"name":"totalAmount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"receiver","type":"address"}],"name":"reward","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"safeWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"}]`
	}

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

	// 1. TxTypeFeeDelegatedValueTransfer
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       roleBased.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGas + params.TxGasFeeDelegated
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFeePayer := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 2. TxTypeFeeDelegatedValueTransferMemo with non-zero values.
	{
		data := []byte{1, 2, 3, 4}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       roleBased.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataNonZeroGas
		intrinsicGas := params.TxGas + gasPayload + params.TxGasFeeDelegated
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFeePayer := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 3. TxTypeFeeDelegatedValueTransferMemo with zero values.
	{
		data := []byte{0, 0, 0, 0}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       roleBased.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataZeroGas
		intrinsicGas := params.TxGas + gasPayload + params.TxGasFeeDelegated
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFeePayer := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}

	// 4. TxTypeFeeDelegatedSmartContractDeploy
	{
		amount := new(big.Int).SetUint64(0)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            common.HexToAddress("12345678"),
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyFeePayer:      roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(common.FromHex(code), true, true)
		assert.Equal(t, nil, err)

		intrinsicGas = intrinsicGas + params.TxGasFeeDelegated
		executionGas := uint64(0x175fd)
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFeePayer := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+executionGas+gasFeePayer, gas)
	}

	// 5. TxTypeFeeDelegatedSmartContractExecution
	{
		amount := new(big.Int).SetUint64(10)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", reservoir.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       contract.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(data, false, true)
		assert.Equal(t, nil, err)

		intrinsicGas = intrinsicGas + params.TxGasFeeDelegated
		executionGas := uint64(0x9ec4)
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFeePayer := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+executionGas+gasFeePayer, gas)
	}

	// 6. TxTypeFeeDelegatedCancel
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedCancel, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGasCancel + params.TxGasFeeDelegated
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		sigValidationGasWithTxKeys := params.TxValidationGasDefault + uint64(len(roleBased.TxKeys))*params.TxValidationGasPerKey
		sigValidationGasWithUpdateKeys := params.TxValidationGasDefault + uint64(len(roleBased.UpdateKeys))*params.TxValidationGasPerKey
		sigValidationGasWithFeeKeys := params.TxValidationGasDefault + uint64(len(roleBased.FeeKeys))*params.TxValidationGasPerKey
		gasFeePayer := sigValidationGasWithTxKeys + sigValidationGasWithUpdateKeys + sigValidationGasWithFeeKeys

		assert.Equal(t, intrinsicGas+gasFeePayer, gas)
	}
}
