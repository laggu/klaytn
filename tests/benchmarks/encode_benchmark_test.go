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
		types.TxValueKeyAmount:   new(big.Int).SetUint64(10),
		types.TxValueKeyGasLimit: uint64(9999999999),
		types.TxValueKeyGasPrice: new(big.Int).SetUint64(25),
	}
	switch txType {
	case types.TxTypeLegacyTransaction:
		txMap[types.TxValueKeyData] = []byte("1234")
		to := common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b")
		txMap[types.TxValueKeyTo] = &to
	case types.TxTypeChainDataPegging:
		txMap[types.TxValueKeyPeggedData] = genTestPeggedData()
	case types.TxTypeAccountCreation:
		k, _ := crypto.GenerateKey()
		txMap[types.TxValueKeyHumanReadable] = true
		txMap[types.TxValueKeyAccountKey] = types.NewAccountKeyPublicWithValue(&k.PublicKey)
	}
	if txType != types.TxTypeLegacyTransaction {
		txMap[types.TxValueKeyFrom] = common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b")
		txMap[types.TxValueKeyTo] = common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	}
	txData, err := types.NewTxInternalDataWithMap(txType, txMap)
	if err != nil {
		panic(err)
	}
	return txData
}

// checkDecoding is called when the benchmark is verbose mode to check if the encoded is performed correctly by decoding it.
func checkDecoding(b *testing.B, txType types.TxType, original types.TxInternalData, encoded []byte) {
	container, err := types.NewTxInternalData(txType)
	if err == nil {
		err = rlp.DecodeBytes(encoded, container)
	}
	if err != nil {
		b.Error(err)
	}
	if original.Equal(container) {
		b.Log("Decoding interface is successful")
	} else {
		b.Errorf("Decoding interface failed. %s", container.String())
	}
}

// benchmarkEncode is an auxiliary function to do encode internal data by `rlp.Encode`.
func benchmarkEncode(b *testing.B, txType types.TxType) {
	testTxData := genTestTxInternalData(txType)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, testTxData); err != nil {
			b.Error(err)
		}
		bs := buffer.Bytes()
		if testing.Verbose() {
			checkDecoding(b, txType, testTxData, bs)
		}
	}
}

// benchmarkEncodeToBytes is an auxiliary function to do encode internal data by `rlp.EncodeToBytes`.
func benchmarkEncodeToBytes(b *testing.B, txType types.TxType) {
	testTxData := genTestTxInternalData(txType)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if bs, err := rlp.EncodeToBytes(testTxData); err != nil {
			b.Error(err)
		} else if testing.Verbose() {
			checkDecoding(b, txType, testTxData, bs)
		}
	}
}

// checkDecodingLegacyTxInterface checks if the encoding legacy transaction internal data in interface list type is correctly
// encoded by decoding it. It is called in verbose mode.
func checkDecodingLegacyTxInterface(b *testing.B, encoded []byte, original types.TxInternalData) (types.TxInternalData, bool) {
	var container struct {
		AccountNonce uint64
		Price        *big.Int
		GasLimit     uint64
		Recipient    *common.Address
		Amount       *big.Int
		Payload      []byte
		V            *big.Int
		R            *big.Int
		S            *big.Int
	}
	if err := rlp.DecodeBytes(encoded, &container); err != nil {
		b.Error(err)
	}
	decoded, err := types.NewTxInternalDataWithMap(types.TxTypeLegacyTransaction,
		txMapData{
			types.TxValueKeyNonce:    container.AccountNonce,
			types.TxValueKeyAmount:   container.Amount,
			types.TxValueKeyGasLimit: container.GasLimit,
			types.TxValueKeyGasPrice: container.Price,
			types.TxValueKeyData:     container.Payload,
			types.TxValueKeyTo:       container.Recipient,
		})
	if err != nil {
		b.Error(err)
	}
	decoded.SetVRS(container.V, container.R, container.S)
	return decoded, original.Equal(decoded)
}

