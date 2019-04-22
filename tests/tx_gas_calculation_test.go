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
	"github.com/ground-x/klaytn/accounts/abi"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/profile"
	"github.com/ground-x/klaytn/params"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
	"time"
)

var code = "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"

type TestAccount interface {
	GetAddr() common.Address
	GetTxKeys() []*ecdsa.PrivateKey
	GetUpdateKeys() []*ecdsa.PrivateKey
	GetFeeKeys() []*ecdsa.PrivateKey
	GetNonce() uint64
	GetAccKey() accountkey.AccountKey
	GetValidationGas(r accountkey.RoleType) uint64
	AddNonce()
}

type genTransaction func(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64)

func TestGasCalculation(t *testing.T) {
	var testFunctions = []struct {
		Name  string
		genTx genTransaction
	}{
		{"LegacyTransaction", genLegacyTransaction},
		{"ValueTransfer", genValueTransfer},
		{"ValueTransferWithMemo", genValueTransferWithMemo},
		{"AccountCreation", genAccountCreation},
		{"AccountUpdate", genAccountUpdate},
		{"SmartContractDeploy", genSmartContractDeploy},
		{"SmartContractDeployWithNil", genSmartContractDeploy},
		{"SmartContractExecution", genSmartContractExecution},
		{"Cancel", genCancel},
		{"ChainDataAnchoring", genChainDataAnchoring},
		{"FeeDelegatedValueTransfer", genFeeDelegatedValueTransfer},
		{"FeeDelegatedValueTransferWithMemo", genFeeDelegatedValueTransferWithMemo},
		{"FeeDelegatedAccountUpdate", genFeeDelegatedAccountUpdate},
		{"FeeDelegatedSmartContractDeploy", genFeeDelegatedSmartContractDeploy},
		{"FeeDelegatedSmartContractDeployWithNil", genFeeDelegatedSmartContractDeploy},
		{"FeeDelegatedSmartContractExecution", genFeeDelegatedSmartContractExecution},
		{"FeeDelegatedCancel", genFeeDelegatedCancel},
		{"FeeDelegatedWithRatioValueTransfer", genFeeDelegatedWithRatioValueTransfer},
		{"FeeDelegatedWithRatioValueTransferWithMemo", genFeeDelegatedWithRatioValueTransferWithMemo},
		{"FeeDelegatedWithRatioAccountUpdate", genFeeDelegatedWithRatioAccountUpdate},
		{"FeeDelegatedWithRatioSmartContractDeploy", genFeeDelegatedWithRatioSmartContractDeploy},
		{"FeeDelegatedWithRatioSmartContractDeployWithNil", genFeeDelegatedWithRatioSmartContractDeploy},
		{"FeeDelegatedWithRatioSmartContractExecution", genFeeDelegatedWithRatioSmartContractExecution},
		{"FeeDelegatedWithRatioCancel", genFeeDelegatedWithRatioCancel},
	}

	var accountTypes = []struct {
		Type    string
		account TestAccount
	}{
		{"Legacy", genLegacyAccount(t)},
		{"KlaytnLegacy", genKlaytnLegacyAccount(t)},
		{"Public", genPublicAccount(t)},
		{"MultiSig", genMultiSigAccount(t)},
		{"RoleBasedWithPublic", genRoleBasedWithPublicAccount(t)},
		{"RoleBasedWithMultiSig", genRoleBasedWithMultiSigAccount(t)},
	}

	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(6, 4)
	assert.Equal(t, nil, err)
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
	var reservoir TestAccount
	reservoir = &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// Preparing step. Send Klay to LegacyAccount.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(params.KLAY)
		tx := types.NewTransaction(reservoir.GetNonce(),
			accountTypes[0].account.GetAddr(), amount, gasLimit, gasPrice, []byte{})

		err := tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.AddNonce()
	}

	// Preparing step. Send Klay to KlaytnAcounts.
	for i := 1; i < len(accountTypes); i++ {
		var txs types.Transactions

		amount := new(big.Int).SetUint64(params.KLAY)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            accountTypes[i].account.GetAddr(),
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    accountTypes[i].account.GetAccKey(),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.AddNonce()
	}

	// For smart contract
	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(0)
		to := contract.GetAddr()

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            &to,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          common.FromHex(code),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.AddNonce()
	}

	for _, f := range testFunctions {
		for _, sender := range accountTypes {
			toAccount := reservoir
			senderRole := accountkey.RoleTransaction

			// LegacyTransaction can be used only with LegacyAccount and KlaytnAccount with AccountKeyLegacy.
			if !strings.Contains(sender.Type, "Legacy") && strings.Contains(f.Name, "Legacy") {
				continue
			}

			if strings.Contains(f.Name, "AccountUpdate") {
				// Sender can't be a LegacyAccount with AccountUpdate
				if sender.Type == "Legacy" {
					continue
				}
				senderRole = accountkey.RoleAccountUpdate
			}

			// Set contract's address with SmartContractExecution
			if strings.Contains(f.Name, "SmartContractExecution") {
				toAccount = contract
			} else if strings.Contains(f.Name, "WithNil") {
				toAccount = nil
			}

			if !strings.Contains(f.Name, "FeeDelegated") {
				// For NonFeeDelegated Transactions
				Name := f.Name + "/" + sender.Type + "Sender"
				t.Run(Name, func(t *testing.T) {
					tx, intrinsic := f.genTx(t, signer, sender.account, toAccount, nil, gasPrice)
					acocuntValidationGas := sender.account.GetValidationGas(senderRole)
					testGasValidation(t, bcdata, tx, intrinsic+acocuntValidationGas)
				})
			} else {
				// For FeeDelegated(WithRatio) Transactions
				for _, payer := range accountTypes {
					Name := f.Name + "/" + sender.Type + "Sender/" + payer.Type + "Payer"
					t.Run(Name, func(t *testing.T) {
						tx, intrinsic := f.genTx(t, signer, sender.account, toAccount, payer.account, gasPrice)
						acocuntsValidationGas := sender.account.GetValidationGas(senderRole) + payer.account.GetValidationGas(accountkey.RoleFeePayer)
						testGasValidation(t, bcdata, tx, intrinsic+acocuntsValidationGas)
					})
				}
			}

		}
	}
}

