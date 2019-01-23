// Copyright 2018 The klaytn Authors
// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from core/state/dump.go (2018/06/04).
// Modified and improved for the klaytn development.

package state

import (
	"encoding/json"
	"fmt"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/ground-x/klaytn/storage/statedb"
)

type DumpAccount struct {
	Balance  string            `json:"balance"`
	Nonce    uint64            `json:"nonce"`
	Root     string            `json:"root"`
	CodeHash string            `json:"codeHash"`
	Code     string            `json:"code"`
	Storage  map[string]string `json:"storage"`
}

type Dump struct {
	Root     string                 `json:"root"`
	Accounts map[string]DumpAccount `json:"accounts"`
}

func (self *StateDB) RawDump() Dump {
	dump := Dump{
		Root:     fmt.Sprintf("%x", self.trie.Hash()),
		Accounts: make(map[string]DumpAccount),
	}

	it := statedb.NewIterator(self.trie.NodeIterator(nil))
	for it.Next() {
		addr := self.trie.GetKey(it.Key)
		serializer := NewAccountSerializer()
		if err := rlp.DecodeBytes(it.Value, serializer); err != nil {
			panic(err)
		}
		data := serializer.account

		obj := self.getStateObject(common.BytesToAddress(addr))
		account := DumpAccount{
			Balance:  data.GetBalance().String(),
			Nonce:    data.GetNonce(),
			Root:     common.Bytes2Hex([]byte{}),
			CodeHash: common.Bytes2Hex([]byte{}),
			Code:     common.Bytes2Hex(obj.Code(self.db)),
			Storage:  make(map[string]string),
		}
		if pa, ok := data.(ProgramAccount); ok {
			account.Root = common.Bytes2Hex(pa.GetStorageRoot().Bytes())
			account.CodeHash = common.Bytes2Hex(pa.GetCodeHash())
		}
		storageTrie := obj.getStorageTrie(self.db)
		storageIt := statedb.NewIterator(storageTrie.NodeIterator(nil))
		for storageIt.Next() {
			account.Storage[common.Bytes2Hex(storageTrie.GetKey(storageIt.Key))] = common.Bytes2Hex(storageIt.Value)
		}
		dump.Accounts[common.Bytes2Hex(addr)] = account
	}
	return dump
}

func (self *StateDB) Dump() []byte {
	json, err := json.MarshalIndent(self.RawDump(), "", "    ")
	if err != nil {
		fmt.Println("dump err", err)
	}

	return json
}
