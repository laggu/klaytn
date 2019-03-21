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
// +build preloadtest

package tests

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ground-x/klaytn/blockchain/state"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/storage/database"
	"github.com/otiai10/copy"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"math/big"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var errNoOriginalDataDir = errors.New("original data directory does not exist, aborting the test")

const (
	dirFor10000Accounts = "testdata1_orig"

	dirForOneMillionAccounts               = "testdata100_orig"
	dirForOneMillionAccountsPartitioned    = "testdata100_partitioned_orig"
	dirForOneMillionAccountsPartitioned4MB = "testdata100_partitioned_4mb_orig"

	dirForFiveMillionAccounts               = "testdata500_orig"
	dirForFiveMillionAccountsPartitioned    = "testdata500_partitioned_orig"
	dirForFiveMillionAccountsPartitioned4MB = "testdata500_partitioned_4mb_orig"
)

// randomIndex is used to access data with random index.
func randomIndex(index, lenAddrs int) int {
	return rand.Intn(lenAddrs)
}

// sequentialIndex is used to access data with sequential index.
func sequentialIndex(index, lenAddrs int) int {
	return index % lenAddrs
}

// makeTxsWithStateDB generates transactions with the nonce retrieved from stateDB.
// stateDB is used only once to initialize nonceMap, and then nonceMap is used instead of stateDB.
func makeTxsWithStateDB(stateDB *state.StateDB, fromAddrs []*common.Address, fromKeys []*ecdsa.PrivateKey, toAddrs []*common.Address, signer types.Signer, numTransactions int, addrPicker func(int, int) int) (types.Transactions, error) {
	if len(fromAddrs) != len(fromKeys) {
		return nil, fmt.Errorf("len(fromAddrs) %v != len(fromKeys) %v", len(fromAddrs), len(fromKeys))
	}

	// Use nonceMap, not to change the nonce of stateDB.
	nonceMap := make(map[common.Address]uint64)
	for _, addr := range fromAddrs {
		nonce := stateDB.GetNonce(*addr)
		nonceMap[*addr] = nonce
	}

	// Generate value transfer transactions from initial account to the given "toAddrs".
	txs := make(types.Transactions, 0, numTransactions)
	lenFromAddrs := len(fromAddrs)
	lenToAddrs := len(toAddrs)
	for i := 0; i < numTransactions; i++ {
		fromIdx := addrPicker(i, lenFromAddrs)
		toIdx := addrPicker(i, lenToAddrs)

		fromAddr := *fromAddrs[fromIdx]
		fromKey := fromKeys[fromIdx]
		fromNonce := nonceMap[fromAddr]

		toAddr := *toAddrs[toIdx]

		tx := types.NewTransaction(fromNonce, toAddr, big.NewInt(10000), 1000000, new(big.Int).SetInt64(0), nil)
		signedTx, err := types.SignTx(tx, signer, fromKey)
		if err != nil {
			return nil, err
		}

		txs = append(txs, signedTx)
		nonceMap[fromAddr]++
	}

	return txs, nil
}

// getOriginalDataDirPath returns the path of original data dir, if exists.
func getOriginalDataDirPath(originalDataDir string) string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Original data directory should be located at github.com/ground-x
	// Therefore, it should be something like github.com/ground-x/testdata150_orig
	grandParentPath := filepath.Dir(filepath.Dir(wd))
	dataDirPath := path.Join(grandParentPath, originalDataDir)

	if _, err := os.Stat(dataDirPath); err != nil {
		return ""
	} else {
		return grandParentPath
	}
}

// settingUpPreLoadedData removes previously used data directory and then copies
// original data directory to the new one for the next run.
func settingUpPreLoadedData(originalDataDir string) (string, error) {
	if grandParentPath := getOriginalDataDirPath(originalDataDir); grandParentPath == "" {
		return "", errNoOriginalDataDir
	} else {
		testDir := strings.Split(originalDataDir, "_")[0]
		os.RemoveAll(path.Join(grandParentPath, testDir))

		originalDataPath := path.Join(grandParentPath, originalDataDir)
		testDataPath := path.Join(grandParentPath, testDir)

		if err := copy.Copy(originalDataPath, testDataPath); err != nil {
			return "", err
		}
		return testDataPath, nil
	}
}

var numAccountsPerFile = 1 * 10000 // numAccountsPerFile is the number of accounts per file to create the pre-prepared data for the preload test.

type preLoadedProfilingTC struct {
	testName          string
	originalDataDir   string
	numTotalTxs       int
	numTotalSenders   int
	numTotalReceivers int
	addrPicker        func(int, int) int
}

