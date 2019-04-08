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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ground-x/klaytn/accounts/abi"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/state"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/blockchain/vm"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/compiler"
	"github.com/ground-x/klaytn/common/profile"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/crypto/sha3"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/ground-x/klaytn/storage/database"
	"github.com/stretchr/testify/assert"
	"math/big"
	"math/rand"
	"strings"
	"testing"
	"time"
)

var (
	to       = common.HexToAddress("7b65B75d204aBed71587c9E519a89277766EE1d0")
	feePayer = common.HexToAddress("5A0043070275d9f6054307Ee7348bD660849D90f")
	nonce    = uint64(1234)
	gasLimit = uint64(1000000)
)

type TestAccountType struct {
	Addr   common.Address
	Keys   []*ecdsa.PrivateKey
	Nonce  uint64
	AccKey accountkey.AccountKey
}

type TestCreateMultisigAccountParam struct {
	Threshold uint
	Weights   []uint
	PrvKeys   []string
}

func genRandomHash() (h common.Hash) {
	hasher := sha3.NewKeccak256()

	r := rand.Uint64()
	rlp.Encode(hasher, r)
	hasher.Sum(h[:0])

	return h
}

// createAnonymousAccount creates an account whose address is derived from the private key.
func createAnonymousAccount(prvKeyHex string) (*TestAccountType, error) {
	key, err := crypto.HexToECDSA(prvKeyHex)
	if err != nil {
		return nil, err
	}

	addr := crypto.PubkeyToAddress(key.PublicKey)

	return &TestAccountType{
		Addr:   addr,
		Keys:   []*ecdsa.PrivateKey{key},
		Nonce:  uint64(0),
		AccKey: accountkey.NewAccountKeyLegacy(),
	}, nil
}

// createDecoupledAccount creates an account whose address is decoupled with its private key.
func createDecoupledAccount(prvKeyHex string, addr common.Address) (*TestAccountType, error) {
	key, err := crypto.HexToECDSA(prvKeyHex)
	if err != nil {
		return nil, err
	}

	return &TestAccountType{
		Addr:   addr,
		Keys:   []*ecdsa.PrivateKey{key},
		Nonce:  uint64(0),
		AccKey: accountkey.NewAccountKeyPublicWithValue(&key.PublicKey),
	}, nil
}

// createHumanReadableAccount creates an account whose address is a human-readable address.
func createHumanReadableAccount(prvKeyHex string, humanReadableAddr string) (*TestAccountType, error) {
	key, err := crypto.HexToECDSA(prvKeyHex)
	if err != nil {
		return nil, err
	}

	addr, err := common.FromHumanReadableAddress(humanReadableAddr)
	if err != nil {
		return nil, err
	}

	return &TestAccountType{
		Addr:   addr,
		Keys:   []*ecdsa.PrivateKey{key},
		Nonce:  uint64(0),
		AccKey: accountkey.NewAccountKeyPublicWithValue(&key.PublicKey),
	}, nil
}

// createMultisigAccount creates an account having multiple of keys.
func createMultisigAccount(threshold uint, weights []uint, prvKeys []string, addr common.Address) (*TestAccountType, error) {
	var err error

	keys := make([]*ecdsa.PrivateKey, len(prvKeys))
	weightedKeys := make(accountkey.WeightedPublicKeys, len(prvKeys))

	for i, p := range prvKeys {
		keys[i], err = crypto.HexToECDSA(p)
		if err != nil {
			return nil, err
		}

		weightedKeys[i] = accountkey.NewWeightedPublicKey(weights[i], (*accountkey.PublicKeySerializable)(&keys[i].PublicKey))
	}

	return &TestAccountType{
		Addr:   addr,
		Keys:   keys,
		Nonce:  uint64(0),
		AccKey: accountkey.NewAccountKeyWeightedMultiSigWithValues(threshold, weightedKeys),
	}, nil
}

// createRoleBasedAccountWithAccountKeyPublic creates an account having keys that have role with AccountKeyPublic.
func createRoleBasedAccountWithAccountKeyPublic(prvKeys []string, addr common.Address) (*TestRoleBasedAccountType, error) {
	var err error

	if len(prvKeys) != 3 {
		return nil, errors.New("Need three key value for create role-based account")
	}

	keys := make([]*ecdsa.PrivateKey, len(prvKeys))

	for i, p := range prvKeys {
		keys[i], err = crypto.HexToECDSA(p)
		if err != nil {
			return nil, err
		}
	}

	accKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
		accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
	})

	return &TestRoleBasedAccountType{
		Addr:       addr,
		TxKeys:     []*ecdsa.PrivateKey{keys[0]},
		UpdateKeys: []*ecdsa.PrivateKey{keys[1]},
		FeeKeys:    []*ecdsa.PrivateKey{keys[2]},
		Nonce:      uint64(0),
		AccKey:     accKey,
	}, nil
}

// createRoleBasedAccountWithAccountKeyWeightedMultisig creates an account having keys that have role with AccountKeyWeightedMultisig.
func createRoleBasedAccountWithAccountKeyWeightedMultiSig(multisigs []TestCreateMultisigAccountParam, addr common.Address) (*TestRoleBasedAccountType, error) {
	var err error

	if len(multisigs) != 3 {
		return nil, errors.New("Need three key value for create role-based account")
	}

	prvKeys := make([][]*ecdsa.PrivateKey, len(multisigs))
	multisigKeys := make([]*accountkey.AccountKeyWeightedMultiSig, len(multisigs))

	for idx, multisig := range multisigs {
		keys := make([]*ecdsa.PrivateKey, len(multisig.PrvKeys))
		weightedKeys := make(accountkey.WeightedPublicKeys, len(multisig.PrvKeys))

		for i, p := range multisig.PrvKeys {
			keys[i], err = crypto.HexToECDSA(p)
			if err != nil {
				return nil, err
			}
			weightedKeys[i] = accountkey.NewWeightedPublicKey(multisig.Weights[i], (*accountkey.PublicKeySerializable)(&keys[i].PublicKey))
		}
		prvKeys[idx] = keys
		multisigKeys[idx] = accountkey.NewAccountKeyWeightedMultiSigWithValues(multisig.Threshold, weightedKeys)
	}

	accKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{multisigKeys[0], multisigKeys[1], multisigKeys[2]})

	return &TestRoleBasedAccountType{
		Addr:       addr,
		TxKeys:     prvKeys[0],
		UpdateKeys: prvKeys[1],
		FeeKeys:    prvKeys[2],
		Nonce:      uint64(0),
		AccKey:     accKey,
	}, nil
}

