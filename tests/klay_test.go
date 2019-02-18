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
	"flag"
	"fmt"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common/profile"
	"github.com/ground-x/klaytn/crypto"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"
)

var txPerBlock int

func init() {
	flag.IntVar(&txPerBlock, "txs-per-block", 1000,
		"Specify the number of transactions per block")
}

////////////////////////////////////////////////////////////////////////////////
// TestValueTransfer
////////////////////////////////////////////////////////////////////////////////
type testOption struct {
	numTransactions    int
	numMaxAccounts     int
	numValidators      int
	numGeneratedBlocks int
	txdata             []byte
	makeTransactions   func(*BCData, *AccountMap, types.Signer, int, *big.Int, []byte) (types.Transactions, error)
}

func makeTransactionsFrom(bcdata *BCData, accountMap *AccountMap, signer types.Signer, numTransactions int,
	amount *big.Int, data []byte) (types.Transactions, error) {
	from := *bcdata.addrs[0]
	privKey := bcdata.privKeys[0]
	toAddrs := bcdata.addrs
	numAddrs := len(toAddrs)

	txs := make(types.Transactions, 0, numTransactions)
	nonce, err := accountMap.GetNonce(from)
	if err != nil {
		return nil, err
	}
	for i := 0; i < numTransactions; i++ {
		a := toAddrs[i%numAddrs]
		txamount := amount
		if txamount == nil {
			txamount = big.NewInt(rand.Int63n(10))
			txamount = txamount.Add(txamount, big.NewInt(1))
		}
		var gasLimit uint64 = 1000000
		gasPrice := new(big.Int).SetInt64(0)

		tx := types.NewTransaction(nonce, *a, txamount, gasLimit, gasPrice, data)
		signedTx, err := types.SignTx(tx, signer, privKey)
		if err != nil {
			return nil, err
		}

		txs = append(txs, signedTx)

		nonce++
	}

	return txs, nil
}

func makeIndependentTransactions(bcdata *BCData, accountMap *AccountMap, signer types.Signer, numTransactions int,
	amount *big.Int, data []byte) (types.Transactions, error) {
	numAddrs := len(bcdata.addrs) / 2
	fromAddrs := bcdata.addrs[:numAddrs]
	toAddrs := bcdata.addrs[numAddrs:]

	var err error
	fromNonces := make([]uint64, numAddrs)
	for i, addr := range fromAddrs {
		fromNonces[i], err = accountMap.GetNonce(*addr)
		if err != nil {
			return nil, err
		}
	}

	txs := make(types.Transactions, 0, numTransactions)

	for i := 0; i < numTransactions; i++ {
		idx := i % numAddrs

		txamount := amount
		if txamount == nil {
			txamount = big.NewInt(rand.Int63n(10))
			txamount = txamount.Add(txamount, big.NewInt(1))
		}
		var gasLimit uint64 = 1000000
		gasPrice := new(big.Int).SetInt64(0)

		tx := types.NewTransaction(fromNonces[idx], *toAddrs[idx], txamount, gasLimit, gasPrice, data)
		signedTx, err := types.SignTx(tx, signer, bcdata.privKeys[idx])
		if err != nil {
			return nil, err
		}

		txs = append(txs, signedTx)

		fromNonces[idx]++
	}

	return txs, nil
}

// makeTransactionsToRandom makes `numTransactions` transactions which transfers a random amount of tokens
// from accounts in `AccountMap` to a randomly generated account.
// It returns the generated transactions if successful, or it returns an error if failed.
func makeTransactionsToRandom(bcdata *BCData, accountMap *AccountMap, signer types.Signer, numTransactions int,
	amount *big.Int, data []byte) (types.Transactions, error) {
	numAddrs := len(bcdata.addrs)
	fromAddrs := bcdata.addrs

	var err error
	fromNonces := make([]uint64, numAddrs)
	for i, addr := range fromAddrs {
		fromNonces[i], err = accountMap.GetNonce(*addr)
		if err != nil {
			return nil, err
		}
	}

	txs := make(types.Transactions, 0, numTransactions)

	for i := 0; i < numTransactions; i++ {
		idx := i % numAddrs

		txamount := amount
		if txamount == nil {
			txamount = big.NewInt(rand.Int63n(10))
			txamount = txamount.Add(txamount, big.NewInt(1))
		}
		var gasLimit uint64 = 1000000
		gasPrice := new(big.Int).SetInt64(0)

		// generate a new address
		k, err := crypto.GenerateKey()
		if err != nil {
			return nil, err
		}
		to := crypto.PubkeyToAddress(k.PublicKey)

		tx := types.NewTransaction(fromNonces[idx], to, txamount, gasLimit, gasPrice, data)
		signedTx, err := types.SignTx(tx, signer, bcdata.privKeys[idx])
		if err != nil {
			return nil, err
		}

		txs = append(txs, signedTx)

		fromNonces[idx]++
	}

	return txs, nil
}