func testGasValidation(t *testing.T, bcdata *BCData, tx *types.Transaction, validationGas uint64) {
	receipt, gas, err := applyTransaction(t, bcdata, tx)
	assert.Equal(t, nil, err)

	assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

	assert.Equal(t, validationGas, gas)
}

func genLegacyTransaction(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	amount := big.NewInt(100000)
	tx := types.NewTransaction(from.GetNonce(), to.GetAddr(), amount, gasLimit, gasPrice, []byte{})

	err := tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	intrinsic := params.TxGas

	return tx, intrinsic
}

func genValueTransfer(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	values := genMapForValueTransfer(from, to, gasPrice)
	tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	intrinsic := params.TxGas

	return tx, intrinsic
}

func genFeeDelegatedValueTransfer(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	values := genMapForValueTransfer(from, to, gasPrice)
	values[types.TxValueKeyFeePayer] = payer.GetAddr()

	tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, values)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	err = tx.SignFeePayerWithKeys(signer, payer.GetFeeKeys())
	assert.Equal(t, nil, err)

	intrinsic := params.TxGas + params.TxGasFeeDelegated

	return tx, intrinsic
}

func genFeeDelegatedWithRatioValueTransfer(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	values := genMapForValueTransfer(from, to, gasPrice)
	values[types.TxValueKeyFeePayer] = payer.GetAddr()
	values[types.TxValueKeyFeeRatioOfFeePayer] = types.FeeRatio(30)

	tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferWithRatio, values)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	err = tx.SignFeePayerWithKeys(signer, payer.GetFeeKeys())
	assert.Equal(t, nil, err)

	intrinsic := params.TxGas + params.TxGasFeeDelegatedWithRatio

	return tx, intrinsic
}

func genValueTransferWithMemo(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	values, gasPayload := genMapForValueTransferWithMemo(from, to, gasPrice)

	tx, err := types.NewTransactionWithMap(types.TxTypeValueTransferMemo, values)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	intrinsic := params.TxGas + gasPayload

	return tx, intrinsic
}

func genFeeDelegatedValueTransferWithMemo(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	values, gasPayload := genMapForValueTransferWithMemo(from, to, gasPrice)
	values[types.TxValueKeyFeePayer] = payer.GetAddr()

	tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	err = tx.SignFeePayerWithKeys(signer, payer.GetFeeKeys())
	assert.Equal(t, nil, err)

	intrinsic := params.TxGas + gasPayload + params.TxGasFeeDelegated

	return tx, intrinsic
}

