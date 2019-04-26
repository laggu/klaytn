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
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/crypto/sha3"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

// TestTransactionSenderTxHash tests SenderTxHash() of all transaction types.
func TestTransactionSenderTxHash(t *testing.T) {
	var txs = []struct {
		Name string
		tx   TxInternalData
	}{
		{"OriginalTx", genLegacyTransaction()},
		{"SmartContractDeploy", genSmartContractDeployTransaction()},
		{"FeeDelegatedSmartContractDeploy", genFeeDelegatedSmartContractDeployTransaction()},
		{"FeeDelegatedSmartContractDeployWithRatio", genFeeDelegatedSmartContractDeployWithRatioTransaction()},
		{"ValueTransfer", genValueTransferTransaction()},
		{"ValueTransferMemo", genValueTransferMemoTransaction()},
		{"FeeDelegatedValueTransferMemo", genFeeDelegatedValueTransferMemoTransaction()},
		{"FeeDelegatedValueTransferMemoWithRatio", genFeeDelegatedValueTransferMemoWithRatioTransaction()},
		{"ChainDataTx", genChainDataTransaction()},
		{"AccountCreation", genAccountCreationTransaction()},
		{"AccountUpdate", genAccountUpdateTransaction()},
		{"FeeDelegatedAccountUpdate", genFeeDelegatedAccountUpdateTransaction()},
		{"FeeDelegatedAccountUpdateWithRatio", genFeeDelegatedAccountUpdateWithRatioTransaction()},
		{"FeeDelegatedValueTransfer", genFeeDelegatedValueTransferTransaction()},
		{"SmartContractExecution", genSmartContractExecutionTransaction()},
		{"FeeDelegatedSmartContractExecution", genFeeDelegatedSmartContractExecutionTransaction()},
		{"FeeDelegatedSmartContractExecutionWithRatio", genFeeDelegatedSmartContractExecutionWithRatioTransaction()},
		{"FeeDelegatedValueTransferWithRatio", genFeeDelegatedValueTransferWithRatioTransaction()},
		{"Cancel", genCancelTransaction()},
		{"FeeDelegatedCancel", genFeeDelegatedCancelTransaction()},
		{"FeeDelegatedCancelWithRatio", genFeeDelegatedCancelWithRatioTransaction()},
	}

	var testcases = []struct {
		Name string
		fn   func(t *testing.T, tx TxInternalData)
	}{
		{"SenderTxHash", testTransactionSenderTxHash},
	}

	txMap := make(map[TxType]TxInternalData)
	for _, test := range testcases {
		for _, tx := range txs {
			txMap[tx.tx.Type()] = tx.tx
			Name := test.Name + "/" + tx.Name
			t.Run(Name, func(t *testing.T) {
				test.fn(t, tx.tx)
			})
		}
	}

	// Below code checks whether serialization for all tx implementations is done or not.
	// If no serialization, make test failed.
	for i := TxTypeLegacyTransaction; i < TxTypeLast; i++ {
		tx, err := NewTxInternalData(i)
		if err == nil {
			if _, ok := txMap[tx.Type()]; !ok {
				t.Errorf("No serialization test for tx %s", tx.Type().String())
			}
		}
	}
}

