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
	"errors"
	"fmt"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/state"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/params"
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
	dirFor10MillionAccounts = "testdata1000_orig"
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
func makeTxsWithStateDB(stateDB *state.StateDB, fromAddrs []*common.Address, fromKeys []*ecdsa.PrivateKey, toAddrs []*common.Address, signer types.Signer, numTransactions int, indexPicker func(int, int) int) (types.Transactions, error) {
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
		fromIdx := indexPicker(i, lenFromAddrs)
		toIdx := indexPicker(i, lenToAddrs)

		fromAddr := *fromAddrs[fromIdx]
		fromKey := fromKeys[fromIdx]
		fromNonce := nonceMap[fromAddr]

		toAddr := *toAddrs[toIdx]

		tx := types.NewTransaction(fromNonce, toAddr, new(big.Int).Mul(big.NewInt(1e3), big.NewInt(params.KLAY)), 1000000, new(big.Int).SetInt64(25000000000), nil)
		signedTx, err := types.SignTx(tx, signer, fromKey)
		if err != nil {
			return nil, err
		}

		txs = append(txs, signedTx)
		nonceMap[fromAddr]++
	}

	return txs, nil
}

// setupTestDir does two things. If it is a data-generating test, it will just
// return the target path. If it is not a data-generating test, it will remove
// previously existing path and then copy the original data to the target path.
func setupTestDir(originalDataDir string, isGenerateTest bool) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Original data directory should be located at github.com/ground-x
	// Therefore, it should be something like github.com/ground-x/testdata150_orig
	grandParentPath := filepath.Dir(filepath.Dir(wd))

	// If it is generating test case, just returns the path.
	if isGenerateTest {
		return path.Join(grandParentPath, originalDataDir), nil
	}

	dataDirPath := path.Join(grandParentPath, originalDataDir)
	if _, err := os.Stat(dataDirPath); err != nil {
		return "", errNoOriginalDataDir
	}

	testDir := strings.Split(originalDataDir, "_orig")[0]

	originalDataPath := path.Join(grandParentPath, originalDataDir)
	testDataPath := path.Join(grandParentPath, testDir)

	os.RemoveAll(testDataPath)
	if err := copy.Copy(originalDataPath, testDataPath); err != nil {
		return "", err
	}
	return testDataPath, nil
}

type preGeneratedTC struct {
	isGenerateTest  bool
	testName        string
	originalDataDir string

	numTotalTxs  int
	numTxsPerGen int

	numTotalSenders   int
	numTotalReceivers int
	indexPicker       func(int, int) int

	dbc           *database.DBConfig
	levelDBOption *opt.Options
	cacheConfig   *blockchain.CacheConfig
}

func BenchmarkPreGeneratedProfiling(b *testing.B) {
	tc1 := &preGeneratedTC{
		isGenerateTest: false, testName: "1000Accounts", originalDataDir: dirFor10MillionAccounts,
		numTotalTxs: 10 * 10000, numTxsPerGen: 20000,
		numTotalSenders: 20000, numTotalReceivers: 20000, indexPicker: randomIndex}

	tc1.cacheConfig = &blockchain.CacheConfig{
		StateDBCaching:   true,
		TxPoolStateCache: true,
		ArchiveMode:      true,
		CacheSize:        256 * 1024 * 1024,
		BlockInterval:    blockchain.DefaultBlockInterval,
	}

	tc1.dbc = generateDefaultDBConfig()

	tc1.levelDBOption = generateDefaultLevelDBOption()

	preGeneratedProfilingTest(b, tc1)
}

// preGeneratedProfilingTest is to check the performance of Klaytn with pre-generated data.
// To run the test, original data directory should be located at "$GOPATH/src/github.com/ground-x/"
func preGeneratedProfilingTest(b *testing.B, tc *preGeneratedTC) {
	if tc.numTotalTxs%tc.numTxsPerGen != 0 {
		b.Fatalf("tc.numTotalTxs %% tc.numTxsPerGen != 0, tc.numTotalTxs: %v, tc.numTxsPerGen: %v",
			tc.numTotalTxs, tc.numTxsPerGen)
	}

	testDataDir, err := setupTestDir(tc.originalDataDir, tc.isGenerateTest)
	if err != nil {
		// If original data directory does not exist for generating test, it is okay.
		if !(tc.isGenerateTest && err == errNoOriginalDataDir) {
			b.Fatal(err)
		}
	}

	if testing.Verbose() {
		enableLog()
	}

	bcData, err := NewBCDataForPreGeneratedTest(testDataDir, tc)
	if err != nil {
		b.Fatal(err)
	}

	defer bcData.db.Close()
	defer bcData.bc.Stop()

	txPool := makeTxPool(bcData, tc.numTotalTxs)
	signer := types.MakeSigner(bcData.bc.Config(), bcData.bc.CurrentHeader().Number)

	timeNow := time.Now()
	f, err := os.Create(tc.testName + "_" + timeNow.Format("2006-01-02-1504") + ".cpu.out")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.StopTimer()

	numTxGenerationRuns := tc.numTotalTxs / tc.numTxsPerGen
	for run := 1; run <= numTxGenerationRuns; run++ {
		var toAddrs []*common.Address
		if tc.isGenerateTest {
			var newPrivateKeys []*ecdsa.PrivateKey
			toAddrs, newPrivateKeys, err = createAccounts(tc.numTxsPerGen)
			writeToFile(toAddrs, newPrivateKeys, run, testDataDir)
		} else {
			toAddrs, err = makeAddrsFromFile(tc.numTotalReceivers, testDataDir, tc.indexPicker)
		}

		if err != nil {
			b.Fatal(err)
		}

		// Generate transactions
		stateDB, err := bcData.bc.State()
		if err != nil {
			b.Fatal(err)
		}

		txs, err := makeTxsWithStateDB(stateDB, bcData.addrs, bcData.privKeys, toAddrs, signer, tc.numTxsPerGen, tc.indexPicker)
		if err != nil {
			b.Fatal(err)
		}

		for _, tx := range txs {
			tx.AsMessageWithAccountKeyPicker(signer, stateDB, bcData.bc.CurrentBlock().NumberU64())
		}

		b.StartTimer()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

		txPool.AddRemotes(txs)

		for {
			if err := bcData.GenABlockWithTxPoolWithoutAccountMap(txPool); err != nil {
				if err == errEmptyPending {
					break
				}
				b.Fatal(err)
			}
		}

		b.StopTimer()
	}
}
