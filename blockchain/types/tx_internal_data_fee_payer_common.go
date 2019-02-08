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
	"crypto/ecdsa"
	"fmt"
	"github.com/ground-x/klaytn/common"
	"math/big"
)

// TxInternalDataFeePayerCommon is a common data structure for fee delegated transactions.
// This structure will be embedded into all fee delegated transaction types.
type TxInternalDataFeePayerCommon struct {
	FeePayer common.Address
	*TxSignature
}

func newTxInternalDataFeePayerCommon() *TxInternalDataFeePayerCommon {
	return &TxInternalDataFeePayerCommon{
		common.Address{},
		NewTxSignature(),
	}
}

func newTxInternalDataFeePayerCommonWithMap(values map[TxValueKeyType]interface{}) (*TxInternalDataFeePayerCommon, error) {
	t := &TxInternalDataFeePayerCommon{
		common.Address{},
		NewTxSignature(),
	}

	if v, ok := values[TxValueKeyFeePayer].(common.Address); ok {
		t.FeePayer = v
	} else {
		return nil, errValueKeyFeePayerMustAddress
	}

	return t, nil
}

func (t *TxInternalDataFeePayerCommon) GetFeePayer() common.Address {
	return t.FeePayer
}

func (t *TxInternalDataFeePayerCommon) SetFeePayerSignature(s *TxSignature) {
	t.TxSignature = s
}

func (t *TxInternalDataFeePayerCommon) GetFeePayerRawSignatureValues() []*big.Int {
	return t.TxSignature.RawSignatureValues()
}

func (t *TxInternalDataFeePayerCommon) RecoverFeePayerPubkey(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) (*ecdsa.PublicKey, error) {
	return t.TxSignature.RecoverPubkey(txhash, homestead, vfunc)
}

func (t *TxInternalDataFeePayerCommon) serializeForSign() []interface{} {
	return []interface{}{t.FeePayer}
}

func (t *TxInternalDataFeePayerCommon) equal(tb *TxInternalDataFeePayerCommon) bool {
	return t.FeePayer == tb.FeePayer &&
		t.TxSignature.equal(tb.TxSignature)
}

func (t *TxInternalDataFeePayerCommon) string() string {
	return fmt.Sprintf("%s %s", t.FeePayer.String(), t.TxSignature.string())
}
