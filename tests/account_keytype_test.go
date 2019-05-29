package tests

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ground-x/klaytn/accounts/abi"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/hexutil"
	"github.com/ground-x/klaytn/common/profile"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/kerrors"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/stretchr/testify/assert"
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"
)

// createDefaultAccount creates a default account with a specific account key type.
func createDefaultAccount(accountKeyType accountkey.AccountKeyType) (*TestAccountType, error) {
	var err error

	// prepare  keys
	keys := genTestKeys(3)
	weights := []uint{1, 1, 1}
	weightedKeys := make(accountkey.WeightedPublicKeys, 3)
	threshold := uint(2)

	for i := range keys {
		weightedKeys[i] = accountkey.NewWeightedPublicKey(weights[i], (*accountkey.PublicKeySerializable)(&keys[i].PublicKey))
	}

	// a role-based key
	roleAccKey := accountkey.AccountKeyRoleBased{
		accountkey.NewAccountKeyPublicWithValue(&keys[accountkey.RoleTransaction].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[accountkey.RoleAccountUpdate].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[accountkey.RoleFeePayer].PublicKey),
	}

	// default account setting
	account := &TestAccountType{
		Addr:   crypto.PubkeyToAddress(keys[0].PublicKey), // default
		Keys:   []*ecdsa.PrivateKey{keys[0]},              // default
		Nonce:  uint64(0),                                 // default
		AccKey: nil,
	}

	// set an account key and a private key
	switch accountKeyType {
	case accountkey.AccountKeyTypeNil:
		account.AccKey, err = accountkey.NewAccountKey(accountKeyType)
	case accountkey.AccountKeyTypeLegacy:
		account.AccKey, err = accountkey.NewAccountKey(accountKeyType)
	case accountkey.AccountKeyTypePublic:
		account.AccKey = accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey)
	case accountkey.AccountKeyTypeFail:
		account.AccKey, err = accountkey.NewAccountKey(accountKeyType)
	case accountkey.AccountKeyTypeWeightedMultiSig:
		account.Keys = keys
		account.AccKey = accountkey.NewAccountKeyWeightedMultiSigWithValues(threshold, weightedKeys)
	case accountkey.AccountKeyTypeRoleBased:
		account.Keys = keys
		account.AccKey = accountkey.NewAccountKeyRoleBasedWithValues(roleAccKey)
	default:
		return nil, kerrors.ErrDifferentAccountKeyType
	}
	if err != nil {
		return nil, err
	}

	return account, err
}