func testTransactionSenderTxHash(t *testing.T, tx TxInternalData) {
	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	if _, ok := tx.(TxInternalDataFeePayer); ok {
		rawTx.SignFeePayer(signer, key)
	}

	switch v := rawTx.data.(type) {
	case *TxInternalDataLegacy:
		assert.Equal(t, rawTx.Hash(), rawTx.SenderTxHash())

	case *TxInternalDataValueTransfer:
		assert.Equal(t, rawTx.Hash(), rawTx.SenderTxHash())

	case *TxInternalDataFeeDelegatedValueTransfer:
		hw := sha3.NewKeccak256()
		rlp.Encode(hw, rawTx.Type())
		rlp.Encode(hw, []interface{}{
			v.AccountNonce,
			v.Price,
			v.GasLimit,
			v.Recipient,
			v.Amount,
			v.From,
			v.TxSignatures,
		})

		h := common.Hash{}

		hw.Sum(h[:0])
		assert.Equal(t, h, rawTx.SenderTxHash())

	case *TxInternalDataFeeDelegatedValueTransferWithRatio:
		hw := sha3.NewKeccak256()
		rlp.Encode(hw, rawTx.Type())
		rlp.Encode(hw, []interface{}{
			v.AccountNonce,
			v.Price,
			v.GasLimit,
			v.Recipient,
			v.Amount,
			v.From,
			v.FeeRatio,
			v.TxSignatures,
		})

		h := common.Hash{}

		hw.Sum(h[:0])
		assert.Equal(t, h, rawTx.SenderTxHash())

	case *TxInternalDataValueTransferMemo:
		assert.Equal(t, rawTx.Hash(), rawTx.SenderTxHash())

	case *TxInternalDataFeeDelegatedValueTransferMemo:
		hw := sha3.NewKeccak256()
		rlp.Encode(hw, rawTx.Type())
		rlp.Encode(hw, []interface{}{
			v.AccountNonce,
			v.Price,
			v.GasLimit,
			v.Recipient,
			v.Amount,
			v.From,
			v.Payload,
			v.TxSignatures,
		})

		h := common.Hash{}

		hw.Sum(h[:0])
		assert.Equal(t, h, rawTx.SenderTxHash())

	case *TxInternalDataFeeDelegatedValueTransferMemoWithRatio:
		hw := sha3.NewKeccak256()
		rlp.Encode(hw, rawTx.Type())
		rlp.Encode(hw, []interface{}{
			v.AccountNonce,
			v.Price,
			v.GasLimit,
			v.Recipient,
			v.Amount,
			v.From,
			v.Payload,
			v.FeeRatio,
			v.TxSignatures,
		})

		h := common.Hash{}

		hw.Sum(h[:0])
		assert.Equal(t, h, rawTx.SenderTxHash())

	case *TxInternalDataAccountCreation:
		assert.Equal(t, rawTx.Hash(), rawTx.SenderTxHash())

	case *TxInternalDataAccountUpdate:
		assert.Equal(t, rawTx.Hash(), rawTx.SenderTxHash())

	case *TxInternalDataFeeDelegatedAccountUpdate:
		serializer := accountkey.NewAccountKeySerializerWithAccountKey(v.Key)
		keyEnc, _ := rlp.EncodeToBytes(serializer)
		hw := sha3.NewKeccak256()
		rlp.Encode(hw, rawTx.Type())
		rlp.Encode(hw, []interface{}{
			v.AccountNonce,
			v.Price,
			v.GasLimit,
			v.From,
			keyEnc,
			v.TxSignatures,
		})

		h := common.Hash{}

		hw.Sum(h[:0])
		assert.Equal(t, h, rawTx.SenderTxHash())

	case *TxInternalDataFeeDelegatedAccountUpdateWithRatio:
		serializer := accountkey.NewAccountKeySerializerWithAccountKey(v.Key)
		keyEnc, _ := rlp.EncodeToBytes(serializer)
		hw := sha3.NewKeccak256()
		rlp.Encode(hw, rawTx.Type())
		rlp.Encode(hw, []interface{}{
			v.AccountNonce,
			v.Price,
			v.GasLimit,
			v.From,
			keyEnc,
			v.FeeRatio,
			v.TxSignatures,
		})

		h := common.Hash{}

		hw.Sum(h[:0])
		assert.Equal(t, h, rawTx.SenderTxHash())

	case *TxInternalDataSmartContractDeploy:
		assert.Equal(t, rawTx.Hash(), rawTx.SenderTxHash())

	case *TxInternalDataFeeDelegatedSmartContractDeploy:
		hw := sha3.NewKeccak256()
		rlp.Encode(hw, rawTx.Type())
		rlp.Encode(hw, []interface{}{
			v.AccountNonce,
			v.Price,
			v.GasLimit,
			v.Recipient,
			v.Amount,
			v.From,
			v.Payload,
			v.HumanReadable,
			v.CodeFormat,
			v.TxSignatures,
		})

		h := common.Hash{}

		hw.Sum(h[:0])
		assert.Equal(t, h, rawTx.SenderTxHash())

	case *TxInternalDataFeeDelegatedSmartContractDeployWithRatio:
		hw := sha3.NewKeccak256()
		rlp.Encode(hw, rawTx.Type())
		rlp.Encode(hw, []interface{}{
			v.AccountNonce,
			v.Price,
			v.GasLimit,
			v.Recipient,
			v.Amount,
			v.From,
			v.Payload,
			v.HumanReadable,
			v.FeeRatio,
			v.CodeFormat,
			v.TxSignatures,
		})

		h := common.Hash{}

		hw.Sum(h[:0])
		assert.Equal(t, h, rawTx.SenderTxHash())

	case *TxInternalDataSmartContractExecution:
		assert.Equal(t, rawTx.Hash(), rawTx.SenderTxHash())

	case *TxInternalDataFeeDelegatedSmartContractExecution:
		hw := sha3.NewKeccak256()
		rlp.Encode(hw, rawTx.Type())
		rlp.Encode(hw, []interface{}{
			v.AccountNonce,
			v.Price,
			v.GasLimit,
			v.Recipient,
			v.Amount,
			v.From,
			v.Payload,
			v.TxSignatures,
		})

		h := common.Hash{}

		hw.Sum(h[:0])
		assert.Equal(t, h, rawTx.SenderTxHash())

	case *TxInternalDataFeeDelegatedSmartContractExecutionWithRatio:
		hw := sha3.NewKeccak256()
		rlp.Encode(hw, rawTx.Type())
		rlp.Encode(hw, []interface{}{
			v.AccountNonce,
			v.Price,
			v.GasLimit,
			v.Recipient,
			v.Amount,
			v.From,
			v.Payload,
			v.FeeRatio,
			v.TxSignatures,
		})

		h := common.Hash{}

		hw.Sum(h[:0])
		assert.Equal(t, h, rawTx.SenderTxHash())

	case *TxInternalDataCancel:
		assert.Equal(t, rawTx.Hash(), rawTx.SenderTxHash())

	case *TxInternalDataFeeDelegatedCancel:
		hw := sha3.NewKeccak256()
		rlp.Encode(hw, rawTx.Type())
		rlp.Encode(hw, []interface{}{
			v.AccountNonce,
			v.Price,
			v.GasLimit,
			v.From,
			v.TxSignatures,
		})

		h := common.Hash{}

		hw.Sum(h[:0])
		assert.Equal(t, h, rawTx.SenderTxHash())

	case *TxInternalDataFeeDelegatedCancelWithRatio:
		hw := sha3.NewKeccak256()
		rlp.Encode(hw, rawTx.Type())
		rlp.Encode(hw, []interface{}{
			v.AccountNonce,
			v.Price,
			v.GasLimit,
			v.From,
			v.FeeRatio,
			v.TxSignatures,
		})

		h := common.Hash{}

		hw.Sum(h[:0])
		assert.Equal(t, h, rawTx.SenderTxHash())

	case *TxInternalDataChainDataAnchoring:
		assert.Equal(t, rawTx.Hash(), rawTx.SenderTxHash())

	default:
		t.Fatal("Undefined tx type.")
	}
}