// checkDecodingValueTransferTxInterface checks if the encoding of transaction internal data in value transfer type
// is done correctly by decoding it. It is called when the benchmark test is verbose mode.
func checkDecodingValueTransferTxInterface(b *testing.B, encoded []byte, original types.TxInternalData) (types.TxInternalData, bool) {
	var container struct {
		AccountNonce uint64
		Price        *big.Int
		GasLimit     uint64
		Recipient    common.Address
		Amount       *big.Int
		From         common.Address
		Type         types.TxType
		V            *big.Int
		R            *big.Int
		S            *big.Int
	}
	if err := rlp.DecodeBytes(encoded, &container); err != nil {
		b.Error(err)
	}

	decoded, err := types.NewTxInternalDataWithMap(types.TxTypeValueTransfer,
		txMapData{
			types.TxValueKeyNonce:    container.AccountNonce,
			types.TxValueKeyAmount:   container.Amount,
			types.TxValueKeyGasLimit: container.GasLimit,
			types.TxValueKeyGasPrice: container.Price,
			types.TxValueKeyFrom:     container.From,
			types.TxValueKeyTo:       container.Recipient,
		})
	if err != nil {
		b.Error(err)
	}
	decoded.SetVRS(container.V, container.R, container.S)
	return decoded, original.Equal(decoded)
}

// checkDecodingChainDataPeggingTxInterface checks if the encoding is correctly done or not in case of chain data pegging type
// transaction internal data in interface list. It is called when the benchmark is verbose mode.
func checkDecodingChainDataPeggingTxInterface(b *testing.B, encoded []byte, original types.TxInternalData) (types.TxInternalData, bool) {
	var container struct {
		AccountNonce uint64
		Price        *big.Int
		GasLimit     uint64
		Recipient    common.Address
		Amount       *big.Int
		From         common.Address
		Type         types.TxType
		PeggedData   []byte
		V            *big.Int
		R            *big.Int
		S            *big.Int
	}
	if err := rlp.DecodeBytes(encoded, &container); err != nil {
		b.Error(err)
	}
	decoded, err := types.NewTxInternalDataWithMap(types.TxTypeChainDataPegging,
		txMapData{
			types.TxValueKeyNonce:      container.AccountNonce,
			types.TxValueKeyAmount:     container.Amount,
			types.TxValueKeyGasLimit:   container.GasLimit,
			types.TxValueKeyGasPrice:   container.Price,
			types.TxValueKeyFrom:       container.From,
			types.TxValueKeyTo:         container.Recipient,
			types.TxValueKeyPeggedData: container.PeggedData,
		})
	if err != nil {
		b.Error(err)
	}
	decoded.SetVRS(container.V, container.R, container.S)
	return decoded, original.Equal(decoded)
}

// checkDecodingAccountCreationTxInterface checks if the encoding is correctly done in case of tx internal data interface list
// of account creation type. It is called when the benchmark is verbose mode.
func checkDecodingAccountCreationTxInterface(b *testing.B, encoded []byte, original types.TxInternalData) (types.TxInternalData, bool) {
	var container struct {
		AccountNonce  uint64
		Price         *big.Int
		GasLimit      uint64
		Recipient     common.Address
		Amount        *big.Int
		From          common.Address
		Type          types.TxType
		HumalReadable bool
		AccountKey    *types.AccountKeySerializer
		V             *big.Int
		R             *big.Int
		S             *big.Int
	}
	if err := rlp.DecodeBytes(encoded, &container); err != nil {
		b.Error(err)
	}
	decoded, err := types.NewTxInternalDataWithMap(types.TxTypeAccountCreation,
		txMapData{
			types.TxValueKeyNonce:         container.AccountNonce,
			types.TxValueKeyAmount:        container.Amount,
			types.TxValueKeyGasLimit:      container.GasLimit,
			types.TxValueKeyGasPrice:      container.Price,
			types.TxValueKeyFrom:          container.From,
			types.TxValueKeyTo:            container.Recipient,
			types.TxValueKeyHumanReadable: container.HumalReadable,
			types.TxValueKeyAccountKey:    container.AccountKey.GetKey(),
		})
	if err != nil {
		b.Error(err)
	}
	decoded.SetVRS(container.V, container.R, container.S)
	return decoded, original.Equal(decoded)
}

