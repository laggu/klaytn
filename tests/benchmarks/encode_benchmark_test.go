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

package benchmarks

import (
	"bytes"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
	"testing"
)

type txMapData map[types.TxValueKeyType]interface{}

// getTestPeggedData returns a sample pegged data.
func genTestPeggedData() []byte {
	blockTxData := &types.ChildChainTxData{
		BlockHash:     common.HexToHash("0"),
		TxHash:        common.HexToHash("1"),
		ParentHash:    common.HexToHash("2"),
		ReceiptHash:   common.HexToHash("3"),
		StateRootHash: common.HexToHash("4"),
		BlockNumber:   big.NewInt(5),
	}
	peggedData, err := rlp.EncodeToBytes(blockTxData)
	if err != nil {
		panic(err)
	}
	return peggedData
}

// genTestTxInternalData creates a `txType` transaction internal data derived from `txMap`.
func genTestTxInternalData(txType types.TxType) types.TxInternalData {
	txMap := txMapData{
		types.TxValueKeyNonce:    uint64(1234),
		types.TxValueKeyTo:       common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b"),
		types.TxValueKeyAmount:   new(big.Int).SetUint64(10),
		types.TxValueKeyGasLimit: uint64(9999999999),
		types.TxValueKeyGasPrice: new(big.Int).SetUint64(25),
	}
	switch txType {
	case types.TxTypeLegacyTransaction:
		txMap[types.TxValueKeyData] = []byte("1234")
	case types.TxTypeValueTransfer:
		txMap[types.TxValueKeyFrom] = common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	case types.TxTypeChainDataPegging:
		txMap[types.TxValueKeyPeggedData] = genTestPeggedData()
	case types.TxTypeAccountCreation:
		k, _ := crypto.GenerateKey()
		txMap[types.TxValueKeyHumanReadable] = true
		txMap[types.TxValueKeyAccountKey] = types.NewAccountKeyPublicWithValue(&k.PublicKey)
	}
	txData, err := types.NewTxInternalDataWithMap(txType, txMap)
	if err != nil {
		panic(err)
	}
	return txData
}

// Auxiliary functions for benchmark tests

// benchmarkEncode is an auxiliary function to do encode internal data by `rlp.Encode`.
func benchmarkEncode(b *testing.B, txType types.TxType) {
	testTxData := genTestTxInternalData(txType)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, testTxData); err != nil {
			b.Error(err)
		}
		buffer.Bytes()
	}
}

// benchmarkEncodeToBytes is an auxiliary function to do encode internal data by `rlp.EncodeToBytes`.
func benchmarkEncodeToBytes(b *testing.B, txType types.TxType) {
	testTxData := genTestTxInternalData(txType)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := rlp.EncodeToBytes(testTxData); err != nil {
			b.Error(err)
		}
	}
}

// benchmarkEncodeInterface is an auxiliary function to do encode interface of transaction internal data.
func benchmarkEncodeInterface(b *testing.B, txType types.TxType) {
	testTxData := genTestTxInternalData(txType)
	txInterfaces := testTxData.SerializeForSign()
	v, r, s := testTxData.GetVRS()
	txInterfaces = append(txInterfaces, v, r, s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, txInterfaces); err != nil {
			b.Error(err)
		}
		buffer.Bytes()
	}
}

// benchmarkEncodeInterfaceOverFields is an auxiliary function to do encoding interface list of transaction internal data
func benchmarkEncodeInterfaceOverFields(b *testing.B, txType types.TxType) {
	testTxData := genTestTxInternalData(txType)
	txInterfaces := testTxData.SerializeForSign()
	v, r, s := testTxData.GetVRS()
	txInterfaces = append(txInterfaces, v, r, s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := new(bytes.Buffer)
		for _, field := range txInterfaces {
			if err := rlp.Encode(buffer, field); err != nil {
				b.Error(err)
			}
		}
		buffer.Bytes()
	}
}

