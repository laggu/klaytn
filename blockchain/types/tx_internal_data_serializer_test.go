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
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
	"testing"
)

// TestTransactionSerialization tests RLP/JSON serialization for TxInternalData
func TestTransactionSerialization(t *testing.T) {
	var txs = []struct {
		Name string
		tx   TxInternalData
	}{
		{"OriginalTx", genLegacyTransaction()},
		{"ValueTransfer", genValueTransferTransaction()},
	}

	var testcases = []struct {
		Name string
		fn   func(t *testing.T, tx TxInternalData)
	}{
		{"RLP", testTransactionRLP},
		{"JSON", testTransactionJSON},
	}

	for _, test := range testcases {
		for _, tx := range txs {
			Name := test.Name + "/" + tx.Name
			t.Run(Name, func(t *testing.T) {
				test.fn(t, tx.tx)
			})
		}
	}
}

func testTransactionRLP(t *testing.T, tx TxInternalData) {
	enc := newTxInternalDataSerializer(tx)

	b, err := rlp.EncodeToBytes(enc)
	if err != nil {
		panic(err)
	}

	dec := newEmptyTxInternalDataSerializer()

	if err := rlp.DecodeBytes(b, &dec); err != nil {
		panic(err)
	}

	if !tx.Equal(dec.tx) {
		t.Fatalf("tx != dec.tx\ntx=%v\ndec.tx=%v", tx, dec.tx)
	}
}

func testTransactionJSON(t *testing.T, tx TxInternalData) {
	enc := newTxInternalDataSerializer(tx)

	b, err := json.Marshal(enc)
	if err != nil {
		panic(err)
	}

	dec := newEmptyTxInternalDataSerializer()

	if err := json.Unmarshal(b, &dec); err != nil {
		panic(err)
	}

	if !tx.Equal(dec.tx) {
		t.Fatalf("tx != dec.tx\ntx=%v\ndec.tx=%v", tx, dec.tx)
	}
}

func genLegacyTransaction() TxInternalData {
	txdata, _ := NewTxInternalDataWithMap(TxTypeLegacyTransaction, map[TxValueKeyType]interface{}{
		TxValueKeyNonce:    uint64(1234),
		TxValueKeyTo:       common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b"),
		TxValueKeyAmount:   new(big.Int).SetUint64(10),
		TxValueKeyGasLimit: uint64(9999999999),
		TxValueKeyGasPrice: new(big.Int).SetUint64(25),
		TxValueKeyData:     []byte("1234"),
	})

	return txdata
}

func genValueTransferTransaction() TxInternalData {
	d, err := NewTxInternalDataWithMap(TxTypeValueTransfer, map[TxValueKeyType]interface{}{
		TxValueKeyNonce:    uint64(1234),
		TxValueKeyTo:       common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b"),
		TxValueKeyAmount:   new(big.Int).SetUint64(10),
		TxValueKeyGasLimit: uint64(9999999999),
		TxValueKeyGasPrice: new(big.Int).SetUint64(25),
		TxValueKeyFrom:     common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b"),
	})

	if err != nil {
		// Since we do not have testing.T here, call panic() instead of t.Fatal().
		panic(err)
	}

	return d
}
