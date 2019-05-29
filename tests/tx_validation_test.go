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

func genMapForTxTypes(from TestAccount, to TestAccount, txType types.TxType) (txValueMap, uint64) {
	var valueMap txValueMap
	gas := uint64(0)
	gasPrice := big.NewInt(25 * params.Ston)
	newAccount, err := createDefaultAccount(accountkey.AccountKeyTypePublic)
	if err != nil {
		return nil, 0
	}
	contractAccount, err := createDefaultAccount(accountkey.AccountKeyTypeFail)
	if err != nil {
		return nil, 0
	}
	contractAccount.Addr, err = common.FromHumanReadableAddress("contract.klaytn")
	if err != nil {
		return nil, 0
	}

	// switch to basic tx type representation and generate a map
	switch toBasicType(txType) {
	case types.TxTypeLegacyTransaction:
		valueMap, gas = genMapForLegacyTransaction(from, to, gasPrice, txType)
	case types.TxTypeValueTransfer:
		valueMap, gas = genMapForValueTransfer(from, to, gasPrice, txType)
	case types.TxTypeValueTransferMemo:
		valueMap, gas = genMapForValueTransferWithMemo(from, to, gasPrice, txType)
	case types.TxTypeAccountCreation:
		valueMap, gas = genMapForCreate(from, newAccount, gasPrice, txType)
	case types.TxTypeAccountUpdate:
		valueMap, gas = genMapForUpdate(from, to, gasPrice, newAccount.AccKey, txType)
	case types.TxTypeSmartContractDeploy:
		valueMap, gas = genMapForDeploy(from, nil, gasPrice, txType)
	case types.TxTypeSmartContractExecution:
		valueMap, gas = genMapForExecution(from, contractAccount, gasPrice, txType)
	case types.TxTypeCancel:
		valueMap, gas = genMapForCancel(from, gasPrice, txType)
	case types.TxTypeChainDataAnchoring:
		valueMap, gas = genMapForChainDataAnchoring(from, gasPrice, txType)
	}

	if txType.IsFeeDelegatedTransaction() {
		valueMap[types.TxValueKeyFeePayer] = from.GetAddr()
	}

	if txType.IsFeeDelegatedWithRatioTransaction() {
		valueMap[types.TxValueKeyFeeRatioOfFeePayer] = types.FeeRatio(30)
	}

	return valueMap, gas
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
		{"invalidCodeFormat", invalidCodeFormat},
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
			valueMap, _ := genMapForTxTypes(reservoir, reservoir, txType)
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

// TestValidationBlockTx generates invalid txs which will be invalidated during block insert process.
func TestValidationBlockTx(t *testing.T) {
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
		{"invalidRecipientProgram", valueTransferToContract},
		{"invalidRecipientNotProgram", executeToEOA},
		{"invalidRecipientExisting", creationToExistingAddr},
		{"invalidCodeFormat", invalidCodeFormat},
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

			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, expectedErr, err)
			if expectedErr == nil {
				assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
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

func invalidCodeFormat(txType types.TxType, values txValueMap) (txValueMap, error) {
	if txType.IsContractDeploy() {
		values[types.TxValueKeyCodeFormat] = params.CodeFormatLast
		return values, kerrors.ErrInvalidCodeFormat
	}
	return values, nil
}

// TestValidationPoolInsert2 generates invalid txs which will be invalidated during txPool insert process.
func TestValidationPoolInsert2(t *testing.T) {
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
		fn   func(*testing.T, types.TxType, *TestAccountType, types.EIP155Signer) (*types.Transaction, error)
	}{
		{"invalidSender", testInvalidSenderSig},
		{"invalidFeePayer", testInvalidFeePayerSig},
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

		for _, invalidCase := range invalidCases {
			tx, expectedErr := invalidCase.fn(t, txType, reservoir, signer)

			if tx != nil {
				err = txpool.AddRemote(tx)
				assert.Equal(t, expectedErr, err)
			}
		}
	}
}

// testInvalidSenderSig generates invalid txs signed by an invalid sender.
func testInvalidSenderSig(t *testing.T, txType types.TxType, reservoir *TestAccountType, signer types.EIP155Signer) (*types.Transaction, error) {
	if !txType.IsLegacyTransaction() {
		newAcc, err := createDefaultAccount(accountkey.AccountKeyTypePublic)
		assert.Equal(t, nil, err)

		valueMap, _ := genMapForTxTypes(reservoir, reservoir, txType)
		tx, err := types.NewTransactionWithMap(txType, valueMap)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, newAcc.Keys)
		assert.Equal(t, nil, err)

		if txType.IsFeeDelegatedTransaction() {
			tx.SignFeePayerWithKeys(signer, reservoir.Keys)
			assert.Equal(t, nil, err)
		}
		return tx, types.ErrInvalidSigSender
	}
	return nil, nil
}