// checkInterfaceDecoding is to check if the transaction internal data is correctly encoded or not by decoding it.
// This function is called only when the benchmark is verbose mode.
func checkInterfaceDecoding(b *testing.B, txType types.TxType, encoded []byte, original types.TxInternalData) {
	var isSuccessful bool
	var decoded types.TxInternalData
	switch txType {
	case types.TxTypeLegacyTransaction:
		decoded, isSuccessful = checkDecodingLegacyTxInterface(b, encoded, original)
	case types.TxTypeValueTransfer:
		decoded, isSuccessful = checkDecodingValueTransferTxInterface(b, encoded, original)
	case types.TxTypeChainDataPegging:
		decoded, isSuccessful = checkDecodingChainDataPeggingTxInterface(b, encoded, original)
	case types.TxTypeAccountCreation:
		decoded, isSuccessful = checkDecodingAccountCreationTxInterface(b, encoded, original)
	}
	if isSuccessful {
		b.Log("Decoding interface is successful")
	} else {
		b.Errorf("Decoding interface failed. %s", decoded.String())
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
		bs := buffer.Bytes()
		if testing.Verbose() {
			checkInterfaceDecoding(b, txType, bs, testTxData)
		}
	}
}

// checkDecodingLegacyInterfaceOverFields is a subfunction to check the encoding is correctly done or not for legacy type.
// It is only called when the benchmark verbose option is set.
func checkDecodingLegacyInterfaceOverFields(b *testing.B, encoded []byte, original types.TxInternalData) (types.TxInternalData, bool) {
	var accountNonce uint64
	var price *big.Int
	var gasLimit uint64
	var recipient *common.Address
	var amount *big.Int
	var payload []byte
	var v *big.Int
	var r *big.Int
	var s *big.Int
	reader := bytes.NewReader(encoded)
	_ = rlp.Decode(reader, &accountNonce)
	_ = rlp.Decode(reader, &price)
	_ = rlp.Decode(reader, &gasLimit)
	_ = rlp.Decode(reader, &recipient)
	_ = rlp.Decode(reader, &amount)
	_ = rlp.Decode(reader, &payload)
	_ = rlp.Decode(reader, &v)
	_ = rlp.Decode(reader, &r)
	_ = rlp.Decode(reader, &s)
	decoded, err := types.NewTxInternalDataWithMap(types.TxTypeLegacyTransaction, txMapData{
		types.TxValueKeyNonce:    accountNonce,
		types.TxValueKeyAmount:   amount,
		types.TxValueKeyGasLimit: gasLimit,
		types.TxValueKeyGasPrice: price,
		types.TxValueKeyData:     payload,
		types.TxValueKeyTo:       recipient,
	})
	if err != nil {
		b.Error(err)
	}
	decoded.SetVRS(v, r, s)
	return decoded, original.Equal(decoded)
}

// checkDecodingValueTransferInterfaceOverFields is a subfunction to check the encoding is correctly done or not for
// value transfer type. It is only called when the benchmark verbose option is set.
func checkDecodingValueTransferInterfaceOverFields(b *testing.B, encoded []byte, original types.TxInternalData) (types.TxInternalData, bool) {
	var accountNonce uint64
	var price *big.Int
	var gasLimit uint64
	var recipient common.Address
	var amount *big.Int
	var from common.Address
	var txType types.TxType
	var v *big.Int
	var r *big.Int
	var s *big.Int
	reader := bytes.NewReader(encoded)
	_ = rlp.Decode(reader, &accountNonce)
	_ = rlp.Decode(reader, &price)
	_ = rlp.Decode(reader, &gasLimit)
	_ = rlp.Decode(reader, &recipient)
	_ = rlp.Decode(reader, &amount)
	_ = rlp.Decode(reader, &from)
	_ = rlp.Decode(reader, &txType)
	_ = rlp.Decode(reader, &v)
	_ = rlp.Decode(reader, &r)
	_ = rlp.Decode(reader, &s)
	decoded, err := types.NewTxInternalDataWithMap(types.TxTypeValueTransfer, txMapData{
		types.TxValueKeyNonce:    accountNonce,
		types.TxValueKeyAmount:   amount,
		types.TxValueKeyGasLimit: gasLimit,
		types.TxValueKeyGasPrice: price,
		types.TxValueKeyFrom:     from,
		types.TxValueKeyTo:       recipient,
	})
	if err != nil {
		b.Error(err)
	}
	decoded.SetVRS(v, r, s)
	return decoded, original.Equal(decoded)
}

