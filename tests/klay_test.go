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
	"time"
	"testing"
	"math/big"
	"math/rand"
	"github.com/ground-x/go-gxplatform/core/types"
	"github.com/ground-x/go-gxplatform/common/profile"
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
		a := toAddrs[i % numAddrs]
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

func TestValueTransfer(t *testing.T) {
	var valueTransferTests = [...]struct {
		name string
		opt testOption
	} {
		{"SingleSenderMultipleRecipient",
		 testOption{1000, 1000, 4, 3, []byte{}, makeTransactionsFrom}},
		{"MultipleSenderMultipleRecipient",
		 testOption{1000, 2000, 4, 3, []byte{}, makeIndependentTransactions}},
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

func BenchmarkValueTransfer(t *testing.B) {
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

	t.ResetTimer()
	for i := 0; i < t.N/txPerBlock; i++ {
		//fmt.Printf("iteration %d tx %d\n", i, opt.numTransactions)
		err := bcdata.GenABlock(accountMap, &opt, txPerBlock, prof)
		if err != nil {
			t.Fatal(err)
		}
	}

	genBlocks := t.N / txPerBlock
	remainTxs := t.N % txPerBlock
	if remainTxs != 0 {
		err := bcdata.GenABlock(accountMap, &opt, remainTxs, prof)
		if err != nil {
			t.Fatal(err)
		}
		genBlocks++
	}
	t.StopTimer()

	bcHeight := int(bcdata.bc.CurrentHeader().Number.Uint64())
	if bcHeight != genBlocks {
		t.Fatalf("generated blocks should be %d, but %d.\n", genBlocks, bcHeight)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}
