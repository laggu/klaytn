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
	"encoding/hex"
	"fmt"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/crypto"
	"github.com/pkg/errors"
	"math/big"
	"os"
	"path"
	"strconv"
	"sync"
	"testing"
)

// BenchmarkGenerateTensOfMillionsAccounts generates given number of 1) accounts which actually
// are put into state trie, 2) their addresses, and 3) their private keys.
func BenchmarkGenerateTensOfMillionsAccounts(b *testing.B) {

	numTransactionsPerTxPool := 10000
	numAccountsPerFile := 10000
	numFilesToGenerate := 5000

	fmt.Printf("numAccountsPerFile: %v, numFilesToGenerate: %v, totalAccountsToGenerate: %v \n",
		numAccountsPerFile, numFilesToGenerate, numAccountsPerFile*numFilesToGenerate)

	if numAccountsPerFile%numTransactionsPerTxPool != 0 {
		b.Fatal("numAccountsPerFile % numTransactionsPerTxPool != 0")
	}

	// Make a new blockchain.
	bcData, err := NewBCData(numAccountsPerFile, numValidators)
	if err != nil {
		b.Fatal(err)
	}

	defer bcData.db.Close()
	defer bcData.bc.Stop()

	// Write Initial "numAccountsPerFile" accounts to file.
	// These accounts will transfer klay to other accounts.
	currDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if err := writeToFile(bcData.addrs, bcData.privKeys, 0, currDir); err != nil {
		b.Fatal(err)
	}

	// make txPool.
	txPool := makeTxPool(bcData, numTransactionsPerTxPool)
	signer := types.MakeSigner(bcData.bc.Config(), bcData.bc.CurrentHeader().Number)

	numTxGenerationRunsPerFile := numAccountsPerFile / numTransactionsPerTxPool

	// Nonce here is globally managed since initial "numAccountsPerFile" will send
	// the same number of transactions to the newly created accounts.
	nonce := 0

	b.ResetTimer()
	for i := 1; i < numFilesToGenerate; i++ {
		b.StopTimer()

		newAddresses, newPrivateKeys, err := createAccounts(numAccountsPerFile)
		if err != nil {
			b.Fatal(err)
		}
		writeToFile(newAddresses, newPrivateKeys, i, currDir)

		// Send seed money to newAddresses,
		// to 1) actually create accounts at state trie and 2) charge the balance.
		for k := 0; k < numTxGenerationRunsPerFile; k++ {
			b.StopTimer()

			// make t.N transactions
			txs, err := makeTxsWithGivenNonce(bcData, newAddresses, signer, numTransactionsPerTxPool, k*numTransactionsPerTxPool, nonce)
			if err != nil {
				b.Fatal(err)
			}

			b.StartTimer()

			txPool.AddRemotes(txs)

			for {
				if err := bcData.GenABlockWithTxPoolWithoutAccountMap(txPool); err != nil {
					if err == errEmptyPending {
						break
					}
					b.Fatal(err)
				}
			}

			nonce++
		}
	}
}

// makeTxsWithGivenNonce generates transactions with the given nonce.
// It assumes that all accounts of fromAddrs have the same nonce.
func makeTxsWithGivenNonce(bcdata *BCData, toAddrs []*common.Address, signer types.Signer, numTransactions, startIndex, nonce int) (types.Transactions, error) {
	fromAddrs := bcdata.addrs

	if len(fromAddrs) != len(toAddrs) {
		return nil, errors.New("len(fromAddrs) != len(toAddrs)")
	}

	if len(fromAddrs) != numTransactions {
		return nil, errors.New("len(fromAddrs) != numTransactions")
	}

	fromAddrs = bcdata.addrs[startIndex : startIndex+numTransactions]
	toAddrs = toAddrs[startIndex : startIndex+numTransactions]

	txs := make(types.Transactions, 0, numTransactions)

	for i := 0; i < numTransactions; i++ {
		idx := i

		var gasLimit uint64 = 1000000
		gasPrice := new(big.Int).SetInt64(0)

		tx := types.NewTransaction(uint64(nonce), *toAddrs[idx], big.NewInt(10000), gasLimit, gasPrice, nil)
		signedTx, err := types.SignTx(tx, signer, bcdata.privKeys[idx+startIndex])
		if err != nil {
			return nil, err
		}

		txs = append(txs, signedTx)
	}

	return txs, nil
}

func writeToFile(addrs []*common.Address, privKeys []*ecdsa.PrivateKey, num int, dir string) error {
	_ = os.Mkdir(path.Join(dir, addressDirectory), os.ModePerm)
	_ = os.Mkdir(path.Join(dir, privateKeyDirectory), os.ModePerm)

	addrsFile, err := os.Create(path.Join(dir, addressDirectory, addressFilePrefix+strconv.Itoa(num)))
	if err != nil {
		return err
	}

	privateKeysFile, err := os.Create(path.Join(dir, privateKeyDirectory, privateKeyFilePrefix+strconv.Itoa(num)))
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}

	wg.Add(2)

	syncSize := len(addrs) / 2

	go func() {
		for i, b := range addrs {
			addrsFile.WriteString(b.String() + "\n")
			if (i+1)%syncSize == 0 {
				addrsFile.Sync()
			}
		}

		addrsFile.Close()
		wg.Done()
	}()

	go func() {
		for i, key := range privKeys {
			privateKeysFile.WriteString(hex.EncodeToString(crypto.FromECDSA(key)) + "\n")
			if (i+1)%syncSize == 0 {
				privateKeysFile.Sync()
			}
		}

		privateKeysFile.Close()
		wg.Done()
	}()

	wg.Wait()
	return nil
}
