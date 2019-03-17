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
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/profile"
	"github.com/ground-x/klaytn/crypto"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

type TestRoleBasedAccountType struct {
	Addr       common.Address
	TxKeys     []*ecdsa.PrivateKey
	UpdateKeys []*ecdsa.PrivateKey
	FeeKeys    []*ecdsa.PrivateKey
	Nonce      uint64
	AccKey     accountkey.AccountKey
}

func genTestKeys(len int) []*ecdsa.PrivateKey {
	keys := make([]*ecdsa.PrivateKey, len)

	for i := 0; i < len; i++ {
		keys[i], _ = crypto.GenerateKey()
	}

	return keys
}

// TestRoleBasedAccount executes transactions to test accounts having a role-based key.
// The scenario is the following:
// 1. Create an account `colin` with a role-key.
// 2. Transfer value using colin.TxKeys.
// 3. Pay tx fee using colin.FeeKeys.
// 4. Update tx key using colin.UpdateKeys.
// 5. Transfer value using updated colin.TxKeys.
func TestRoleBasedAccount(t *testing.T) {
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
	reservoir := &TestRoleBasedAccountType{
		Addr:       *bcdata.addrs[0],
		TxKeys:     []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		UpdateKeys: []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		FeeKeys:    []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce:      uint64(0),
	}

	// Create an account having a role-based key
	colinAddr, err := common.FromHumanReadableAddress("colin")
	keys := genTestKeys(3)
	accKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
		accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
		accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
	})
	colin := &TestRoleBasedAccountType{
		Addr:       colinAddr,
		TxKeys:     []*ecdsa.PrivateKey{keys[0]},
		UpdateKeys: []*ecdsa.PrivateKey{keys[1]},
		FeeKeys:    []*ecdsa.PrivateKey{keys[2]},
		Nonce:      uint64(0),
		AccKey:     accKey,
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. Create an account `colin` with a role-key.
	{
		var txs types.Transactions

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

		err = tx.SignWithKeys(signer, reservoir.TxKeys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 2. Transfer value using colin.TxKeys.
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

		err = tx.SignWithKeys(signer, colin.TxKeys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1
	}

	// 3. Pay tx fee using colin.FeeKeys.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    colin.Nonce,
			types.TxValueKeyFrom:     colin.Addr,
			types.TxValueKeyFeePayer: colin.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, colin.TxKeys)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayerWithKeys(signer, colin.FeeKeys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1
	}

	// 4. Update tx key using colin.UpdateKeys.
	{
		var txs types.Transactions

		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		if err != nil {
			t.Fatal(err)
		}

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    colin.Nonce,
			types.TxValueKeyFrom:     colin.Addr,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyAccountKey: accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
				accountkey.NewAccountKeyPublicWithValue(&newKey.PublicKey),
				accountkey.NewAccountKeyNil(),
			}),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, colin.UpdateKeys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1

		colin.TxKeys = []*ecdsa.PrivateKey{newKey}
	}

	// 5. Transfer value using updated colin.TxKeys.
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

		err = tx.SignWithKeys(signer, colin.TxKeys)
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

// TestAccountUpdateRoleBasedNil tests TxInternalDataAccountUpdate with the following scenario:
// 1. Create an account colin using TxTypeAccountCreation.
// 2. Update key to RoleBasedKey. If anyone is AccountKeyNil, it should fail. First nil key.
// 3. Update key to RoleBasedKey. If anyone is AccountKeyNil, it should fail. Second nil key.
func TestAccountUpdateRoleBasedNil(t *testing.T) {
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

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("colinAddr = ", colin.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. Create an account colin using TxTypeAccountCreation.
	{
		var txs types.Transactions

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

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 2. Update key to RoleBasedKey. If anyone is AccountKeyNil, it should fail. First nil key.
	{
		keys := genTestKeys(3)
		accKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyNil(),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
		})

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      colin.Nonce,
			types.TxValueKeyFrom:       colin.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: accKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, colin.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrAccountKeyNilUninitializable, receipt.Status)
	}

	// 3. Update key to RoleBasedKey. If anyone is AccountKeyNil, it should fail. Second nil key.
	{
		keys := genTestKeys(3)
		accKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
			accountkey.NewAccountKeyNil(),
			accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
		})

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      colin.Nonce,
			types.TxValueKeyFrom:       colin.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: accKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, colin.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrAccountKeyNilUninitializable, receipt.Status)
	}
}

// TestAccountUpdateRoleBasedWrongLength tests if the role-based key having wrong length.
// It tests the following scenario:
// 1. Create an account colin using TxTypeAccountCreation.
// 2. Update key to RoleBasedKey with four roles. It should fail.
// 3. Update key to RoleBasedKey with zero role. It should fail.
func TestAccountUpdateRoleBasedWrongLength(t *testing.T) {
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

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("colinAddr = ", colin.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. Create an account colin using TxTypeAccountCreation.
	{
		var txs types.Transactions

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

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 2. Update key to RoleBasedKey with four roles. It should fail.
	{
		keys := genTestKeys(4)
		accKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[3].PublicKey),
		})

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      colin.Nonce,
			types.TxValueKeyFrom:       colin.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: accKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, colin.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrLengthTooLong, receipt.Status)
	}

	// 3. Update key to RoleBasedKey with zero role. It should fail.
	{
		accKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{})

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      colin.Nonce,
			types.TxValueKeyFrom:       colin.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: accKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, colin.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrZeroLength, receipt.Status)
	}
}