func genFeeDelegatedWithRatioValueTransferWithMemo(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	values, gasPayload := genMapForValueTransferWithMemo(from, to, gasPrice)
	values[types.TxValueKeyFeePayer] = payer.GetAddr()
	values[types.TxValueKeyFeeRatioOfFeePayer] = types.FeeRatio(30)

	tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemoWithRatio, values)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	err = tx.SignFeePayerWithKeys(signer, payer.GetFeeKeys())
	assert.Equal(t, nil, err)

	intrinsic := params.TxGas + gasPayload + params.TxGasFeeDelegatedWithRatio

	return tx, intrinsic
}

func genAccountCreation(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	newAccount, gasKey, readable := genNewAccountWithGas(t, from)

	amount := big.NewInt(100000)
	tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, map[types.TxValueKeyType]interface{}{
		types.TxValueKeyNonce:         from.GetNonce(),
		types.TxValueKeyFrom:          from.GetAddr(),
		types.TxValueKeyTo:            newAccount.GetAddr(),
		types.TxValueKeyAmount:        amount,
		types.TxValueKeyGasLimit:      gasLimit,
		types.TxValueKeyGasPrice:      gasPrice,
		types.TxValueKeyHumanReadable: readable,
		types.TxValueKeyAccountKey:    newAccount.GetAccKey(),
	})
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	intrinsic := params.TxGasAccountCreation + gasKey

	return tx, intrinsic
}

func genAccountUpdate(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	newAccount, gasKey, _ := genNewAccountWithGas(t, from)

	values := genMapForUpdate(from, to, gasPrice, newAccount.GetAccKey())

	tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetUpdateKeys())
	assert.Equal(t, nil, err)

	intrinsic := params.TxGasAccountUpdate + gasKey

	return tx, intrinsic
}

func genFeeDelegatedAccountUpdate(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	newAccount, gasKey, _ := genNewAccountWithGas(t, from)

	values := genMapForUpdate(from, to, gasPrice, newAccount.GetAccKey())
	values[types.TxValueKeyFeePayer] = payer.GetAddr()

	tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdate, values)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetUpdateKeys())
	assert.Equal(t, nil, err)

	err = tx.SignFeePayerWithKeys(signer, payer.GetFeeKeys())
	assert.Equal(t, nil, err)

	intrinsic := params.TxGasAccountUpdate + gasKey + params.TxGasFeeDelegated

	return tx, intrinsic
}

func genFeeDelegatedWithRatioAccountUpdate(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	newAccount, gasKey, _ := genNewAccountWithGas(t, from)

	values := genMapForUpdate(from, to, gasPrice, newAccount.GetAccKey())
	values[types.TxValueKeyFeePayer] = payer.GetAddr()
	values[types.TxValueKeyFeeRatioOfFeePayer] = types.FeeRatio(30)

	tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdateWithRatio, values)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetUpdateKeys())
	assert.Equal(t, nil, err)

	err = tx.SignFeePayerWithKeys(signer, payer.GetFeeKeys())
	assert.Equal(t, nil, err)

	intrinsic := params.TxGasAccountUpdate + gasKey + params.TxGasFeeDelegatedWithRatio

	return tx, intrinsic
}

func genSmartContractDeploy(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	values, intrinsicGas := genMapForDeploy(t, from, to, gasPrice, types.TxTypeSmartContractDeploy)

	tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	return tx, intrinsicGas
}

func genFeeDelegatedSmartContractDeploy(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	values, intrinsicGas := genMapForDeploy(t, from, to, gasPrice, types.TxTypeFeeDelegatedSmartContractDeploy)
	values[types.TxValueKeyFeePayer] = payer.GetAddr()

	tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, values)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	err = tx.SignFeePayerWithKeys(signer, payer.GetFeeKeys())
	assert.Equal(t, nil, err)

	return tx, intrinsicGas
}

func genFeeDelegatedWithRatioSmartContractDeploy(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	values, intrinsicGas := genMapForDeploy(t, from, to, gasPrice, types.TxTypeFeeDelegatedSmartContractDeployWithRatio)
	values[types.TxValueKeyFeePayer] = payer.GetAddr()
	values[types.TxValueKeyFeeRatioOfFeePayer] = types.FeeRatio(30)

	tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeployWithRatio, values)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	err = tx.SignFeePayerWithKeys(signer, payer.GetFeeKeys())
	assert.Equal(t, nil, err)

	return tx, intrinsicGas
}

