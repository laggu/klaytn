package tests

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ground-x/klaytn/accounts/abi"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/hexutil"
	"github.com/ground-x/klaytn/common/math"
	"github.com/ground-x/klaytn/common/profile"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/kerrors"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/stretchr/testify/assert"
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
	prvKeysHex := []string{
		"bb113e82881499a7a361e8354a5b68f6c6885c7bcba09ea2b0891480396c322e",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae59438e989",
		"c32c471b732e2f56103e2f8e8cfd52792ef548f05f326e546a7d1fbf9d0419ec",
	}

	keys := make([]*ecdsa.PrivateKey, len(prvKeysHex))
	weights := []uint{1, 1, 1}
	weightedKeys := make(accountkey.WeightedPublicKeys, len(prvKeysHex))
	threshold := uint(2)

	for i, p := range prvKeysHex {
		keys[i], err = crypto.HexToECDSA(p)
		if err != nil {
			return nil, err
		}
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
// If txType is a kind of account update, it will return a updatedKey value.
// Otherwise, it will return (tx, nil, nil).
// For contract execution Txs, TxValueKeyTo value is set to "contract" as a default.
// The address "contact" should exist before calling this function.
func generateDefaultTx(sender *TestAccountType, recipient *TestAccountType, txType types.TxType) (*types.Transaction, *ecdsa.PrivateKey, error) {
	gasPrice := new(big.Int).SetUint64(25)
	gasLimit := uint64(1000000000)
	amount := new(big.Int).SetUint64(1000000000)

	// generate a random private key for account creation/update Txs or contract deploy Txs
	newKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, err
	}

	newAccKey := accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey)
	newAddr := crypto.PubkeyToAddress(newKey.PublicKey)

	// a new private key will be returned
	var retNewKey *ecdsa.PrivateKey = nil

	// a default recipient address of smart contract execution to "contract"
	var contractAddr common.Address
	contractAddr.SetBytesFromFront([]byte("contract"))

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
		values[types.TxValueKeyTo] = newAddr
		values[types.TxValueKeyAmount] = amount
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyHumanReadable] = false
		values[types.TxValueKeyAccountKey] = newAccKey
	case types.TxTypeAccountUpdate:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyAccountKey] = newAccKey
	case types.TxTypeFeeDelegatedAccountUpdate:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyAccountKey] = newAccKey
		values[types.TxValueKeyFeePayer] = recipient.Addr
	case types.TxTypeFeeDelegatedAccountUpdateWithRatio:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = gasPrice
		values[types.TxValueKeyAccountKey] = newAccKey
		values[types.TxValueKeyFeePayer] = recipient.Addr
		values[types.TxValueKeyFeeRatioOfFeePayer] = ratio
	case types.TxTypeSmartContractDeploy:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = newAddr
		values[types.TxValueKeyAmount] = amountZero
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = amountZero
		values[types.TxValueKeyData] = dataCode
		values[types.TxValueKeyHumanReadable] = false
	case types.TxTypeFeeDelegatedSmartContractDeploy:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = newAddr
		values[types.TxValueKeyAmount] = amountZero
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = amountZero
		values[types.TxValueKeyData] = dataCode
		values[types.TxValueKeyHumanReadable] = false
		values[types.TxValueKeyFeePayer] = recipient.Addr
	case types.TxTypeFeeDelegatedSmartContractDeployWithRatio:
		values[types.TxValueKeyNonce] = sender.Nonce
		values[types.TxValueKeyFrom] = sender.Addr
		values[types.TxValueKeyTo] = newAddr
		values[types.TxValueKeyAmount] = amountZero
		values[types.TxValueKeyGasLimit] = gasLimit
		values[types.TxValueKeyGasPrice] = amountZero
		values[types.TxValueKeyData] = dataCode
		values[types.TxValueKeyHumanReadable] = false
		values[types.TxValueKeyFeePayer] = recipient.Addr
		values[types.TxValueKeyFeeRatioOfFeePayer] = ratio
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
	}

	tx, err := types.NewTransactionWithMap(txType, values)
	if err != nil {
		return nil, nil, err
	}

	// the function returns a private key for account update Txs
	if txType.IsAccountUpdate() {
		retNewKey = newKey
	}

	return tx, retNewKey, err
}

// expectedTestResultForDefaultTx returns expected validity of tx which generated from (accountKeyType, txType) pair.
func expectedTestResultForDefaultTx(accountKeyType accountkey.AccountKeyType, txType types.TxType) bool {
	// TODO-Klaytn-Core. legacy tx will be prohibited to all account types execpt legacy account type
	switch accountKeyType {
	//case accountkey.AccountKeyTypeNil:                     // not supported type
	case accountkey.AccountKeyTypeLegacy:
		return true
	case accountkey.AccountKeyTypePublic:
		return true
	case accountkey.AccountKeyTypeFail:
		if txType.IsLegacyTransaction() {
			return true
		}
		return false
	case accountkey.AccountKeyTypeWeightedMultiSig:
		return true
	case accountkey.AccountKeyTypeRoleBased:
		return true
	}

	return false
}

