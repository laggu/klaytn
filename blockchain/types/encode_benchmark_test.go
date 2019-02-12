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
	"bytes"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
	"testing"
)

type txMapData map[TxValueKeyType]interface{}

// getTestAnchoredData returns a sample anchored data.
func genTestAnchoredData() []byte {
	blockTxData := &ChainHashes{
		BlockHash:     common.HexToHash("0"),
		TxHash:        common.HexToHash("1"),
		ParentHash:    common.HexToHash("2"),
		ReceiptHash:   common.HexToHash("3"),
		StateRootHash: common.HexToHash("4"),
		BlockNumber:   big.NewInt(5),
	}
	anchoredData, err := rlp.EncodeToBytes(blockTxData)
	if err != nil {
		panic(err)
	}
	return anchoredData
}

// checkDecoding is called when the benchmark is verbose mode to check if the encoded is performed correctly by decoding it.
func checkDecoding(b *testing.B, txType TxType, original TxInternalData, encoded []byte) {
	container, err := NewTxInternalData(txType)
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
func benchmarkEncode(b *testing.B, testTxData TxInternalData) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, testTxData); err != nil {
			b.Error(err)
		}
		bs := buffer.Bytes()
		if testing.Verbose() {
			checkDecoding(b, testTxData.Type(), testTxData, bs)
		}
	}
}

// benchmarkEncodeToBytes is an auxiliary function to do encode internal data by `rlp.EncodeToBytes`.
func benchmarkEncodeToBytes(b *testing.B, testTxData TxInternalData) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if bs, err := rlp.EncodeToBytes(testTxData); err != nil {
			b.Error(err)
		} else if testing.Verbose() {
			checkDecoding(b, testTxData.Type(), testTxData, bs)
		}
	}
}

// checkDecodingLegacyTxInterface checks if the encoding legacy transaction internal data in interface list type is correctly
// encoded by decoding it. It is called in verbose mode.
func checkDecodingLegacyTxInterface(b *testing.B, encoded []byte, original TxInternalData) (TxInternalData, bool) {
	var container struct {
		AccountNonce uint64
		Price        *big.Int
		GasLimit     uint64
		Recipient    common.Address
		Amount       *big.Int
		Payload      []byte
		V            *big.Int
		R            *big.Int
		S            *big.Int
	}
	if err := rlp.DecodeBytes(encoded, &container); err != nil {
		b.Error(err)
	}
	decoded, err := NewTxInternalDataWithMap(TxTypeLegacyTransaction,
		txMapData{
			TxValueKeyNonce:    container.AccountNonce,
			TxValueKeyAmount:   container.Amount,
			TxValueKeyGasLimit: container.GasLimit,
			TxValueKeyGasPrice: container.Price,
			TxValueKeyData:     container.Payload,
			TxValueKeyTo:       container.Recipient,
		})
	if err != nil {
		b.Error(err)
	}
	decoded.SetSignature(TxSignatures{&TxSignature{container.V, container.R, container.S}})
	return decoded, original.Equal(decoded)
}

// checkDecodingValueTransferTxInterface checks if the encoding of transaction internal data in value transfer type
// is done correctly by decoding it. It is called when the benchmark test is verbose mode.
func checkDecodingValueTransferTxInterface(b *testing.B, encoded []byte, original TxInternalData) (TxInternalData, bool) {
	var container struct {
		Type         TxType
		AccountNonce uint64
		Price        *big.Int
		GasLimit     uint64
		Recipient    common.Address
		Amount       *big.Int
		From         common.Address
		V            *big.Int
		R            *big.Int
		S            *big.Int
	}
	if err := rlp.DecodeBytes(encoded, &container); err != nil {
		b.Error(err)
	}

	decoded, err := NewTxInternalDataWithMap(TxTypeValueTransfer,
		txMapData{
			TxValueKeyNonce:    container.AccountNonce,
			TxValueKeyAmount:   container.Amount,
			TxValueKeyGasLimit: container.GasLimit,
			TxValueKeyGasPrice: container.Price,
			TxValueKeyFrom:     container.From,
			TxValueKeyTo:       container.Recipient,
		})
	if err != nil {
		b.Error(err)
	}
	decoded.SetSignature(TxSignatures{&TxSignature{container.V, container.R, container.S}})
	return decoded, original.Equal(decoded)
}