// TestAccountUpdatedWithExistingKey creates two different accounts which have the same PubKey.
// A user can sign two different accounts with a private key.
// Step 1. Create an EOA account
// Step 2. Create a decoupled EOA account
// Step 3. Update a pubKey of the decoupled account to the same key with the eoa account
// Step 4. Sign value transfer transactions of two accounts with the same key
// Expected result: PASS
func TestAccountUpdatedWithExistingKey(t *testing.T) {
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

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	var txs types.Transactions
	amount := new(big.Int).Mul(big.NewInt(100), new(big.Int).SetUint64(params.KLAY))

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// set an eoa account and an decoupled account
	prvKeyHex := "98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab"
	key, err := crypto.HexToECDSA(prvKeyHex)
	assert.Equal(t, nil, err)
	eoaAddr := crypto.PubkeyToAddress(key.PublicKey)
	eoa, err := createDecoupledAccount(prvKeyHex, eoaAddr)
	assert.Equal(t, nil, err)

	decoupled, err := createDecoupledAccount("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
		common.HexToAddress("0x75c3098be5e4b63fbac05838daaee378dd48098d"))
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoir Addr = ", reservoir.Addr.String())
		fmt.Println("EOA decoupled Addr = ", decoupled.Addr.String())
		fmt.Println("EOA Addr = ", eoa.Addr.String())
	}

	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// Step 1. Create an EOA account
	values := map[types.TxValueKeyType]interface{}{
		types.TxValueKeyNonce:         reservoir.Nonce,
		types.TxValueKeyFrom:          reservoir.Addr,
		types.TxValueKeyTo:            eoa.Addr,
		types.TxValueKeyAmount:        amount,
		types.TxValueKeyGasLimit:      gasLimit,
		types.TxValueKeyGasPrice:      gasPrice,
		types.TxValueKeyHumanReadable: false,
		types.TxValueKeyAccountKey:    eoa.AccKey,
	}
	tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
	assert.Equal(t, nil, err)
	err = tx.SignWithKeys(signer, reservoir.Keys)
	assert.Equal(t, nil, err)
	txs = append(txs, tx)
	reservoir.Nonce += 1

	// Step 2. Create a decoupled EOA account
	values = map[types.TxValueKeyType]interface{}{
		types.TxValueKeyNonce:         reservoir.Nonce,
		types.TxValueKeyFrom:          reservoir.Addr,
		types.TxValueKeyTo:            decoupled.Addr,
		types.TxValueKeyAmount:        amount,
		types.TxValueKeyGasLimit:      gasLimit,
		types.TxValueKeyGasPrice:      gasPrice,
		types.TxValueKeyHumanReadable: false,
		types.TxValueKeyAccountKey:    decoupled.AccKey,
	}
	tx, err = types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
	assert.Equal(t, nil, err)
	err = tx.SignWithKeys(signer, reservoir.Keys)
	assert.Equal(t, nil, err)
	txs = append(txs, tx)
	reservoir.Nonce += 1

	if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
		t.Fatal(err)
	}

	// Step 3. Update a pubKey of the decoupled account to the same key with the eoa account
	txs = txs[:0]
	values = map[types.TxValueKeyType]interface{}{
		types.TxValueKeyNonce:      decoupled.Nonce,
		types.TxValueKeyFrom:       decoupled.Addr,
		types.TxValueKeyGasLimit:   gasLimit,
		types.TxValueKeyGasPrice:   gasPrice,
		types.TxValueKeyAccountKey: eoa.AccKey,
	}
	tx, err = types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
	assert.Equal(t, nil, err)
	err = tx.SignWithKeys(signer, decoupled.Keys)
	assert.Equal(t, nil, err)
	txs = append(txs, tx)
	decoupled.Nonce += 1
	decoupled.Keys = eoa.Keys
	decoupled.AccKey = eoa.AccKey

	if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
		t.Fatal(err)
	}

	// Step 4. Deployed transactions signed by two different accounts which have the same key
	txs = txs[:0]
	values = map[types.TxValueKeyType]interface{}{
		types.TxValueKeyNonce:    eoa.Nonce,
		types.TxValueKeyFrom:     eoa.Addr,
		types.TxValueKeyTo:       reservoir.Addr,
		types.TxValueKeyAmount:   big.NewInt(100000), // smaller than total amount
		types.TxValueKeyGasLimit: gasLimit,
		types.TxValueKeyGasPrice: gasPrice,
	}
	tx, err = types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
	assert.Equal(t, nil, err)
	err = tx.SignWithKeys(signer, eoa.Keys)
	assert.Equal(t, nil, err)
	txs = append(txs, tx)
	eoa.Nonce += 1

	values = map[types.TxValueKeyType]interface{}{
		types.TxValueKeyNonce:    decoupled.Nonce,
		types.TxValueKeyFrom:     decoupled.Addr,
		types.TxValueKeyTo:       reservoir.Addr,
		types.TxValueKeyAmount:   big.NewInt(100000), // smaller than total amount
		types.TxValueKeyGasLimit: gasLimit,
		types.TxValueKeyGasPrice: gasPrice,
	}
	tx, err = types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
	assert.Equal(t, nil, err)
	err = tx.SignWithKeys(signer, eoa.Keys) // eoa.Keys == decoupled.Keys
	assert.Equal(t, nil, err)
	txs = append(txs, tx)
	decoupled.Nonce += 1

	if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
		t.Fatal(err)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// Below code failed since it is trying to create an account with a human-readable address "1colin".