func genSmartContractExecution(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	values, intrinsicGas := genMapForExecution(t, from, to, gasPrice, types.TxTypeSmartContractExecution)

	tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, values)

	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	return tx, intrinsicGas
}

func genFeeDelegatedSmartContractExecution(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	values, intrinsicGas := genMapForExecution(t, from, to, gasPrice, types.TxTypeFeeDelegatedSmartContractExecution)
	values[types.TxValueKeyFeePayer] = payer.GetAddr()

	tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractExecution, values)

	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	err = tx.SignFeePayerWithKeys(signer, payer.GetFeeKeys())
	assert.Equal(t, nil, err)

	return tx, intrinsicGas
}

func genFeeDelegatedWithRatioSmartContractExecution(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	values, intrinsicGas := genMapForExecution(t, from, to, gasPrice, types.TxTypeFeeDelegatedSmartContractExecutionWithRatio)
	values[types.TxValueKeyFeePayer] = payer.GetAddr()
	values[types.TxValueKeyFeeRatioOfFeePayer] = types.FeeRatio(30)

	tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractExecutionWithRatio, values)

	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	err = tx.SignFeePayerWithKeys(signer, payer.GetFeeKeys())
	assert.Equal(t, nil, err)

	return tx, intrinsicGas
}

func genCancel(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	values := genMapForCancel(from, gasPrice)

	tx, err := types.NewTransactionWithMap(types.TxTypeCancel, values)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	intrinsic := params.TxGasCancel

	return tx, intrinsic
}

func genFeeDelegatedCancel(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	values := genMapForCancel(from, gasPrice)
	values[types.TxValueKeyFeePayer] = payer.GetAddr()

	tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedCancel, values)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	err = tx.SignFeePayerWithKeys(signer, payer.GetFeeKeys())
	assert.Equal(t, nil, err)

	intrinsic := params.TxGasCancel + params.TxGasFeeDelegated

	return tx, intrinsic
}

func genFeeDelegatedWithRatioCancel(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	values := genMapForCancel(from, gasPrice)
	values[types.TxValueKeyFeePayer] = payer.GetAddr()
	values[types.TxValueKeyFeeRatioOfFeePayer] = types.FeeRatio(30)

	tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedCancelWithRatio, values)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	err = tx.SignFeePayerWithKeys(signer, payer.GetFeeKeys())
	assert.Equal(t, nil, err)

	intrinsic := params.TxGasCancel + params.TxGasFeeDelegatedWithRatio

	return tx, intrinsic
}

func genChainDataAnchoring(t *testing.T, signer types.Signer, from TestAccount, to TestAccount, payer TestAccount, gasPrice *big.Int) (*types.Transaction, uint64) {
	anchoredData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	tx, err := types.NewTransactionWithMap(types.TxTypeChainDataAnchoring, map[types.TxValueKeyType]interface{}{
		types.TxValueKeyNonce:        from.GetNonce(),
		types.TxValueKeyFrom:         from.GetAddr(),
		types.TxValueKeyGasLimit:     gasLimit,
		types.TxValueKeyGasPrice:     gasPrice,
		types.TxValueKeyAnchoredData: anchoredData,
	})

	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, from.GetTxKeys())
	assert.Equal(t, nil, err)

	gasAnchoring := params.ChainDataAnchoringGas * (uint64)(len(anchoredData))
	intrinsic := params.TxChainDataAnchoringGas + gasAnchoring

	return tx, intrinsic
}

// Generate map functions
func genMapForValueTransfer(from TestAccount, to TestAccount, gasPrice *big.Int) map[types.TxValueKeyType]interface{} {
	amount := big.NewInt(100000)

	values := map[types.TxValueKeyType]interface{}{
		types.TxValueKeyNonce:    from.GetNonce(),
		types.TxValueKeyFrom:     from.GetAddr(),
		types.TxValueKeyTo:       to.GetAddr(),
		types.TxValueKeyAmount:   amount,
		types.TxValueKeyGasLimit: gasLimit,
		types.TxValueKeyGasPrice: gasPrice,
	}
	return values
}

