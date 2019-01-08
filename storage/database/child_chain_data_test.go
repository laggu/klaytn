// Copyright 2018 The go-klaytn Authors
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

package database

import (
	"github.com/ground-x/go-gxplatform/common"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestChildChainData_ReadAndWrite_ChildChainTxHash(t *testing.T) {
	dir, err := ioutil.TempDir("", "klaytn-test-child-chain-data")
	if err != nil {
		t.Fatalf("cannot create temporary directory: %v", err)
	}
	defer os.RemoveAll(dir)

	dbm, err := NewDBManager(dir, LEVELDB, true, 32, 32)
	if err != nil {
		t.Fatalf("cannot create DBManager: %v", err)
	}
	defer dbm.Close()

	ccBlockHash := common.HexToHash("0x0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e")
	ccTxHash := common.HexToHash("0x0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f")

	// Before writing the data into DB, nil should be returned.
	ccTxHashFromDB := dbm.ConvertChildChainBlockHashToParentChainTxHash(ccBlockHash)
	assert.Equal(t, common.Hash{}, ccTxHashFromDB)

	// After writing the data into DB, data should be returned.
	dbm.WriteChildChainTxHash(ccBlockHash, ccTxHash)
	ccTxHashFromDB = dbm.ConvertChildChainBlockHashToParentChainTxHash(ccBlockHash)
	assert.NotNil(t, ccTxHashFromDB)
	assert.Equal(t, ccTxHash, ccTxHashFromDB)

	ccBlockHashFake := common.HexToHash("0x0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a")
	// Invalid information should not return the data.
	ccTxHashFromDB = dbm.ConvertChildChainBlockHashToParentChainTxHash(ccBlockHashFake)
	assert.Equal(t, common.Hash{}, ccTxHashFromDB)
}