// checkDecodingChainDataPeggingTxInterface checks if the encoding is correctly done or not in case of chain data pegging type
// transaction internal data in interface list. It is called when the benchmark is verbose mode.
func checkDecodingChainDataPeggingTxInterface(b *testing.B, encoded []byte, original TxInternalData) (TxInternalData, bool) {
	var container struct {
		Type         TxType
		AccountNonce uint64
		Price        *big.Int
		GasLimit     uint64
		Recipient    common.Address
		Amount       *big.Int
		From         common.Address
		PeggedData   []byte
		V            *big.Int
		R            *big.Int
		S            *big.Int
	}
	if err := rlp.DecodeBytes(encoded, &container); err != nil {
		b.Error(err)
	}
	decoded, err := NewTxInternalDataWithMap(TxTypeChainDataAnchoring,
		txMapData{
			TxValueKeyNonce:        container.AccountNonce,
			TxValueKeyAmount:       container.Amount,
			TxValueKeyGasLimit:     container.GasLimit,
			TxValueKeyGasPrice:     container.Price,
			TxValueKeyFrom:         container.From,
			TxValueKeyTo:           container.Recipient,
			TxValueKeyAnchoredData: container.PeggedData,
		})
	if err != nil {
		b.Error(err)
	}
	decoded.SetSignature(TxSignatures{&TxSignature{container.V, container.R, container.S}})
	return decoded, original.Equal(decoded)
}

// checkDecodingAccountCreationTxInterface checks if the encoding is correctly done in case of tx internal data interface list
// of account creation type. It is called when the benchmark is verbose mode.
func checkDecodingAccountCreationTxInterface(b *testing.B, encoded []byte, original TxInternalData) (TxInternalData, bool) {
	var container struct {
		Type              TxType
		AccountNonce      uint64
		Price             *big.Int
		GasLimit          uint64
		Recipient         common.Address
		Amount            *big.Int
		From              common.Address
		HumalReadable     bool
		V                 *big.Int
		R                 *big.Int
		S                 *big.Int
		EncodedAccountKey []byte
	}
	if err := rlp.DecodeBytes(encoded, &container); err != nil {
		b.Error(err)
	}
	var serializer *AccountKeySerializer
	if err := rlp.DecodeBytes(container.EncodedAccountKey, &serializer); err != nil {
		b.Error(err)
	}
	decoded, err := NewTxInternalDataWithMap(TxTypeAccountCreation,
		txMapData{
			TxValueKeyNonce:         container.AccountNonce,
			TxValueKeyAmount:        container.Amount,
			TxValueKeyGasLimit:      container.GasLimit,
			TxValueKeyGasPrice:      container.Price,
			TxValueKeyFrom:          container.From,
			TxValueKeyTo:            container.Recipient,
			TxValueKeyHumanReadable: container.HumalReadable,
			TxValueKeyAccountKey:    serializer.GetKey(),
		})
	if err != nil {
		b.Error(err)
	}
	decoded.SetSignature(TxSignatures{&TxSignature{container.V, container.R, container.S}})
	return decoded, original.Equal(decoded)
}

// checkInterfaceDecoding is to check if the transaction internal data is correctly encoded or not by decoding it.
// This function is called only when the benchmark is verbose mode.
func checkInterfaceDecoding(b *testing.B, txType TxType, encoded []byte, original TxInternalData) {
	var isSuccessful bool
	var decoded TxInternalData
	switch txType {
	case TxTypeLegacyTransaction:
		decoded, isSuccessful = checkDecodingLegacyTxInterface(b, encoded, original)
	case TxTypeValueTransfer:
		decoded, isSuccessful = checkDecodingValueTransferTxInterface(b, encoded, original)
	case TxTypeChainDataAnchoring:
		decoded, isSuccessful = checkDecodingChainDataPeggingTxInterface(b, encoded, original)
	case TxTypeAccountCreation:
		decoded, isSuccessful = checkDecodingAccountCreationTxInterface(b, encoded, original)
	}
	if isSuccessful {
		b.Log("Decoding interface is successful")
	} else {
		b.Errorf("Decoding interface failed. %s", decoded.String())
	}
}

