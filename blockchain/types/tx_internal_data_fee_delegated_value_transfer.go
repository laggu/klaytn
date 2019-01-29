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
	"fmt"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
)

// TxInternalDataFeeDelegatedValueTransfer represents a fee-delegated transaction
// transferring KLAY with a fee payer.
// FeePayer should be RLP-encoded after the signature of the sender.
type TxInternalDataFeeDelegatedValueTransfer struct {
	*TxInternalDataCommon
	*TxSignature
	*TxInternalDataFeePayerCommon
}

func NewTxInternalDataFeeDelegatedValueTransfer() *TxInternalDataFeeDelegatedValueTransfer {
	return &TxInternalDataFeeDelegatedValueTransfer{
		newTxInternalDataCommon(),
		NewTxSignature(),
		newTxInternalDataFeePayerCommon(),
	}
}

func NewTxInternalDataFeeDelegatedValueTransferWithMap(values map[TxValueKeyType]interface{}) *TxInternalDataFeeDelegatedValueTransfer {
	return &TxInternalDataFeeDelegatedValueTransfer{
		newTxInternalDataCommonWithMap(values),
		NewTxSignature(),
		newTxInternalDataFeePayerCommonWithMap(values),
	}
}

func (t *TxInternalDataFeeDelegatedValueTransfer) Type() TxType {
	return TxTypeFeeDelegatedValueTransfer
}

func (t *TxInternalDataFeeDelegatedValueTransfer) Equal(b TxInternalData) bool {
	tb, ok := b.(*TxInternalDataFeeDelegatedValueTransfer)
	if !ok {
		return false
	}

	return t.TxInternalDataCommon.equal(tb.TxInternalDataCommon) &&
		t.TxSignature.equal(tb.TxSignature) &&
		t.TxInternalDataFeePayerCommon.equal(tb.TxInternalDataFeePayerCommon)
}

func (t *TxInternalDataFeeDelegatedValueTransfer) SetSignature(s *TxSignature) {
	t.TxSignature = s
}

func (t *TxInternalDataFeeDelegatedValueTransfer) String() string {
	ser := newTxInternalDataSerializerWithValues(t)
	tx := Transaction{data: t}
	enc, _ := rlp.EncodeToBytes(ser)
	return fmt.Sprintf(`
	TX(%x)
	Type:          %s%s
	Signature:     %s
	FeePayer:      %s
	Hex:           %x
`,
		tx.Hash(),
		t.Type().String(),
		t.TxInternalDataCommon.string(),
		t.TxSignature.string(),
		t.TxInternalDataFeePayerCommon.string(),
		enc)
}

func (t *TxInternalDataFeeDelegatedValueTransfer) IntrinsicGas() (uint64, error) {
	return params.TxGas + params.TxGasFeeDelegated, nil
}

func (t *TxInternalDataFeeDelegatedValueTransfer) SerializeForSign() []interface{} {
	infs := []interface{}{t.Type()}
	return append(infs, t.TxInternalDataCommon.serializeForSign())
}
