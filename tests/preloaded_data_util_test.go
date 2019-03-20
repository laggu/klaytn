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
	"bufio"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/consensus/istanbul"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/governance"
	"github.com/ground-x/klaytn/params"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"math"
	"math/big"
	"os"
	"path"
	"strconv"
	"sync"
)

const (
	addressDirectory    = "addrs"
	privateKeyDirectory = "privatekeys"

	addressFilePrefix    = "addrs_"
	privateKeyFilePrefix = "privateKeys_"
)

// getDataDirName returns a name of directory from the given parameters.
func getDataDirName(numFilesToGenerate int, ldbOption *opt.Options) string {
	dataDirectory := fmt.Sprintf("testdata%v", numFilesToGenerate)

	if ldbOption == nil {
		return dataDirectory
	}

	dataDirectory += fmt.Sprintf("NoSyncIs%s", strconv.FormatBool(ldbOption.NoSync))

	// Below codes can be used if necessary.
	//dataDirectory += fmt.Sprintf("_BlockCacheCapacity%vMB", ldbOption.BlockCacheCapacity / opt.MiB)
	//dataDirectory += fmt.Sprintf("_CompactionTableSize%vMB", ldbOption.CompactionTableSize / opt.MiB)
	//dataDirectory += fmt.Sprintf("_CompactionTableSizeMultiplier%v", int(ldbOption.CompactionTableSizeMultiplier))

	return dataDirectory
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

func readAddrsFromFile(dir string, num int) ([]*common.Address, error) {
	var addrs []*common.Address

	addrsFile, err := os.Open(path.Join(dir, addressDirectory, addressFilePrefix+strconv.Itoa(num)))
	if err != nil {
		return nil, err
	}

	defer addrsFile.Close()

	scanner := bufio.NewScanner(addrsFile)
	for scanner.Scan() {
		keyStr := scanner.Text()
		addr := common.HexToAddress(keyStr)
		addrs = append(addrs, &addr)
	}

	return addrs, nil
}

func readPrivateKeysFromFile(dir string, num int) ([]*ecdsa.PrivateKey, error) {
	var privKeys []*ecdsa.PrivateKey
	privateKeysFile, err := os.Open(path.Join(dir, privateKeyDirectory, privateKeyFilePrefix+strconv.Itoa(num)))
	if err != nil {
		return nil, err
	}

	defer privateKeysFile.Close()

	scanner := bufio.NewScanner(privateKeysFile)
	for scanner.Scan() {
		keyStr := scanner.Text()

		key, err := hex.DecodeString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}

		if pk, err := crypto.ToECDSA(key); err != nil {
			return nil, fmt.Errorf("%v", err)
		} else {
			privKeys = append(privKeys, pk)
		}
	}

	return privKeys, nil
}

func readAddrsAndPrivateKeysFromFile(dir string, num int) ([]*common.Address, []*ecdsa.PrivateKey, error) {
	addrs, err := readAddrsFromFile(dir, num)
	if err != nil {
		return nil, nil, err
	}

	privateKeys, err := readPrivateKeysFromFile(dir, num)
	if err != nil {
		return nil, nil, err
	}

	return addrs, privateKeys, nil
}

// makeAddrsFromFile extracts the address stored in file by numAccounts.
func makeAddrsFromFile(numAccounts int, fileDir string) ([]*common.Address, error) {
	addrs := make([]*common.Address, 0, numAccounts)

	remain := numAccounts
	fileIndex := 0
	for remain > 0 {
		// Read recipient addresses from file.
		addrsPerFile, err := readAddrsFromFile(fileDir, fileIndex)

		if err != nil {
			return nil, err
		}

		partSize := int(math.Min(float64(len(addrsPerFile)), float64(remain)))
		addrs = append(addrs, addrsPerFile[:partSize]...)
		remain -= partSize
		fileIndex++
	}

	return addrs, nil
}

// makeAddrsAndPrivKeysFromFile extracts the address and private key stored in file by numAccounts.
func makeAddrsAndPrivKeysFromFile(numAccounts int, fileDir string) ([]*common.Address, []*ecdsa.PrivateKey, error) {
	addrs := make([]*common.Address, 0, numAccounts)
	privKeys := make([]*ecdsa.PrivateKey, 0, numAccounts)

	remain := numAccounts
	fileIndex := 0
	for remain > 0 {
		// Read addresses and private keys from file.
		addrsPerFile, privKeysPerFile, err := readAddrsAndPrivateKeysFromFile(fileDir, fileIndex)

		if err != nil {
			return nil, nil, err
		}

		partSize := int(math.Min(float64(len(addrsPerFile)), float64(remain)))
		addrs = append(addrs, addrsPerFile[:partSize]...)
		privKeys = append(privKeys, privKeysPerFile[:partSize]...)
		remain -= partSize
		fileIndex++
	}

	return addrs, privKeys, nil
}

// generateGovernaceDataForTest returns *governance.Governance for test.
func generateGovernaceDataForTest() *governance.Governance {
	return governance.NewGovernance(&params.ChainConfig{
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
}

// getValidatorAddrsAndKeys returns the first `numValidators` addresses and private keys
// for validators.
func getValidatorAddrsAndKeys(addrs []*common.Address, privateKeys []*ecdsa.PrivateKey, numValidators int) ([]common.Address, []*ecdsa.PrivateKey) {
	validatorAddresses := make([]common.Address, numValidators)
	validatorPrivateKeys := make([]*ecdsa.PrivateKey, numValidators)

	for i := 0; i < numValidators; i++ {
		validatorPrivateKeys[i] = privateKeys[i]
		validatorAddresses[i] = *addrs[i]
	}

	return validatorAddresses, validatorPrivateKeys
}