func genMapForValueTransferWithMemo(from TestAccount, to TestAccount, gasPrice *big.Int) (map[types.TxValueKeyType]interface{}, uint64) {
	nonZeroData := []byte{1, 2, 3, 4}
	zeroData := []byte{0, 0, 0, 0}

	data := append(nonZeroData, zeroData...)

	amount := big.NewInt(100000)

	values := map[types.TxValueKeyType]interface{}{
		types.TxValueKeyNonce:    from.GetNonce(),
		types.TxValueKeyFrom:     from.GetAddr(),
		types.TxValueKeyTo:       to.GetAddr(),
		types.TxValueKeyAmount:   amount,
		types.TxValueKeyGasLimit: gasLimit,
		types.TxValueKeyGasPrice: gasPrice,
		types.TxValueKeyData:     data,
	}

	gasPayload := uint64(len(nonZeroData))*params.TxDataNonZeroGas + uint64(len(zeroData))*params.TxDataZeroGas

	return values, gasPayload
}

func genMapForUpdate(from TestAccount, to TestAccount, gasPrice *big.Int, newKeys accountkey.AccountKey) map[types.TxValueKeyType]interface{} {
	values := map[types.TxValueKeyType]interface{}{
		types.TxValueKeyNonce:      from.GetNonce(),
		types.TxValueKeyFrom:       from.GetAddr(),
		types.TxValueKeyGasLimit:   gasLimit,
		types.TxValueKeyGasPrice:   gasPrice,
		types.TxValueKeyAccountKey: newKeys,
	}
	return values
}

func genMapForDeploy(t *testing.T, from TestAccount, to TestAccount, gasPrice *big.Int, txType types.TxType) (map[types.TxValueKeyType]interface{}, uint64) {
	amount := new(big.Int).SetUint64(0)
	values := map[types.TxValueKeyType]interface{}{
		types.TxValueKeyNonce:         from.GetNonce(),
		types.TxValueKeyAmount:        amount,
		types.TxValueKeyGasLimit:      gasLimit,
		types.TxValueKeyGasPrice:      gasPrice,
		types.TxValueKeyHumanReadable: false,
		types.TxValueKeyFrom:          from.GetAddr(),
		types.TxValueKeyData:          common.FromHex(code),
	}

	if to == nil {
		values[types.TxValueKeyTo] = (*common.Address)(nil)
	} else {
		addr := common.HexToAddress("12345678")
		values[types.TxValueKeyTo] = &addr
	}

	intrinsicGas := getIntrinsicGas(txType)
	intrinsicGas += uint64(0x175fd)

	gasPayloadWithGas, err := types.IntrinsicGasPayload(intrinsicGas, common.FromHex(code))
	assert.Equal(t, nil, err)

	return values, gasPayloadWithGas
}

func genMapForExecution(t *testing.T, from TestAccount, to TestAccount, gasPrice *big.Int, txType types.TxType) (map[types.TxValueKeyType]interface{}, uint64) {
	abiStr := `[{"constant":true,"inputs":[],"name":"totalAmount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"receiver","type":"address"}],"name":"reward","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"safeWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"}]`
	abii, err := abi.JSON(strings.NewReader(string(abiStr)))
	assert.Equal(t, nil, err)

	data, err := abii.Pack("reward", to.GetAddr())
	assert.Equal(t, nil, err)

	amount := new(big.Int).SetUint64(10)

	values := map[types.TxValueKeyType]interface{}{
		types.TxValueKeyNonce:    from.GetNonce(),
		types.TxValueKeyFrom:     from.GetAddr(),
		types.TxValueKeyTo:       to.GetAddr(),
		types.TxValueKeyAmount:   amount,
		types.TxValueKeyGasLimit: gasLimit,
		types.TxValueKeyGasPrice: gasPrice,
		types.TxValueKeyData:     data,
	}

	intrinsicGas := getIntrinsicGas(txType)
	intrinsicGas += uint64(0x9ec4)

	gasPayloadWithGas, err := types.IntrinsicGasPayload(intrinsicGas, data)
	assert.Equal(t, nil, err)

	return values, gasPayloadWithGas
}

func genMapForCancel(from TestAccount, gasPrice *big.Int) map[types.TxValueKeyType]interface{} {
	values := map[types.TxValueKeyType]interface{}{
		types.TxValueKeyNonce:    from.GetNonce(),
		types.TxValueKeyFrom:     from.GetAddr(),
		types.TxValueKeyGasLimit: gasLimit,
		types.TxValueKeyGasPrice: gasPrice,
	}
	return values
}

