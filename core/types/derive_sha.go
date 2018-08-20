// Copyright 2015 The go-ethereum Authors
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

package types

import (
	"bytes"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/rlp"
	"github.com/ground-x/go-gxplatform/trie"
	"github.com/ground-x/go-gxplatform/crypto/sha3"
)

type DerivableList interface {
	Len() int
	GetRlp(i int) []byte
}

func DeriveSha(list DerivableList) common.Hash {
	keybuf := new(bytes.Buffer)
	trie := new(trie.Trie)
	for i := 0; i < list.Len(); i++ {
		keybuf.Reset()
		rlp.Encode(keybuf, uint(i))
		trie.Update(keybuf.Bytes(), list.GetRlp(i))
	}
	return trie.Hash()
}

// An alternative implementation of DeriveSha()
// This function generates a hash of `DerivableList` by simulating merkle tree generation
// TODO-GX: Replace calls of DeriveSha() with this function
func DeriveShaSimple(list DerivableList) common.Hash {
	hasher := sha3.NewKeccak256()

	encoded := make([][]byte, list.Len())
	for i := 0; i < list.Len(); i++ {
		hasher.Reset()
		hasher.Write(list.GetRlp((i)))
		encoded[i] = hasher.Sum(nil)
	}

	for len(encoded) > 1 {
		// make even numbers
		if len(encoded) % 2 == 1 {
			encoded = append(encoded, encoded[len(encoded)-1])
		}

		for i := 0; i < len(encoded) / 2; i++ {
			hasher.Reset()
			hasher.Write(encoded[2*i])
			hasher.Write(encoded[2*i+1])

			encoded[i] = hasher.Sum(nil)
		}

		encoded = encoded[0:len(encoded)/2]
	}

	return common.BytesToHash(encoded[0])
}

// An alternative implementation of DeriveSha()
// This function generates a hash of `DerivableList` as below:
// 1. make a byte slice by concatenating RLP-encoded items
// 2. make a hash of the byte slice.
// TODO-GX: Replace calls of DeriveSha() with this function
func DeriveShaConcat(list DerivableList) (hash common.Hash) {
	hasher := sha3.NewKeccak256()
	keybuf := new(bytes.Buffer)

	for i := 0; i < list.Len(); i++ {
		keybuf.Write(list.GetRlp(i))
	}
	hasher.Write(keybuf.Bytes())
	hasher.Sum(hash[:0])

	return hash
}