// generateDefaultTx returns a Tx with default values of txTypes.
// If txType is a kind of account update, it will return an account to update.
// Otherwise, it will return (tx, nil, nil).
// For contract execution Txs, TxValueKeyTo value is set to "contract" as a default.
// The address "contact" should exist before calling this function.
func generateDefaultTx(sender *TestAccountType, recipient *TestAccountType, txType types.TxType) (*types.Transaction, *TestAccountType, error) {
	gasPrice := new(big.Int).SetUint64(25 * params.Ston)
	gasLimit := uint64(10000000)
	amount := new(big.Int).SetUint64(1)

	// generate a new account for account creation/update Txs or contract deploy Txs
	senderAccType := accountkey.AccountKeyTypeLegacy
	if sender.AccKey != nil {
		senderAccType = sender.AccKey.Type()
	}
	newAcc, err := createDefaultAccount(senderAccType)
	if err != nil {
		return nil, nil, err
	}

	// a default recipient address of smart contract execution to "contract"
	var contractAddr common.Address
	contractAddr.SetBytesFromFront([]byte("contract.klaytn"))

	// Smart contract data for TxTypeSmartContractDeploy, TxTypeSmartContractExecution Txs
	var code string
	var abiStr string

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

	abii, err := abi.JSON(strings.NewReader(string(abiStr)))
	if err != nil {
		return nil, nil, err
	}

	dataABI, err := abii.Pack("reward", recipient.Addr)
	if err != nil {
		return nil, nil, err
	}

	// generate a legacy tx
	if txType == types.TxTypeLegacyTransaction {
		tx := types.NewTransaction(sender.Nonce, recipient.Addr, amount, gasLimit, gasPrice, []byte{})
		return tx, nil, nil
	}

	// Default valuesMap setting
	amountZero := new(big.Int).SetUint64(0)
	ratio := types.FeeRatio(30)
	dataMemo := []byte("hello")
	dataAnchor := []byte{0x11, 0x22}
	dataCode := common.FromHex(code)
	values := map[types.TxValueKeyType]interface{}{}

	switch txType {
	case types.TxTypeValueTransfer:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = recipient.Addr
		values[types.TxValueKeyAmount] = amount
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
	case types.TxTypeFeeDelegatedValueTransfer:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = recipient.Addr
		values[types.TxValueKeyAmount] = amount
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyFeePayer] = recipient.Addr
	case types.TxTypeFeeDelegatedValueTransferWithRatio:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = recipient.Addr
		values[types.TxValueKeyAmount] = amount
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyFeePayer] = recipient.Addr
		values[types.TxValueKeyFeeRatioOfFeePayer] = ratio
	case types.TxTypeValueTransferMemo:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = recipient.Addr
		values[types.TxValueKeyAmount] = amount
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyData] = dataMemo
	case types.TxTypeFeeDelegatedValueTransferMemo:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = recipient.Addr
		values[types.TxValueKeyAmount] = amount
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyData] = dataMemo
		values[types.TxValueKeyFeePayer] = recipient.Addr
	case types.TxTypeFeeDelegatedValueTransferMemoWithRatio:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = recipient.Addr
		values[types.TxValueKeyAmount] = amount
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyData] = dataMemo
		values[types.TxValueKeyFeePayer] = recipient.Addr
		values[types.TxValueKeyFeeRatioOfFeePayer] = ratio
	case types.TxTypeAccountCreation:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = newAcc.Addr
		values[types.TxValueKeyAmount] = amount
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyHumanReadable] = false
		values[types.TxValueKeyAccountKey] = newAcc.AccKey
	case types.TxTypeAccountUpdate:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyAccountKey] = newAcc.AccKey
	case types.TxTypeFeeDelegatedAccountUpdate:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyAccountKey] = newAcc.AccKey
		values[types.TxValueKeyFeePayer] = recipient.Addr
	case types.TxTypeFeeDelegatedAccountUpdateWithRatio:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyAccountKey] = newAcc.AccKey
		values[types.TxValueKeyFeePayer] = recipient.Addr
		values[types.TxValueKeyFeeRatioOfFeePayer] = ratio
	case types.TxTypeSmartContractDeploy:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = &newAcc.Addr
		values[types.TxValueKeyAmount] = amountZero
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = amountZero
		values[types.TxValueKeyData] = dataCode
		values[types.TxValueKeyHumanReadable] = false
		values[types.TxValueKeyCodeFormat] = params.CodeFormatEVM
	case types.TxTypeFeeDelegatedSmartContractDeploy:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = &newAcc.Addr
		values[types.TxValueKeyAmount] = amountZero
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = amountZero
		values[types.TxValueKeyData] = dataCode
		values[types.TxValueKeyHumanReadable] = false
		values[types.TxValueKeyFeePayer] = recipient.Addr
		values[types.TxValueKeyCodeFormat] = params.CodeFormatEVM
	case types.TxTypeFeeDelegatedSmartContractDeployWithRatio:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = &newAcc.Addr
		values[types.TxValueKeyAmount] = amountZero
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = amountZero
		values[types.TxValueKeyData] = dataCode
		values[types.TxValueKeyHumanReadable] = false
		values[types.TxValueKeyFeePayer] = recipient.Addr
		values[types.TxValueKeyFeeRatioOfFeePayer] = ratio
		values[types.TxValueKeyCodeFormat] = params.CodeFormatEVM
	case types.TxTypeSmartContractExecution:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = contractAddr
		values[types.TxValueKeyAmount] = amountZero
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = amountZero
		values[types.TxValueKeyData] = dataABI
	case types.TxTypeFeeDelegatedSmartContractExecution:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = contractAddr
		values[types.TxValueKeyAmount] = amountZero
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = amountZero
		values[types.TxValueKeyData] = dataABI
		values[types.TxValueKeyFeePayer] = recipient.Addr
	case types.TxTypeFeeDelegatedSmartContractExecutionWithRatio:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = contractAddr
		values[types.TxValueKeyAmount] = amountZero
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = amountZero
		values[types.TxValueKeyData] = dataABI
		values[types.TxValueKeyFeePayer] = recipient.Addr
		values[types.TxValueKeyFeeRatioOfFeePayer] = ratio
	case types.TxTypeCancel:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
	case types.TxTypeFeeDelegatedCancel:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyFeePayer] = recipient.Addr
	case types.TxTypeFeeDelegatedCancelWithRatio:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyFeePayer] = recipient.Addr
		values[types.TxValueKeyFeeRatioOfFeePayer] = ratio
	case types.TxTypeChainDataAnchoring:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyAnchoredData] = dataAnchor
	}

	tx, err := types.NewTransactionWithMap(txType, values)
	if err != nil {
		return nil, nil, err
	}

	// the function returns an updated sender account for account update Txs
	if txType.IsAccountUpdate() {
		// For the account having a legacy key, its private key will not be updated since it is coupled with its address.
		if newAcc.AccKey.Type().IsLegacyAccountKey() {
			newAcc.Keys = sender.Keys
		}
		newAcc.Addr = sender.Addr
		newAcc.Nonce = sender.Nonce
		return tx, newAcc, err
	}

	return tx, nil, err
}

// expectedTestResultForDefaultTx returns expected validity of tx which generated from (accountKeyType, txType) pair.
func expectedTestResultForDefaultTx(accountKeyType accountkey.AccountKeyType, txType types.TxType) error {
	switch accountKeyType {
	//case accountkey.AccountKeyTypeNil:                     // not supported type
	case accountkey.AccountKeyTypeFail:
		if txType.IsAccountUpdate() {
			return kerrors.ErrAccountKeyFailNotUpdatable
		}
		return types.ErrInvalidSigSender
	}
	return nil
}

func signTxWithVariousKeyTypes(signer types.EIP155Signer, tx *types.Transaction, sender *TestAccountType) (*types.Transaction, error) {
	var err error
	txType := tx.Type()
	accKeyType := sender.AccKey.Type()

	if accKeyType == accountkey.AccountKeyTypeWeightedMultiSig {
		if txType.IsLegacyTransaction() {
			err = tx.SignWithKeys(signer, []*ecdsa.PrivateKey{sender.Keys[0]})
		} else {
			err = tx.SignWithKeys(signer, sender.Keys)
		}
	} else if accKeyType == accountkey.AccountKeyTypeRoleBased {
		if txType.IsAccountUpdate() {
			err = tx.SignWithKeys(signer, []*ecdsa.PrivateKey{sender.Keys[accountkey.RoleAccountUpdate]})
		} else {
			err = tx.SignWithKeys(signer, []*ecdsa.PrivateKey{sender.Keys[accountkey.RoleTransaction]})
		}
	} else {
		err = tx.SignWithKeys(signer, sender.Keys)
	}
	return tx, err
}