// benchmarkEncodeInterfaceAccountCreation is an auxiliary function to do encoding for AccountCreation type transaction
func benchmarkEncodeInterfaceAccountCreation(b *testing.B, testTxData *TxInternalDataAccountCreation) {
	txInterfaces := testTxData.SerializeForSign()
	txInterfaces = txInterfaces[:len(txInterfaces)-1]
	sigs := testTxData.RawSignatureValues()
	v, r, s := sigs[0], sigs[1], sigs[2]
	txInterfaces = append(txInterfaces, v, r, s)
	var keyEnc []byte
	txInterfaces = append(txInterfaces, keyEnc)
	size := len(txInterfaces)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serializer := NewAccountKeySerializerWithAccountKey(testTxData.Key)
		keyEnc, _ = rlp.EncodeToBytes(serializer)
		txInterfaces[size-1] = keyEnc
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, txInterfaces); err != nil {
			b.Error(err)
		}
		bs := buffer.Bytes()
		if testing.Verbose() {
			checkInterfaceDecoding(b, testTxData.Type(), bs, testTxData)
		}
	}
}

// benchmarkEncodeInterface is an auxiliary function to do encode interface of transaction internal data.
func benchmarkEncodeInterface(b *testing.B, testTxData TxInternalData) {
	if data, ok := testTxData.(*TxInternalDataAccountCreation); ok {
		benchmarkEncodeInterfaceAccountCreation(b, data)
		return
	}
	txInterfaces := testTxData.SerializeForSign()
	sigs := testTxData.RawSignatureValues()
	v, r, s := sigs[0], sigs[1], sigs[2]
	txInterfaces = append(txInterfaces, v, r, s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, txInterfaces); err != nil {
			b.Error(err)
		}
		bs := buffer.Bytes()
		if testing.Verbose() {
			checkInterfaceDecoding(b, testTxData.Type(), bs, testTxData)
		}
	}
}

// checkDecodingLegacyInterfaceOverFields is a subfunction to check the encoding is correctly done or not for legacy type.
// It is only called when the benchmark verbose option is set.
func checkDecodingLegacyInterfaceOverFields(b *testing.B, encoded []byte, original TxInternalData) (TxInternalData, bool) {
	var accountNonce uint64
	var price *big.Int
	var gasLimit uint64
	var recipient common.Address
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
	decoded, err := NewTxInternalDataWithMap(TxTypeLegacyTransaction, txMapData{
		TxValueKeyNonce:    accountNonce,
		TxValueKeyGasPrice: price,
		TxValueKeyGasLimit: gasLimit,
		TxValueKeyTo:       recipient,
		TxValueKeyAmount:   amount,
		TxValueKeyData:     payload,
	})
	if err != nil {
		b.Error(err)
	}
	decoded.SetSignature(TxSignatures{&TxSignature{v, r, s}})
	return decoded, original.Equal(decoded)
}

// checkDecodingValueTransferInterfaceOverFields is a subfunction to check the encoding is correctly done or not for
// value transfer type. It is only called when the benchmark verbose option is set.
func checkDecodingValueTransferInterfaceOverFields(b *testing.B, encoded []byte, original TxInternalData) (TxInternalData, bool) {
	var accountNonce uint64
	var price *big.Int
	var gasLimit uint64
	var recipient common.Address
	var amount *big.Int
	var from common.Address
	var txType TxType
	var v *big.Int
	var r *big.Int
	var s *big.Int
	reader := bytes.NewReader(encoded)
	_ = rlp.Decode(reader, &txType)
	_ = rlp.Decode(reader, &accountNonce)
	_ = rlp.Decode(reader, &price)
	_ = rlp.Decode(reader, &gasLimit)
	_ = rlp.Decode(reader, &recipient)
	_ = rlp.Decode(reader, &amount)
	_ = rlp.Decode(reader, &from)
	_ = rlp.Decode(reader, &v)
	_ = rlp.Decode(reader, &r)
	_ = rlp.Decode(reader, &s)
	decoded, err := NewTxInternalDataWithMap(TxTypeValueTransfer, txMapData{
		TxValueKeyNonce:    accountNonce,
		TxValueKeyAmount:   amount,
		TxValueKeyGasLimit: gasLimit,
		TxValueKeyGasPrice: price,
		TxValueKeyFrom:     from,
		TxValueKeyTo:       recipient,
	})
	if err != nil {
		b.Error(err)
	}
	decoded.SetSignature(TxSignatures{&TxSignature{v, r, s}})
	return decoded, original.Equal(decoded)
}

