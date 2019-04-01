package tests

import (
	"crypto/ecdsa"
	"github.com/ground-x/klaytn/accounts/abi"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/profile"
	"github.com/ground-x/klaytn/params"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
	"time"
)

// TestFeeDelegatedTransactionGasWithKlaytnLegacyAndRoleBasedWithPublicPayer checks gas calculations
// using KlaytnAccount with AccountKeyLegacy sender and AccountKeyRoleBased with AccountKeyPublic fee payer
// for fee delegated transaction types such as:
// 1. TxTypeFeeDelegatedValueTransfer
// 2. TxTypeFeeDelegatedValueTransferMemo with non-zero values.
// 3. TxTypeFeeDelegatedValueTransferMemo with zero values.
// 4. TxTypeFeeDelegatedSmartContractDeploy
// 5. TxTypeFeeDelegatedSmartContractExecution
// 6. TxTypeFeeDelegatedCancel
func TestFeeDelegatedTransactionGasWithKlaytnLegacyAndRoleBasedWithPublicPayer(t *testing.T) {
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
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// Preparing step. Create an account KlaytnAccount with AccountKeyLegacy.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(params.KLAY)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    anon.AccKey,
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

	// Preparing step. Create an account KlaytnAccount with AccountKeyLegacy.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(params.KLAY)
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
			types.TxValueKeyNonce:    anon.Nonce,
			types.TxValueKeyFrom:     anon.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGas + params.TxGasFeeDelegated
		// TODO-Klaytn-Gas Need to revise gas fee calculation.
		gasFrom := params.TxValidationGasDefault
		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFeePayer := 3 * gasAccountKeyPublicForSigValidation

		assert.Equal(t, intrinsicGas+gasFrom+gasFeePayer, gas)
	}

	// 2. TxTypeFeeDelegatedValueTransferMemo with non-zero values.
	{
		data := []byte{1, 2, 3, 4}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    anon.Nonce,
			types.TxValueKeyFrom:     anon.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataNonZeroGas
		intrinsicGas := params.TxGas + gasPayload + params.TxGasFeeDelegated
		// TODO-Klaytn-Gas Need to revise gas fee calculation.
		gasFrom := params.TxValidationGasDefault
		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFeePayer := 3 * gasAccountKeyPublicForSigValidation

		assert.Equal(t, intrinsicGas+gasFrom+gasFeePayer, gas)
	}

	// 3. TxTypeFeeDelegatedValueTransferMemo with zero values.
	{
		data := []byte{0, 0, 0, 0}
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    anon.Nonce,
			types.TxValueKeyFrom:     anon.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   big.NewInt(100000),
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		gasPayload := uint64(len(data)) * params.TxDataZeroGas
		intrinsicGas := params.TxGas + gasPayload + params.TxGasFeeDelegated
		// TODO-Klaytn-Gas Need to revise gas fee calculation.
		gasFrom := params.TxValidationGasDefault
		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFeePayer := 3 * gasAccountKeyPublicForSigValidation

		assert.Equal(t, intrinsicGas+gasFrom+gasFeePayer, gas)
	}

	// 4. TxTypeFeeDelegatedSmartContractDeploy
	{
		amount := new(big.Int).SetUint64(0)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         anon.Nonce,
			types.TxValueKeyFrom:          anon.Addr,
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

		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(common.FromHex(code), true, true)
		assert.Equal(t, nil, err)

		intrinsicGas = intrinsicGas + params.TxGasFeeDelegated
		// TODO-Klaytn-Gas Need to revise gas fee calculation.
		gasFrom := params.TxValidationGasDefault
		executionGas := uint64(0x175fd)
		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFeePayer := 3 * gasAccountKeyPublicForSigValidation

		assert.Equal(t, intrinsicGas+gasFrom+executionGas+gasFeePayer, gas)
	}

	// 5. TxTypeFeeDelegatedSmartContractExecution
	{
		amount := new(big.Int).SetUint64(10)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", roleBased.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    anon.Nonce,
			types.TxValueKeyFrom:     anon.Addr,
			types.TxValueKeyTo:       contract.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas, err := types.IntrinsicGas(data, false, true)
		assert.Equal(t, nil, err)

		intrinsicGas = intrinsicGas + params.TxGasFeeDelegated
		// TODO-Klaytn-Gas Need to revise gas fee calculation.
		gasFrom := params.TxValidationGasDefault
		executionGas := uint64(0x9ec4)
		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFeePayer := 3 * gasAccountKeyPublicForSigValidation

		assert.Equal(t, intrinsicGas+gasFrom+executionGas+gasFeePayer, gas)
	}

	// 6. TxTypeFeeDelegatedCancel
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    anon.Nonce,
			types.TxValueKeyFrom:     anon.Addr,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedCancel, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, roleBased.FeeKeys)
		assert.Equal(t, nil, err)

		receipt, gas, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)

		assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		intrinsicGas := params.TxGasCancel + params.TxGasFeeDelegated
		// TODO-Klaytn-Gas Need to revise gas fee calculation.
		gasFrom := params.TxValidationGasDefault
		gasAccountKeyPublicForSigValidation := params.TxValidationGasDefault + 1*params.TxValidationGasPerKey
		// TODO-Klaytn-Gas Gas calculation logic has to be changed to calculate only the gas value for the key used in signing.
		gasFeePayer := 3 * gasAccountKeyPublicForSigValidation

		assert.Equal(t, intrinsicGas+gasFrom+gasFeePayer, gas)
	}
}
