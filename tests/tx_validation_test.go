package tests

import (
	"crypto/ecdsa"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/profile"
	"github.com/ground-x/klaytn/kerrors"
	"github.com/ground-x/klaytn/params"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

type txValueMap map[types.TxValueKeyType]interface{}

type testTxType struct {
	name   string
	txType types.TxType
}

func toBasicType(txType types.TxType) types.TxType {
	return txType &^ ((1 << types.SubTxTypeBits) - 1)
}

func genMapForTxTypes(from TestAccount, to TestAccount, txType types.TxType) txValueMap {
	var valueMap txValueMap
	gasPrice := big.NewInt(25 * params.Ston)
	newAccount, err := createDefaultAccount(accountkey.AccountKeyTypePublic)
	if err != nil {
		return nil
	}
	contractAccount, err := createDefaultAccount(accountkey.AccountKeyTypeFail)
	if err != nil {
		return nil
	}
	contractAccount.Addr, err = common.FromHumanReadableAddress("contract.klaytn")
	if err != nil {
		return nil
	}

	// switch to basic tx type representation and generate a map
	switch toBasicType(txType) {
	case types.TxTypeLegacyTransaction:
		valueMap, _ = genMapForLegacyTransaction(from, to, gasPrice, txType)
	case types.TxTypeValueTransfer:
		valueMap, _ = genMapForValueTransfer(from, to, gasPrice, txType)
	case types.TxTypeValueTransferMemo:
		valueMap, _ = genMapForValueTransferWithMemo(from, to, gasPrice, txType)
	case types.TxTypeAccountCreation:
		valueMap, _ = genMapForCreate(from, newAccount, gasPrice, txType)
	case types.TxTypeAccountUpdate:
		valueMap, _ = genMapForUpdate(from, to, gasPrice, newAccount.AccKey, txType)
	case types.TxTypeSmartContractDeploy:
		valueMap, _ = genMapForDeploy(from, nil, gasPrice, txType)
	case types.TxTypeSmartContractExecution:
		valueMap, _ = genMapForExecution(from, contractAccount, gasPrice, txType)
	case types.TxTypeCancel:
		valueMap, _ = genMapForCancel(from, gasPrice, txType)
	case types.TxTypeChainDataAnchoring:
		valueMap, _ = genMapForChainDataAnchoring(from, gasPrice, txType)
	}

	if txType.IsFeeDelegatedTransaction() {
		valueMap[types.TxValueKeyFeePayer] = from.GetAddr()
	}

	if txType.IsFeeDelegatedWithRatioTransaction() {
		valueMap[types.TxValueKeyFeeRatioOfFeePayer] = types.FeeRatio(30)
	}

	return valueMap
}

// TestValidationPoolInsert generates invalid txs which will be invalidated during txPool insert process.
func TestValidationPoolInsert(t *testing.T) {
	var testTxTypes = []testTxType{
		{"LegacyTransaction", types.TxTypeLegacyTransaction},
		{"ValueTransfer", types.TxTypeValueTransfer},
		{"ValueTransferWithMemo", types.TxTypeValueTransferMemo},
		{"AccountCreation", types.TxTypeAccountCreation},
		{"AccountUpdate", types.TxTypeAccountUpdate},
		{"SmartContractDeploy", types.TxTypeSmartContractDeploy},
		{"SmartContractExecution", types.TxTypeSmartContractExecution},
		{"Cancel", types.TxTypeCancel},
		{"ChainDataAnchoring", types.TxTypeChainDataAnchoring},
		{"FeeDelegatedValueTransfer", types.TxTypeFeeDelegatedValueTransfer},
		{"FeeDelegatedValueTransferWithMemo", types.TxTypeFeeDelegatedValueTransferMemo},
		{"FeeDelegatedAccountUpdate", types.TxTypeFeeDelegatedAccountUpdate},
		{"FeeDelegatedSmartContractDeploy", types.TxTypeFeeDelegatedSmartContractDeploy},
		{"FeeDelegatedSmartContractExecution", types.TxTypeFeeDelegatedSmartContractExecution},
		{"FeeDelegatedCancel", types.TxTypeFeeDelegatedCancel},
		{"FeeDelegatedWithRatioValueTransfer", types.TxTypeFeeDelegatedValueTransferWithRatio},
		{"FeeDelegatedWithRatioValueTransferWithMemo", types.TxTypeFeeDelegatedValueTransferMemoWithRatio},
		{"FeeDelegatedWithRatioAccountUpdate", types.TxTypeFeeDelegatedAccountUpdateWithRatio},
		{"FeeDelegatedWithRatioSmartContractDeploy", types.TxTypeFeeDelegatedSmartContractDeployWithRatio},
		{"FeeDelegatedWithRatioSmartContractExecution", types.TxTypeFeeDelegatedSmartContractExecutionWithRatio},
		{"FeeDelegatedWithRatioCancel", types.TxTypeFeeDelegatedCancelWithRatio},
	}

	var invalidCases = []struct {
		Name string
		fn   func(types.TxType, txValueMap) (txValueMap, error)
	}{
		{"invalidNonce", decreaseNonce},
		{"invalidGasLimit", decreaseGasLimit},
		{"invalidTxSize", exceedSizeLimit},
		{"invalidRecipientProgram", valueTransferToContract},
		{"invalidRecipientNotProgram", executeToEOA},
		{"invalidRecipientExisting", creationToExistingAddr},
	}

	prof := profile.NewProfiler()

	// Initialize blockchain
	bcdata, err := NewBCData(6, 4)
	if err != nil {
		t.Fatal(err)
	}
	defer bcdata.Shutdown()

	// Initialize address-balance map for verification
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// for contract execution txs
	contract, err := createHumanReadableAccount(getRandomPrivateKeyString(t), "contract")

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// deploy a contract for contract execution tx type
	{
		var txs types.Transactions

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            &(contract.Addr),
			types.TxValueKeyAmount:        big.NewInt(0),
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      big.NewInt(25 * params.Ston),
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.AddNonce()
	}

	// test for all tx types
	for _, testTxType := range testTxTypes {
		txType := testTxType.txType

		// generate invalid txs and check the return error
		for _, invalidCase := range invalidCases {
			// generate a new tx and mutate it
			valueMap := genMapForTxTypes(reservoir, reservoir, txType)
			invalidMap, expectedErr := invalidCase.fn(txType, valueMap)

			tx, err := types.NewTransactionWithMap(txType, invalidMap)
			assert.Equal(t, nil, err)

			err = tx.SignWithKeys(signer, reservoir.Keys)
			assert.Equal(t, nil, err)

			if txType.IsFeeDelegatedTransaction() {
				tx.SignFeePayerWithKeys(signer, reservoir.Keys)
				assert.Equal(t, nil, err)
			}

			err = txpool.AddRemote(tx)
			assert.Equal(t, expectedErr, err)
			if expectedErr == nil {
				reservoir.Nonce += 1
			}
		}
	}
}

// decreaseNonce changes nonce to zero.
func decreaseNonce(txType types.TxType, values txValueMap) (txValueMap, error) {
	values[types.TxValueKeyNonce] = uint64(0)

	return values, blockchain.ErrNonceTooLow
}

// decreaseGasLimit changes gasLimit to 12345678
func decreaseGasLimit(txType types.TxType, values txValueMap) (txValueMap, error) {
	(*big.Int).SetUint64(values[types.TxValueKeyGasPrice].(*big.Int), 12345678)

	return values, blockchain.ErrInvalidUnitPrice
}

// exceedSizeLimit assigns tx data bigger than MaxTxDataSize.
func exceedSizeLimit(txType types.TxType, values txValueMap) (txValueMap, error) {
	invalidData := make([]byte, blockchain.MaxTxDataSize+1)

	if values[types.TxValueKeyData] != nil {
		values[types.TxValueKeyData] = invalidData
		return values, blockchain.ErrOversizedData
	}

	if values[types.TxValueKeyAnchoredData] != nil {
		values[types.TxValueKeyAnchoredData] = invalidData
		return values, blockchain.ErrOversizedData
	}

	return values, nil
}

// valueTransferToContract changes recipient address of value transfer txs to the contract address, "contract.klaytn".
func valueTransferToContract(txType types.TxType, values txValueMap) (txValueMap, error) {
	programAddr, err := common.FromHumanReadableAddress("contract.klaytn")
	if err != nil {
		return nil, nil
	}

	txType = toBasicType(txType)
	if txType == types.TxTypeValueTransfer || txType == types.TxTypeValueTransferMemo {
		values[types.TxValueKeyTo] = programAddr
		return values, kerrors.ErrNotForProgramAccount
	}

	return values, nil
}

// creationToExistingAddr changes the recipient of account creating txs to the existing address, "contract.klaytn".
func creationToExistingAddr(txType types.TxType, values txValueMap) (txValueMap, error) {
	existingAddr, err := common.FromHumanReadableAddress("contract.klaytn")
	if err != nil {
		return nil, nil
	}

	if txType.IsAccountCreation() {
		values[types.TxValueKeyTo] = existingAddr
		values[types.TxValueKeyHumanReadable] = true
		return values, kerrors.ErrAccountAlreadyExists
	}

	if txType.IsContractDeploy() {
		values[types.TxValueKeyTo] = &existingAddr
		values[types.TxValueKeyHumanReadable] = true
		return values, kerrors.ErrAccountAlreadyExists
	}

	return values, nil
}

// executeToEOA changes the recipient of contract execution txs to an EOA address (the same with the sender).
func executeToEOA(txType types.TxType, values txValueMap) (txValueMap, error) {
	if toBasicType(txType) == types.TxTypeSmartContractExecution {
		values[types.TxValueKeyTo] = values[types.TxValueKeyFrom].(common.Address)
		return values, kerrors.ErrNotProgramAccount
	}

	return values, nil
}