// Generate TestAccount functions
func genLegacyAccount(t *testing.T) TestAccount {
	// For LegacyAccount
	legacyAccount, err := createAnonymousAccount("c32c471b732e2f56103e2f8e8cfd52792ef548f05f326e546a7d1fbf9d0419ec")
	assert.Equal(t, nil, err)

	return legacyAccount
}

func genKlaytnLegacyAccount(t *testing.T) TestAccount {
	// For KlaytnLegacy
	klaytnLegacy, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	return klaytnLegacy
}

func genPublicAccount(t *testing.T) TestAccount {
	// For AccountKeyPublic
	publicAccount, err := createDecoupledAccount("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
		common.HexToAddress("0x75c3098be5e4b63fbac05838daaee378dd48098d"))
	assert.Equal(t, nil, err)

	return publicAccount
}

func genMultiSigAccount(t *testing.T) TestAccount {
	// For AccountKeyWeightedMultiSig
	multisigAccount, err := createMultisigAccount(uint(2),
		[]uint{1, 1, 1},
		[]string{"bb113e82881499a7a361e8354a5b68f6c6885c7bcba09ea2b0891480396c322e",
			"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae59438e989",
			"c32c471b732e2f56103e2f8e8cfd52792ef548f05f326e546a7d1fbf9d0419ec"},
		common.HexToAddress("0xbbfa38050bf3167c887c086758f448ce067ea8ea"))
	assert.Equal(t, nil, err)

	return multisigAccount
}

func genRoleBasedWithPublicAccount(t *testing.T) TestAccount {
	// For AccountKeyRoleBased With AccountKeyPublic
	roleBasedWithPublicAddr, err := common.FromHumanReadableAddress("roleBasedPublic")
	assert.Equal(t, nil, err)

	roleBasedWithPublic, err := createRoleBasedAccountWithAccountKeyPublic(
		[]string{"98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab",
			"c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
			"41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867"}, roleBasedWithPublicAddr)
	assert.Equal(t, nil, err)

	return roleBasedWithPublic
}

func genRoleBasedWithMultiSigAccount(t *testing.T) TestAccount {
	// For AccountKeyRoleBased With AccountKeyWeightedMultiSig
	p := genMultiSigParamForRoleBased()

	roleBasedWithMultiSigAddr, err := common.FromHumanReadableAddress("roleBasedMultiSig")
	assert.Equal(t, nil, err)

	roleBasedWithMultiSig, err := createRoleBasedAccountWithAccountKeyWeightedMultiSig(
		[]TestCreateMultisigAccountParam{p[0], p[1], p[2]}, roleBasedWithMultiSigAddr)
	assert.Equal(t, nil, err)

	return roleBasedWithMultiSig
}

// Generate new Account functions for testing AccountCreation and AccountUpdate
func genNewAccountWithGas(t *testing.T, testAccount TestAccount) (TestAccount, uint64, bool) {
	var newAccount TestAccount
	gas := uint64(0)
	readable := false

	// AccountKeyLegacy
	if testAccount.GetAccKey() == nil || testAccount.GetAccKey().Type() == accountkey.AccountKeyTypeLegacy {
		anon, err := createAnonymousAccount("4c6482f6de5c70dc69b07ab378d3ad59f4c9bc030426fa686842e90296b2d2ea")
		assert.Equal(t, nil, err)

		return anon, gas, readable
	}

	// humanReadableAddress for newAccount
	newAccountAddress, err := common.FromHumanReadableAddress("newAccountAddress")
	assert.Equal(t, nil, err)
	readable = true

	switch testAccount.GetAccKey().Type() {
	case accountkey.AccountKeyTypePublic:
		publicAccount, err := createDecoupledAccount("4e5a6f01c67a6f795f74c06b9cdde15ee90b69b20f20c42f5481506cbbc54abe", newAccountAddress)
		assert.Equal(t, nil, err)

		newAccount = publicAccount
		gas += uint64(len(newAccount.GetTxKeys())) * params.TxAccountCreationGasPerKey
	case accountkey.AccountKeyTypeWeightedMultiSig:
		multisigAccount, err := createMultisigAccount(uint(2), []uint{1, 1, 1},
			[]string{"d5597077c93cebd487cd33d7268a99595dbb7811be8118f2d82d530208e77397",
				"c28959a6d845ab274b46c94edf08c5ba115fa3aaa6b1f23a8be4d78b948a7693",
				"ab4e26754e40c5fb115e09a076b33ba8bc4526b09e9461baa36ec7bd645b96e4"}, newAccountAddress)
		assert.Equal(t, nil, err)

		newAccount = multisigAccount
		gas += uint64(len(newAccount.GetTxKeys())) * params.TxAccountCreationGasPerKey
	case accountkey.AccountKeyTypeRoleBased:
		p := genMultiSigParamForRoleBased()

		newRoleBasedWithMultiSig, err := createRoleBasedAccountWithAccountKeyWeightedMultiSig(
			[]TestCreateMultisigAccountParam{p[2], p[0], p[1]}, newAccountAddress)
		assert.Equal(t, nil, err)

		newAccount = newRoleBasedWithMultiSig
		gas = uint64(len(newAccount.GetTxKeys())) * params.TxAccountCreationGasPerKey
		gas += uint64(len(newAccount.GetUpdateKeys())) * params.TxAccountCreationGasPerKey
		gas += uint64(len(newAccount.GetFeeKeys())) * params.TxAccountCreationGasPerKey
	}

	return newAccount, gas, readable
}

