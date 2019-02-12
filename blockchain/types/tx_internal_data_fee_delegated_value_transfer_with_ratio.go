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

// TxInternalDataFeeDelegatedValueTransferWithRatio represents a fee-delegated value transfer transaction with a
// specified fee ratio between the sender and the fee payer.
// The ratio is a fee payer's ratio in percentage.
// For example, if it is 20, 20% of tx fee will be paid by the fee payer.
// 80% of tx fee will be paid by the sender.
type TxInternalDataFeeDelegatedValueTransferWithRatio struct {
	*TxInternalDataCommon
	FeeRatio uint8
	TxSignatures
	*TxInternalDataFeePayerCommon
}

func NewTxInternalDataFeeDelegatedValueTransferWithRatio() *TxInternalDataFeeDelegatedValueTransferWithRatio {
	return &TxInternalDataFeeDelegatedValueTransferWithRatio{
		newTxInternalDataCommon(),
		0,
		NewTxSignatures(),
		newTxInternalDataFeePayerCommon(),
	}
}

func NewTxInternalDataFeeDelegatedValueTransferWithRatioWithMap(values map[TxValueKeyType]interface{}) (*TxInternalDataFeeDelegatedValueTransferWithRatio, error) {
	c, err := newTxInternalDataCommonWithMap(values)
	if err != nil {
		return nil, err
	}
	f, err := newTxInternalDataFeePayerCommonWithMap(values)
	if err != nil {
		return nil, err
	}
	t := &TxInternalDataFeeDelegatedValueTransferWithRatio{c, 0, NewTxSignatures(), f}

	if v, ok := values[TxValueKeyFeeRatioOfFeePayer].(uint8); ok {
		t.FeeRatio = v
	}

	return t, nil
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) Type() TxType {
	return TxTypeFeeDelegatedValueTransferWithRatio
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) GetFeeRatio() uint8 {
	return t.FeeRatio
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) Equal(a TxInternalData) bool {
	ta, ok := a.(*TxInternalDataFeeDelegatedValueTransferWithRatio)
	if !ok {
		return false
	}

	return t.TxInternalDataCommon.equal(ta.TxInternalDataCommon) &&
		t.FeeRatio == ta.FeeRatio &&
		t.TxSignatures.equal(ta.TxSignatures) &&
		t.TxInternalDataFeePayerCommon.equal(ta.TxInternalDataFeePayerCommon)
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) SetSignature(s TxSignatures) {
	t.TxSignatures = s
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) String() string {
	ser := newTxInternalDataSerializerWithValues(t)
	tx := Transaction{data: t}
	enc, _ := rlp.EncodeToBytes(ser)
	return fmt.Sprintf(`
	TX(%x)
	Type:          %s%s
	Signature:     %s
	FeePayer:      %s
	FeeRatio:      %d
	Hex:           %x
`,
		tx.Hash(),
		t.Type().String(),
		t.TxInternalDataCommon.string(),
		t.TxSignatures.string(),
		t.TxInternalDataFeePayerCommon.string(),
		t.FeeRatio,
		enc)

}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) IntrinsicGas() (uint64, error) {
	return params.TxGas + params.TxGasFeeDelegatedWithRatio, nil
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) SerializeForSign() []interface{} {
	infs := []interface{}{t.Type()}
	infs = append(infs, t.TxInternalDataCommon.serializeForSign()...)

	return append(infs, t.FeeRatio)
}