// testInvalidFeePayerSig generates invalid txs signed by an invalid fee payer.
func testInvalidFeePayerSig(t *testing.T, txType types.TxType, reservoir *TestAccountType, signer types.EIP155Signer) (*types.Transaction, error) {
	if txType.IsFeeDelegatedTransaction() {
		newAcc, err := createDefaultAccount(accountkey.AccountKeyTypePublic)
		assert.Equal(t, nil, err)

		valueMap, _ := genMapForTxTypes(reservoir, reservoir, txType)
		tx, err := types.NewTransactionWithMap(txType, valueMap)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		tx.SignFeePayerWithKeys(signer, newAcc.Keys)
		assert.Equal(t, nil, err)

		return tx, blockchain.ErrInvalidFeePayer
	}
	return nil, nil
}

// TestLegacyTxFromNonLegacyAcc generates legacy tx from non-legacy account, and it will be invalidated during txPool insert process.
func TestLegacyTxFromNonLegacyAcc(t *testing.T) {
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

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	var txs types.Transactions
	acc1, err := createDefaultAccount(accountkey.AccountKeyTypePublic)

	valueMap, _ := genMapForTxTypes(reservoir, reservoir, types.TxTypeAccountCreation)
	valueMap[types.TxValueKeyTo] = acc1.Addr
	valueMap[types.TxValueKeyAccountKey] = acc1.AccKey

	tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, valueMap)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, reservoir.Keys)
	assert.Equal(t, nil, err)

	txs = append(txs, tx)

	if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
		t.Fatal(err)
	}
	reservoir.AddNonce()

	valueMap, _ = genMapForTxTypes(acc1, reservoir, types.TxTypeLegacyTransaction)
	tx, err = types.NewTransactionWithMap(types.TxTypeLegacyTransaction, valueMap)
	assert.Equal(t, nil, err)

	err = tx.SignWithKeys(signer, acc1.Keys)
	assert.Equal(t, nil, err)

	err = txpool.AddRemote(tx)
	assert.Equal(t, kerrors.ErrLegacyTransactionMustBeWithLegacyKey, err)
}

