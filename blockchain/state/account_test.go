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

package state

import (
	"fmt"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/crypto/sha3"
	"github.com/ground-x/go-gxplatform/ser/rlp"
	"math/big"
	"math/rand"
	"testing"
)

// TestAccountSerialization tests serialization of various account types.
func TestAccountSerialization(t *testing.T) {
	var accs = []struct {
		Name string
		acc  Account
	}{
		{"LegacyAccount", genLegacyAccount()},
	}
	var testcases = []struct {
		Name string
		fn   func(t *testing.T, acc Account)
	}{
		{"RLP", testAccountRLP},
	}
	for _, test := range testcases {
		for _, acc := range accs {
			Name := test.Name + "/" + acc.Name
			t.Run(Name, func(t *testing.T) {
				test.fn(t, acc.acc)
			})
		}
	}
}

func testAccountRLP(t *testing.T, acc Account) {
	enc := NewAccountSerializerWithAccount(acc)

	b, err := rlp.EncodeToBytes(enc)
	if err != nil {
		panic(err)
	}

	dec := NewAccountSerializer()

	if err := rlp.DecodeBytes(b, &dec); err != nil {
		panic(err)
	}

	if !acc.Equal(dec.account) {
		fmt.Println("acc")
		fmt.Println(acc)
		fmt.Println("dec.account")
		fmt.Println(dec.account)
		t.Errorf("acc != dec.account")
	}
}

func genRandomHash() (h common.Hash) {
	hasher := sha3.NewKeccak256()

	r := rand.Uint64()
	rlp.Encode(hasher, r)
	hasher.Sum(h[:0])

	return h
}

func genLegacyAccount() *LegacyAccount {
	return newLegacyAccountWithMap(map[AccountValueKeyType]interface{}{
		AccountValueKeyNonce:       rand.Uint64(),
		AccountValueKeyBalance:     big.NewInt(rand.Int63n(10000)),
		AccountValueKeyStorageRoot: genRandomHash(),
		AccountValueKeyCodeHash:    genRandomHash().Bytes(),
	})
}