// "1colin" is an invalid human-readable address because it starts with a number "1".
func TestHumanReadableAddress(t *testing.T) {
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
	var colinAddr common.Address
	colinAddr.SetBytesFromFront([]byte("1colin"))

	colin, err := createDecoupledAccount("ed580f5bd71a2ee4dae5cb43e331b7d0318596e561e6add7844271ed94156b20", colinAddr)
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("colinAddr = ", colin.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// Create an account "1colin" using TxTypeAccountCreation.
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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrNotHumanReadableAddress, receipt.Status)

		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationWithNilKey is for account creation with a Nil Key.
// An account should not be created with a key type of AccountKeyNil.
func TestAccountCreationWithNilKey(t *testing.T) {
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

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// Create account with Nil key
	{
		// an account has a Nil key
		addr, err := common.FromHumanReadableAddress("addrNilKey")
		assert.Equal(t, nil, err)
		prvKeyHex := "c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c"
		key, err := crypto.HexToECDSA(prvKeyHex)
		assert.Equal(t, nil, err)

		anon, err := &TestAccountType{
			Addr:   addr,
			Keys:   []*ecdsa.PrivateKey{key},
			Nonce:  uint64(0),
			AccKey: accountkey.NewAccountKeyNil(),
		}, nil
		assert.Equal(t, nil, err)

		if testing.Verbose() {
			fmt.Println("reservoirAddr = ", reservoir.Addr.String())
			fmt.Println("anonAddr = ", anon.Addr.String())
		}

		var txs types.Transactions
		amount := new(big.Int).SetUint64(100000000000)
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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrAccountKeyNilUninitializable, receipt.Status)

		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestFeeDelegatedWithSmallBalance tests the case that an account having a small amount of tokens transfers
// all the tokens to another account with a fee payer.
// This kinds of transactions were discarded in TxPool.promoteExecutable() because the total cost of
// the transaction is larger than the amount of tokens the sender has.
// Since we provide fee-delegated transactions, it is not true in the above case.
// This test code should succeed.
func TestFeeDelegatedWithSmallBalance(t *testing.T) {
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

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := big.NewInt(25 * params.Ston)

	// 1. Transfer (reservoir -> anon) using a legacy transaction.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000)
		tx := types.NewTransaction(reservoir.Nonce,
			anon.Addr, amount, gasLimit, gasPrice, []byte{})

		err := tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 2. Transfer (anon -> reservoir) using a TxTypeFeeDelegatedValueTransfer
	{
		amount := new(big.Int).SetUint64(10000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    anon.Nonce,
			types.TxValueKeyFrom:     anon.Addr,
			types.TxValueKeyFeePayer: reservoir.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		p := makeTxPool(bcdata, 10)

		p.AddRemote(tx)

		if err := bcdata.GenABlockWithTxpool(accountMap, p, prof); err != nil {
			t.Fatal(err)
		}
		anon.Nonce += 1
	}

	state, err := bcdata.bc.State()
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(0), state.GetBalance(anon.Addr).Uint64())
}

// TestTransactionScenario tests a following scenario:
// 1. Transfer (reservoir -> anon) using a legacy transaction.
// 2. Create an account decoupled using TxTypeAccountCreation.
// 3. Transfer (reservoir -> decoupled) using TxTypeValueTransfer.
// 4. Transfer (decoupled -> reservoir) using TxTypeValueTransfer.
// 5. Create an account colin using TxTypeAccountCreation.
// 6. Transfer (colin-> reservoir) using TxTypeValueTransfer.
// 7. ChainDataAnchoring (reservoir -> reservoir) using TxTypeChainDataAnchoring.
// 8. Transfer (colin-> reservoir) using TxTypeFeeDelegatedValueTransfer with a fee payer (reservoir).
// 9. Transfer (colin-> reservoir) using TxTypeFeeDelegatedValueTransferWithRatio with a fee payer (reservoir) and a ratio of 30.
// 10. Transfer (reservoir -> decoupled) using TxTypeValueTransferMemo.
// 11. Transfer (reservoir -> decoupled) using TxTypeFeeDelegatedValueTransferMemo.
// 12. Transfer (reservoir -> decoupled) using TxTypeFeeDelegatedValueTransferMemoWithRatio.
func TestTransactionScenario(t *testing.T) {
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

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
		fmt.Println("decoupledAddr = ", decoupled.Addr.String())
		fmt.Println("colinAddr = ", colin.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)
	amount := new(big.Int).Mul(big.NewInt(100), new(big.Int).SetUint64(params.KLAY))

	// 1. Transfer (reservoir -> anon) using a legacy transaction.
	{
		var txs types.Transactions

		tx := types.NewTransaction(reservoir.Nonce,
			anon.Addr, amount, gasLimit, gasPrice, []byte{})

		err := tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 2. Create an account decoupled using TxTypeAccountCreation.
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

	// Create the same account decoupled. This should be failed.
	{
		amount := new(big.Int).SetUint64(100000000)
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

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrAddressAlreadyExists, receipt.Status)
	}

	// 3. Transfer (reservoir -> decoupled) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(params.KLAY)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
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

	// 4. Transfer (decoupled -> reservoir) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    decoupled.Nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		decoupled.Nonce += 1
	}

	// 5. Create an account colin using TxTypeAccountCreation.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(params.KLAY)
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

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 6. Transfer (colin-> reservoir) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    colin.Nonce,
			types.TxValueKeyFrom:     colin.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, colin.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1
	}

	// 7. ChainDataAnchoring (reservoir -> reservoir) using TxTypeChainDataAnchoring.
	{
		scData := types.NewChainHashes(bcdata.bc.CurrentBlock())
		dataAnchoredRLP, _ := rlp.EncodeToBytes(scData)

		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:        reservoir.Nonce,
			types.TxValueKeyFrom:         reservoir.Addr,
			types.TxValueKeyTo:           reservoir.Addr,
			types.TxValueKeyAmount:       amount,
			types.TxValueKeyGasLimit:     gasLimit,
			types.TxValueKeyGasPrice:     gasPrice,
			types.TxValueKeyAnchoredData: dataAnchoredRLP,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeChainDataAnchoring, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 8. Transfer (colin-> reservoir) using TxTypeFeeDelegatedValueTransfer with a fee payer (reservoir).
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    colin.Nonce,
			types.TxValueKeyFrom:     colin.Addr,
			types.TxValueKeyFeePayer: reservoir.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, colin.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1
	}

	// 9. Transfer (colin-> reservoir) using TxTypeFeeDelegatedValueTransferWithRatio with a fee payer (reservoir) and a ratio of 30.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              colin.Nonce,
			types.TxValueKeyFrom:               colin.Addr,
			types.TxValueKeyFeePayer:           reservoir.Addr,
			types.TxValueKeyTo:                 reservoir.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, colin.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1
	}

	// 10. Transfer (reservoir -> decoupled) using TxTypeValueTransferMemo.
	{
		var txs types.Transactions
		data := []byte("hello")

		amount := new(big.Int).SetUint64(10000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1

		blkTxs := bcdata.bc.CurrentBlock().Transactions()
		assert.Equal(t, 1, blkTxs.Len())
		assert.Equal(t, types.TxTypeValueTransferMemo, blkTxs[0].Type())
		assert.Equal(t, data, blkTxs[0].Data())
	}

	// 11. Transfer (reservoir -> decoupled) using TxTypeFeeDelegatedValueTransferMemo.
	{
		var txs types.Transactions
		data := []byte("hello")

		amount := new(big.Int).SetUint64(10000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
			types.TxValueKeyFeePayer: colin.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemo, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, colin.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1

		blkTxs := bcdata.bc.CurrentBlock().Transactions()
		assert.Equal(t, 1, blkTxs.Len())
		assert.Equal(t, types.TxTypeFeeDelegatedValueTransferMemo, blkTxs[0].Type())
		assert.Equal(t, data, blkTxs[0].Data())
	}

	// 12. Transfer (reservoir -> decoupled) using TxTypeFeeDelegatedValueTransferMemoWithRatio.
	{
		var txs types.Transactions
		data := []byte("hello")

		amount := new(big.Int).SetUint64(10000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 decoupled.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyData:               data,
			types.TxValueKeyFeePayer:           colin.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferMemoWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, colin.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1

		blkTxs := bcdata.bc.CurrentBlock().Transactions()
		assert.Equal(t, 1, blkTxs.Len())
		assert.Equal(t, types.TxTypeFeeDelegatedValueTransferMemoWithRatio, blkTxs[0].Type())
		assert.Equal(t, data, blkTxs[0].Data())
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestSmartContractDeployNonHumanReadableAddress checks that the smart contract is deployed to the given address.
// Since the address is an invalid human-readable address and humanReadable == false, it should succeed.
func TestSmartContractDeployNonHumanReadableAddressSuccess(t *testing.T) {
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

	// assign invalid human readable address
	contract.Addr.SetBytesFromFront([]byte("1contract"))

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
		amount := new(big.Int).SetUint64(0)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            contract.Addr,
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

		// check receipt
		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	}
}

// TestSmartContractDeployNonHumanReadableAddress checks that the smart contract is deployed to the given address.
// Since the address is an invalid human-readable address, the transaction should fail.
func TestSmartContractDeployNonHumanReadableAddressFail(t *testing.T) {
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

	// assign invalid human readable address
	contract.Addr.SetBytesFromFront([]byte("1contract"))

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

		// check receipt
		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrNotHumanReadableAddress, receipt.Status)
	}
}

// TestSmartContractDeployAddress checks that the smart contract is deployed to the given address or not by
// checking receipt.ContractAddress.
func TestSmartContractDeployAddress(t *testing.T) {
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

		// check receipt
		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, contract.Addr, receipt.ContractAddress)
	}
}