// Return multisig parameters for creating RoleBased with MultiSig
func genMultiSigParamForRoleBased() []TestCreateMultisigAccountParam {
	var params []TestCreateMultisigAccountParam
	param1 := TestCreateMultisigAccountParam{
		Threshold: uint(2),
		Weights:   []uint{1, 1, 1},
		PrvKeys: []string{"8dfd76bea5baf283b89026a97d72b07b015d9ab7552d03ace04c16ee7170b6dc",
			"7c2de9bc5273e626790ac1fd85fac3789fe1ba98e137ee745a57c2a27919b50b",
			"50e51200c969344a60ad8441536d86565f8094f82dbcdd799b9f52e99f268b06"},
	}
	params = append(params, param1)

	param2 := TestCreateMultisigAccountParam{
		Threshold: uint(2),
		Weights:   []uint{1, 1, 1},
		PrvKeys: []string{"d0b83aec6c8aa48333b6a1c5b6fe801f6e719adb7b7e74f6e2f4d6337c9fcfc6",
			"fef8f3687afed116e0227248b0d315207693d0e624b66661e4a68812e207ccd7",
			"f5383f13bd1aea8681b5d3177e07dace822f0fe79087afde13ab5bc32ae467f7"},
	}
	params = append(params, param2)

	param3 := TestCreateMultisigAccountParam{
		Threshold: uint(2),
		Weights:   []uint{1, 1, 1},
		PrvKeys: []string{"2740eb27d7f71d863b7a71d4261b95559501e3a1207fab6d053faf19ef6b4fb5",
			"626edd10bf26b5584673a702246d0fb1806bdc5f86c3e17aa912e30bd4bae07b",
			"8435615b2e09d3725a1cdc616e123b8ee49e5c5f1044188135a9a0a37e214c01"},
	}
	params = append(params, param3)

	return params
}

func getIntrinsicGas(txType types.TxType) uint64 {
	var intrinsic uint64

	switch txType {
	case types.TxTypeLegacyTransaction:
		intrinsic = params.TxGas
	case types.TxTypeValueTransfer:
		intrinsic = params.TxGas
	case types.TxTypeFeeDelegatedValueTransfer:
		intrinsic = params.TxGas + params.TxGasFeeDelegated
	case types.TxTypeFeeDelegatedValueTransferWithRatio:
		intrinsic = params.TxGas + params.TxGasFeeDelegatedWithRatio
	case types.TxTypeValueTransferMemo:
		intrinsic = params.TxGas
	case types.TxTypeFeeDelegatedValueTransferMemo:
		intrinsic = params.TxGas + params.TxGasFeeDelegated
	case types.TxTypeFeeDelegatedValueTransferMemoWithRatio:
		intrinsic = params.TxGas + params.TxGasFeeDelegatedWithRatio
	case types.TxTypeAccountCreation:
		intrinsic = params.TxGasAccountCreation
	case types.TxTypeAccountUpdate:
		intrinsic = params.TxGasAccountUpdate
	case types.TxTypeFeeDelegatedAccountUpdate:
		intrinsic = params.TxGasAccountUpdate + params.TxGasFeeDelegated
	case types.TxTypeFeeDelegatedAccountUpdateWithRatio:
		intrinsic = params.TxGasAccountUpdate + params.TxGasFeeDelegatedWithRatio
	case types.TxTypeSmartContractDeploy:
		intrinsic = params.TxGasContractCreation
	case types.TxTypeFeeDelegatedSmartContractDeploy:
		intrinsic = params.TxGasContractCreation + params.TxGasFeeDelegated
	case types.TxTypeFeeDelegatedSmartContractDeployWithRatio:
		intrinsic = params.TxGasContractCreation + params.TxGasFeeDelegatedWithRatio
	case types.TxTypeSmartContractExecution:
		intrinsic = params.TxGas
	case types.TxTypeFeeDelegatedSmartContractExecution:
		intrinsic = params.TxGas + params.TxGasFeeDelegated
	case types.TxTypeFeeDelegatedSmartContractExecutionWithRatio:
		intrinsic = params.TxGas + params.TxGasFeeDelegatedWithRatio
	case types.TxTypeChainDataAnchoring:
		intrinsic = params.TxChainDataAnchoringGas
	case types.TxTypeCancel:
		intrinsic = params.TxGasCancel
	case types.TxTypeFeeDelegatedCancel:
		intrinsic = params.TxGasCancel + params.TxGasFeeDelegated
	case types.TxTypeFeeDelegatedCancelWithRatio:
		intrinsic = params.TxGasCancel + params.TxGasFeeDelegatedWithRatio
	}

	return intrinsic
}