// checkDecodingChainDataPeggingInterfaceOverFields is a subfunction to check if the encoding is correctly performed
// for the data pegging transaction type. It is only called when the benchmark verbose option is set.
func checkDecodingChainDataPeggingInterfaceOverFields(b *testing.B, encoded []byte, original TxInternalData) (TxInternalData, bool) {
	var accountNonce uint64
	var price *big.Int
	var gasLimit uint64
	var recipient common.Address
	var amount *big.Int
	var from common.Address
	var txType TxType
	var peggedData []byte
	var v *big.Int
	var r *big.Int
	var s *big.Int
	reader := bytes.NewReader(encoded)
	_ = rlp.Decode(reader, &txType)
	_ = rlp.Decode(reader, &accountNonce)
	_ = rlp.Decode(reader, &price)
	_ = rlp.Decode(reader, &gasLimit)
	_ = rlp.Decode(reader, &recipient)
	_ = rlp.Decode(reader, &amount)
	_ = rlp.Decode(reader, &from)
	_ = rlp.Decode(reader, &peggedData)
	_ = rlp.Decode(reader, &v)
	_ = rlp.Decode(reader, &r)
	_ = rlp.Decode(reader, &s)
	decoded, err := NewTxInternalDataWithMap(TxTypeChainDataAnchoring, txMapData{
		TxValueKeyNonce:        accountNonce,
		TxValueKeyAmount:       amount,
		TxValueKeyGasLimit:     gasLimit,
		TxValueKeyGasPrice:     price,
		TxValueKeyTo:           recipient,
		TxValueKeyFrom:         from,
		TxValueKeyAnchoredData: peggedData,
	})
	if err != nil {
		b.Error(err)
	}
	decoded.SetSignature(TxSignatures{&TxSignature{v, r, s}})
	return decoded, original.Equal(decoded)
}

// checkDecodingAccountCreationInterfaceOverFields is a subfunction to check if the encoding is correctly done or not
// for the account creation type of transaction internal data. It is only called when the benchmark verbose option is set.
func checkDecodingAccountCreationInterfaceOverFields(b *testing.B, encoded []byte, original TxInternalData) (TxInternalData, bool) {
	var accountNonce uint64
	var price *big.Int
	var gasLimit uint64
	var recipient common.Address
	var amount *big.Int
	var from common.Address
	var txType TxType
	var humanReadable bool
	var encodedKeySerializer []byte
	var v *big.Int
	var r *big.Int
	var s *big.Int
	var accountKeySerializer *AccountKeySerializer
	reader := bytes.NewReader(encoded)
	_ = rlp.Decode(reader, &txType)
	_ = rlp.Decode(reader, &accountNonce)
	_ = rlp.Decode(reader, &price)
	_ = rlp.Decode(reader, &gasLimit)
	_ = rlp.Decode(reader, &recipient)
	_ = rlp.Decode(reader, &amount)
	_ = rlp.Decode(reader, &from)
	_ = rlp.Decode(reader, &humanReadable)
	_ = rlp.Decode(reader, &v)
	_ = rlp.Decode(reader, &r)
	_ = rlp.Decode(reader, &s)
	_ = rlp.Decode(reader, &encodedKeySerializer)
	_ = rlp.DecodeBytes(encodedKeySerializer, &accountKeySerializer)
	decoded, err := NewTxInternalDataWithMap(TxTypeAccountCreation, txMapData{
		TxValueKeyNonce:         accountNonce,
		TxValueKeyAmount:        amount,
		TxValueKeyGasLimit:      gasLimit,
		TxValueKeyGasPrice:      price,
		TxValueKeyTo:            recipient,
		TxValueKeyFrom:          from,
		TxValueKeyHumanReadable: humanReadable,
		TxValueKeyAccountKey:    accountKeySerializer.GetKey(),
	})
	if err != nil {
		b.Error(err)
	}
	decoded.SetSignature(TxSignatures{&TxSignature{v, r, s}})
	return decoded, original.Equal(decoded)
}