// TestDefaultTxsWithDefaultAccountKey tests most of transactions types with most of account key types.
// The test creates a default account for each account key type, and generates default Tx for each Tx type.
// AccountKeyTypeNil is excluded because it cannot be used for account creation.
func TestDefaultTxsWithDefaultAccountKey(t *testing.T) {
	gasPrice := new(big.Int).SetUint64(25 * params.Ston)
	gasLimit := uint64(100000000)

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

	// smart contact account
	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	// smart contract code
	var code string

	if isCompilerAvailable() {
		filename := string("../contracts/reward/contract/KlaytnReward.sol")
		codes, _ := compileSolidity(filename)
		code = codes[0]
	} else {
		// Falling back to use compiled code.
		code = "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// create a smart contract account for contract execution test
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(0)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            &contract.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      uint64(50 * uint64(params.Ston)),
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		err = bcdata.GenABlockWithTransactions(accountMap, txs, prof)
		assert.Equal(t, nil, err)

		reservoir.Nonce += 1
	}
	// select account key types to be tested
	accountKeyTypes := []accountkey.AccountKeyType{
		//accountkey.AccountKeyTypeNil, // not supported type
		accountkey.AccountKeyTypeLegacy,
		accountkey.AccountKeyTypePublic,
		accountkey.AccountKeyTypeFail,
		accountkey.AccountKeyTypeWeightedMultiSig,
		accountkey.AccountKeyTypeRoleBased,
	}

	// select txtypes to be tested
	txTypes := []types.TxType{
		types.TxTypeLegacyTransaction,

		types.TxTypeValueTransfer,
		types.TxTypeValueTransferMemo,
		types.TxTypeSmartContractDeploy,
		types.TxTypeSmartContractExecution,
		types.TxTypeAccountCreation,
		types.TxTypeAccountUpdate,
		types.TxTypeCancel,

		types.TxTypeFeeDelegatedValueTransfer,
		types.TxTypeFeeDelegatedValueTransferMemo,
		types.TxTypeFeeDelegatedSmartContractDeploy,
		types.TxTypeFeeDelegatedSmartContractExecution,
		types.TxTypeFeeDelegatedAccountUpdate,
		types.TxTypeFeeDelegatedCancel,

		types.TxTypeFeeDelegatedValueTransferWithRatio,
		types.TxTypeFeeDelegatedValueTransferMemoWithRatio,
		types.TxTypeFeeDelegatedSmartContractDeployWithRatio,
		types.TxTypeFeeDelegatedSmartContractExecutionWithRatio,
		types.TxTypeFeeDelegatedAccountUpdateWithRatio,
		types.TxTypeFeeDelegatedCancelWithRatio,

		types.TxTypeChainDataAnchoring,
	}

	// tests for all accountKeyTypes
	for _, accountKeyType := range accountKeyTypes {
		// a sender account
		sender, err := createDefaultAccount(accountKeyType)
		assert.Equal(t, nil, err)

		if testing.Verbose() {
			fmt.Println("reservoirAddr = ", reservoir.Addr.String())
			fmt.Println("senderAddr = ", sender.Addr.String())
		}

		// create an account decoupled using TxTypeAccountCreation.
		{
			var txs types.Transactions

			amount := new(big.Int).SetUint64(params.KLAY)
			values := map[types.TxValueKeyType]interface{}{
				types.TxValueKeyNonce:         reservoir.Nonce,
				types.TxValueKeyFrom:          reservoir.Addr,
				types.TxValueKeyTo:            sender.Addr,
				types.TxValueKeyAmount:        amount,
				types.TxValueKeyGasLimit:      gasLimit,
				types.TxValueKeyGasPrice:      gasPrice,
				types.TxValueKeyHumanReadable: false,
				types.TxValueKeyAccountKey:    sender.AccKey,
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

		// tests for all txTypes
		for _, txType := range txTypes {
			// skip if tx type is legacy transaction and sender is not legacy.
			if txType == types.TxTypeLegacyTransaction &&
				!sender.AccKey.Type().IsLegacyAccountKey() {
				continue
			}

			if testing.Verbose() {
				fmt.Println("Testing... accountKeyType: ", accountKeyType, ", txType: ", txType)
			}

			// generate a default transaction
			tx, _, err := generateDefaultTx(sender, reservoir, txType)
			assert.Equal(t, nil, err)

			// sign a tx
			tx, err = signTxWithVariousKeyTypes(signer, tx, sender)
			assert.Equal(t, nil, err)

			if txType.IsFeeDelegatedTransaction() {
				err = tx.SignFeePayerWithKeys(signer, reservoir.Keys)
				assert.Equal(t, nil, err)
			}

			expectedError := expectedTestResultForDefaultTx(accountKeyType, txType)

			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, expectedError, err)

			if err == nil {
				assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
			}
		}
	}
	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationHumanReadableFail tests account creation with an invalid address.
// For a human-readable account, only alphanumeric characters are allowed in the address.
// In addition, the first character should be an alphabet and the address should contain ".klaytn" suffix.
// The valid length are 12 ~ 20 including the suffix.
// 1. Non-alphanumeric characters in the address.
// 2. The first character of the address is a number.
// 3. The too short address.
// 4. The too long address.
// 5. Invalid suffix.
// 6. A valid address, "humanReadable.klaytn"
func TestAccountCreationHumanReadableFail(t *testing.T) {
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

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	key, err := crypto.HexToECDSA("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// 1. Non-alphanumeric characters in the address.
	{
		readable := &TestAccountType{
			Addr:   common.HexToAddress("68756d616e5265616461626c655f2e6b6c6179746e"), // humanReadable_.klaytn
			Keys:   []*ecdsa.PrivateKey{key},
			Nonce:  uint64(0),
			AccKey: accountkey.NewAccountKeyPublicWithValue(&key.PublicKey),
		}

		if testing.Verbose() {
			fmt.Println("reservoirAddr = ", reservoir.Addr.String())
			fmt.Println("anonAddr = ", readable.Addr.String())
		}

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            readable.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    readable.AccKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrNotHumanReadableAddress, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrNotHumanReadableAddress, err)
		}
	}

	// 2. The first character of the address is a number.
	{
		readable := &TestAccountType{
			Addr:   common.HexToAddress("30756d616e5265616461626c652e6b6c6179746e"), // 0umanReadable.klaytn
			Keys:   []*ecdsa.PrivateKey{key},
			Nonce:  uint64(0),
			AccKey: accountkey.NewAccountKeyPublicWithValue(&key.PublicKey),
		}

		if testing.Verbose() {
			fmt.Println("reservoirAddr = ", reservoir.Addr.String())
			fmt.Println("anonAddr = ", readable.Addr.String())
		}

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            readable.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    readable.AccKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrNotHumanReadableAddress, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrNotHumanReadableAddress, err)
		}
	}

	// 3. The too short address.
	{
		readable := &TestAccountType{
			Addr:   common.HexToAddress("413233342e6b6c6179746e"), // A234.klaytn
			Keys:   []*ecdsa.PrivateKey{key},
			Nonce:  uint64(0),
			AccKey: accountkey.NewAccountKeyPublicWithValue(&key.PublicKey),
		}

		if testing.Verbose() {
			fmt.Println("reservoirAddr = ", reservoir.Addr.String())
			fmt.Println("anonAddr = ", readable.Addr.String())
		}

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            readable.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    readable.AccKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrNotHumanReadableAddress, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrNotHumanReadableAddress, err)
		}
	}

	// 4. The too long address.
	{
		readable := &TestAccountType{
			Addr:   common.HexToAddress("41323334353637383930313233342e6b6c6179746e"), // A234567890123.klaytn
			Keys:   []*ecdsa.PrivateKey{key},
			Nonce:  uint64(0),
			AccKey: accountkey.NewAccountKeyPublicWithValue(&key.PublicKey),
		}

		if testing.Verbose() {
			fmt.Println("reservoirAddr = ", reservoir.Addr.String())
			fmt.Println("anonAddr = ", readable.Addr.String())
		}

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            readable.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    readable.AccKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrNotHumanReadableAddress, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrNotHumanReadableAddress, err)
		}
	}

	// 5. Invalid suffix.
	{
		readable := &TestAccountType{
			Addr:   common.HexToAddress("68756d616e5265616461626c652e616263646566"), // humanReadable.abcdef
			Keys:   []*ecdsa.PrivateKey{key},
			Nonce:  uint64(0),
			AccKey: accountkey.NewAccountKeyPublicWithValue(&key.PublicKey),
		}

		if testing.Verbose() {
			fmt.Println("reservoirAddr = ", reservoir.Addr.String())
			fmt.Println("anonAddr = ", readable.Addr.String())
		}

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            readable.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    readable.AccKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrNotHumanReadableAddress, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrNotHumanReadableAddress, err)
		}
	}

	// 6. A valid address, "humanReadable.klaytn"
	{
		readable := &TestAccountType{
			Addr:   common.HexToAddress("68756d616e5265616461626c652e6b6c6179746e"), // humanReadable.klaytn
			Keys:   []*ecdsa.PrivateKey{key},
			Nonce:  uint64(0),
			AccKey: accountkey.NewAccountKeyPublicWithValue(&key.PublicKey),
		}

		if testing.Verbose() {
			fmt.Println("reservoirAddr = ", reservoir.Addr.String())
			fmt.Println("anonAddr = ", readable.Addr.String())
		}

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            readable.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    readable.AccKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationMultiSigKeyMaxKey tests multiSig account creation with maximum private keys.
// A multiSig account supports maximum 10 different private keys.
// Create a multiSig account with 11 different private keys (more than 10 -> failed)
func TestAccountCreationMultiSigKeyMaxKey(t *testing.T) {
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

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// multisig setting
	threshold := uint(10)
	weights := []uint{1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 0, 1}
	multisigAddr := common.HexToAddress("0xbbfa38050bf3167c887c086758f448ce067ea8ea")
	prvKeys := []string{
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380000",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380001",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380002",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380003",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380004",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380005",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380006",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594300007",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594300008",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594300009",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594300010",
	}

	// multi-sig account
	multisig, err := createMultisigAccount(threshold, weights, prvKeys, multisigAddr)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("multisigAddr = ", multisig.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// Create a multiSig account with 11 different private keys (more than 10 -> failed)
	{
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

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrMaxKeysExceed, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, kerrors.ErrMaxKeysExceed, err)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
		}

		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationMultiSigKeyBigThreshold tests multiSig account creation with abnormal threshold.
// When a multisig account is created, a threshold value should be less or equal to the total weight of private keys.
// If not, the account cannot creates any valid signatures.
// The test creates a multisig account with a threshold (10) and the total weight (6). (failed case)
func TestAccountCreationMultiSigKeyBigThreshold(t *testing.T) {
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

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// multisig setting
	threshold := uint(10)
	weights := []uint{1, 2, 3}
	multisigAddr := common.HexToAddress("0xbbfa38050bf3167c887c086758f448ce067ea8ea")
	prvKeys := []string{
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380000",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380001",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380002",
	}

	// multi-sig account
	multisig, err := createMultisigAccount(threshold, weights, prvKeys, multisigAddr)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("multisigAddr = ", multisig.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// creates a multisig account with a threshold (10) and the total weight (6). (failed case)
	{
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

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrUnsatisfiableThreshold, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrUnsatisfiableThreshold, err)
		}
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationMultiSigKeyDupPrvKeys tests multiSig account creation with duplicated private keys.
// A multisig account has all different private keys, therefore account creation with duplicated private keys should be failed.
// The case when two same private keys are used in creation processes.
func TestAccountCreationMultiSigKeyDupPrvKeys(t *testing.T) {
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

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// the case when two same private keys are used in creation processes.
	threshold := uint(2)
	weights := []uint{1, 1, 2}
	multisigAddr := common.HexToAddress("0xbbfa38050bf3167c887c086758f448ce067ea8ea")
	prvKeys := []string{
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380000",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380001",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380001",
	}

	// multi-sig account
	multisig, err := createMultisigAccount(threshold, weights, prvKeys, multisigAddr)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("multisigAddr = ", multisig.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// the case when two same private keys are used in creation processes.
	{
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

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrDuplicatedKey, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrDuplicatedKey, err)
		}
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationMultiSigKeyWeightOverflow tests multiSig account creation with weight overflow.
// If the sum of weights is overflowed, the test should fail.
func TestAccountCreationMultiSigKeyWeightOverflow(t *testing.T) {
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

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// Simply check & set the maximum value of uint
	MAX := uint(math.MaxUint32)
	if strconv.IntSize == 64 {
		MAX = math.MaxUint64
	}

	// multisig setting
	threshold := uint(MAX)
	weights := []uint{MAX / 2, MAX / 2, MAX / 2}
	multisigAddr := common.HexToAddress("0xbbfa38050bf3167c887c086758f448ce067ea8ea")
	prvKeys := []string{
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380000",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380001",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380002",
	}

	// multi-sig account
	multisig, err := createMultisigAccount(threshold, weights, prvKeys, multisigAddr)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("multisigAddr = ", multisig.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// creates a multisig account with a threshold, uint(MAX), and the total weight, uint(MAX/2)*3. (failed case)
	{
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

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrWeightedSumOverflow, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrWeightedSumOverflow, err)
		}
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationRoleBasedKeyNested tests account creation with nested RoleBasedKey.
// Nested RoleBasedKey is not allowed in Klaytn.
// The test should fail to the account creation
// 1. A key for the first role, RoleTransaction, is nested
// 2. A key for the second role, RoleAccountUpdate, is nested.
// 3. A key for the third role, RoleFeePayer, is nested.
func TestAccountCreationRoleBasedKeyNested(t *testing.T) {
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

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	anon, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)
	anon2, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78990001")
	assert.Equal(t, nil, err)
	anon3, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78990002")
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// 1. A key for the first role, RoleTransaction, is nested
	{
		keys := genTestKeys(3)
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
		})

		keys2 := genTestKeys(2)
		nestedKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			roleKey,
			accountkey.NewAccountKeyPublicWithValue(&keys2[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys2[1].PublicKey),
		})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    nestedKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrNestedCompositeType, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrNestedCompositeType, err)
		}
	}

	// 2. A key for the second role, RoleAccountUpdate, is nested.
	{
		keys := genTestKeys(3)
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
		})

		keys2 := genTestKeys(2)
		nestedKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys2[0].PublicKey),
			roleKey,
			accountkey.NewAccountKeyPublicWithValue(&keys2[1].PublicKey),
		})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon2.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    nestedKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrNestedCompositeType, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrNestedCompositeType, err)
		}
	}

	// 3. A key for the third role, RoleFeePayer, is nested.
	{
		keys := genTestKeys(3)
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
		})

		keys2 := genTestKeys(2)
		nestedKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys2[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys2[1].PublicKey),
			roleKey,
		})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon3.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    nestedKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrNestedCompositeType, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrNestedCompositeType, err)
		}
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationRoleBasedKeyInvalidNumKey tests account creation with a RoleBased key which contains invalid number of sub-keys.
// A RoleBased key can contain 1 ~ 3 sub-keys, otherwise it will fail to the account creation.
// 1. try to create an account with a RoleBased key which contains 4 sub-keys.
// 2. try to create an account with a RoleBased key which contains 0 sub-key.
func TestAccountCreationRoleBasedKeyInvalidNumKey(t *testing.T) {
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

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	anon, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// 1. try to create an account with a RoleBased key which contains 4 sub-keys.
	{
		keys := genTestKeys(4)
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[3].PublicKey),
		})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrLengthTooLong, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrLengthTooLong, err)
		}
	}

	// 2. try to create an account with a RoleBased key which contains 0 sub-key.
	{
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrZeroLength, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrZeroLength, err)
		}
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationRoleBasedKeyInvalidTypeKey tests account creation with a RoleBased key contains types of sub-keys.
// As a sub-key type, a RoleBased key can have AccountKeyFail keys but not AccountKeyNil keys.
// 1. a RoleBased key contains an AccountKeyNil type sub-key as a first sub-key. (fail)
// 2. a RoleBased key contains an AccountKeyNil type sub-key as a second sub-key. (fail)
// 3. a RoleBased key contains an AccountKeyNil type sub-key as a third sub-key. (fail)
// 4. a RoleBased key contains an AccountKeyFail type sub-key as a first sub-key. (success)
// 5. a RoleBased key contains an AccountKeyFail type sub-key as a second sub-key. (success)
// 6. a RoleBased key contains an AccountKeyFail type sub-key as a third sub-key. (success)
func TestAccountCreationRoleBasedKeyInvalidTypeKey(t *testing.T) {
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

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	anon, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)
	keys := genTestKeys(2)

	// 1. a RoleBased key contains an AccountKeyNil type sub-key as a first sub-key. (fail)
	{
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyNil(),
			accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
		})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrAccountKeyNilUninitializable, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrAccountKeyNilUninitializable, err)
		}
	}

	// 2. a RoleBased key contains an AccountKeyNil type sub-key as a second sub-key. (fail)
	{
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
			accountkey.NewAccountKeyNil(),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
		})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrAccountKeyNilUninitializable, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrAccountKeyNilUninitializable, err)
		}
	}

	// 3. a RoleBased key contains an AccountKeyNil type sub-key as a third sub-key. (fail)
	{
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
			accountkey.NewAccountKeyNil(),
		})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrAccountKeyNilUninitializable, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrAccountKeyNilUninitializable, err)
		}
	}

	// 4. a RoleBased key contains an AccountKeyFail type sub-key as a first sub-key. (success)
	{
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyFail(),
			accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
		})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	}

	// 5. a RoleBased key contains an AccountKeyFail type sub-key as a second sub-key. (success)
	{
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
			accountkey.NewAccountKeyFail(),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
		})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	}

	// 6. a RoleBased key contains an AccountKeyFail type sub-key as a third sub-key. (success)
	{
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
			accountkey.NewAccountKeyFail(),
		})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationUpdateRoleBasedKey tests account creation and update with a roleBased key.