// Implement TestAccount interface for TestAccountType
func (t *TestAccountType) GetAddr() common.Address {
	return t.Addr
}

func (t *TestAccountType) GetTxKeys() []*ecdsa.PrivateKey {
	return t.Keys
}

func (t *TestAccountType) GetUpdateKeys() []*ecdsa.PrivateKey {
	return t.Keys
}

func (t *TestAccountType) GetFeeKeys() []*ecdsa.PrivateKey {
	return t.Keys
}

func (t *TestAccountType) GetNonce() uint64 {
	return t.Nonce
}

func (t *TestAccountType) GetAccKey() accountkey.AccountKey {
	return t.AccKey
}

// Return SigValidationGas depends on AccountType
func (t *TestAccountType) GetValidationGas(r accountkey.RoleType) uint64 {
	if t.GetAccKey() == nil {
		return 0
	}

	var gas uint64

	switch t.GetAccKey().Type() {
	case accountkey.AccountKeyTypeLegacy:
		gas = 0
	case accountkey.AccountKeyTypePublic:
		gas = (1 - 1) * params.TxValidationGasPerKey
	case accountkey.AccountKeyTypeWeightedMultiSig:
		gas = uint64(len(t.GetTxKeys())-1) * params.TxValidationGasPerKey
	}

	return gas
}

func (t *TestAccountType) AddNonce() {
	t.Nonce += 1
}

// Implement TestAccount interface for TestRoleBasedAccountType
func (t *TestRoleBasedAccountType) GetAddr() common.Address {
	return t.Addr
}

func (t *TestRoleBasedAccountType) GetTxKeys() []*ecdsa.PrivateKey {
	return t.TxKeys
}

func (t *TestRoleBasedAccountType) GetUpdateKeys() []*ecdsa.PrivateKey {
	return t.UpdateKeys
}

func (t *TestRoleBasedAccountType) GetFeeKeys() []*ecdsa.PrivateKey {
	return t.FeeKeys
}

func (t *TestRoleBasedAccountType) GetNonce() uint64 {
	return t.Nonce
}

func (t *TestRoleBasedAccountType) GetAccKey() accountkey.AccountKey {
	return t.AccKey
}

// Return SigValidationGas depends on AccountType
func (t *TestRoleBasedAccountType) GetValidationGas(r accountkey.RoleType) uint64 {
	if t.GetAccKey() == nil {
		return 0
	}

	var gas uint64

	switch r {
	case accountkey.RoleTransaction:
		gas = uint64(len(t.GetTxKeys())-1) * params.TxValidationGasPerKey
	case accountkey.RoleAccountUpdate:
		gas = uint64(len(t.GetUpdateKeys())-1) * params.TxValidationGasPerKey
	case accountkey.RoleFeePayer:
		gas = uint64(len(t.GetFeeKeys())-1) * params.TxValidationGasPerKey
	}

	return gas
}

func (t *TestRoleBasedAccountType) AddNonce() {
	t.Nonce += 1
}