// TestSmartContractScenario tests the following scenario:
// 1. Deploy smart contract (reservoir -> contract)
// 2. Check the the smart contract is deployed well.
// 3. Execute "reward" function with amountToSend
// 4. Validate "reward" function is executed correctly by executing "balanceOf".
func TestSmartContractScenario(t *testing.T) {
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

	// 3. Execute "reward" function with amountToSend
	amountToSend := new(big.Int).SetUint64(10)
	{
		var txs types.Transactions

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

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 4. Validate "reward" function is executed correctly by executing "balanceOf".
	{
		amount := new(big.Int).SetUint64(0)

		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("balanceOf", reservoir.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       contract.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		ret, err := callContract(bcdata, tx)
		assert.Equal(t, nil, err)

		balance := new(big.Int)
		abii.Unpack(&balance, "balanceOf", ret)

		assert.Equal(t, amountToSend, balance)
		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestSmartContractSign tests value transfer and fee delegation of smart contract accounts.
// It performs the following scenario:
// 1. Deploy smart contract (reservoir -> contract)
// 2. Check the the smart contract is deployed well.
// 3. Try value transfer. It should be failed.
// 4. Try fee delegation. It should be failed.
func TestSmartContractSign(t *testing.T) {
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

	reservoir2 := &TestAccountType{
		Addr:  *bcdata.addrs[1],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[1]},
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

	// 3. Try value transfer. It should be failed.
	{
		amount := new(big.Int).SetUint64(100000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    contract.Nonce,
			types.TxValueKeyFrom:     contract.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, contract.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
		assert.Equal(t, types.ErrInvalidSigSender, err)
	}

	// 4. Try fee delegation. It should be failed.
	{
		amount := new(big.Int).SetUint64(1000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       reservoir2.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyFeePayer: contract.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, contract.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
		assert.Equal(t, types.ErrInvalidSigFeePayer, err)
	}
}

// TestFeeDelegatedSmartContractScenario tests the following scenario:
// 1. Deploy smart contract (reservoir -> contract) with fee-delegation.
// 2. Check the the smart contract is deployed well.
// 3. Execute "reward" function with amountToSend with fee-delegation.
// 4. Validate "reward" function is executed correctly by executing "balanceOf".
func TestFeeDelegatedSmartContractScenario(t *testing.T) {
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

	reservoir2 := &TestAccountType{
		Addr:  *bcdata.addrs[1],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[1]},
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

	// 1. Deploy smart contract (reservoir -> contract) with fee-delegation.
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
			types.TxValueKeyFeePayer:      reservoir2.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, reservoir2.Keys)
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

	// 3. Execute "reward" function with amountToSend with fee-delegation.
	amountToSend := new(big.Int).SetUint64(10)
	{
		var txs types.Transactions

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
			types.TxValueKeyFeePayer: reservoir2.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, reservoir2.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 4. Validate "reward" function is executed correctly by executing "balanceOf".
	{
		amount := new(big.Int).SetUint64(0)

		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("balanceOf", reservoir.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       contract.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		ret, err := callContract(bcdata, tx)
		assert.Equal(t, nil, err)

		balance := new(big.Int)
		abii.Unpack(&balance, "balanceOf", ret)

		assert.Equal(t, amountToSend, balance)
		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestFeeDelegatedSmartContractScenarioWithRatio tests the following scenario:
// 1. Deploy smart contract (reservoir -> contract) with fee-delegation.
// 2. Check the the smart contract is deployed well.
// 3. Execute "reward" function with amountToSend with fee-delegation.
// 4. Validate "reward" function is executed correctly by executing "balanceOf".
func TestFeeDelegatedSmartContractScenarioWithRatio(t *testing.T) {
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

	reservoir2 := &TestAccountType{
		Addr:  *bcdata.addrs[1],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[1]},
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

	// 1. Deploy smart contract (reservoir -> contract) with fee-delegation.
	{
		var txs types.Transactions

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
			types.TxValueKeyFeePayer:           reservoir2.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeployWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, reservoir2.Keys)
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

	// 3. Execute "reward" function with amountToSend with fee-delegation.
	amountToSend := new(big.Int).SetUint64(10)
	{
		var txs types.Transactions

		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", reservoir.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 contract.Addr,
			types.TxValueKeyAmount:             amountToSend,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyData:               data,
			types.TxValueKeyFeePayer:           reservoir2.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractExecutionWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, reservoir2.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 4. Validate "reward" function is executed correctly by executing "balanceOf".
	{
		amount := new(big.Int).SetUint64(0)

		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("balanceOf", reservoir.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       contract.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyData:     data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		ret, err := callContract(bcdata, tx)
		assert.Equal(t, nil, err)

		balance := new(big.Int)
		abii.Unpack(&balance, "balanceOf", ret)

		assert.Equal(t, amountToSend, balance)
		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationWithFailKey creates an account with an AccountKeyFail key.
// AccountKeyFail type is for smart contract accounts, so all txs signed by the account should be failed.
// Expected result: PASS for account creation
//                  FAIL for value transfer (commented out now)
func TestAccountCreationWithFailKey(t *testing.T) {
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

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	prvKeyHex := "c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c"
	key, err := crypto.HexToECDSA(prvKeyHex)
	assert.Equal(t, nil, err)
	addr := crypto.PubkeyToAddress(key.PublicKey)

	// anon has an AccountKeyFail key
	anon, err := &TestAccountType{
		Addr:   addr,
		Keys:   []*ecdsa.PrivateKey{key},
		Nonce:  uint64(0),
		AccKey: accountkey.NewAccountKeyFail(),
	}, nil
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("anon.AccKey = ", anon.AccKey)
		fmt.Println("anonAddr = ", anon.Addr.String())
	}

	// Create anon with a fail-type key
	{
		var txs types.Transactions
		amount := new(big.Int).SetUint64(200000000000)
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
		reservoir.Nonce += 1

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
	}

	// Transfer (anon -> reservoir) should be failed
	{
		amount := new(big.Int).SetUint64(100000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    anon.Nonce,
			types.TxValueKeyFrom:     anon.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		_, _, err = applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigSender, err)
	}

	// Updating from AccountKeyFail to RoleBasedKey should fail.
	{
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      anon.Nonce,
			types.TxValueKeyFrom:       anon.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: genAccountKeyRoleBased(),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		r, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigSender, err)
		assert.Equal(t, (*types.Receipt)(nil), r)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationWithLegacyKey creates accounts with a legacy type of key and an address derived from a PubKey
// The test creates an EOA, however the value of AccKey field is not related with a PubKey
// The PubKey of the account is regenerated from the signature like legacy accounts
// Expected result: PASS
func TestAccountCreationWithLegacyKey(t *testing.T) {
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

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// an account with Legacy Key
	prvKeyHex := "c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c"
	key, err := crypto.HexToECDSA(prvKeyHex)
	assert.Equal(t, nil, err)
	addr := crypto.PubkeyToAddress(key.PublicKey)

	anon, err := &TestAccountType{
		Addr:   addr,
		Keys:   []*ecdsa.PrivateKey{key},
		Nonce:  uint64(0),
		AccKey: accountkey.NewAccountKeyLegacy(),
	}, nil
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("anon.AccKey = ", anon.AccKey)
		fmt.Println("Addr = ", anon.Addr.String())
	}

	// Create anon with a legacy key
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
		reservoir.Nonce += 1

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
	}

	// Transfer (anon -> reservoir) to check whether a PubKey is regenerated from the signature
	{
		var txs types.Transactions
		amount := new(big.Int).SetUint64(100000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    anon.Nonce,
			types.TxValueKeyFrom:     anon.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)
		anon.Nonce += 1

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationWithLegacyKeyNReadableAddr creates accounts with a legacy type of key and a human readable address.
// The test creates an EOA, but the value of AccKey field is not related with a PubKey.
// Since there is no information related with a PubKey in Tx, all Txs created by the account should not be validated.
func TestAccountCreationWithLegacyKeyNReadableAddr(t *testing.T) {
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

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// an account with Legacy Key
	prvKeyHex := "c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c"
	key, err := crypto.HexToECDSA(prvKeyHex)
	assert.Equal(t, nil, err)

	// Set human readable address
	addr, err := common.FromHumanReadableAddress("addrLegacyKey")
	assert.Equal(t, nil, err)

	anon, err := &TestAccountType{
		Addr:   addr,
		Keys:   []*ecdsa.PrivateKey{key},
		Nonce:  uint64(0),
		AccKey: accountkey.NewAccountKeyLegacy(),
	}, nil
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("anon.AccKey = ", anon.AccKey)
		fmt.Println("Addr = ", anon.Addr.String())
	}

	// Create anon with a legacy key
	{
		var txs types.Transactions
		amount := new(big.Int).SetUint64(300000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    anon.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)
		reservoir.Nonce += 1

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
	}

	// Transfer (anon -> reservoir) to check validity of the anon's private key
	{
		amount := new(big.Int).SetUint64(100000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    anon.Nonce,
			types.TxValueKeyFrom:     anon.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		_, _, err = applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigSender, err)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountUpdate tests a following scenario:
// 1. Transfer (reservoir -> anon) using a legacy transaction.
// 2. Key update of anon using AccountUpdate
// 3. Create an account decoupled using TxTypeAccountCreation.
// 4. Key update of decoupled using AccountUpdate
// 5. Transfer (decoupled -> reservoir) using TxTypeValueTransfer.
// 6. Create an account colin using TxTypeAccountCreation.
// 7. Key update of colin using AccountUpdate with multisig keys.
// 8. Transfer (colin-> reservoir) using TxTypeValueTransfer.
func TestAccountUpdate(t *testing.T) {
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

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
		fmt.Println("decoupledAddr = ", decoupled.Addr.String())
		fmt.Println("colinAddr = ", colin.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// 1. Transfer (reservoir -> anon) using a legacy transaction.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(params.KLAY)
		tx := types.NewTransaction(reservoir.Nonce,
			anon.Addr, amount, gasLimit, gasPrice, []byte{})

		err := tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 2. Key update of anon using AccountUpdate
	// This test should be failed because a legacy account does not have attribute `key`.
	{
		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		if err != nil {
			t.Fatal(err)
		}

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      anon.Nonce,
			types.TxValueKeyFrom:       anon.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		r, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		// TODO-Klaytn-Accounts: need to return more detailed status instead of ReceiptStatusErrDefault.
		assert.Equal(t, types.ReceiptStatusErrDefault, r.Status)

		// This should be failed.
	}

	// 3. Create an account decoupled using TxTypeAccountCreation.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(params.KLAY)
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

	// 4. Key update of decoupled using AccountUpdate
	{
		var txs types.Transactions

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
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		decoupled.Nonce += 1

		decoupled.Keys = []*ecdsa.PrivateKey{newKey}
	}

	// 5. Transfer (decoupled -> reservoir) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    decoupled.Nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		decoupled.Nonce += 1
	}

	// 6. Create an account colin using TxTypeAccountCreation.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(params.KLAY)
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

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 7. Key update of colin using AccountUpdate with multisig keys.
	{
		var txs types.Transactions

		k1, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		if err != nil {
			t.Fatal(err)
		}
		k2, err := crypto.HexToECDSA("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c")
		if err != nil {
			panic(err)
		}
		threshold := uint(2)
		keys := accountkey.WeightedPublicKeys{
			accountkey.NewWeightedPublicKey(1, (*accountkey.PublicKeySerializable)(&k1.PublicKey)),
			accountkey.NewWeightedPublicKey(1, (*accountkey.PublicKeySerializable)(&k2.PublicKey)),
		}
		newKey := accountkey.NewAccountKeyWeightedMultiSigWithValues(threshold, keys)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      colin.Nonce,
			types.TxValueKeyFrom:       colin.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: newKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, colin.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1

		colin.Keys = []*ecdsa.PrivateKey{k1, k2}
		colin.AccKey = newKey
	}

	// 8. Transfer (colin-> reservoir) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    colin.Nonce,
			types.TxValueKeyFrom:     colin.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, colin.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestFeeDelegatedAccountUpdate tests a following scenario:
// 1. Transfer (reservoir -> anon) using a legacy transaction.
// 2. Key update of anon using AccountUpdate
// 3. Create an account decoupled using TxTypeAccountCreation.
// 4. Key update of decoupled using TxTypeFeeDelegatedAccountUpdate
// 5. Transfer (decoupled -> reservoir) using TxTypeValueTransfer.
// 6. Create an account colin using TxTypeAccountCreation.
// 7. Key update of colin using TxTypeFeeDelegatedAccountUpdate
// 8. Transfer (colin-> reservoir) using TxTypeValueTransfer.
func TestFeeDelegatedAccountUpdate(t *testing.T) {
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

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
		fmt.Println("decoupledAddr = ", decoupled.Addr.String())
		fmt.Println("colinAddr = ", colin.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// 1. Transfer (reservoir -> anon) using a legacy transaction.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(100000000000)
		tx := types.NewTransaction(reservoir.Nonce,
			anon.Addr, amount, gasLimit, gasPrice, []byte{})

		err := tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 2. Key update of anon using AccountUpdate
	// This test should be failed because a legacy account does not have attribute `key`.
	{
		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		if err != nil {
			t.Fatal(err)
		}

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      anon.Nonce,
			types.TxValueKeyFrom:       anon.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey),
			types.TxValueKeyFeePayer:   reservoir.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		r, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		// TODO-Klaytn-Accounts: need to return more detailed status instead of ReceiptStatusErrDefault.
		assert.Equal(t, types.ReceiptStatusErrDefault, r.Status)

		// This should be failed.
	}

	// 3. Create an account decoupled using TxTypeAccountCreation.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(params.KLAY)
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

	// 4. Key update of decoupled using TxTypeFeeDelegatedAccountUpdate
	{
		var txs types.Transactions

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
			types.TxValueKeyFeePayer:   reservoir.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		decoupled.Nonce += 1

		decoupled.Keys = []*ecdsa.PrivateKey{newKey}
	}

	// 5. Transfer (decoupled -> reservoir) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    decoupled.Nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		decoupled.Nonce += 1
	}

	// 6. Create an account colin using TxTypeAccountCreation.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(params.KLAY)
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

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 7. Key update of colin using TxTypeFeeDelegatedAccountUpdate
	{
		var txs types.Transactions

		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		if err != nil {
			t.Fatal(err)
		}

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      colin.Nonce,
			types.TxValueKeyFrom:       colin.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey),
			types.TxValueKeyFeePayer:   reservoir.Addr,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, colin.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1

		colin.Keys = []*ecdsa.PrivateKey{newKey}
	}

	// 8. Transfer (colin-> reservoir) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    colin.Nonce,
			types.TxValueKeyFrom:     colin.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, colin.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestFeeDelegatedAccountUpdateWithRatio tests a following scenario:
// 1. Transfer (reservoir -> anon) using a legacy transaction.
// 2. Key update of anon using AccountUpdate
// 3. Create an account decoupled using TxTypeAccountCreation.
// 4. Key update of decoupled using TxTypeFeeDelegatedAccountUpdateWithRatio.
// 5. Transfer (decoupled -> reservoir) using TxTypeValueTransfer.
// 6. Create an account colin using TxTypeAccountCreation.
// 7. Key update of colin using TxTypeFeeDelegatedAccountUpdateWithRatio.
// 8. Transfer (colin-> reservoir) using TxTypeValueTransfer.
func TestFeeDelegatedAccountUpdateWithRatio(t *testing.T) {
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

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
		fmt.Println("decoupledAddr = ", decoupled.Addr.String())
		fmt.Println("colinAddr = ", colin.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// 1. Transfer (reservoir -> anon) using a legacy transaction.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(params.KLAY)
		tx := types.NewTransaction(reservoir.Nonce,
			anon.Addr, amount, gasLimit, gasPrice, []byte{})

		err := tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 2. Key update of anon using AccountUpdate
	// This test should be failed because a legacy account does not have attribute `key`.
	{
		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		if err != nil {
			t.Fatal(err)
		}

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              anon.Nonce,
			types.TxValueKeyFrom:               anon.Addr,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyAccountKey:         accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey),
			types.TxValueKeyFeePayer:           reservoir.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdateWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		r, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		// TODO-Klaytn-Accounts: need to return more detailed status instead of ReceiptStatusErrDefault.
		assert.Equal(t, types.ReceiptStatusErrDefault, r.Status)

		// This should be failed.
	}

	// 3. Create an account decoupled using TxTypeAccountCreation.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(params.KLAY)
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

	// 4. Key update of decoupled using TxTypeFeeDelegatedAccountUpdateWithRatio.
	{
		var txs types.Transactions

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
			types.TxValueKeyFeePayer:           reservoir.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdateWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		decoupled.Nonce += 1

		decoupled.Keys = []*ecdsa.PrivateKey{newKey}
	}

	// 5. Transfer (decoupled -> reservoir) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    decoupled.Nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		decoupled.Nonce += 1
	}

	// 6. Create an account colin using TxTypeAccountCreation.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(params.KLAY)
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

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 7. Key update of colin using TxTypeFeeDelegatedAccountUpdateWithRatio.
	{
		var txs types.Transactions

		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		if err != nil {
			t.Fatal(err)
		}

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              colin.Nonce,
			types.TxValueKeyFrom:               colin.Addr,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyAccountKey:         accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey),
			types.TxValueKeyFeePayer:           reservoir.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedAccountUpdateWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, colin.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1

		colin.Keys = []*ecdsa.PrivateKey{newKey}
	}

	// 8. Transfer (colin-> reservoir) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    colin.Nonce,
			types.TxValueKeyFrom:     colin.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, colin.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestMultisigScenario tests a test case for a multi-sig accounts.
// 1. Create an account multisig using TxTypeAccountCreation.
// 2. Transfer (multisig -> reservoir) using TxTypeValueTransfer.
// 3. Transfer (multisig -> reservoir) using TxTypeValueTransfer with only two keys.
// 4. FAILED-CASE: Transfer (multisig -> reservoir) using TxTypeValueTransfer with only one key.
func TestMultisigScenario(t *testing.T) {
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

	// multi-sig account
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
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// 1. Create an account multisig using TxTypeAccountCreation.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(params.KLAY)
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

	// 2. Transfer (multisig -> reservoir) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    multisig.Nonce,
			types.TxValueKeyFrom:     multisig.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, multisig.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		multisig.Nonce += 1
	}

	// 3. Transfer (multisig -> reservoir) using TxTypeValueTransfer with only two keys.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    multisig.Nonce,
			types.TxValueKeyFrom:     multisig.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, multisig.Keys[0:2])
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		multisig.Nonce += 1
	}

	// 4. FAILED-CASE: Transfer (multisig -> reservoir) using TxTypeValueTransfer with only one key.
	{
		amount := new(big.Int).SetUint64(1000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    multisig.Nonce,
			types.TxValueKeyFrom:     multisig.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, multisig.Keys[:1])
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, types.ErrInvalidSigSender, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestValidateSender tests ValidateSender with all transaction types.
func TestValidateSender(t *testing.T) {
	// anonymous account
	anon, err := createAnonymousAccount("1da6dfcb52128060cdd2108edb786ca0aff4ef1fa537574286eeabe5c2ebd5ca")
	assert.Equal(t, nil, err)

	// decoupled account
	decoupled, err := createDecoupledAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab",
		common.HexToAddress("0x5104711f7faa9e2dadf593e43db1577a2887636f"))
	assert.Equal(t, nil, err)

	initialBalance := big.NewInt(1000000)

	statedb, _ := state.New(common.Hash{}, state.NewDatabase(database.NewMemoryDBManager()))
	statedb.CreateEOA(anon.Addr, false, anon.AccKey)
	statedb.SetNonce(anon.Addr, nonce)
	statedb.SetBalance(anon.Addr, initialBalance)

	statedb.CreateEOA(decoupled.Addr, false, decoupled.AccKey)
	statedb.SetNonce(decoupled.Addr, rand.Uint64())
	statedb.SetBalance(decoupled.Addr, initialBalance)

	signer := types.MakeSigner(params.BFTTestChainConfig, big.NewInt(32))
	gasPrice := new(big.Int).SetUint64(0)
	gasLimit := uint64(2500000)
	amount := new(big.Int).SetUint64(10000)

	// LegacyTransaction
	{
		amount := new(big.Int).SetUint64(100000000000)
		tx := types.NewTransaction(anon.Nonce,
			decoupled.Addr, amount, gasLimit, gasPrice, []byte{})

		err := tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, anon.Addr, actualFrom)
	}

	// TxTypeValueTransfer
	{
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    nonce,
			types.TxValueKeyFrom:     anon.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		})
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, anon.Addr, actualFrom)
	}

	// TxTypeValueTransfer
	{
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		})
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, decoupled.Addr, actualFrom)
	}

	// TxTypeSmartContractDeploy
	{
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         nonce,
			types.TxValueKeyFrom:          decoupled.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			// The binary below is a compiled binary of contracts/reward/contract/KlaytnReward.sol.
			types.TxValueKeyData: common.Hex2Bytes("608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"),
		})
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, decoupled.Addr, actualFrom)
	}

	// TxTypeSmartContractExecution
	{
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyTo:       anon.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			// A abi-packed bytes calling "reward" of contracts/reward/contract/KlaytnReward.sol with an address "bc5951f055a85f41a3b62fd6f68ab7de76d299b2".
			types.TxValueKeyData: common.Hex2Bytes("6353586b000000000000000000000000bc5951f055a85f41a3b62fd6f68ab7de76d299b2"),
		})
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, decoupled.Addr, actualFrom)
	}

	// TxTypeChainDataAnchoring
	{
		dummyBlock := types.NewBlock(&types.Header{
			GasLimit: gasLimit,
		}, nil, nil, nil)

		scData := types.NewChainHashes(dummyBlock)
		dataAnchoredRLP, _ := rlp.EncodeToBytes(scData)

		amount := new(big.Int).SetUint64(0)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:        anon.Nonce,
			types.TxValueKeyFrom:         anon.Addr,
			types.TxValueKeyTo:           anon.Addr,
			types.TxValueKeyAmount:       amount,
			types.TxValueKeyGasLimit:     gasLimit,
			types.TxValueKeyGasPrice:     gasPrice,
			types.TxValueKeyAnchoredData: dataAnchoredRLP,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeChainDataAnchoring, values)
		if err != nil {
			t.Fatal(err)
		}
		err = tx.SignWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, anon.Addr, actualFrom)
	}

	// TxTypeFeeDelegatedValueTransfer
	{
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyFeePayer: anon.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		})
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, decoupled.Addr, actualFrom)

		actualFeePayer, _, err := types.ValidateFeePayer(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, anon.Addr, actualFeePayer)
	}

	// TxTypeFeeDelegatedValueTransferWithRatio
	{
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferWithRatio, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              nonce,
			types.TxValueKeyFrom:               decoupled.Addr,
			types.TxValueKeyFeePayer:           anon.Addr,
			types.TxValueKeyTo:                 decoupled.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
		})
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, decoupled.Keys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, anon.Keys)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, decoupled.Addr, actualFrom)

		actualFeePayer, _, err := types.ValidateFeePayer(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, anon.Addr, actualFeePayer)
	}
}