// A roleBased key can contain three types of sub-keys, but two sub-keys are used in this test.
// 0. create an account creator's account, "accountK".
// 1. "accountK" creates a role-based account, "roleBasedAccount", with a human-readable address.
// 2. "roleBasedAccount" updates its transaction key.
func TestAccountCreationUpdateRoleBasedKey(t *testing.T) {
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

	// raw transaction to be printed in verbose mode
	var txCreation *types.Transaction
	var txUpdate *types.Transaction

	// key, address setting
	private_key := "4f047422233bca64fb5973489dc274a0ade80d270749fb7c4e6231a08cc6bd30"
	user_key_1 := "e62544af405e9e6512ebbef81721ba7428a7914dadacb44bea2a86426d8a8b96"
	user_key_2 := "227ffb0a420d70f344be9410a9918e411dd8bc1c46ee0e966751db4a6c086de3"
	user_key_3 := "cc2e738550d8df28ad840d7aa8bfb87bf21798e3f3cbd953e0fdc1dea39bc14f"
	humanReadableAddr, err := common.FromHumanReadableAddress("humanReadable" + ".klaytn")
	if err != nil {
		t.Fatal(err)
	}

	// "accountK" will create an role-based account, "roleBasedAccount"
	accountK, err := createAnonymousAccount(private_key)
	assert.Equal(t, nil, err)

	// prepare private keys and account keys for a role-based account named "roleBasedAccount"
	txKey, err := crypto.HexToECDSA(user_key_1)
	if err != nil {
		t.Fatal(err)
	}

	updateKey, err := crypto.HexToECDSA(user_key_2)
	if err != nil {
		t.Fatal(err)
	}

	newTxKey, err := crypto.HexToECDSA(user_key_3)
	if err != nil {
		t.Fatal(err)
	}

	// account key for "roleBasedAccount"
	roleBasedAccKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
		accountkey.NewAccountKeyPublicWithValue(&txKey.PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&updateKey.PublicKey),
	})

	// new account key will replace the old account key
	newRoleBasedAccKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
		accountkey.NewAccountKeyPublicWithValue(&newTxKey.PublicKey),
		accountkey.NewAccountKeyNil(),
	})

	// "roleBasedAccount" initial setting
	roleBasedAccount := &TestRoleBasedAccountType{
		Addr:       humanReadableAddr,
		TxKeys:     []*ecdsa.PrivateKey{txKey},
		UpdateKeys: []*ecdsa.PrivateKey{updateKey},
		Nonce:      uint64(0),
		AccKey:     roleBasedAccKey,
	}

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("accountK = ", accountK.Addr.String())
		fmt.Println("roleBasedAccount = ", roleBasedAccount.Addr.String())
	}

	//signer := types.NewEIP155Signer(new(big.Int).SetUint64(1001))    // signer with Baobab chainID
	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// 0. create an account creator's account, "accountK".
	{
		var txs types.Transactions
		amount := new(big.Int).Mul(big.NewInt(5000), new(big.Int).SetUint64(params.KLAY))
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            accountK.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    accountK.AccKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}

		reservoir.Nonce += 1
	}

	// 1. "accountK" creates a role-based account, "roleBasedAccount", with a human-readable address.
	{
		var txs types.Transactions
		amount := new(big.Int).Mul(big.NewInt(2500), new(big.Int).SetUint64(params.KLAY)) // 2500 KLAY to pay for accountUpdate tx
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         accountK.Nonce,
			types.TxValueKeyFrom:          accountK.Addr,
			types.TxValueKeyTo:            roleBasedAccount.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    roleBasedAccount.AccKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)

		err = tx.SignWithKeys(signer, accountK.Keys)
		assert.Equal(t, nil, err)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}

		accountK.Nonce += 1

		txCreation = tx
	}

	// 2. "roleBasedAccount" updates its transaction key.
	{
		var txs types.Transactions
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      roleBasedAccount.Nonce,
			types.TxValueKeyFrom:       roleBasedAccount.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: newRoleBasedAccKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)

		err = tx.SignWithKeys(signer, roleBasedAccount.UpdateKeys)
		assert.Equal(t, nil, err)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}

		roleBasedAccount.TxKeys = []*ecdsa.PrivateKey{newTxKey}
		roleBasedAccount.AccKey = newRoleBasedAccKey
		roleBasedAccount.Nonce += 1

		txUpdate = tx
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()

		// print information of update tx ('HEX' field data is rawTransaction data)
		fmt.Println("\n[Account creation tx]")
		fmt.Println(txCreation)

		// how to make raw transaction data from tx
		rawTxCreation, err := rlp.EncodeToBytes(txCreation)
		if err != nil {
			t.Fatal(err)
		}
		rawTxCreationString := hexutil.Encode(rawTxCreation)
		fmt.Println("\tRAW TX: ", rawTxCreationString)

		// print information of update tx
		fmt.Println("\n[Account update raw tx]")
		fmt.Println(txUpdate)

		// how to make raw transaction data from tx
		rawTxUpdate, err := rlp.EncodeToBytes(txUpdate)
		if err != nil {
			t.Fatal(err)
		}
		rawTxUpdateString := hexutil.Encode(rawTxUpdate)
		fmt.Println("\tRAW TX: ", rawTxUpdateString)
	}
}