// TestDefaultTxsWithDefaultAccountKey tests most of transactions types with most of account key types.
// The test creates a default account for each account key type, and generates default Tx for each Tx type.
// AccountKeyTypeNil is excluded because it cannot be used for account creation.
// TxTypeChainDataAnchoring is excluded because it is not fully developed.
func TestDefaultTxsWithDefaultAccountKey(t *testing.T) {
	gasPrice := new(big.Int).SetUint64(25)
	gasLimit := uint64(1000000000)

	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()

	// select account key types to be tested
	accountKeyTypes := []accountkey.AccountKeyType{
		//accountkey.AccountKeyTypeNil,               // not supported type
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

		// types.TxTypeChainDataAnchoring,             // not supported type
	}

	// tests for all accountKeyTypes
	for _, accountKeyType := range accountKeyTypes {
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

		// a sender account
		sender, err := createDefaultAccount(accountKeyType)
		assert.Equal(t, nil, err)

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

		if testing.Verbose() {
			fmt.Println("reservoirAddr = ", reservoir.Addr.String())
			fmt.Println("senderAddr = ", sender.Addr.String())
		}

		signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

		// create an account decoupled using TxTypeAccountCreation.
		{
			var txs types.Transactions

			amount := new(big.Int).SetUint64(1000000000000)
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

		// create a smart contract account for contract execution test
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

			err = bcdata.GenABlockWithTransactions(accountMap, txs, prof)
			assert.Equal(t, nil, err)

			reservoir.Nonce += 1
		}

		{
			// tests for all txTypes
			for _, txType := range txTypes {
				var txs types.Transactions
				var tx *types.Transaction
				var newKey *ecdsa.PrivateKey = nil

				// generate a default transaction
				tx, newKey, err = generateDefaultTx(sender, reservoir, txType)
				assert.Equal(t, nil, err)

				// sign a tx
				if accountKeyType == accountkey.AccountKeyTypeWeightedMultiSig {
					if txType == types.TxTypeLegacyTransaction {
						err = tx.SignWithKeys(signer, []*ecdsa.PrivateKey{sender.Keys[0]})
					} else {
						err = tx.SignWithKeys(signer, sender.Keys)
					}
				} else if accountKeyType == accountkey.AccountKeyTypeRoleBased {
					if txType == types.TxTypeAccountUpdate {
						err = tx.SignWithKeys(signer, []*ecdsa.PrivateKey{sender.Keys[accountkey.RoleAccountUpdate]})
					} else {
						err = tx.SignWithKeys(signer, []*ecdsa.PrivateKey{sender.Keys[accountkey.RoleTransaction]})
					}
				} else {
					err = tx.SignWithKeys(signer, sender.Keys)
				}
				assert.Equal(t, nil, err)

				// isFeePay to notice outside of this function that the tx required delegator's signature
				if txType.IsFeeDelegatedTransaction() {
					err = tx.SignFeePayerWithKeys(signer, reservoir.Keys)
					assert.Equal(t, nil, err)
				}

				txs = append(txs, tx)

				isSuccess := expectedTestResultForDefaultTx(accountKeyType, txType)

				if testing.Verbose() {
					fmt.Println("Testing... accountKeyType: ", accountKeyType, ", txType: ", txType, "isSuccess: ", isSuccess)
				}

				if isSuccess {
					err = bcdata.GenABlockWithTransactions(accountMap, txs, prof)
					assert.Equal(t, nil, err)

					// update sender's key
					if newKey != nil {
						sender.AccKey = accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey)
						sender.Keys = []*ecdsa.PrivateKey{newKey}
					}
					sender.Nonce += 1
				} else {
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

// TestAccountCreationMaxMultiSigKey tests multiSig account creation with maximum private keys.
// A multiSig account supports maximum 10 different private keys.
// Create a multiSig account with 11 different private keys (more than 10 -> failed)
func TestAccountCreationMaxMultiSigKey(t *testing.T) {
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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, kerrors.ErrMaxKeysExceed, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)

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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrUnsatisfiableThreshold, receipt.Status)

		reservoir.Nonce += 1
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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrNestedRoleBasedKey, receipt.Status)
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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrNestedRoleBasedKey, receipt.Status)

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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrNestedRoleBasedKey, receipt.Status)
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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrLengthTooLong, receipt.Status)
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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrZeroLength, receipt.Status)
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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrDuplicatedKey, receipt.Status)

		reservoir.Nonce += 1
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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrWeightedSumOverflow, receipt.Status)

		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationHumanReadableFail tests account creation with an invalid address.