// checkDecodingChainDataPeggingInterfaceOverFields is a subfunction to check if the encoding is correctly performed
// for the data pegging transaction type. It is only called when the benchmark verbose option is set.
func checkDecodingChainDataPeggingInterfaceOverFields(b *testing.B, encoded []byte, original types.TxInternalData) (types.TxInternalData, bool) {
	var accountNonce uint64
	var price *big.Int
	var gasLimit uint64
	var recipient common.Address
	var amount *big.Int
	var from common.Address
	var txType types.TxType
	var peggedData []byte
	var v *big.Int
	var r *big.Int
	var s *big.Int
	reader := bytes.NewReader(encoded)
	_ = rlp.Decode(reader, &accountNonce)
	_ = rlp.Decode(reader, &price)
	_ = rlp.Decode(reader, &gasLimit)
	_ = rlp.Decode(reader, &recipient)
	_ = rlp.Decode(reader, &amount)
	_ = rlp.Decode(reader, &from)
	_ = rlp.Decode(reader, &txType)
	_ = rlp.Decode(reader, &peggedData)
	_ = rlp.Decode(reader, &v)
	_ = rlp.Decode(reader, &r)
	_ = rlp.Decode(reader, &s)
	decoded, err := types.NewTxInternalDataWithMap(types.TxTypeChainDataPegging, txMapData{
		types.TxValueKeyNonce:      accountNonce,
		types.TxValueKeyAmount:     amount,
		types.TxValueKeyGasLimit:   gasLimit,
		types.TxValueKeyGasPrice:   price,
		types.TxValueKeyTo:         recipient,
		types.TxValueKeyFrom:       from,
		types.TxValueKeyPeggedData: peggedData,
	})
	if err != nil {
		b.Error(err)
	}
	decoded.SetVRS(v, r, s)
	return decoded, original.Equal(decoded)
}

// checkDecodingAccountCreationInterfaceOverFields is a subfunction to check if the encoding is correctly done or not
// for the account creation type of transaction internal data. It is only called when the benchmark verbose option is set.
func checkDecodingAccountCreationInterfaceOverFields(b *testing.B, encoded []byte, original types.TxInternalData) (types.TxInternalData, bool) {
	var accountNonce uint64
	var price *big.Int
	var gasLimit uint64
	var recipient common.Address
	var amount *big.Int
	var from common.Address
	var txType types.TxType
	var humanReadable bool
	var accountKeySerializer *types.AccountKeySerializer
	var v *big.Int
	var r *big.Int
	var s *big.Int
	reader := bytes.NewReader(encoded)
	_ = rlp.Decode(reader, &accountNonce)
	_ = rlp.Decode(reader, &price)
	_ = rlp.Decode(reader, &gasLimit)
	_ = rlp.Decode(reader, &recipient)
	_ = rlp.Decode(reader, &amount)
	_ = rlp.Decode(reader, &from)
	_ = rlp.Decode(reader, &txType)
	_ = rlp.Decode(reader, &humanReadable)
	_ = rlp.Decode(reader, &accountKeySerializer)
	_ = rlp.Decode(reader, &v)
	_ = rlp.Decode(reader, &r)
	_ = rlp.Decode(reader, &s)
	decoded, err := types.NewTxInternalDataWithMap(types.TxTypeAccountCreation, txMapData{
		types.TxValueKeyNonce:         accountNonce,
		types.TxValueKeyAmount:        amount,
		types.TxValueKeyGasLimit:      gasLimit,
		types.TxValueKeyGasPrice:      price,
		types.TxValueKeyTo:            recipient,
		types.TxValueKeyFrom:          from,
		types.TxValueKeyHumanReadable: humanReadable,
		types.TxValueKeyAccountKey:    accountKeySerializer.GetKey(),
	})
	if err != nil {
		b.Error(err)
	}
	decoded.SetVRS(v, r, s)
	return decoded, original.Equal(decoded)
}