// TestAccountUpdateWithRoleBasedKey tests account update with a roleBased key.
// A roleBased key contains three types of sub-keys, and only RoleAccountUpdate key is used for update.
// Other sub-keys are not used for the account update.
// 0. create an account with a roleBased key.
// 1. try to update the account with a RoleTransaction key. (fail)
// 2. try to update the account with a RoleFeePayer key. (fail)
// 3. try to update the account with a RoleAccountUpdate key. (success)
func TestAccountUpdateRoleBasedKey(t *testing.T) {
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

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	anon, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	// generate a roleBased key
	keys := genTestKeys(3)
	roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
		accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
	})

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// 0. create an account with a roleBased key.
	{
		var txs types.Transactions
		amount := new(big.Int).Mul(big.NewInt(2500), new(big.Int).SetUint64(params.KLAY))
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}

		reservoir.Nonce += 1
	}

	// 1. try to update the account with a RoleTransaction key. (fail)
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      anon.Nonce,
			types.TxValueKeyFrom:       anon.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: roleKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, []*ecdsa.PrivateKey{keys[accountkey.RoleTransaction]})
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, types.ErrInvalidSigSender, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, types.ErrInvalidSigSender, err)
		}
	}

	// 2. try to update the account with a RoleFeePayer key. (fail)
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      anon.Nonce,
			types.TxValueKeyFrom:       anon.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: roleKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, []*ecdsa.PrivateKey{keys[accountkey.RoleFeePayer]})
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, types.ErrInvalidSigSender, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, types.ErrInvalidSigSender, err)
		}
	}

	// 3. try to update the account with a RoleAccountUpdate key. (success)
	{
		var txs types.Transactions
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      anon.Nonce,
			types.TxValueKeyFrom:       anon.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: roleKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)

		err = tx.SignWithKeys(signer, []*ecdsa.PrivateKey{keys[accountkey.RoleAccountUpdate]})
		assert.Equal(t, nil, err)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}

		anon.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountUpdateRoleBasedKeyNested tests account update with a nested RoleBasedKey.