// For a human-readable account, only alphanumeric characters are allowed in the address.
// In addition, the first character should be an alphabet.
// 1. Non-alphanumeric characters in the address.
// 2. The first character of the address is a number.
// 3. A valid address, "humanReadable"
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
			Addr:   common.HexToAddress("68756d616e5265616461626c655f000000000000"), // humanReadable_
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
		assert.Equal(t, types.ReceiptStatusErrNotHumanReadableAddress, receipt.Status)
	}

	// 2. The first character of the address is a number.
	{
		readable := &TestAccountType{
			Addr:   common.HexToAddress("3068756d616e5265616461626c65000000000000"), // 0humanReadable
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
		assert.Equal(t, types.ReceiptStatusErrNotHumanReadableAddress, receipt.Status)
	}

	// 3. A valid address, "humanReadable"
	{
		readable := &TestAccountType{
			Addr:   common.HexToAddress("68756d616e5265616461626c6500000000000000"), // humanReadable
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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrAccountKeyNilUninitializable, receipt.Status)
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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrAccountKeyNilUninitializable, receipt.Status)
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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrAccountKeyNilUninitializable, receipt.Status)
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
	humanReadableAddr, err := common.FromHumanReadableAddress("humanReadableAddr")
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
		amount := new(big.Int).Mul(big.NewInt(100), new(big.Int).SetUint64(params.KLAY))
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
		amount := new(big.Int).SetUint64(params.KLAY) // 1 Klay to pay for accountUpdate tx
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         accountK.Nonce,
			types.TxValueKeyFrom:          accountK.Addr,
			types.TxValueKeyTo:            roleBasedAccount.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
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
// 2. try to update the account with a RoleAccountUpdate key. (success)
// 3. try to update the account with a RoleFeePayer key. (fail)
func TestAccountUpdateWithRoleBasedKey(t *testing.T) {
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
		amount := new(big.Int).Mul(big.NewInt(100), new(big.Int).SetUint64(params.KLAY))
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

		err = tx.SignWithKeys(signer, []*ecdsa.PrivateKey{keys[accountkey.RoleTransaction]})
		assert.Equal(t, nil, err)

		err = bcdata.GenABlockWithTransactions(accountMap, txs, prof)
		assert.Equal(t, types.ErrInvalidSigSender, err)
	}

	// 2. try to update the account with a RoleAccountUpdate key. (success)
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

	// 3. try to update the account with a RoleFeePayer key. (fail)
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

		err = tx.SignWithKeys(signer, []*ecdsa.PrivateKey{keys[accountkey.RoleFeePayer]})
		assert.Equal(t, nil, err)

		err = bcdata.GenABlockWithTransactions(accountMap, txs, prof)
		assert.Equal(t, types.ErrInvalidSigSender, err)

		anon.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountTransferWithRoleBasedTxKey tests Txs signed by a RoleTransaction key contained in a role-based key.
