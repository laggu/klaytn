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
	"github.com/ground-x/klaytn/blockchain/state"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/profile"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/crypto/sha3"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/ground-x/klaytn/storage/database"
	"github.com/stretchr/testify/assert"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

type TestAccountType struct {
	Addr  common.Address
	Key   *ecdsa.PrivateKey
	Nonce uint64
}

func genRandomHash() (h common.Hash) {
	hasher := sha3.NewKeccak256()

	r := rand.Uint64()
	rlp.Encode(hasher, r)
	hasher.Sum(h[:0])

	return h
}

// createAnonymousAccount creates an account whose address is derived from the private key.
func createAnonymousAccount() (*TestAccountType, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	addr := crypto.PubkeyToAddress(key.PublicKey)

	return &TestAccountType{
		Addr:  addr,
		Key:   key,
		Nonce: uint64(0),
	}, nil
}

// createDecoupledAccount creates an account whose address is decoupled with its private key.
func createDecoupledAccount() (*TestAccountType, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	b := genRandomHash().Bytes()[:common.AddressLength]
	addr := common.BytesToAddress(b)

	return &TestAccountType{
		Addr:  addr,
		Key:   key,
		Nonce: uint64(0),
	}, nil
}

func createHumanReadableAccount(humanReadableAddr string) (*TestAccountType, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	addr, err := common.FromHumanReadableAddress(humanReadableAddr)
	if err != nil {
		return nil, err
	}

	return &TestAccountType{
		Addr:  addr,
		Key:   key,
		Nonce: uint64(0),
	}, nil
}