// checkInterfaceOverFieldsDecoding is to check if a transaction internal data is encoded correctly or not.
// This function is only called when the benchmark verbose option is set.
func checkInterfaceOverFieldsDecoding(b *testing.B, txType types.TxType, encoded []byte, original types.TxInternalData) {
	var isSuccessful bool
	var decoded types.TxInternalData
	switch txType {
	case types.TxTypeLegacyTransaction:
		decoded, isSuccessful = checkDecodingLegacyInterfaceOverFields(b, encoded, original)
	case types.TxTypeValueTransfer:
		decoded, isSuccessful = checkDecodingValueTransferInterfaceOverFields(b, encoded, original)
	case types.TxTypeChainDataPegging:
		decoded, isSuccessful = checkDecodingChainDataPeggingInterfaceOverFields(b, encoded, original)
	case types.TxTypeAccountCreation:
		decoded, isSuccessful = checkDecodingAccountCreationInterfaceOverFields(b, encoded, original)
	}
	if isSuccessful {
		b.Log("Decoding interface is successful")
	} else {
		b.Errorf("Decoding interface failed. %s", decoded.String())
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
		bs := buffer.Bytes()
		if testing.Verbose() {
			checkInterfaceOverFieldsDecoding(b, txType, bs, testTxData)
		}
	}
}

// genCommonDefaultData generates a common internal transaction data
func genCommonDefaultData() types.TxInternalDataCommon {
	// receipient is missing due to the different type between legacy and the other types
	return types.TxInternalDataCommon{
		AccountNonce: uint64(1234),
		Price:        new(big.Int).SetUint64(25),
		GasLimit:     uint64(9999999999),
		Amount:       new(big.Int).SetUint64(10),
		Hash:         &common.Hash{},
	}
}

// isSameTxInternalDataCommon returns the two operands are the same internal data common or not.
func isSameTxInternalDataCommon(original types.TxInternalDataCommon, decoded types.TxInternalDataCommon) bool {
	v1, r1, s1 := decoded.GetVRS()
	v2, r2, s2 := original.GetVRS()
	return decoded.GetAccountNonce() == original.GetAccountNonce() &&
		decoded.GetPrice().Cmp(original.GetPrice()) == 0 &&
		decoded.GetGasLimit() == original.GetGasLimit() &&
		decoded.GetFrom() == original.GetFrom() &&
		decoded.GetAmount().Cmp(original.GetAmount()) == 0 &&
		v1.Cmp(v2) == 0 && r1.Cmp(r2) == 0 && s1.Cmp(s2) == 0
}

// checkDecodingSeparateFieldsLegacy is a subfunction to check the encoding is correctly performed by decoding it
// in case of the legacy transaction internal data.
func checkDecodingSeparateFieldsLegacy(b *testing.B, encoded []byte, original types.TxInternalDataCommon, originalExtra []byte) {
	var commonData types.TxInternalDataCommon
	var extra []byte
	reader := bytes.NewReader(encoded)
	_ = rlp.Decode(reader, &commonData)
	_ = rlp.Decode(reader, &extra)

	if bytes.Compare(originalExtra, extra) == 0 && isSameTxInternalDataCommon(original, commonData) {
		b.Log("Decoding interface is successful")
	} else {
		b.Errorf("Decoding interface failed.")
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
		bs := buffer.Bytes()
		if testing.Verbose() {
			checkDecodingSeparateFieldsLegacy(b, bs, commonData, extra)
		}
	}
}

