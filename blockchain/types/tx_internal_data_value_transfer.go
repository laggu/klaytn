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

// TxInternalDataValueTransfer represents a transaction transferring KLAY.
// No more attributes required than attributes in TxInternalDataCommon.
type TxInternalDataValueTransfer struct {
	*TxInternalDataCommon
	*TxSignature
}

func newTxInternalDataValueTransfer() *TxInternalDataValueTransfer {
	return &TxInternalDataValueTransfer{newTxInternalDataCommon(), NewTxSignature()}
}

func newTxInternalDataValueTransferWithMap(values map[TxValueKeyType]interface{}) *TxInternalDataValueTransfer {
	return &TxInternalDataValueTransfer{newTxInternalDataCommonWithMap(values), NewTxSignature()}
}

func (t *TxInternalDataValueTransfer) Type() TxType {
	return TxTypeValueTransfer
}

func (t *TxInternalDataValueTransfer) Equal(b TxInternalData) bool {
	tb, ok := b.(*TxInternalDataValueTransfer)
	if !ok {
		return false
	}

	return t.TxInternalDataCommon.equal(tb.TxInternalDataCommon) &&
		t.TxSignature.equal(tb.TxSignature)
}

func (t *TxInternalDataValueTransfer) String() string {
	ser := newTxInternalDataSerializerWithValues(t)
	tx := Transaction{data: t}
	enc, _ := rlp.EncodeToBytes(ser)
	return fmt.Sprintf(`
	TX(%x)
	Type:          %s%s
	Signature:     %s
	Hex:           %x
`,
		tx.Hash(),
		t.Type().String(),
		t.TxInternalDataCommon.string(),
		t.TxSignature.string(),
		enc)
}

func (t *TxInternalDataValueTransfer) SetSignature(s *TxSignature) {
	t.TxSignature = s
}

func (t *TxInternalDataValueTransfer) IntrinsicGas() (uint64, error) {
	// TxInternalDataValueTransfer does not have payload, and it
	// is not account creation. Hence, its intrinsic gas is determined by
	// params.TxGas. Refer to types.IntrinsicGas().
	return params.TxGas, nil
}

func (t *TxInternalDataValueTransfer) SerializeForSign() []interface{} {
	infs := []interface{}{t.Type()}
	return append(infs, t.TxInternalDataCommon.serializeForSign()...)
}
