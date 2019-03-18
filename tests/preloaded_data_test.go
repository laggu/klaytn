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
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/blockchain/vm"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/consensus/istanbul"
	"github.com/ground-x/klaytn/governance"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/storage/database"
	"github.com/ground-x/klaytn/work"
	"github.com/otiai10/copy"
	"math/big"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	istanbulBackend "github.com/ground-x/klaytn/consensus/istanbul/backend"
)

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

// makeTxsWithStateDB generates transactions with the nonce retrieved from stateDB.
// stateDB is used only once to initialize nonceMap, and then nonceMap is used instead of stateDB.
func makeTxsWithStateDB(bcdata *BCData, toAddrs []*common.Address, signer types.Signer, numTransactions int) (types.Transactions, error) {
	fromAddrs := bcdata.addrs
	if len(fromAddrs) != len(toAddrs) {
		return nil, fmt.Errorf("len(fromAddrs) %v != len(toAddrs) %v", len(fromAddrs), len(toAddrs))
	}

	stateDB, err := bcdata.bc.State()
	if err != nil {
		return nil, err
	}

	// Use nonceMap, not to change the nonce of stateDB.
	nonceMap := make(map[common.Address]uint64)
	for _, addr := range fromAddrs {
		nonce := stateDB.GetNonce(*addr)
		nonceMap[*addr] = nonce
	}

	// Generate value transfer transactions from initial account to the given "toAddrs".
	txs := make(types.Transactions, 0, numTransactions)
	lenAddrs := len(fromAddrs)
	for i := 0; i < numTransactions; i++ {
		idx := i % lenAddrs
		fromAddr := *fromAddrs[idx]
		toAddr := *toAddrs[idx]
		nonce := nonceMap[fromAddr]

		var gasLimit uint64 = 1000000
		gasPrice := new(big.Int).SetInt64(0)

		tx := types.NewTransaction(nonce, toAddr, big.NewInt(10000), gasLimit, gasPrice, nil)
		signedTx, err := types.SignTx(tx, signer, bcdata.privKeys[idx])
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

var numFiles = 25
var numAccountsPerFile = 1 * 10000
var numValidators = 4

func TestPreLoad10000Accounts(t *testing.T) {
	preLoadedTest(t, "10000Accounts", dirFor10000Accounts, numAccountsPerFile*1)
}

func TestPreLoadOneMillionAccounts(t *testing.T) {
	preLoadedTest(t, "1500000Accounts", dirForOneMillionAccounts, numAccountsPerFile*1)
}

// preLoadedTest is to check the performance of Klaytn with pre-loaded data.
// To run the test, original data directory should be located at "$GOPATH/src/github.com/ground-x/"
func preLoadedTest(t *testing.T, testName, originalDataDir string, numTransactionsPerFile int) {
	readTestDir, err := settingUpPreLoadedData(originalDataDir)
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

	bcData, err := NewBCDataFromPreLoadedData(readTestDir, numValidators)
	if err != nil {
		t.Fatal(err)
	}
	defer bcData.db.Close()
	defer bcData.bc.Stop()

	txPool := makeTxPool(bcData, numTransactionsPerFile)
	signer := types.MakeSigner(bcData.bc.Config(), bcData.bc.CurrentHeader().Number)

	timeNow := time.Now()
	f, err := os.Create(testName + "_" + timeNow.Format("2006-01-02 15:04:05") + ".cpu.out")
	if err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= numFiles; i++ {
		// Read recipient addresses from file.
		toAddrs, err := readAddrsFromFile(readTestDir, i)
		if err != nil {
			t.Fatal(err)
		}

		// Generate transactions
		txs, err := makeTxsWithStateDB(bcData, toAddrs, signer, numTransactionsPerFile)
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

		// NOTE-Klaytn-Tests If you want to get profile, please set numFiles as 1, to start profile only once.
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

		txPool.AddRemotes(txs)
		for {
			if err := bcData.GenABlockWithTxPoolWithoutAccountMap(txPool); err != nil {
				if err == errEmptyPending {
					break
				}
				t.Fatal(err)
			}
		}
	}
}

// GenABlockWithTxPoolWithoutAccountMap basically does the same thing with GenABlockWithTxPool,
// however, it does not accept AccountMap which validates the outcome with stateDB.
// This is to remove the overhead of AccountMap management.
func (bcdata *BCData) GenABlockWithTxPoolWithoutAccountMap(txPool *blockchain.TxPool) error {
	signer := types.MakeSigner(bcdata.bc.Config(), bcdata.bc.CurrentHeader().Number)

	pending, err := txPool.Pending()
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		return errEmptyPending
	}

	// TODO-Klaytn-Issue136 gasPrice
	pooltxs := types.NewTransactionsByPriceAndNonce(signer, pending)

	// Set the block header
	header, err := bcdata.prepareHeader()
	if err != nil {
		return err
	}

	stateDB, err := bcdata.bc.State()
	if err != nil {
		return err
	}

	gp := new(blockchain.GasPool)
	gp = gp.AddGas(GasLimit)
	task := work.NewTask(bcdata.bc.Config(), signer, stateDB, gp, header)
	task.ApplyTransactions(pooltxs, bcdata.bc, *bcdata.rewardBase)
	newtxs := task.Transactions()
	receipts := task.Receipts()

	if len(newtxs) == 0 {
		return errEmptyPending
	}

	// Finalize the block.
	b, err := bcdata.engine.Finalize(bcdata.bc, header, stateDB, newtxs, []*types.Header{}, receipts)
	if err != nil {
		return err
	}

	// Seal the block.
	b, err = sealBlock(b, bcdata.validatorPrivKeys)
	if err != nil {
		return err
	}

	// Insert the block into the blockchain.
	if n, err := bcdata.bc.InsertChain(types.Blocks{b}); err != nil {
		return fmt.Errorf("err = %s, n = %d\n", err, n)
	}

	fmt.Println("blockNum", b.NumberU64(), "numTransactions", len(newtxs))

	return nil
}

// TestGenerateTensOfMillionsAccounts generates given number of 1) accounts which actually
// are put into state trie, 2) their addresses, and 3) their private keys.
func TestGenerateTensOfMillionsAccounts(t *testing.T) {

	numTransactionsPerTxPool := 10000
	numAccountsPerFile := 10000
	numFilesToGenerate := 5000

	fmt.Printf("numAccountsPerFile: %v, numFilesToGenerate: %v, totalAccountsToGenerate: %v \n",
		numAccountsPerFile, numFilesToGenerate, numAccountsPerFile*numFilesToGenerate)

	if numAccountsPerFile%numTransactionsPerTxPool != 0 {
		t.Fatal("numAccountsPerFile % numTransactionsPerTxPool != 0")
	}

	// Make a new blockchain.
	bcData, err := NewBCData(numAccountsPerFile, numValidators)
	if err != nil {
		t.Fatal(err)
	}

	defer bcData.db.Close()
	defer bcData.bc.Stop()

	// Write Initial "numAccountsPerFile" accounts to file.
	// These accounts will transfer klay to other accounts.
	if err := writeToFile(bcData.addrs, bcData.privKeys, 0); err != nil {
		t.Fatal(err)
	}

	// make txPool.
	txPool := makeTxPool(bcData, numTransactionsPerTxPool)
	signer := types.MakeSigner(bcData.bc.Config(), bcData.bc.CurrentHeader().Number)

	numTxGenerationRunsPerFile := numAccountsPerFile / numTransactionsPerTxPool

	// Nonce here is globally managed since initial "numAccountsPerFile" will send
	// the same number of transactions to the newly created accounts.
	nonce := 0
	for i := 1; i <= numFilesToGenerate; i++ {
		newAddresses, newPrivateKeys, err := createAccounts(numAccountsPerFile)
		if err != nil {
			t.Fatal(err)
		}
		writeToFile(newAddresses, newPrivateKeys, i)

		// Send seed money to newAddresses,
		// to 1) actually create accounts at state trie and 2) charge the balance.
		for k := 0; k < numTxGenerationRunsPerFile; k++ {
			// make t.N transactions
			txs, err := makeTxsWithGivenNonce(bcData, newAddresses, signer, numTransactionsPerTxPool, k*numTransactionsPerTxPool, nonce)
			if err != nil {
				t.Fatal(err)
			}

			txPool.AddRemotes(txs)

			for {
				if err := bcData.GenABlockWithTxPoolWithoutAccountMap(txPool); err != nil {
					if err == errEmptyPending {
						break
					}
					t.Fatal(err)
				}
			}

			nonce++
		}
	}
}

// NewBCDataFromPreLoadedData returns a new BCData pointer constructed from the existing data.
func NewBCDataFromPreLoadedData(dbDir string, numValidators int) (*BCData, error) {
	// Remove transactions.rlp if exists
	if _, err := os.Stat(transactionsJournalFilename); err == nil {
		os.RemoveAll(transactionsJournalFilename)
	}

	////////////////////////////////////////////////////////////////////////////////
	// Create a database
	chainDB := NewDatabase(path.Join(dbDir, "chaindata"), database.LevelDB)

	addrs, privKeys, err := readAddrsAndPrivateKeysFromFile(dbDir, 0)
	if err != nil {
		return nil, err
	}

	////////////////////////////////////////////////////////////////////////////////
	// Set the genesis address
	genesisAddr := *addrs[0]

	////////////////////////////////////////////////////////////////////////////////
	// Use first 4 accounts as validators
	validatorPrivKeys := make([]*ecdsa.PrivateKey, numValidators)
	validatorAddresses := make([]common.Address, numValidators)
	for i := 0; i < numValidators; i++ {
		validatorPrivKeys[i] = privKeys[i]
		validatorAddresses[i] = *addrs[i]
	}

	////////////////////////////////////////////////////////////////////////////////
	// Create a governance
	gov := governance.NewGovernance(&params.ChainConfig{
		ChainID:       big.NewInt(2018),
		UnitPrice:     25000000000,
		DeriveShaImpl: 0,
		Istanbul: &params.IstanbulConfig{
			Epoch:          istanbul.DefaultConfig.Epoch,
			ProposerPolicy: uint64(istanbul.DefaultConfig.ProposerPolicy),
			SubGroupSize:   istanbul.DefaultConfig.SubGroupSize,
		},
		Governance: governance.GetDefaultGovernanceConfig(params.UseIstanbul),
	})

	////////////////////////////////////////////////////////////////////////////////
	// Setup istanbul consensus backend
	engine := istanbulBackend.New(genesisAddr, istanbul.DefaultConfig, validatorPrivKeys[0], chainDB, gov)
	////////////////////////////////////////////////////////////////////////////////
	// Make a blockchain
	trieConfig := &blockchain.CacheConfig{
		ArchiveMode:   true,
		CacheSize:     256 * 1024 * 1024,
		BlockInterval: blockchain.DefaultBlockInterval,
	}

	stored := chainDB.ReadBlockByNumber(0)
	if stored == nil {
		return nil, errors.New("chainDB.ReadBlockByNumber(0) == nil")
	}

	storedcfg := chainDB.ReadChainConfig(stored.Hash())
	if storedcfg == nil {
		return nil, errors.New("storedcfg == nil")
	}

	bc, err := blockchain.NewBlockChain(chainDB, trieConfig, storedcfg, engine, vm.Config{})
	if err != nil {
		return nil, err
	}

	return &BCData{bc, addrs, privKeys, chainDB,
		&genesisAddr, validatorAddresses,
		validatorPrivKeys, engine}, nil
}

// BenchmarkRandomStateTrieRead randomly reads stateObjects
// to check the read performance for given pre-generated data.
func BenchmarkRandomStateTrieRead(b *testing.B) {
	randomStateTrieRead(b, dirForOneMillionAccounts)
}

func randomStateTrieRead(b *testing.B, originalDataDir string) {
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

		bcData, err := NewBCDataFromPreLoadedData(readTestDir, numValidators)
		if err != nil {
			b.Fatal(err)
		}

		stateDB, err := bcData.bc.State()
		if err != nil {
			b.Fatal(err)
		}

		for i := 0; i < addressFileSamplingSize; i++ {
			b.StopTimer()

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
		}

		b.StopTimer()
		bcData.db.Close()
		bcData.bc.Stop()
	}
}