// checkDecodingSeparateFieldsValueTransfer is a subfunction to check if the encoding is correctly done by decoding it.
func checkDecodingSeparateFieldsValueTransfer(b *testing.B, encoded []byte, original types.TxInternalDataCommon, originalExtra common.Address) {
	var commonData types.TxInternalDataCommon
	var extra common.Address
	reader := bytes.NewReader(encoded)
	_ = rlp.Decode(reader, &commonData)
	_ = rlp.Decode(reader, &extra)

	if originalExtra == extra && isSameTxInternalDataCommon(original, commonData) {
		b.Log("Decoding interface is successful")
	} else {
		b.Errorf("Decoding interface failed.")
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
		bs := buffer.Bytes()
		if testing.Verbose() {
			checkDecodingSeparateFieldsValueTransfer(b, bs, commonData, extra)
		}
	}
}

// checkDecodingSeparateFieldsChainDataPegging is a subfunction to check if the encoding is correctly done or not
// in case of chain data pegging transaction internal data. It is only called when the benchmark verbose option is set.
func checkDecodingSeparateFieldsChainDataPegging(b *testing.B, encoded []byte, original types.TxInternalDataCommon, originalExtra []byte) {
	var commonData types.TxInternalDataCommon
	var extra []byte
	reader := bytes.NewReader(encoded)
	_ = rlp.Decode(reader, &commonData)
	_ = rlp.Decode(reader, &extra)

	if bytes.Compare(extra, originalExtra) == 0 && isSameTxInternalDataCommon(original, commonData) {
		b.Log("Decoding interface is successful")
	} else {
		b.Errorf("Decoding interface failed.")
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
		bs := buffer.Bytes()
		if testing.Verbose() {
			checkDecodingSeparateFieldsChainDataPegging(b, bs, commonData, extra)
		}
	}
}

// checkDecodingSeparateFieldsAccountCreation is to check if the encoding the account creation transaction internal data is correctly done or not.
// This function is only called when the benchmark verbose option is set.
func checkDecodingSeparateFieldsAccountCreation(b *testing.B, encoded []byte, original types.TxInternalDataCommon, originalHumanReadable bool, originalAccountKey *types.AccountKeyPublic) {
	var commonData types.TxInternalDataCommon
	var humanReadable bool
	var accountKeySerializer *types.AccountKeySerializer
	reader := bytes.NewReader(encoded)
	_ = rlp.Decode(reader, &commonData)
	_ = rlp.Decode(reader, &humanReadable)
	_ = rlp.Decode(reader, &accountKeySerializer)

	if originalHumanReadable == humanReadable &&
		originalAccountKey.Equal(accountKeySerializer.GetKey()) &&
		isSameTxInternalDataCommon(original, commonData) {
		b.Log("Decoding interface is successful")
	} else {
		b.Errorf("Decoding interface failed.")
	}
}

// encodeSeparateFieldsAccountCreation encodes both common data and extra field separately.
func encodeSeparateFieldsAccountCreation(b *testing.B) {
	commonData := genCommonDefaultData()
	k, _ := crypto.GenerateKey()
	accountKey := types.NewAccountKeyPublicWithValue(&k.PublicKey)
	accountKeySerializer := types.NewAccountKeySerializerWithAccountKey(accountKey)
	humanReadable := true
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, commonData); err != nil {
			b.Error(err)
		}
		if err := rlp.Encode(buffer, humanReadable); err != nil {
			b.Error(err)
		}
		if err := rlp.Encode(buffer, accountKeySerializer); err != nil {
			b.Error(err)
		}
		bs := buffer.Bytes()
		if testing.Verbose() {
			checkDecodingSeparateFieldsAccountCreation(b, bs, commonData, humanReadable, accountKey)
		}
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