// checkInterfaceOverFieldsDecoding is to check if a transaction internal data is encoded correctly or not.
// This function is only called when the benchmark verbose option is set.
func checkInterfaceOverFieldsDecoding(b *testing.B, txType TxType, encoded []byte, original TxInternalData) {
	var isSuccessful bool
	var decoded TxInternalData
	switch txType {
	case TxTypeLegacyTransaction:
		decoded, isSuccessful = checkDecodingLegacyInterfaceOverFields(b, encoded, original)
	case TxTypeValueTransfer:
		decoded, isSuccessful = checkDecodingValueTransferInterfaceOverFields(b, encoded, original)
	case TxTypeChainDataAnchoring:
		decoded, isSuccessful = checkDecodingChainDataPeggingInterfaceOverFields(b, encoded, original)
	case TxTypeAccountCreation:
		decoded, isSuccessful = checkDecodingAccountCreationInterfaceOverFields(b, encoded, original)
	}
	if isSuccessful {
		b.Log("Decoding interface is successful")
	} else {
		b.Errorf("Decoding interface failed. %s", decoded.String())
	}
}

// benchmarkEncodeInterfaceOverFields is an auxiliary function to do encoding interface list of account creation transaction data
func benchmarkEncodeInterfaceOverFieldsAccountCreation(b *testing.B, testTxData *TxInternalDataAccountCreation) {
	txInterfaces := testTxData.SerializeForSign()
	txInterfaces = txInterfaces[:len(txInterfaces)-1]
	sigs := testTxData.RawSignatureValues()
	v, r, s := sigs[0], sigs[1], sigs[2]
	txInterfaces = append(txInterfaces, v, r, s)
	var keyEnc []byte
	txInterfaces = append(txInterfaces, keyEnc)
	size := len(txInterfaces)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serializer := NewAccountKeySerializerWithAccountKey(testTxData.Key)
		keyEnc, _ = rlp.EncodeToBytes(serializer)
		txInterfaces[size-1] = keyEnc
		buffer := new(bytes.Buffer)
		for _, field := range txInterfaces {
			if err := rlp.Encode(buffer, field); err != nil {
				b.Error(err)
			}
		}
		bs := buffer.Bytes()
		if testing.Verbose() {
			checkInterfaceOverFieldsDecoding(b, testTxData.Type(), bs, testTxData)
		}
	}
}

// benchmarkEncodeInterfaceOverFields is an auxiliary function to do encoding interface list of transaction internal data
func benchmarkEncodeInterfaceOverFields(b *testing.B, testTxData TxInternalData) {
	if data, ok := testTxData.(*TxInternalDataAccountCreation); ok {
		benchmarkEncodeInterfaceOverFieldsAccountCreation(b, data)
		return
	}
	txInterfaces := testTxData.SerializeForSign()
	sigs := testTxData.RawSignatureValues()
	v, r, s := sigs[0], sigs[1], sigs[2]
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
			checkInterfaceOverFieldsDecoding(b, testTxData.Type(), bs, testTxData)
		}
	}
}

// genCommonDefaultData generates a common internal transaction data
func genCommonDefaultData() TxInternalDataCommon {
	return TxInternalDataCommon{
		AccountNonce: uint64(1234),
		Price:        new(big.Int).SetUint64(25),
		Recipient:    common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b"),
		GasLimit:     uint64(9999999999),
		Amount:       new(big.Int).SetUint64(10),
		Hash:         &common.Hash{},
	}
}