// TestTransactionScenario tests a following scenario:
// 1. Transfer (reservoir -> anon) using a legacy transaction.
// 2. Create an account decoupled using TxTypeAccountCreation.
// 3. Transfer (reservoir -> decoupled) using TxTypeValueTransfer.
// 4. Transfer (decoupled -> reservoir) using TxTypeValueTransfer.
// 5. Create an account colin using TxTypeAccountCreation.
// 6. Transfer (colin-> reservoir) using TxTypeValueTransfer.
// 7. ChainDataAnchoring (reservoir -> reservoir) using TxTypeChainDataAnchoring.
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
		Key:   bcdata.privKeys[0],
		Nonce: uint64(0),
	}

	// anonymous account
	anon, err := createAnonymousAccount()
	assert.Equal(t, nil, err)

	// decoupled account
	decoupled, err := createDecoupledAccount()
	assert.Equal(t, nil, err)

	colin, err := createHumanReadableAccount("colin")
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
		fmt.Println("decoupledAddr = ", decoupled.Addr.String())
		fmt.Println("colinAddr = ", colin.Addr.String())
	}

	gasPrice := new(big.Int).SetUint64(25)
	gasLimit := uint64(2500000)

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. Transfer (reservoir -> anon) using a legacy transaction.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(100000000000)
		tx := types.NewTransaction(reservoir.Nonce,
			anon.Addr, amount, gasLimit, gasPrice, []byte{})

		err := tx.Sign(signer, reservoir.Key)
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

		amount := new(big.Int).SetUint64(100000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      reservoir.Nonce,
			types.TxValueKeyFrom:       reservoir.Addr,
			types.TxValueKeyTo:         decoupled.Addr,
			types.TxValueKeyAmount:     amount,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: types.NewAccountKeyPublicWithValue(&decoupled.Key.PublicKey),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, reservoir.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// Create the same account decoupled. This should be failed.
	// TODO-Klaytn: make a test case for this error case.
	//{
	//	var txs types.Transactions
	//
	//	amount := new(big.Int).SetUint64(100000000)
	//	values := map[types.TxValueKeyType]interface{}{
	//		types.TxValueKeyNonce:      reservoir.Nonce,
	//		types.TxValueKeyFrom:       reservoir.Addr,
	//		types.TxValueKeyTo:         decoupled.Addr,
	//		types.TxValueKeyAmount:     amount,
	//		types.TxValueKeyGasLimit:   gasLimit,
	//		types.TxValueKeyGasPrice:   gasPrice,
	//		types.TxValueKeyAccountKey: types.NewAccountKeyPublicWithValue(&decoupled.Key.PublicKey),
	//	}
	//	tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
	//	assert.Equal(t, nil, err)
	//
	//	signedTx, err := types.SignTx(tx, signer, reservoir.Key)
	//	assert.Equal(t, nil, err)
	//
	//	txs = append(txs, signedTx)
	//
	//	if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
	//		t.Fatal(err)
	//	}
	//	reservoir.Nonce += 1
	//}

	// 3. Transfer (reservoir -> decoupled) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000000)
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

		err = tx.Sign(signer, reservoir.Key)
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

		amount := new(big.Int).SetUint64(10000)
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

		err = tx.Sign(signer, decoupled.Key)
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

		amount := new(big.Int).SetUint64(100000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      reservoir.Nonce,
			types.TxValueKeyFrom:       reservoir.Addr,
			types.TxValueKeyTo:         colin.Addr,
			types.TxValueKeyAmount:     amount,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: types.NewAccountKeyPublicWithValue(&colin.Key.PublicKey),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, reservoir.Key)
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

		err = tx.Sign(signer, colin.Key)
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
		if err != nil {
			t.Fatal(err)
		}
		signedTx, err := types.SignTx(tx, signer, reservoir.Key)
		if err != nil {
			t.Fatal(err)
		}
		txs = append(txs, signedTx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestValidateSender tests ValidateSender with all transaction types.
func TestValidateSender(t *testing.T) {
	// anonymous account
	anon, err := createAnonymousAccount()
	assert.Equal(t, nil, err)

	// decoupled account
	decoupled, err := createDecoupledAccount()
	assert.Equal(t, nil, err)

	statedb, _ := state.New(common.Hash{}, state.NewDatabase(database.NewMemoryDBManager()))
	statedb.CreateAccountWithMap(anon.Addr, state.ExternallyOwnedAccountType,
		map[state.AccountValueKeyType]interface{}{
			state.AccountValueKeyNonce:         rand.Uint64(),
			state.AccountValueKeyBalance:       big.NewInt(rand.Int63n(10000)),
			state.AccountValueKeyHumanReadable: false,
			state.AccountValueKeyAccountKey:    types.NewAccountKeyNil(),
		})

	statedb.CreateAccountWithMap(decoupled.Addr, state.ExternallyOwnedAccountType,
		map[state.AccountValueKeyType]interface{}{
			state.AccountValueKeyNonce:         rand.Uint64(),
			state.AccountValueKeyBalance:       big.NewInt(rand.Int63n(10000)),
			state.AccountValueKeyHumanReadable: false,
			state.AccountValueKeyAccountKey:    types.NewAccountKeyPublicWithValue(&decoupled.Key.PublicKey),
		})

	signer := types.MakeSigner(params.BFTTestChainConfig, big.NewInt(32))
	gasPrice := new(big.Int).SetUint64(0)
	gasLimit := uint64(2500000)
	amount := new(big.Int).SetUint64(10000)

	// LegacyTransaction
	{
		amount := new(big.Int).SetUint64(100000000000)
		tx := types.NewTransaction(anon.Nonce,
			decoupled.Addr, amount, gasLimit, gasPrice, []byte{})

		err := tx.Sign(signer, anon.Key)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, anon.Addr, actualFrom)
	}

	// TxTypeValueTransfer
	{
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    0,
			types.TxValueKeyFrom:     anon.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		})
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, anon.Key)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, anon.Addr, actualFrom)
	}

	// TxTypeValueTransfer
	{
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    0,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		})
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, decoupled.Key)
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

		var txs types.Transactions

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
		signedTx, err := types.SignTx(tx, signer, anon.Key)
		if err != nil {
			t.Fatal(err)
		}
		txs = append(txs, signedTx)

		actualFrom, _, err := types.ValidateSender(signer, signedTx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, anon.Addr, actualFrom)
	}
}
