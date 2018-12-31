// Copyright 2018 The go-klaytn Authors
//
// This file is part of the go-klaytn library.
//
// The go-klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-klaytn library. If not, see <http://www.gnu.org/licenses/>.

package tests

import (
	"flag"
	"github.com/ground-x/go-gxplatform/blockchain"
	"github.com/ground-x/go-gxplatform/blockchain/types"
	"github.com/ground-x/go-gxplatform/common/profile"
	"github.com/ground-x/go-gxplatform/crypto"
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
	nonce := accountMap.GetNonce(from)
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

	fromNonces := make([]uint64, numAddrs)
	for i, addr := range fromAddrs {
		fromNonces[i] = accountMap.GetNonce(*addr)
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

	fromNonces := make([]uint64, numAddrs)
	for i, addr := range fromAddrs {
		fromNonces[i] = accountMap.GetNonce(*addr)
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

// BenchmarkValueTransfer measures TPS without txpool operations and network traffics
// while creating a block. As a disclaimer, this function does not tell that Klaytn
// can perform this amount of TPS in real environment.
func BenchmarkValueTransfer(t *testing.B) {
	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()
	opt := testOption{t.N, 2000, 4, 1, []byte{}, makeIndependentTransactions}

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