func isCompilerAvailable() bool {
	solc, err := compiler.SolidityVersion("")
	if err != nil {
		fmt.Println("Solidity version check failed. Skipping this test", err)
		return false
	}
	if solc.Version != "0.4.24" && solc.Version != "0.4.25" {
		if testing.Verbose() {
			fmt.Println("solc version mismatch. Supported versions are 0.4.24 and 0.4.25.", "version", solc.Version)
		}
		return false
	}

	return true
}

func compileSolidity(filename string) (code []string, abiStr []string) {
	contracts, err := compiler.CompileSolidity("", filename)
	if err != nil {
		panic(err)
	}

	code = make([]string, 0, len(contracts))
	abiStr = make([]string, 0, len(contracts))

	for _, c := range contracts {
		abiBytes, err := json.Marshal(c.Info.AbiDefinition)
		if err != nil {
			panic(err)
		}

		code = append(code, c.Code)
		abiStr = append(abiStr, string(abiBytes))
	}

	return
}

// applyTransaction setups variables to call block.ApplyTransaction() for tests.
// It directly returns values from block.ApplyTransaction().
func applyTransaction(t *testing.T, bcdata *BCData, tx *types.Transaction) (*types.Receipt, uint64, error) {
	state, err := bcdata.bc.State()
	assert.Equal(t, nil, err)

	vmConfig := &vm.Config{
		JumpTable: vm.ConstantinopleInstructionSet,
	}
	parent := bcdata.bc.CurrentBlock()
	num := parent.Number()
	author := bcdata.addrs[0]
	gp := new(blockchain.GasPool).AddGas(parent.GasLimit())
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     num.Add(num, common.Big1),
		GasLimit:   blockchain.CalcGasLimit(parent),
		Extra:      parent.Extra(),
		Time:       new(big.Int).Add(parent.Time(), common.Big1),
		Difficulty: big.NewInt(0),
	}
	usedGas := uint64(0)
	return blockchain.ApplyTransaction(bcdata.bc.Config(), bcdata.bc, author, gp, state, header, tx, &usedGas, vmConfig)
}

func genAccountKeyRoleBased() accountkey.AccountKey {
	k1, err := crypto.HexToECDSA("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	if err != nil {
		panic(err)
	}
	txKey := accountkey.NewAccountKeyPublicWithValue(&k1.PublicKey)

	k2, err := crypto.HexToECDSA("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c")
	if err != nil {
		panic(err)
	}
	threshold := uint(2)
	keys := accountkey.WeightedPublicKeys{
		accountkey.NewWeightedPublicKey(1, (*accountkey.PublicKeySerializable)(&k1.PublicKey)),
		accountkey.NewWeightedPublicKey(1, (*accountkey.PublicKeySerializable)(&k2.PublicKey)),
	}
	updateKey := accountkey.NewAccountKeyWeightedMultiSigWithValues(threshold, keys)

	k3, err := crypto.HexToECDSA("ed580f5bd71a2ee4dae5cb43e331b7d0318596e561e6add7844271ed94156b20")
	if err != nil {
		panic(err)
	}
	feeKey := accountkey.NewAccountKeyPublicWithValue(&k3.PublicKey)

	return accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{txKey, updateKey, feeKey})
}