// TestInvalidBalance tests generates invalid txs which don't have enough KLAY, and will be invalidated during txPool insert process.
func TestInvalidBalance(t *testing.T) {
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

	// test account will be lack of KLAY
	testAcc, err := createDefaultAccount(accountkey.AccountKeyTypeLegacy)
	assert.Equal(t, nil, err)

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	gasLimit := uint64(100000000000)
	gasPrice := big.NewInt(25 * params.Ston)
	amount := uint64(25 * params.Ston)
	cost := new(big.Int).Mul(new(big.Int).SetUint64(gasLimit), gasPrice)
	cost.Add(cost, new(big.Int).SetUint64(amount))

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

	// generate a test account with a specific amount of KLAY
	{
		var txs types.Transactions

		valueMapForCreation, _ := genMapForTxTypes(reservoir, reservoir, types.TxTypeAccountCreation)
		valueMapForCreation[types.TxValueKeyTo] = testAcc.Addr
		valueMapForCreation[types.TxValueKeyAccountKey] = testAcc.AccKey
		valueMapForCreation[types.TxValueKeyAmount] = cost

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, valueMapForCreation)
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

		if !txType.IsFeeDelegatedTransaction() {
			// tx with a specific amount or a gasLimit requiring more KLAY than the sender has.
			{
				valueMap, _ := genMapForTxTypes(testAcc, reservoir, txType)
				if valueMap[types.TxValueKeyAmount] != nil {
					valueMap[types.TxValueKeyAmount] = new(big.Int).SetUint64(amount)
					valueMap[types.TxValueKeyGasLimit] = gasLimit + 1 // requires 1 more gas
				} else {
					valueMap[types.TxValueKeyGasLimit] = gasLimit + (amount / gasPrice.Uint64()) + 1 // requires 1 more gas
				}

				tx, err := types.NewTransactionWithMap(txType, valueMap)
				assert.Equal(t, nil, err)

				err = tx.SignWithKeys(signer, testAcc.Keys)
				assert.Equal(t, nil, err)

				err = txpool.AddRemote(tx)
				assert.Equal(t, blockchain.ErrInsufficientFundsFrom, err)
			}

			// tx with a specific amount or a gasLimit requiring the exact KLAY the sender has.
			{
				valueMap, _ := genMapForTxTypes(testAcc, reservoir, txType)
				if valueMap[types.TxValueKeyAmount] != nil {
					valueMap[types.TxValueKeyAmount] = new(big.Int).SetUint64(amount)
					valueMap[types.TxValueKeyGasLimit] = gasLimit
				} else {
					valueMap[types.TxValueKeyGasLimit] = gasLimit + (amount / gasPrice.Uint64())
				}

				tx, err := types.NewTransactionWithMap(txType, valueMap)
				assert.Equal(t, nil, err)

				err = tx.SignWithKeys(signer, testAcc.Keys)
				assert.Equal(t, nil, err)

				// Since `txpool.AddRemote` does not make a block,
				// the sender can send txs to txpool in multiple times (by the for loop) with limited KLAY.
				err = txpool.AddRemote(tx)
				assert.Equal(t, nil, err)
				testAcc.AddNonce()
			}
		}

		if txType.IsFeeDelegatedTransaction() && !txType.IsFeeDelegatedWithRatioTransaction() {
			// tx with a specific amount requiring more KLAY than the sender has.
			{
				valueMap, _ := genMapForTxTypes(testAcc, reservoir, txType)
				if valueMap[types.TxValueKeyAmount] != nil {
					valueMap[types.TxValueKeyFeePayer] = reservoir.Addr
					valueMap[types.TxValueKeyAmount] = new(big.Int).Add(cost, new(big.Int).SetUint64(1)) // requires 1 more amount

					tx, err := types.NewTransactionWithMap(txType, valueMap)
					assert.Equal(t, nil, err)

					err = tx.SignWithKeys(signer, testAcc.Keys)
					assert.Equal(t, nil, err)

					tx.SignFeePayerWithKeys(signer, reservoir.Keys)
					assert.Equal(t, nil, err)

					err = txpool.AddRemote(tx)
					assert.Equal(t, blockchain.ErrInsufficientFundsFrom, err)
				}
			}

			// tx with a specific gasLimit (or amount) requiring more KLAY than the feePayer has.
			{
				valueMap, _ := genMapForTxTypes(reservoir, reservoir, txType)
				valueMap[types.TxValueKeyFeePayer] = testAcc.Addr
				valueMap[types.TxValueKeyGasLimit] = gasLimit + (amount / gasPrice.Uint64()) + 1 // requires 1 more gas

				tx, err := types.NewTransactionWithMap(txType, valueMap)
				assert.Equal(t, nil, err)

				err = tx.SignWithKeys(signer, reservoir.Keys)
				assert.Equal(t, nil, err)

				tx.SignFeePayerWithKeys(signer, testAcc.Keys)
				assert.Equal(t, nil, err)

				err = txpool.AddRemote(tx)
				assert.Equal(t, blockchain.ErrInsufficientFundsFeePayer, err)
			}

			// tx with a specific amount requiring the exact KLAY the sender has.
			{
				valueMap, _ := genMapForTxTypes(testAcc, reservoir, txType)
				if valueMap[types.TxValueKeyAmount] != nil {
					valueMap[types.TxValueKeyFeePayer] = reservoir.Addr
					valueMap[types.TxValueKeyAmount] = cost

					tx, err := types.NewTransactionWithMap(txType, valueMap)
					assert.Equal(t, nil, err)

					err = tx.SignWithKeys(signer, testAcc.Keys)
					assert.Equal(t, nil, err)

					tx.SignFeePayerWithKeys(signer, reservoir.Keys)
					assert.Equal(t, nil, err)

					// Since `txpool.AddRemote` does not make a block,
					// the sender can send txs to txpool in multiple times (by the for loop) with limited KLAY.
					err = txpool.AddRemote(tx)
					assert.Equal(t, nil, err)
					testAcc.AddNonce()
				}
			}

			// tx with a specific gasLimit (or amount) requiring the exact KLAY the feePayer has.
			{
				valueMap, _ := genMapForTxTypes(reservoir, reservoir, txType)
				valueMap[types.TxValueKeyFeePayer] = testAcc.Addr
				valueMap[types.TxValueKeyGasLimit] = gasLimit + (amount / gasPrice.Uint64())

				tx, err := types.NewTransactionWithMap(txType, valueMap)
				assert.Equal(t, nil, err)

				err = tx.SignWithKeys(signer, reservoir.Keys)
				assert.Equal(t, nil, err)

				tx.SignFeePayerWithKeys(signer, testAcc.Keys)
				assert.Equal(t, nil, err)

				// Since `txpool.AddRemote` does not make a block,
				// the sender can send txs to txpool in multiple times (by the for loop) with limited KLAY.
				err = txpool.AddRemote(tx)
				assert.Equal(t, nil, err)
				reservoir.AddNonce()
			}
		}

		if txType.IsFeeDelegatedWithRatioTransaction() {
			// tx with a specific amount and a gasLimit requiring more KLAY than the sender has.
			{
				valueMap, _ := genMapForTxTypes(testAcc, reservoir, txType)
				valueMap[types.TxValueKeyFeePayer] = reservoir.Addr
				valueMap[types.TxValueKeyFeeRatioOfFeePayer] = types.FeeRatio(90)
				if valueMap[types.TxValueKeyAmount] != nil {
					valueMap[types.TxValueKeyAmount] = new(big.Int).SetUint64(amount)
					// Gas testAcc will charge = tx gasLimit * sender's feeRatio
					// = (gasLimit + 1) * 10 * (100 - 90) * 0.01 = gasLimit + 1
					valueMap[types.TxValueKeyGasLimit] = (gasLimit + 1) * 10 // requires 1 more gas
				} else {
					// Gas testAcc will charge = tx gasLimit * sender's feeRatio
					// = (gasLimit + (amount / gasPrice.Uint64()) + 1) * 10 * (100 - 90) * 0.01 = gasLimit + (amount / gasPrice.Uint64()) + 1
					valueMap[types.TxValueKeyGasLimit] = (gasLimit + (amount / gasPrice.Uint64()) + 1) * 10 // requires 1 more gas
				}

				tx, err := types.NewTransactionWithMap(txType, valueMap)
				assert.Equal(t, nil, err)

				err = tx.SignWithKeys(signer, testAcc.Keys)
				assert.Equal(t, nil, err)

				tx.SignFeePayerWithKeys(signer, reservoir.Keys)
				assert.Equal(t, nil, err)

				err = txpool.AddRemote(tx)
				assert.Equal(t, blockchain.ErrInsufficientFundsFrom, err)
			}

			// tx with a specific amount and a gasLimit requiring more KLAY than the feePayer has.
			{
				valueMap, _ := genMapForTxTypes(reservoir, reservoir, txType)
				valueMap[types.TxValueKeyFeePayer] = testAcc.Addr
				valueMap[types.TxValueKeyFeeRatioOfFeePayer] = types.FeeRatio(10)
				// Gas testAcc will charge = tx gasLimit * fee-payer's feeRatio
				// = (gasLimit + (amount / gasPrice.Uint64()) + 1) * 10 * 10 * 0.01 = gasLimit + (amount / gasPrice.Uint64()) + 1
				valueMap[types.TxValueKeyGasLimit] = (gasLimit + (amount / gasPrice.Uint64()) + 1) * 10 // requires 1 more gas

				tx, err := types.NewTransactionWithMap(txType, valueMap)
				assert.Equal(t, nil, err)

				err = tx.SignWithKeys(signer, reservoir.Keys)
				assert.Equal(t, nil, err)

				tx.SignFeePayerWithKeys(signer, testAcc.Keys)
				assert.Equal(t, nil, err)

				err = txpool.AddRemote(tx)
				assert.Equal(t, blockchain.ErrInsufficientFundsFeePayer, err)
			}

			// tx with a specific amount and a gasLimit requiring the exact KLAY the sender has.
			{
				valueMap, _ := genMapForTxTypes(testAcc, reservoir, txType)
				valueMap[types.TxValueKeyFeePayer] = reservoir.Addr
				valueMap[types.TxValueKeyFeeRatioOfFeePayer] = types.FeeRatio(90)
				if valueMap[types.TxValueKeyAmount] != nil {
					valueMap[types.TxValueKeyAmount] = new(big.Int).SetUint64(amount)
					// Gas testAcc will charge = tx gasLimit * sender's feeRatio
					// = gasLimit * 10 * (100 - 90) * 0.01 = gasLimit
					valueMap[types.TxValueKeyGasLimit] = gasLimit * 10
				} else {
					// Gas testAcc will charge = tx gasLimit * sender's feeRatio
					// = (gasLimit + (amount / gasPrice.Uint64())) * 10 * (100 - 90) * 0.01 = gasLimit + (amount / gasPrice.Uint64())
					valueMap[types.TxValueKeyGasLimit] = (gasLimit + (amount / gasPrice.Uint64())) * 10
				}

				tx, err := types.NewTransactionWithMap(txType, valueMap)
				assert.Equal(t, nil, err)

				err = tx.SignWithKeys(signer, testAcc.Keys)
				assert.Equal(t, nil, err)

				tx.SignFeePayerWithKeys(signer, reservoir.Keys)
				assert.Equal(t, nil, err)

				// Since `txpool.AddRemote` does not make a block,
				// the sender can send txs to txpool in multiple times (by the for loop) with limited KLAY.
				err = txpool.AddRemote(tx)
				assert.Equal(t, nil, err)
				testAcc.AddNonce()
			}

			// tx with a specific amount and a gasLimit requiring the exact KLAY the feePayer has.
			{
				valueMap, _ := genMapForTxTypes(reservoir, reservoir, txType)
				valueMap[types.TxValueKeyFeePayer] = testAcc.Addr
				valueMap[types.TxValueKeyFeeRatioOfFeePayer] = types.FeeRatio(10)
				// Gas testAcc will charge = tx gasLimit * fee-payer's feeRatio
				// = (gasLimit + (amount / gasPrice.Uint64())) * 10 * 10 * 0.01 = gasLimit + (amount / gasPrice.Uint64())
				valueMap[types.TxValueKeyGasLimit] = (gasLimit + (amount / gasPrice.Uint64())) * 10

				tx, err := types.NewTransactionWithMap(txType, valueMap)
				assert.Equal(t, nil, err)

				err = tx.SignWithKeys(signer, reservoir.Keys)
				assert.Equal(t, nil, err)

				tx.SignFeePayerWithKeys(signer, testAcc.Keys)
				assert.Equal(t, nil, err)

				// Since `txpool.AddRemote` does not make a block,
				// the sender can send txs to txpool in multiple times (by the for loop) with limited KLAY.
				err = txpool.AddRemote(tx)
				assert.Equal(t, nil, err)
				reservoir.AddNonce()
			}
		}
	}
}