// makeTransactionsToRandom makes `numTransactions` transactions which transfers a random amount of KLAY
// from accounts in `AccountMap` to a randomly generated account.
// It returns the generated transactions if successful, or it returns an error if failed.
func makeNewTransactionsToRandom(bcdata *BCData, accountMap *AccountMap, signer types.Signer, numTransactions int,
	amount *big.Int, data []byte) (types.Transactions, error) {
	numAddrs := len(bcdata.addrs)
	fromAddrs := bcdata.addrs

	var err error
	fromNonces := make([]uint64, numAddrs)
	for i, addr := range fromAddrs {
		fromNonces[i], err = accountMap.GetNonce(*addr)
		if err != nil {
			return nil, err
		}
	}

	txs := make(types.Transactions, 0, numTransactions)

	for i := 0; i < numTransactions; i++ {
		idx := i % numAddrs

		txamount := amount
		if txamount == nil {
			txamount = big.NewInt(rand.Int63n(10))
			txamount = txamount.Add(txamount, big.NewInt(1))
		}
		var gasLimit uint64 = 1000000
		gasPrice := new(big.Int).SetInt64(0)

		// generate a new address
		k, err := crypto.GenerateKey()
		if err != nil {
			return nil, err
		}
		to := crypto.PubkeyToAddress(k.PublicKey)

		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    fromNonces[idx],
			types.TxValueKeyTo:       to,
			types.TxValueKeyAmount:   txamount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyFrom:     *bcdata.addrs[idx],
		})
		signedTx, err := types.SignTx(tx, signer, bcdata.privKeys[idx])
		if err != nil {
			return nil, err
		}

		txs = append(txs, signedTx)

		fromNonces[idx]++
	}

	return txs, nil
}

// makeNewTransactionsToRing makes `numTransactions` transactions which transfers a fixed amount of KLAY
// from account with index i to account with index (i+1). To have same amount of balance before and after the test,
// total number of transactions should be the multiple of number of addresses.
// It returns the generated transactions if successful, or it returns an error if failed.
func makeNewTransactionsToRing(bcdata *BCData, accountMap *AccountMap, signer types.Signer, numTransactions int,
	amount *big.Int, data []byte) (types.Transactions, error) {
	numAddrs := len(bcdata.addrs)
	fromAddrs := bcdata.addrs

	if numTransactions%numAddrs != 0 {
		return nil, fmt.Errorf("numTranasctions should be divided by numAddrs! numTransactions: %v, numAddrs: %v", numTransactions, numAddrs)
	}

	var err error
	fromNonces := make([]uint64, numAddrs)
	for i, addr := range fromAddrs {
		fromNonces[i], err = accountMap.GetNonce(*addr)
		if err != nil {
			return nil, err
		}
	}

	txs := make(types.Transactions, 0, numTransactions)
	txAmount := amount
	if txAmount == nil {
		txAmount = big.NewInt(rand.Int63n(10))
		txAmount = txAmount.Add(txAmount, big.NewInt(1))
	}
	var gasLimit uint64 = 1000000
	gasPrice := new(big.Int).SetInt64(0)
	for i := 0; i < numTransactions; i++ {
		fromIdx := i % numAddrs

		toIdx := (fromIdx + 1) % numAddrs
		toAddr := *bcdata.addrs[toIdx]

		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    fromNonces[fromIdx],
			types.TxValueKeyTo:       toAddr,
			types.TxValueKeyAmount:   txAmount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			types.TxValueKeyFrom:     *bcdata.addrs[fromIdx],
		})

		signedTx, err := types.SignTx(tx, signer, bcdata.privKeys[fromIdx])
		if err != nil {
			return nil, err
		}

		txs = append(txs, signedTx)
		fromNonces[fromIdx]++
	}

	return txs, nil
}

func TestValueTransfer(t *testing.T) {
	var nBlocks int = 3
	var txPerBlock int = 10

	if i, err := strconv.ParseInt(os.Getenv("NUM_BLOCKS"), 10, 32); err == nil {
		nBlocks = int(i)
	}

	if i, err := strconv.ParseInt(os.Getenv("TXS_PER_BLOCK"), 10, 32); err == nil {
		txPerBlock = int(i)
	}

	var valueTransferTests = [...]struct {
		name string
		opt  testOption
	}{
		{"SingleSenderMultipleRecipient",
			testOption{txPerBlock, 1000, 4, nBlocks, []byte{}, makeTransactionsFrom}},
		{"MultipleSenderMultipleRecipient",
			testOption{txPerBlock, 2000, 4, nBlocks, []byte{}, makeIndependentTransactions}},
		{"MultipleSenderRandomRecipient",
			testOption{txPerBlock, 2000, 4, nBlocks, []byte{}, makeTransactionsToRandom}},
		{"MultipleSenderMultipleRecipientRingTx",
			testOption{2000, 1000, 4, nBlocks, []byte{}, makeNewTransactionsToRing}},

		// Below test cases execute one transaction per a block.
		{"SingleSenderMultipleRecipientSingleTxPerBlock",
			testOption{1, 1000, 4, 10, []byte{}, makeTransactionsFrom}},
		{"MultipleSenderMultipleRecipientSingleTxPerBlock",
			testOption{1, 2000, 4, 10, []byte{}, makeIndependentTransactions}},
		{"MultipleSenderRandomRecipientSingleTxPerBlock",
			testOption{1, 2000, 4, 10, []byte{}, makeTransactionsToRandom}},
	}

	for _, test := range valueTransferTests {
		t.Run(test.name, func(t *testing.T) {
			testValueTransfer(t, &test.opt)
		})
	}
}

