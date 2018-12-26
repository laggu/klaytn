// Copyright 2018 The go-klaytn Authors
//
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

package grpc

import (
	"github.com/ground-x/go-gxplatform/accounts/abi/bind"
	"github.com/ground-x/go-gxplatform/cmd/sol2proto/protobuf"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/crypto"
	"math/big"
	"os"
)

type TransactOptsFn func(m *protobuf.TransactOpts) *bind.TransactOpts

// DefaultTransactOptsFn
func DefaultTransactOptsFn(m *protobuf.TransactOpts) *bind.TransactOpts {
	privateKey, err := crypto.ToECDSA(common.Hex2Bytes(m.PrivateKey))
	if err != nil {
		os.Exit(-1)
	}
	auth := bind.NewKeyedTransactor(privateKey)
	if m.GasLimit < 0 {
		// get system suggested gas limit
		auth.GasLimit = uint64(0)
	} else {
		auth.GasLimit = uint64(m.GasLimit)
	}

	if m.GasPrice < 0 {
		// get system suggested gas price
		auth.GasPrice = nil
	} else {
		auth.GasPrice = big.NewInt(m.GasPrice)
	}

	if m.Nonce < 0 {
		// get system account nonce
		auth.Nonce = nil
	} else {
		auth.Nonce = big.NewInt(m.Nonce)
	}
	auth.Value = big.NewInt(m.Value)
	return auth
}

// BigIntArrayToBytes converts []*big.Int to [][]byte
func BigIntArrayToBytes(ints []*big.Int) (b [][]byte) {
	for _, i := range ints {
		if i == nil {
			b = append(b, nil)
		} else {
			b = append(b, i.Bytes())
		}
	}
	return
}

// BytesToBigIntArray converts [][]byte to []*big.Int
func BytesToBigIntArray(b [][]byte) (ints []*big.Int) {
	for _, i := range b {
		if i == nil {
			ints = append(ints, new(big.Int).SetInt64(0))
		} else {
			ints = append(ints, new(big.Int).SetBytes(i))
		}
	}
	return
}

// BytesToBytes32 converts []byte to [32]byte
func BytesToBytes32(b []byte) (bs [32]byte) {
	copyLen := len(b)
	if copyLen == 0 {
		return
	} else if copyLen > 32 {
		copyLen = 32
	}
	copy(bs[:], b[:copyLen])
	return
}

// BytesArrayToBytes32Array converts [][]byte to [][32]byte
func BytesArrayToBytes32Array(b [][]byte) (bs [][32]byte) {
	bs = make([][32]byte, len(b))
	for i := 0; i < len(b); i++ {
		bs[i] = BytesToBytes32(b[i])
	}
	return
}