// checkDecodingSeparateFieldsLegacy is a subfunction to check the encoding is correctly performed by decoding it
// in case of the legacy transaction internal data.
func checkDecodingSeparateFieldsLegacy(b *testing.B, encoded []byte, original TxInternalDataCommon, originalExtra []byte, originalSignature *TxSignature) {
	var commonData TxInternalDataCommon
	var extra []byte
	var signature *TxSignature
	reader := bytes.NewReader(encoded)
	_ = rlp.Decode(reader, &commonData)
	_ = rlp.Decode(reader, &extra)
	_ = rlp.Decode(reader, &signature)

	if bytes.Compare(originalExtra, extra) == 0 && original.equal(&commonData) && signature.equal(originalSignature) {
		b.Log("Decoding interface is successful")
	} else {
		b.Errorf("Decoding interface failed.")
	}
}

// encodeSeparateFieldsLegacy encodes both common data and extra field separately.
func encodeSeparateFieldsLegacy(b *testing.B) {
	commonData := genCommonDefaultData()
	extra := []byte("1234")
	signature := NewTxSignature()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, commonData); err != nil {
			b.Error(err)
		}
		if err := rlp.Encode(buffer, extra); err != nil {
			b.Error(err)
		}
		if err := rlp.Encode(buffer, signature); err != nil {
			b.Error(err)
		}
		bs := buffer.Bytes()
		if testing.Verbose() {
			checkDecodingSeparateFieldsLegacy(b, bs, commonData, extra, signature)
		}
	}
}

// checkDecodingSeparateFieldsValueTransfer is a subfunction to check if the encoding is correctly done by decoding it.
func checkDecodingSeparateFieldsValueTransfer(b *testing.B, encoded []byte, original TxInternalDataCommon, originalExtra common.Address, originalSignature *TxSignature) {
	var commonData TxInternalDataCommon
	var extra common.Address
	var signature *TxSignature
	reader := bytes.NewReader(encoded)
	_ = rlp.Decode(reader, &commonData)
	_ = rlp.Decode(reader, &extra)
	_ = rlp.Decode(reader, &signature)
	if originalExtra == extra && original.equal(&commonData) && originalSignature.equal(signature) {
		b.Log("Decoding interface is successful")
	} else {
		b.Errorf("Decoding interface failed.")
	}
}

// encodeSeparateFieldsValueTransfer encodes both common data and extra field separately.
func encodeSeparateFieldsValueTransfer(b *testing.B) {
	commonData := genCommonDefaultData()
	extra := common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	signature := NewTxSignature()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, commonData); err != nil {
			b.Error(err)
		}
		if err := rlp.Encode(buffer, extra); err != nil {
			b.Error(err)
		}
		if err := rlp.Encode(buffer, signature); err != nil {
			b.Error(err)
		}
		bs := buffer.Bytes()
		if testing.Verbose() {
			checkDecodingSeparateFieldsValueTransfer(b, bs, commonData, extra, signature)
		}
	}
}

// checkDecodingSeparateFieldsChainDataPegging is a subfunction to check if the encoding is correctly done or not
// in case of chain data pegging transaction internal data. It is only called when the benchmark verbose option is set.
func checkDecodingSeparateFieldsChainDataPegging(b *testing.B, encoded []byte, original TxInternalDataCommon, originalExtra []byte, originalSignature *TxSignature) {
	var commonData TxInternalDataCommon
	var extra []byte
	var signature *TxSignature
	reader := bytes.NewReader(encoded)
	_ = rlp.Decode(reader, &commonData)
	_ = rlp.Decode(reader, &extra)
	_ = rlp.Decode(reader, &signature)
	if bytes.Compare(extra, originalExtra) == 0 && original.equal(&commonData) && originalSignature.equal(signature) {
		b.Log("Decoding interface is successful")
	} else {
		b.Errorf("Decoding interface failed.")
	}
}