func testValueTransfer(t *testing.T, opt *testOption) {
	prof := profile.NewProfiler()

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(opt.numMaxAccounts, opt.numValidators)
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

	for i := 0; i < opt.numGeneratedBlocks; i++ {
		//fmt.Printf("iteration %d\n", i)
		err := bcdata.GenABlock(accountMap, opt, opt.numTransactions, prof)
		if err != nil {
			t.Fatal(err)
		}
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

func TestValueTransferRing(t *testing.T) {
	var valueTransferTests = [...]struct {
		name string
		opt  testOption
	}{
		{"RingTxValueTransfer2000TxsPerBlock1000Accounts",
			testOption{2000, 1000, 4, 10, []byte{}, makeNewTransactionsToRing}},
		{"RingTxValueTransfer3000TxsPerBlock1000Accounts",
			testOption{3000, 1000, 4, 10, []byte{}, makeNewTransactionsToRing}},
		{"RingTxValueTransfer4000TxsPerBlock1000Accounts",
			testOption{4000, 1000, 4, 10, []byte{}, makeNewTransactionsToRing}},
	}

	for _, test := range valueTransferTests {
		t.Run(test.name, func(t *testing.T) {
			testValueTransferRing(t, &test.opt)
		})
	}
}

func testValueTransferRing(t *testing.T, opt *testOption) {
	prof := profile.NewProfiler()

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(opt.numMaxAccounts, opt.numValidators)
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

	statedb, err := bcdata.bc.State()
	if err != nil {
		t.Fatal(err)
	}
	preBalance := statedb.GetBalance(*bcdata.addrs[0])

	for i := 0; i < opt.numGeneratedBlocks; i++ {
		//fmt.Printf("iteration %d\n", i)
		err := bcdata.GenABlock(accountMap, opt, opt.numTransactions, prof)
		if err != nil {
			t.Fatal(err)
		}
	}

	statedb, err = bcdata.bc.State()
	if err != nil {
		t.Fatal(err)
	}
	postBalance := statedb.GetBalance(*bcdata.addrs[0])

	if preBalance.Cmp(postBalance) != 0 {
		t.Fatal("Different balance after ring transactions")
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// BenchmarkValueTransfer measures TPS without network traffics
// while creating a block. As a disclaimer, this function does not tell that Klaytn
// can perform this amount of TPS in real environment.
func BenchmarkValueTransfer(t *testing.B) {
	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()
	opt := testOption{t.N, 2000, 4, 1, []byte{}, makeTransactionsToRandom}

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(opt.numMaxAccounts, opt.numValidators)
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

	// make txpool
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.AccountSlots = uint64(t.N)
	txpoolconfig.AccountQueue = uint64(t.N)
	txpoolconfig.GlobalSlots = 2 * uint64(t.N)
	txpoolconfig.GlobalQueue = 2 * uint64(t.N)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)
	signer := types.MakeSigner(bcdata.bc.Config(), bcdata.bc.CurrentHeader().Number)

	// make t.N transactions
	txs, err := makeIndependentTransactions(bcdata, accountMap, signer, t.N, nil, []byte{})
	if err != nil {
		t.Fatal(err)
	}

	t.ResetTimer()

	txpool.AddRemotes(txs)

	for {
		if err := bcdata.GenABlockWithTxpool(accountMap, txpool, prof); err != nil {
			if err == errEmptyPending {
				break
			}
			t.Fatal(err)
		}
	}
	t.StopTimer()

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// BenchmarkNewValueTransfer measures TPS without network traffics
// while creating a block. As a disclaimer, this function does not tell that Klaytn
// can perform this amount of TPS in real environment.
func BenchmarkNewValueTransfer(t *testing.B) {
	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()
	opt := testOption{t.N, 2000, 4, 1, []byte{}, makeNewTransactionsToRandom}

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(opt.numMaxAccounts, opt.numValidators)
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

	// make txpool
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.AccountSlots = uint64(t.N)
	txpoolconfig.AccountQueue = uint64(t.N)
	txpoolconfig.GlobalSlots = 2 * uint64(t.N)
	txpoolconfig.GlobalQueue = 2 * uint64(t.N)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)
	signer := types.MakeSigner(bcdata.bc.Config(), bcdata.bc.CurrentHeader().Number)

	// make t.N transactions
	txs, err := makeIndependentTransactions(bcdata, accountMap, signer, t.N, nil, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	txpool.AddRemotes(txs)

	t.ResetTimer()
	for {
		if err := bcdata.GenABlockWithTxpool(accountMap, txpool, prof); err != nil {
			if err == errEmptyPending {
				break
			}
			t.Fatal(err)
		}
	}
	t.StopTimer()

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}
