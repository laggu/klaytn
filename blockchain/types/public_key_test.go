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

package types

import (
	"encoding/json"
	"github.com/ground-x/go-gxplatform/ser/rlp"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

// TestPublicKeyRLP tests RLP encoding/decoding of PublicKeySerializable.
func TestPublicKeyRLP(t *testing.T) {
	k := newEmptyPublicKeySerializable()
	k.X.SetUint64(10)
	k.Y.SetUint64(20)

	b, err := rlp.EncodeToBytes(k)
	if err != nil {
		t.Fatal(err)
	}

	dec := newEmptyPublicKeySerializable()

	if err := rlp.DecodeBytes(b, &dec); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, k, dec)
	if !k.Equal(dec) {
		t.Fatal("k != dec")
	}
}

// TestPublicKeyRLP tests JSON encoding/decoding of PublicKeySerializable.
func TestPublicKeyJSON(t *testing.T) {
	k := newEmptyPublicKeySerializable()
	k.X.SetUint64(10)
	k.Y.SetUint64(20)

	b, err := json.Marshal(k)
	if err != nil {
		t.Fatal(err)
	}

	dec := newEmptyPublicKeySerializable()

	if err := json.Unmarshal(b, &dec); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, k, dec)
	if !k.Equal(dec) {
		t.Fatal("k != dec")
	}
}

// TestPublicKeyRLP tests DeepCopy() of PublicKeySerializable.
func TestPublicKeyDeepCopy(t *testing.T) {
	k := newEmptyPublicKeySerializable()
	k.X.SetUint64(10)
	k.Y.SetUint64(20)

	newK := k.DeepCopy()

	newK.X.SetUint64(30)
	newK.Y.SetUint64(40)

	assert.Equal(t, k.X, big.NewInt(10))
	assert.Equal(t, k.Y, big.NewInt(20))
	assert.Equal(t, newK.X, big.NewInt(30))
	assert.Equal(t, newK.Y, big.NewInt(40))
}