// encodeSeparateFieldsChainDataAnchoring encodes both common data and extra field separately.
func encodeSeparateFieldsChainDataAnchoring(b *testing.B) {
	commonData := genCommonDefaultData()
	extra := genTestAnchoredData()
	signature := NewTxSignature()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, commonData); err != nil {
			b.Error(err)
		}
		if err := rlp.Encode(buffer, extra); err != nil {
			b.Error(err)
		}
		if err := rlp.Encode(buffer, signature); err != nil {
			b.Error(err)
		}
		bs := buffer.Bytes()
		if testing.Verbose() {
			checkDecodingSeparateFieldsChainDataPegging(b, bs, commonData, extra, signature)
		}
	}
}

// checkDecodingSeparateFieldsAccountCreation is to check if the encoding the account creation transaction internal data is correctly done or not.
// This function is only called when the benchmark verbose option is set.
func checkDecodingSeparateFieldsAccountCreation(b *testing.B, encoded []byte, original TxInternalDataCommon, originalHumanReadable bool, originalAccountKey *AccountKeyPublic, originalSignature *TxSignature) {
	var commonData TxInternalDataCommon
	var humanReadable bool
	var accountKeySerializer *AccountKeySerializer
	var signature *TxSignature
	reader := bytes.NewReader(encoded)
	_ = rlp.Decode(reader, &commonData)
	_ = rlp.Decode(reader, &humanReadable)
	_ = rlp.Decode(reader, &accountKeySerializer)
	_ = rlp.Decode(reader, &signature)
	if originalHumanReadable == humanReadable && originalAccountKey.Equal(accountKeySerializer.GetKey()) && original.equal(&commonData) && originalSignature.equal(signature) {
		b.Log("Decoding interface is successful")
	} else {
		b.Errorf("Decoding interface failed.")
	}
}

// encodeSeparateFieldsAccountCreation encodes both common data and extra field separately.
func encodeSeparateFieldsAccountCreation(b *testing.B) {
	commonData := genCommonDefaultData()
	k, _ := crypto.GenerateKey()
	humanReadable := true
	signature := NewTxSignature()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		accountKey := NewAccountKeyPublicWithValue(&k.PublicKey)
		accountKeySerializer := NewAccountKeySerializerWithAccountKey(accountKey)
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
		if err := rlp.Encode(buffer, signature); err != nil {
			b.Error(err)
		}
		bs := buffer.Bytes()
		if testing.Verbose() {
			checkDecodingSeparateFieldsAccountCreation(b, bs, commonData, humanReadable, accountKey, signature)
		}
	}
}

// benchmarkEncodeExtraSeparateFields is an auxiliary function to do encoding a common struct and an extra field separately.
func benchmarkEncodeExtraSeparateFields(b *testing.B, testTxData TxInternalData) {
	switch testTxData.Type() {
	case TxTypeLegacyTransaction:
		encodeSeparateFieldsLegacy(b)
	case TxTypeValueTransfer:
		encodeSeparateFieldsValueTransfer(b)
	case TxTypeChainDataAnchoring:
		encodeSeparateFieldsChainDataAnchoring(b)
	case TxTypeAccountCreation:
		encodeSeparateFieldsAccountCreation(b)
	}
}

// Main benchmark function
func BenchmarkRLPEncoding(b *testing.B) {
	var options = []struct {
		Name        string
		Subfunction func(*testing.B, TxInternalData)
	}{
		{"Encode", benchmarkEncode},
		{"EncodeToBytes", benchmarkEncodeToBytes},
		{"EncodeInterface", benchmarkEncodeInterface},
		{"EncodeInterfaceList", benchmarkEncodeInterfaceOverFields},
		{"EncodeExtraSeparateFields", benchmarkEncodeExtraSeparateFields},
	}
	var testMaterials = []struct {
		TypeName string
		tx       TxInternalData
	}{
		{"legacyTx", genLegacyTransaction()},
		{"valueTransferTx", genValueTransferTransaction()},
		{"chainDataAnchoringTx", genChainDataTransaction()},
		{"accountCreationTx", genAccountCreationTransaction()},
	}
	for _, option := range options {
		for _, data := range testMaterials {
			name := option.Name + "/" + data.TypeName
			b.Run(name, func(b *testing.B) {
				option.Subfunction(b, data.tx)
			})
		}
	}
}