// genCommonDefaultData generates a common internal transaction data
func genCommonDefaultData() types.TxInternalDataCommon {
	return types.TxInternalDataCommon{
		AccountNonce: uint64(1234),
		Price:        new(big.Int).SetUint64(25),
		GasLimit:     uint64(9999999999),
		Recipient:    common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b"),
		Amount:       new(big.Int).SetUint64(10),
		Hash:         &common.Hash{},
	}
}

// encodeSeparateFieldsLegacy encodes both common data and extra field separately.
func encodeSeparateFieldsLegacy(b *testing.B) {
	commonData := genCommonDefaultData()
	extra := []byte("1234")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, commonData); err != nil {
			b.Error(err)
		}
		if err := rlp.Encode(buffer, extra); err != nil {
			b.Error(err)
		}
		buffer.Bytes()
	}
}

// encodeSeparateFieldsValueTransfer encodes both common data and extra field separately.
func encodeSeparateFieldsValueTransfer(b *testing.B) {
	commonData := genCommonDefaultData()
	extra := common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, commonData); err != nil {
			b.Error(err)
		}
		if err := rlp.Encode(buffer, extra); err != nil {
			b.Error(err)
		}
		buffer.Bytes()
	}
}

// encodeSeparateFieldsChainDataPegging encodes both common data and extra field separately.
func encodeSeparateFieldsChainDataPegging(b *testing.B) {
	commonData := genCommonDefaultData()
	extra := genTestPeggedData()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, commonData); err != nil {
			b.Error(err)
		}
		if err := rlp.Encode(buffer, extra); err != nil {
			b.Error(err)
		}
		buffer.Bytes()
	}
}

// encodeSeparateFieldsAccountCreation encodes both common data and extra field separately.
func encodeSeparateFieldsAccountCreation(b *testing.B) {
	commonData := genCommonDefaultData()
	k, _ := crypto.GenerateKey()
	extra := struct {
		HumanReadable bool
		AccountKey    *types.AccountKeyPublic
	}{
		true,
		types.NewAccountKeyPublicWithValue(&k.PublicKey),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, commonData); err != nil {
			b.Error(err)
		}
		if err := rlp.Encode(buffer, extra); err != nil {
			b.Error(err)
		}
		buffer.Bytes()
	}
}

// benchmarkEncodeExtraSeparateFields is an auxiliary function to do encoding a common struct and an extra field separately.
func benchmarkEncodeExtraSeparateFields(b *testing.B, txType types.TxType) {
	switch txType {
	case types.TxTypeLegacyTransaction:
		encodeSeparateFieldsLegacy(b)
	case types.TxTypeValueTransfer:
		encodeSeparateFieldsValueTransfer(b)
	case types.TxTypeChainDataPegging:
		encodeSeparateFieldsChainDataPegging(b)
	case types.TxTypeAccountCreation:
		encodeSeparateFieldsAccountCreation(b)
	}
}

// Main benchmark function
func BenchmarkRLPEncoding(b *testing.B) {
	var options = []struct {
		Name        string
		Subfunction func(*testing.B, types.TxType)
	}{
		{"Encode", benchmarkEncode},
		{"EncodeToBytes", benchmarkEncodeToBytes},
		{"EncodeInterface", benchmarkEncodeInterface},
		{"EncodeInterfaceList", benchmarkEncodeInterfaceOverFields},
		{"EncodeExtraSeparateFields", benchmarkEncodeExtraSeparateFields},
	}
	var testMaterials = []struct {
		TypeName string
		Type     types.TxType
	}{
		{"legacyTx", types.TxTypeLegacyTransaction},
		{"valueTransferTx", types.TxTypeValueTransfer},
		{"chainDataPeggingTx", types.TxTypeChainDataPegging},
		{"accountCreationTx", types.TxTypeAccountCreation},
	}
	for _, option := range options {
		for _, data := range testMaterials {
			name := option.Name + "/" + data.TypeName
			b.Run(name, func(b *testing.B) {
				option.Subfunction(b, data.Type)
			})
		}
	}
}
