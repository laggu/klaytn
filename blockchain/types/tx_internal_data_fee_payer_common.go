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
	"github.com/ground-x/klaytn/common"
	"math/big"
)

// TxInternalDataFeePayerCommon is a common data structure for fee delegated transactions.
// This structure will be embedded into all fee delegated transaction types.
type TxInternalDataFeePayerCommon struct {
	feePayer common.Address
	*TxSignature
}

func newTxInternalDataFeePayerCommon() *TxInternalDataFeePayerCommon {
	return &TxInternalDataFeePayerCommon{
		common.Address{},
		NewTxSignature(),
	}
}

func newTxInternalDataFeePayerCommonWithMap(values map[TxValueKeyType]interface{}) *TxInternalDataFeePayerCommon {
	t := &TxInternalDataFeePayerCommon{}

	if v, ok := values[TxValueKeyFeePayer].(common.Address); ok {
		t.feePayer = v
	}

	return t
}

func (t *TxInternalDataFeePayerCommon) GetFeePayer() common.Address {
	return t.feePayer
}

func (t *TxInternalDataFeePayerCommon) SetFeePayerSignature(s *TxSignature) {
	t.TxSignature = s
}

func (t *TxInternalDataFeePayerCommon) GetFeePayerVRS() (*big.Int, *big.Int, *big.Int) {
	return t.TxSignature.GetVRS()
}

func (t *TxInternalDataFeePayerCommon) serializeForSign() []interface{} {
	return []interface{}{t.feePayer}
}

func (t *TxInternalDataFeePayerCommon) equal(tb *TxInternalDataFeePayerCommon) bool {
	return t.feePayer == tb.feePayer &&
		t.TxSignature.equal(tb.TxSignature)
}

func (t *TxInternalDataFeePayerCommon) string() string {
	return t.TxSignature.string()
}