// Nested RoleBasedKey is not allowed in Klaytn.
// 1. Create an account with a RoleBasedKey.
// 2. Update an accountKey with a nested RoleBasedKey
func TestAccountUpdateRoleBasedKeyNested(t *testing.T) {
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

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// roleBasedKeys and a nested roleBasedKey
	roleKey, err := createDefaultAccount(accountkey.AccountKeyTypeRoleBased)
	assert.Equal(t, nil, err)

	roleKey2, err := createDefaultAccount(accountkey.AccountKeyTypeRoleBased)
	assert.Equal(t, nil, err)

	nestedAccKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
		roleKey2.AccKey,
	})

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("roleAddr = ", roleKey.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// 1. Create an account with a RoleBasedKey.
	{
		var txs types.Transactions

		amount := new(big.Int).Mul(big.NewInt(2500), new(big.Int).SetUint64(params.KLAY))
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            roleKey.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleKey.AccKey,
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

	// 2. Update an accountKey with a nested RoleBasedKey.
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      roleKey.Nonce,
			types.TxValueKeyFrom:       roleKey.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: nestedAccKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, []*ecdsa.PrivateKey{roleKey.Keys[accountkey.RoleAccountUpdate]})
		assert.Equal(t, nil, err)

		// For tx pool validation test
		{
			err = txpool.AddRemote(tx)
			assert.Equal(t, kerrors.ErrNestedCompositeType, err)
		}

		// For block tx validation test
		{
			receipt, _, err := applyTransaction(t, bcdata, tx)
			assert.Equal(t, (*types.Receipt)(nil), receipt)
			assert.Equal(t, kerrors.ErrNestedCompositeType, err)
		}
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestRoleBasedKeySendTx tests signing transactions with a role-based key.
// A role-based key contains three types of sub-keys: RoleTransaction, RoleAccountUpdate, RoleFeePayer.
// Only RoleTransaction can generate valid signature as a sender except account update txs.
// RoleAccountUpdate can generate valid signature for account update txs.
func TestRoleBasedKeySendTx(t *testing.T) {
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

	gasPrice := new(big.Int).SetUint64(25 * params.Ston)

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// main account with a role-based key
	roleBased, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	// smart contract account
	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	// generate a role-based key
	prvKeys := genTestKeys(3)
	roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
		accountkey.NewAccountKeyPublicWithValue(&prvKeys[0].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&prvKeys[1].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&prvKeys[2].PublicKey),
	})

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("roleBasedAddr = ", roleBased.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	txTypes := []types.TxType{
		//types.TxTypeLegacyTransaction, // accounts with role-based key cannot a send legacy tx.

		types.TxTypeValueTransfer,
		types.TxTypeValueTransferMemo,
		types.TxTypeSmartContractDeploy,
		types.TxTypeSmartContractExecution,
		types.TxTypeAccountCreation,
		types.TxTypeAccountUpdate,
		types.TxTypeCancel,

		types.TxTypeFeeDelegatedValueTransfer,
		types.TxTypeFeeDelegatedValueTransferMemo,
		types.TxTypeFeeDelegatedSmartContractDeploy,
		types.TxTypeFeeDelegatedSmartContractExecution,
		types.TxTypeFeeDelegatedAccountUpdate,
		types.TxTypeFeeDelegatedCancel,

		types.TxTypeFeeDelegatedValueTransferWithRatio,
		types.TxTypeFeeDelegatedValueTransferMemoWithRatio,
		types.TxTypeFeeDelegatedSmartContractDeployWithRatio,
		types.TxTypeFeeDelegatedSmartContractExecutionWithRatio,
		types.TxTypeFeeDelegatedAccountUpdateWithRatio,
		types.TxTypeFeeDelegatedCancelWithRatio,

		types.TxTypeChainDataAnchoring,
	}

	// deploy a contract to test smart contract execution.
	{
		var txs types.Transactions
		valueMap := genMapForTxTypes(reservoir, reservoir, types.TxTypeSmartContractDeploy)
		valueMap[types.TxValueKeyTo] = &contract.Addr
		valueMap[types.TxValueKeyHumanReadable] = true

		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, valueMap)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}

		reservoir.Nonce += 1
	}

	// create an roleBased account with a role-based key.
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
			types.TxValueKeyAccountKey:    roleKey,
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

	// test fee delegation txs for each role of role-based key.
	// only RoleFeePayer type can generate valid signature as a fee payer.
	for keyType, key := range prvKeys {
		for _, txType := range txTypes {
			valueMap := genMapForTxTypes(roleBased, reservoir, txType)
			valueMap[types.TxValueKeyGasLimit] = uint64(1000000)

			if txType.IsFeeDelegatedTransaction() {
				valueMap[types.TxValueKeyFeePayer] = reservoir.Addr
			}

			// Currently, test VM is not working properly when the GasPrice is not 0.
			basicType := toBasicType(txType)
			if keyType == int(accountkey.RoleTransaction) {
				if basicType == types.TxTypeSmartContractDeploy || basicType == types.TxTypeSmartContractExecution {
					valueMap[types.TxValueKeyGasPrice] = new(big.Int).SetUint64(0)
				}
			}

			tx, err := types.NewTransactionWithMap(txType, valueMap)
			assert.Equal(t, nil, err)

			err = tx.SignWithKeys(signer, []*ecdsa.PrivateKey{key})
			assert.Equal(t, nil, err)

			if txType.IsFeeDelegatedTransaction() {
				err = tx.SignFeePayerWithKeys(signer, reservoir.Keys)
				assert.Equal(t, nil, err)
			}

			// Only RoleTransaction can generate valid signature as a sender except account update txs.
			// RoleAccountUpdate can generate valid signature for account update txs.
			if keyType == int(accountkey.RoleAccountUpdate) && txType.IsAccountUpdate() ||
				keyType == int(accountkey.RoleTransaction) && !txType.IsAccountUpdate() {
				// Do not make a block since account update tx can change sender's keys.
				receipt, _, err := applyTransaction(t, bcdata, tx)
				assert.Equal(t, nil, err)
				assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
			} else {
				// For tx pool validation test
				{
					err = txpool.AddRemote(tx)
					assert.Equal(t, types.ErrInvalidSigSender, err)
				}

				// For block tx validation test
				{
					receipt, _, err := applyTransaction(t, bcdata, tx)
					assert.Equal(t, types.ErrInvalidSigSender, err)
					assert.Equal(t, (*types.Receipt)(nil), receipt)
				}
			}
		}
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestRoleBasedKeyFeeDelegation tests fee delegation with a role-based key.
// A role-based key contains three types of sub-keys: RoleTransaction, RoleAccountUpdate, RoleFeePayer.
// Only RoleFeePayer can sign txs as a fee payer.
func TestRoleBasedKeyFeeDelegation(t *testing.T) {
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

	gasPrice := new(big.Int).SetUint64(25 * params.Ston)

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// main account with a role-based key
	roleBased, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	// smart contract account
	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	// generate a role-based key
	prvKeys := genTestKeys(3)
	roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
		accountkey.NewAccountKeyPublicWithValue(&prvKeys[0].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&prvKeys[1].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&prvKeys[2].PublicKey),
	})

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("roleBasedAddr = ", roleBased.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	feeTxTypes := []types.TxType{
		types.TxTypeFeeDelegatedValueTransfer,
		types.TxTypeFeeDelegatedValueTransferMemo,
		types.TxTypeFeeDelegatedSmartContractDeploy,
		types.TxTypeFeeDelegatedSmartContractExecution,
		types.TxTypeFeeDelegatedAccountUpdate,
		types.TxTypeFeeDelegatedCancel,

		types.TxTypeFeeDelegatedValueTransferWithRatio,
		types.TxTypeFeeDelegatedValueTransferMemoWithRatio,
		types.TxTypeFeeDelegatedSmartContractDeployWithRatio,
		types.TxTypeFeeDelegatedSmartContractExecutionWithRatio,
		types.TxTypeFeeDelegatedAccountUpdateWithRatio,
		types.TxTypeFeeDelegatedCancelWithRatio,
	}

	// deploy a contract to test smart contract execution.
	{
		var txs types.Transactions
		valueMap := genMapForTxTypes(reservoir, reservoir, types.TxTypeSmartContractDeploy)
		valueMap[types.TxValueKeyTo] = &contract.Addr
		valueMap[types.TxValueKeyHumanReadable] = true

		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, valueMap)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}

		reservoir.Nonce += 1
	}

	// create an roleBased account with a role-based key.
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
			types.TxValueKeyAccountKey:    roleKey,
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

	// test fee delegation txs for each role of role-based key.
	// only RoleFeePayer type can generate valid signature as a fee payer.
	for keyType, key := range prvKeys {
		for _, txType := range feeTxTypes {
			valueMap := genMapForTxTypes(reservoir, reservoir, txType)
			valueMap[types.TxValueKeyFeePayer] = roleBased.GetAddr()
			valueMap[types.TxValueKeyGasLimit] = uint64(1000000)

			// Currently, test VM is not working properly when the GasPrice is not 0.
			basicType := toBasicType(txType)
			if keyType == int(accountkey.RoleFeePayer) {
				if basicType == types.TxTypeSmartContractDeploy || basicType == types.TxTypeSmartContractExecution {
					valueMap[types.TxValueKeyGasPrice] = new(big.Int).SetUint64(0)
				}
			}

			tx, err := types.NewTransactionWithMap(txType, valueMap)
			assert.Equal(t, nil, err)

			err = tx.SignWithKeys(signer, reservoir.Keys)
			assert.Equal(t, nil, err)

			err = tx.SignFeePayerWithKeys(signer, []*ecdsa.PrivateKey{key})
			assert.Equal(t, nil, err)

			if keyType == int(accountkey.RoleFeePayer) {
				// Do not make a block since account update tx can change sender's keys.
				receipt, _, err := applyTransaction(t, bcdata, tx)
				assert.Equal(t, nil, err)
				assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
			} else {
				// For tx pool validation test
				{
					err = txpool.AddRemote(tx)
					assert.Equal(t, blockchain.ErrInvalidFeePayer, err)
				}

				// For block tx validation test
				{
					receipt, _, err := applyTransaction(t, bcdata, tx)
					assert.Equal(t, types.ErrInvalidSigFeePayer, err)
					assert.Equal(t, (*types.Receipt)(nil), receipt)
				}
			}
		}
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}
