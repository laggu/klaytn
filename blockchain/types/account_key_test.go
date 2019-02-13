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

package types

import (
	"encoding/json"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/ser/rlp"
	"testing"
)

func TestAccountKeySerialization(t *testing.T) {
	var keys = []struct {
		Name string
		k    AccountKey
	}{
		{"Legacy", genAccountKeyLegacy()},
		{"Public", genAccountKeyPublic()},
		{"Fail", genAccountKeyFail()},
		{"WeightedMultisig", genAccountKeyWeightedMultisig()},
	}

	var testcases = []struct {
		Name string
		fn   func(t *testing.T, k AccountKey)
	}{
		{"RLP", testAccountKeyRLP},
		{"JSON", testAccountKeyJSON},
	}
	for _, test := range testcases {
		for _, key := range keys {
			Name := test.Name + "/" + key.Name
			t.Run(Name, func(t *testing.T) {
				test.fn(t, key.k)
			})
		}
	}
}

func testAccountKeyRLP(t *testing.T, k AccountKey) {
	enc := NewAccountKeySerializerWithAccountKey(k)

	b, err := rlp.EncodeToBytes(enc)
	if err != nil {
		t.Fatal(err)
	}

	dec := NewAccountKeySerializer()

	if err := rlp.DecodeBytes(b, &dec); err != nil {
		t.Fatal(err)
	}

	switch k.Type() {
	case AccountKeyTypeFail:
		if k.Equal(dec.key) {
			t.Errorf("AlwaysFail key returns true! k != dec.key\nk=%v\ndec.key=%v", k, dec.key)
		}
	default:
		if !k.Equal(dec.key) {
			t.Errorf("AlwaysFail key returns true! k != dec.key\nk=%v\ndec.key=%v", k, dec.key)
		}
	}
}

func testAccountKeyJSON(t *testing.T, k AccountKey) {
	enc := NewAccountKeySerializerWithAccountKey(k)

	b, err := json.Marshal(enc)
	if err != nil {
		t.Fatal(err)
	}

	dec := NewAccountKeySerializer()

	if err := json.Unmarshal(b, &dec); err != nil {
		t.Fatal(err)
	}

	switch k.Type() {
	case AccountKeyTypeFail:
		if k.Equal(dec.key) {
			t.Errorf("AlwaysFail key returns true! k != dec.key\nk=%v\ndec.key=%v", k, dec.key)
		}
	default:
		if !k.Equal(dec.key) {
			t.Errorf("AlwaysFail key returns true! k != dec.key\nk=%v\ndec.key=%v", k, dec.key)
		}
	}
}

func genAccountKeyLegacy() AccountKey {
	return NewAccountKeyLegacy()
}

func genAccountKeyPublic() AccountKey {
	k, _ := crypto.GenerateKey()
	return NewAccountKeyPublicWithValue(&k.PublicKey)
}

func genAccountKeyFail() AccountKey {
	return NewAccountKeyFail()
}

func genAccountKeyWeightedMultisig() AccountKey {
	threshold := uint(3)
	numKeys := 4
	keys := make(WeightedPublicKeys, numKeys)

	for i := 0; i < numKeys; i++ {
		k, _ := crypto.GenerateKey()
		keys[i] = NewWeightedPublicKey(1, (*PublicKeySerializable)(&k.PublicKey))
	}

	return NewAccountKeyWeightedMultiSigWithValues(threshold, keys)
}