// A role-based key contains three types of sub-keys: RoleTransaction, RoleAccountUpdate, RoleFeePayer.
// A RoleTransaction key can sign for any transactions except accountUpdate Txs and FeeDelegating Txs.
// 1. create an account with a role-based key.
// 2. TxTypeValueTransfer signed by a RoleTransaction key.
// 3. TxTypeAccountCreation signed by a RoleTransaction key.
// 4. TxTypeSmartContractDeploy signed by a RoleTransaction key.
// 5. TxTypeSmartContractExecution signed by a RoleTransaction key.
// 6. TxTypeCancel signed by a RoleTransaction key.
// 7. TxTypeChainDataAnchoringTransaction signed by a RoleTransaction key.
func TestAccountTransferWithRoleBasedTxKey(t *testing.T) {
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

	gasPrice := new(big.Int).SetUint64(0)
	gasLimit := uint64(2500000000)

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

	// main account with a role-based key
	roleBased, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	// anonymous account
	anon, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78990000")
	assert.Equal(t, nil, err)

	// smart contract account
	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	// Smart contract used to test TxTypeSmartContractDeploy, TxTypeSmartContractExecution transactions
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

	// generate a role-based key
	keys := genTestKeys(3)
	roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
		accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
	})

	// a RoleTransaction key for signing
	txKey := []*ecdsa.PrivateKey{keys[accountkey.RoleTransaction]}

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("roleBasedAddr = ", roleBased.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. create an account with a role-based key.
	{
		var txs types.Transactions
		amount := new(big.Int).SetUint64(1000000000000000)
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

		txs = append(txs, tx)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}

		reservoir.Nonce += 1
	}

	// 2. TxTypeValueTransfer signed by a RoleTransaction key.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		err = bcdata.GenABlockWithTransactions(accountMap, txs, prof)
		assert.Equal(t, nil, err)

		roleBased.Nonce += 1
	}

	// 3. TxTypeAccountCreation signed by a RoleTransaction key.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         roleBased.Nonce,
			types.TxValueKeyFrom:          roleBased.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    anon.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		err = bcdata.GenABlockWithTransactions(accountMap, txs, prof)
		assert.Equal(t, nil, err)

		roleBased.Nonce += 1
	}

	// 4. TxTypeSmartContractDeploy signed by a RoleTransaction key.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(0)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         roleBased.Nonce,
			types.TxValueKeyFrom:          roleBased.Addr,
			types.TxValueKeyTo:            contract.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          common.FromHex(code),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		err = bcdata.GenABlockWithTransactions(accountMap, txs, prof)
		assert.Equal(t, nil, err)

		roleBased.Nonce += 1
	}

	// 5. TxTypeSmartContractExecution signed by a RoleTransaction key.
	{
		amount := new(big.Int).SetUint64(0)
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

		err = tx.SignWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	}

	// 6. TxTypeCancel signed by a RoleTransaction key.
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeCancel, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountTransferWithRoleBasedUpdateKey tests Txs signed by a RoleAccountUpdate key contained in a role-based key.
// A role-based key contains three types of sub-keys: RoleTransaction, RoleAccountUpdate, RoleFeePayer.
// Since a RoleAccountUpdate key can sign only for accountUpdate Txs, following cases will fail.
// 1. create an account with a role-based key.
// 2. TxTypeValueTransfer signed by a RoleAccountUpdate key.
// 3. TxTypeAccountCreation signed by a RoleAccountUpdate key.
// 4. TxTypeSmartContractDeploy signed by a RoleAccountUpdate key.
// 5. TxTypeSmartContractExecution signed by a RoleAccountUpdate key.
// 6. TxTypeCancel signed by a RoleAccountUpdate key.
// The logic below tests validity of the  signature, but didn't test functionality of the TX execution.
func TestAccountTransferWithRoleBasedUpdateKey(t *testing.T) {
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

	gasPrice := new(big.Int).SetUint64(0)
	gasLimit := uint64(2500000000)

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

	// main account with a role-based key
	roleBased, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	// anonymous account
	anon, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78990000")
	assert.Equal(t, nil, err)

	// smart contract account
	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	// smart contract used to test TxTypeSmartContractDeploy, TxTypeSmartContractExecution transactions
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

	// generate a role-based key
	keys := genTestKeys(3)
	roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
		accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
	})

	// a RoleAccountUpdate key for signing
	upKey := []*ecdsa.PrivateKey{keys[accountkey.RoleAccountUpdate]}

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("roleBasedAddr = ", roleBased.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. create an account with a role-based key.
	{
		amount := new(big.Int).SetUint64(1000000000000000)
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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	}

	// 2. TxTypeValueTransfer signed by a RoleAccountUpdate key.
	{
		amount := new(big.Int).SetUint64(1000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigSender, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 3. TxTypeAccountCreation signed by a RoleAccountUpdate key.
	{
		amount := new(big.Int).SetUint64(1000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         roleBased.Nonce,
			types.TxValueKeyFrom:          roleBased.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    anon.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigSender, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 4. TxTypeSmartContractDeploy signed by a RoleAccountUpdate key.
	{
		amount := new(big.Int).SetUint64(0)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         roleBased.Nonce,
			types.TxValueKeyFrom:          roleBased.Addr,
			types.TxValueKeyTo:            contract.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          common.FromHex(code),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigSender, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 5. TxTypeSmartContractExecution signed by a RoleAccountUpdate key.
	{
		amount := new(big.Int).SetUint64(0)
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

		err = tx.SignWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigSender, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 6. TxTypeCancel signed by a RoleAccountUpdate key.
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeCancel, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigSender, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountTransferWithRoleBasedFeePayerKey tests Txs signed by a RoleAccountUpdate key contained in a role-based key.
// A role-based key contains three types of sub-keys: RoleTransaction, RoleAccountUpdate, RoleFeePayer.
// Since a RoleFeePayer key can sign only for fee delegation Txs, following cases will fail.
// 1. create an account with a role-based key.
// 2. TxTypeValueTransfer signed by a RoleFeePayer key.
// 3. TxTypeAccountCreation signed by a RoleFeePayer key.
// 4. TxTypeSmartContractDeploy signed by a RoleFeePayer key.
// 5. TxTypeSmartContractDeploy signed by a RoleTransaction key to test smart contract execution.
// 6. TxTypeSmartContractExecution signed by a RoleFeePayer key.
// 7. TxTypeCancel signed by a RoleFeePayer key.
func TestAccountTransferWithRoleBasedFeePayerKey(t *testing.T) {
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

	gasPrice := new(big.Int).SetUint64(0)
	gasLimit := uint64(2500000000)

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

	// main account with a role-based key
	roleBased, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	// anonymous account
	anon, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78990000")
	assert.Equal(t, nil, err)

	// smart contract account
	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	// smart contract used to test TxTypeSmartContractDeploy, TxTypeSmartContractExecution transactions
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

	// generate a role-based key
	keys := genTestKeys(3)
	roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
		accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
	})

	// a RoleFeePayer key, a RoleTransaction key for signing
	feeKey := []*ecdsa.PrivateKey{keys[accountkey.RoleFeePayer]}
	txKey := []*ecdsa.PrivateKey{keys[accountkey.RoleTransaction]}

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("roleBasedAddr = ", roleBased.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. create an account with a role-based key.
	{
		// transfer value to a decoupled account for an accountUpdate Tx
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000000)
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

	// 2. TxTypeValueTransfer signed by a RoleFeePayer key.
	{
		amount := new(big.Int).SetUint64(1000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigSender, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 3. TxTypeAccountCreation signed by a RoleFeePayer key.
	{
		amount := new(big.Int).SetUint64(1000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         roleBased.Nonce,
			types.TxValueKeyFrom:          roleBased.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    anon.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigSender, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 4. TxTypeSmartContractDeploy signed by a RoleFeePayer key.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(0)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         roleBased.Nonce,
			types.TxValueKeyFrom:          roleBased.Addr,
			types.TxValueKeyTo:            contract.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          common.FromHex(code),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		err = bcdata.GenABlockWithTransactions(accountMap, txs, prof)
		assert.Equal(t, types.ErrInvalidSigSender, err)
	}

	// 5. TxTypeSmartContractDeploy signed by a RoleTransaction key to test smart contract execution.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(0)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         roleBased.Nonce,
			types.TxValueKeyFrom:          roleBased.Addr,
			types.TxValueKeyTo:            contract.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          common.FromHex(code),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}

		reservoir.Nonce += 1
	}

	// 6. TxTypeSmartContractExecution signed by a RoleFeePayer key.
	{
		amount := new(big.Int).SetUint64(0)
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

		err = tx.SignWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigSender, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 7. TxTypeCancel signed by a RoleFeePayer key.
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    roleBased.Nonce,
			types.TxValueKeyFrom:     roleBased.Addr,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeCancel, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigSender, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountFeeDelegationWithRoleBasedUpdateKey tests fee delegation with a RoleAccountUpdate key contained in a role-based key.
// A role-based key contains three types of sub-keys: RoleTransaction, RoleAccountUpdate, RoleFeePayer.
// Since a RoleAccountUpdate key cannot sign for fee delegation Txs, following cases will fail.
// 1. create a decoupled account.
// 2. create an roleBased account with a role-based key.
// 3. TxTypeFeeDelegatedValueTransfer signed by a RoleAccountUpdate key.
// 4. TxTypeFeeDelegatedValueTransferWithRatio signed by a RoleAccountUpdate key.
// 5. TxTypeFeeDelegatedValueTransferMemo signed by a RoleAccountUpdate key.
// 6. TxTypeFeeDelegatedValueTransferMemoWithRatio signed by a RoleAccountUpdate key.
// 7. TxTypeFeeDelegatedAccountUpdate signed by a RoleAccountUpdate key.
// 8. TxTypeFeeDelegatedAccountUpdateWithRatio signed by a RoleAccountUpdate key.
// 9. TxTypeFeeDelegatedSmartContractDeploy signed by a RoleAccountUpdate key.
// 10. TxTypeFeeDelegatedSmartContractDeployWithRatio signed by a RoleAccountUpdate key.
// 10-1. TxTypeSmartContractDeploy for smart contract execution tests
// 11. TxTypeFeeDelegatedSmartContractExecution signed by a RoleAccountUpdate key.
// 12. TxTypeFeeDelegatedSmartContractExecutionWithRatio signed by a RoleAccountUpdate key.
// 13. TxTypeFeeDelegatedCancel signed by a RoleAccountUpdate key.
// 14. TxTypeFeeDelegatedCancelWithRatio signed by a RoleAccountUpdate key.
func TestAccountFeeDelegationWithRoleBasedUpdateKey(t *testing.T) {
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

	gasPrice := new(big.Int).SetUint64(0)
	gasLimit := uint64(2500000000)

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

	// main account with a role-based key
	roleBased, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	// decoupled account
	decoupled, err := createDecoupledAccount("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
		common.HexToAddress("0x75c3098be5e4b63fbac05838daaee378dd48098d"))
	assert.Equal(t, nil, err)

	// smart contract account
	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	// smart contract code and abi
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

	// generate a role-based key
	keys := genTestKeys(3)
	roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
		accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
	})

	// a RoleAccountUpdate key for signing
	upKey := []*ecdsa.PrivateKey{keys[accountkey.RoleAccountUpdate]}

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("roleBasedAddr = ", roleBased.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. create a decoupled account.
	{
		// transfer value to a decoupled account for an accountUpdate Tx
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000000)
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

	// 2. create an roleBased account with a role-based key.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000000)
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

	// 3. TxTypeFeeDelegatedValueTransfer signed by a RoleAccountUpdate key.
	{
		amount := new(big.Int).SetUint64(1000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyFeePayer: roleBased.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 4. TxTypeFeeDelegatedValueTransferWithRatio signed by a RoleAccountUpdate key.
	{
		amount := new(big.Int).SetUint64(1000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyTo:                 decoupled.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 5. TxTypeFeeDelegatedValueTransferMemo signed by a RoleAccountUpdate key.
	{
		amount := new(big.Int).SetUint64(1000000000)
		data := []byte("hello")
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 6. TxTypeFeeDelegatedValueTransferMemoWithRatio signed by a RoleAccountUpdate key.
	{
		amount := new(big.Int).SetUint64(1000000000)
		data := []byte("hello")
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 decoupled.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyData:               data,
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemoWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 7. TxTypeFeeDelegatedAccountUpdate signed by a RoleAccountUpdate key.
	{
		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		if err != nil {
			t.Fatal(err)
		}

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      decoupled.Nonce,
			types.TxValueKeyFrom:       decoupled.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey),
			types.TxValueKeyFeePayer:   roleBased.Addr,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 8. TxTypeFeeDelegatedAccountUpdateWithRatio signed by a RoleAccountUpdate key.
	{
		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		if err != nil {
			t.Fatal(err)
		}

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              decoupled.Nonce,
			types.TxValueKeyFrom:               decoupled.Addr,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyAccountKey:         accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey),
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdateWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 9. TxTypeFeeDelegatedSmartContractDeploy signed by a RoleAccountUpdate key.
	{
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
			types.TxValueKeyFeePayer:      roleBased.Addr,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 10. TxTypeFeeDelegatedSmartContractDeployWithRatio signed by a RoleAccountUpdate key.
	{
		amount := new(big.Int).SetUint64(0)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 contract.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyHumanReadable:      true,
			types.TxValueKeyData:               common.FromHex(code),
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeployWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 10-1. TxTypeSmartContractDeploy for smart contract execution tests
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
	// 11. TxTypeFeeDelegatedSmartContractExecution signed by a RoleAccountUpdate key.
	{
		amount := new(big.Int).SetUint64(0)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", roleBased.Addr)
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

		err = tx.SignFeePayerWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 12. TxTypeFeeDelegatedSmartContractExecutionWithRatio signed by a RoleAccountUpdate key.
	{
		amount := new(big.Int).SetUint64(0)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", roleBased.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 contract.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyData:               data,
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractExecutionWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 13. TxTypeFeeDelegatedCancel signed by a RoleAccountUpdate key.
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

		err = tx.SignFeePayerWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 14. TxTypeFeeDelegatedCancelWithRatio signed by a RoleAccountUpdate key.
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedCancelWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, upKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountFeeDelegationWithRoleBasedTxKey tests fee delegation with a RoleTransaction key contained in a role-based key.
// A role-based key contains three types of sub-keys: RoleTransaction, RoleAccountUpdate, RoleFeePayer.
// Since a RoleTransaction key cannot sign for fee delegation Txs, following cases will fail.
// 1. create a decoupled account.
// 2. create an roleBased account with a role-based key.
// 3. TxTypeFeeDelegatedValueTransfer signed by a RoleTransaction key.
// 4. TxTypeFeeDelegatedValueTransferWithRatio signed by a RoleTransaction key.
// 5. TxTypeFeeDelegatedValueTransferMemo signed by a RoleTransaction key.
// 6. TxTypeFeeDelegatedValueTransferMemoWithRatio signed by a RoleTransaction key.
// 7. TxTypeFeeDelegatedAccountUpdate signed by a RoleTransaction key.
// 8. TxTypeFeeDelegatedAccountUpdateWithRatio signed by a RoleTransaction key.
// 9. TxTypeFeeDelegatedSmartContractDeploy signed by a RoleTransaction key.
// 10. TxTypeFeeDelegatedSmartContractDeployWithRatio signed by a RoleTransaction key.
// 10-1. TxTypeSmartContractDeploy for smart contract execution tests
// 11. TxTypeFeeDelegatedSmartContractExecution signed by a RoleTransaction key.
// 12. TxTypeFeeDelegatedSmartContractExecutionWithRatio signed by a RoleTransaction key.
// 13. TxTypeFeeDelegatedCancel signed by a RoleTransaction key.
// 14. TxTypeFeeDelegatedCancelWithRatio signed by a RoleTransaction key.
func TestAccountFeeDelegationWithRoleBasedTxKey(t *testing.T) {
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

	gasPrice := new(big.Int).SetUint64(0)
	gasLimit := uint64(2500000000)

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

	// main account with a role-based key
	roleBased, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	// decoupled account
	decoupled, err := createDecoupledAccount("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
		common.HexToAddress("0x75c3098be5e4b63fbac05838daaee378dd48098d"))
	assert.Equal(t, nil, err)

	// smart contract account
	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	// smart contract code and abi
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

	// generate a role-based key
	keys := genTestKeys(3)
	roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
		accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
	})

	// a RoleTransaction key for signing
	txKey := []*ecdsa.PrivateKey{keys[accountkey.RoleTransaction]}

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("roleBasedAddr = ", roleBased.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. create a decoupled account.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000000)
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

	// 2. create an roleBased account with a role-based key.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000000)
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

	// 3. TxTypeFeeDelegatedValueTransfer signed by a RoleTransaction key.
	{
		amount := new(big.Int).SetUint64(1000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyFeePayer: roleBased.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 4. TxTypeFeeDelegatedValueTransferWithRatio signed by a RoleTransaction key.
	{
		amount := new(big.Int).SetUint64(1000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyTo:                 decoupled.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 5. TxTypeFeeDelegatedValueTransferMemo signed by a RoleTransaction key.
	{
		amount := new(big.Int).SetUint64(1000000000)
		data := []byte("hello")
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 6. TxTypeFeeDelegatedValueTransferMemoWithRatio signed by a RoleTransaction key.
	{
		amount := new(big.Int).SetUint64(1000000000)
		data := []byte("hello")
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 decoupled.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyData:               data,
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemoWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 7. TxTypeFeeDelegatedAccountUpdate signed by a RoleTransaction key.
	{
		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		if err != nil {
			t.Fatal(err)
		}

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      decoupled.Nonce,
			types.TxValueKeyFrom:       decoupled.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey),
			types.TxValueKeyFeePayer:   roleBased.Addr,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 8. TxTypeFeeDelegatedAccountUpdateWithRatio signed by a RoleTransaction key.
	{
		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		if err != nil {
			t.Fatal(err)
		}

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              decoupled.Nonce,
			types.TxValueKeyFrom:               decoupled.Addr,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyAccountKey:         accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey),
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdateWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 9. TxTypeFeeDelegatedSmartContractDeploy signed by a RoleTransaction key.
	{
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
			types.TxValueKeyFeePayer:      roleBased.Addr,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 10. TxTypeFeeDelegatedSmartContractDeployWithRatio signed by a RoleTransaction key.
	{
		amount := new(big.Int).SetUint64(0)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 contract.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyHumanReadable:      true,
			types.TxValueKeyData:               common.FromHex(code),
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeployWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 10-1. TxTypeSmartContractDeploy for smart contract execution tests
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
	// 11. TxTypeFeeDelegatedSmartContractExecution signed by a RoleTransaction key.
	{
		amount := new(big.Int).SetUint64(0)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", roleBased.Addr)
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

		err = tx.SignFeePayerWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 12. TxTypeFeeDelegatedSmartContractExecutionWithRatio signed by a RoleTransaction key.
	{
		amount := new(big.Int).SetUint64(0)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", roleBased.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 contract.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyData:               data,
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractExecutionWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 13. TxTypeFeeDelegatedCancel signed by a RoleTransaction key.
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

		err = tx.SignFeePayerWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 14. TxTypeFeeDelegatedCancelWithRatio signed by a RoleTransaction key.
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedCancelWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, txKey)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountFeeDelegationWithRoleBasedFeePayerKey tests fee delegation with a RoleFeePayer key contained in a role-based key.
// A role-based key contains three types of sub-keys: RoleTransaction, RoleAccountUpdate, RoleFeePayer.
// Since a RoleFeePayer key can sign for fee delegation Txs, all following cases will success.
// 1. create a decoupled account.
// 2. create an roleBased account with a role-based key.
// 3. TxTypeFeeDelegatedValueTransfer signed by a RoleFeePayer key.
// 4. TxTypeFeeDelegatedValueTransferWithRatio signed by a RoleFeePayer key.
// 5. TxTypeFeeDelegatedValueTransferMemo signed by a RoleFeePayer key.
// 6. TxTypeFeeDelegatedValueTransferMemoWithRatio signed by a RoleFeePayer key.
// 7. TxTypeFeeDelegatedAccountUpdate signed by a RoleFeePayer key.
// 8. TxTypeFeeDelegatedAccountUpdateWithRatio signed by a RoleFeePayer key.
// 9. TxTypeFeeDelegatedSmartContractDeploy signed by a RoleFeePayer key.
// 10. TxTypeFeeDelegatedSmartContractDeployWithRatio signed by a RoleFeePayer key.
// 11. TxTypeFeeDelegatedSmartContractExecution signed by a RoleFeePayer key.
// 12. TxTypeFeeDelegatedSmartContractExecutionWithRatio signed by a RoleFeePayer key.
// 13. TxTypeFeeDelegatedCancel signed by a RoleFeePayer key.
// 14. TxTypeFeeDelegatedCancelWithRatio signed by a RoleFeePayer key.
func TestAccountFeeDelegationWithRoleBasedFeePayerKey(t *testing.T) {
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

	gasPrice := new(big.Int).SetUint64(0)
	gasLimit := uint64(2500000000)

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

	// main account with a role-based key
	roleBased, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	// decoupled account
	decoupled, err := createDecoupledAccount("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
		common.HexToAddress("0x75c3098be5e4b63fbac05838daaee378dd48098d"))
	assert.Equal(t, nil, err)

	// two smart contract accounts
	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	contract2, err := createHumanReadableAccount("6080604052600436106100615763ffffffff7c0100000000000000000000f7c0",
		"contract2")
	assert.Equal(t, nil, err)

	// smart contract code and abi
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

	// generate a role-based key
	keys := genTestKeys(3)
	roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
		accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
	})

	// a RoleFeePayer key for signing
	feeKey := []*ecdsa.PrivateKey{keys[accountkey.RoleFeePayer]}

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("roleBasedAddr = ", roleBased.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. create a decoupled account.
	{
		// transfer value to a decoupled account for an accountUpdate Tx
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000000)
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

	// 2. create an roleBased account with a role-based key.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000000)
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

	// 3. TxTypeFeeDelegatedValueTransfer signed by a RoleTransaction key.
	{
		var txs types.Transactions
		amount := new(big.Int).SetUint64(1000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyFeePayer: roleBased.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}

		reservoir.Nonce += 1
	}

	// 4. TxTypeFeeDelegatedValueTransferWithRatio signed by a RoleFeePayer key.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyTo:                 decoupled.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 5. TxTypeFeeDelegatedValueTransferMemo signed by a RoleFeePayer key.
	{
		var txs types.Transactions
		amount := new(big.Int).SetUint64(1000000000)
		data := []byte("hello")
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: roleBased.Addr,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}

		reservoir.Nonce += 1
	}

	// 6. TxTypeFeeDelegatedValueTransferMemoWithRatio signed by a RoleFeePayer key.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000)
		data := []byte("hello")
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 decoupled.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyData:               data,
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemoWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 7. TxTypeFeeDelegatedAccountUpdate signed by a RoleFeePayer key.
	{
		var txs types.Transactions
		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		if err != nil {
			t.Fatal(err)
		}

		decoupled.AccKey = accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      decoupled.Nonce,
			types.TxValueKeyFrom:       decoupled.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: decoupled.AccKey,
			types.TxValueKeyFeePayer:   roleBased.Addr,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}

		decoupled.Nonce += 1
		decoupled.Keys = []*ecdsa.PrivateKey{newKey}
	}

	// 8. TxTypeFeeDelegatedAccountUpdateWithRatio signed by a RoleFeePayer key.
	{
		var txs types.Transactions

		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		if err != nil {
			t.Fatal(err)
		}

		decoupled.AccKey = accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              decoupled.Nonce,
			types.TxValueKeyFrom:               decoupled.Addr,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyAccountKey:         decoupled.AccKey,
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdateWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		decoupled.Nonce += 1
		decoupled.Keys = []*ecdsa.PrivateKey{newKey}
	}

	// 9. TxTypeFeeDelegatedSmartContractDeploy signed by a RoleFeePayer key.
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
			types.TxValueKeyFeePayer:      roleBased.Addr,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 10. TxTypeFeeDelegatedSmartContractDeployWithRatio signed by a RoleFeePayer key.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(0)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 contract2.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyHumanReadable:      true,
			types.TxValueKeyData:               common.FromHex(code),
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeployWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 11. TxTypeFeeDelegatedSmartContractExecution signed by a RoleFeePayer key.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(0)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", roleBased.Addr)
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

		err = tx.SignFeePayerWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}

		reservoir.Nonce += 1
	}

	// 12. TxTypeFeeDelegatedSmartContractExecutionWithRatio signed by a RoleFeePayer key.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(0)
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", roleBased.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 contract.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyData:               data,
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractExecutionWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 13. TxTypeFeeDelegatedCancel signed by a RoleFeePayer key.
	{
		var txs types.Transactions

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

		err = tx.SignFeePayerWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}

		reservoir.Nonce += 1
	}

	// 14. TxTypeFeeDelegatedCancelWithRatio signed by a RoleFeePayer key.
	{
		var txs types.Transactions

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyFeePayer:           roleBased.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedCancelWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, feeKey)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}