func TestPreLoadedProfiling(t *testing.T) {
	tc1 := &preLoadedProfilingTC{"100AccountsRandom", dirForOneMillionAccounts,
		20000, 20000, 20000, randomIndex}

	tc2 := &preLoadedProfilingTC{"100AccountsRandom", dirForOneMillionAccounts,
		20000, 20000, 20000, sequentialIndex}

	preLoadedProfilingTest(t, tc1, generateDefaultDBConfig(), generateDefaultLevelDBOption())
	preLoadedProfilingTest(t, tc2, generateDefaultDBConfig(), generateDefaultLevelDBOption())
}

// preLoadedProfilingTest is to check the performance of Klaytn with pre-generated data.
// To run the test, original data directory should be located at "$GOPATH/src/github.com/ground-x/"
func preLoadedProfilingTest(t *testing.T, tc *preLoadedProfilingTC, dbc *database.DBConfig, levelDBOption *opt.Options) {
	testDataDir, err := settingUpPreLoadedData(tc.originalDataDir)
	if err != nil {
		// If original data directory does not exist, test does nothing.
		if err == errNoOriginalDataDir {
			return
		}
		t.Fatal(err)
	}

	if testing.Verbose() {
		enableLog()
	}

	bcData, err := NewBCDataForPreLoadedTest(testDataDir, tc.numTotalSenders, dbc, levelDBOption, false)
	if err != nil {
		t.Fatal(err)
	}

	defer bcData.db.Close()
	defer bcData.bc.Stop()

	txPool := makeTxPool(bcData, tc.numTotalTxs)
	signer := types.MakeSigner(bcData.bc.Config(), bcData.bc.CurrentHeader().Number)

	toAddrs, err := makeAddrsFromFile(tc.numTotalReceivers, testDataDir)
	if err != nil {
		t.Fatal(err)
	}

	// Generate transactions
	stateDB, err := bcData.bc.State()
	if err != nil {
		t.Fatal(err)
	}

	txs, err := makeTxsWithStateDB(stateDB, bcData.addrs, bcData.privKeys, toAddrs, signer, tc.numTotalTxs, tc.addrPicker)
	if err != nil {
		t.Fatal(err)
	}

	state, err := bcData.bc.State()
	if err != nil {
		t.Fatal(err)
	}

	for _, tx := range txs {
		tx.AsMessageWithAccountKeyPicker(signer, state)
	}

	timeNow := time.Now()
	f, err := os.Create(tc.testName + "_" + timeNow.Format("2006-01-02 15:04:05") + ".cpu.out")
	if err != nil {
		t.Fatal(err)
	}

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	txPool.AddRemotes(txs)
	for {
		if err := bcData.GenABlockWithTxPoolWithoutAccountMap(txPool); err != nil {
			if err == errEmptyPending {
				fmt.Println("sumTx", sumTx)
				break
			}
			t.Fatal(err)
		}
	}
}

// BenchmarkRandomStateTrieRead randomly reads stateObjects
// to check the read performance for given pre-generated data.
func BenchmarkRandomStateTrieRead(b *testing.B) {
	randomStateTrieRead(b, dirForOneMillionAccounts, 25, generateDefaultDBConfig(), generateDefaultLevelDBOption())
}

func randomStateTrieRead(b *testing.B, originalDataDir string, numFiles int, dbc *database.DBConfig, levelDBOption *opt.Options) {
	readTestDir, err := settingUpPreLoadedData(originalDataDir)
	if err != nil {
		// If original data directory does not exist, test does nothing.
		if err == errNoOriginalDataDir {
			return
		}
		b.Fatal(err)
	}

	fmt.Println("testDirectoryPath", readTestDir)

	if testing.Verbose() {
		enableLog()
	}

	b.ResetTimer()

	// Totally, (b.N x addressSamplingSizeFromAFile x addressFileSamplingSize) of getStateObject occur.
	// Each run is independent since blockchain is built for each run.
	addressSamplingSizeFromAFile := 500
	addressFileSamplingSize := 10

	rand.Seed(time.Now().UnixNano())
	for run := 0; run < b.N; run++ {
		b.StopTimer()

		bcData, err := NewBCDataForPreLoadedTest(readTestDir, numValidatorsForTest, dbc, levelDBOption, false)
		if err != nil {
			b.Fatal(err)
		}

		stateDB, err := bcData.bc.State()
		if err != nil {
			b.Fatal(err)
		}

		for i := 0; i < addressFileSamplingSize; i++ {
			// Choose a random file to read addresses from.
			fileIndex := rand.Intn(numFiles)

			// Read recipient addresses from file.
			addrs, err := readAddrsFromFile(readTestDir, fileIndex)
			if err != nil {
				b.Fatal(err)
			}

			b.StartTimer()

			lenAddrs := len(addrs)
			for k := 0; k < addressSamplingSizeFromAFile; k++ {
				if common.Big0.Cmp(stateDB.GetBalance(*addrs[rand.Intn(lenAddrs)])) == 0 {
					b.Fatal("zero balance!!")
				}
			}

			b.StopTimer()
		}

		b.StopTimer()
		bcData.db.Close()
		bcData.bc.Stop()
	}
}
