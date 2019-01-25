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
	"fmt"
	"github.com/ground-x/klaytn/common/math"
	"github.com/ground-x/klaytn/kerrors"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
)

// TxInternalDataChainDataPegging represents the transaction pegging child chain data.
type TxInternalDataChainDataPegging struct {
	*TxInternalDataCommon

	PeggedData []byte
}

func newTxInternalDataChainDataPegging() *TxInternalDataChainDataPegging {
	return &TxInternalDataChainDataPegging{newTxInternalDataCommon(), []byte{}}
}

func newTxInternalDataChainDataPeggingWithMap(values map[TxValueKeyType]interface{}) *TxInternalDataChainDataPegging {
	var peggedData []byte
	if v, ok := values[TxValueKeyPeggedData].([]byte); ok {
		peggedData = v
	}

	return &TxInternalDataChainDataPegging{newTxInternalDataCommonWithMap(values), peggedData}
}

func (t *TxInternalDataChainDataPegging) Type() TxType {
	return TxTypeChainDataPegging
}

func (t *TxInternalDataChainDataPegging) Equal(b TxInternalData) bool {
	tb, ok := b.(*TxInternalDataChainDataPegging)
	if !ok {
		return false
	}

	return t.TxInternalDataCommon.equal(tb.TxInternalDataCommon) && bytes.Equal(t.PeggedData, tb.PeggedData)
}

func (t *TxInternalDataChainDataPegging) String() string {
	ser := newTxInternalDataSerializerWithValues(t)
	enc, _ := rlp.EncodeToBytes(ser)
	dataPeggedRLP, _ := rlp.EncodeToBytes(t.PeggedData)

	return fmt.Sprintf(`%s
	Hex:      %x
	Type:     %s
	PeggedData:	%s
`,
		t.TxInternalDataCommon.string(),
		enc,
		t.Type().String(),
		dataPeggedRLP)
}

func (t *TxInternalDataChainDataPegging) SerializeForSign() []interface{} {
	infs := []interface{}{t.Type()}
	return append(infs,
		t.TxInternalDataCommon.serializeForSign(),
		t.PeggedData)
}

func (t *TxInternalDataChainDataPegging) IntrinsicGas() (uint64, error) {
	nByte := (uint64)(len(t.PeggedData))

	// Make sure we don't exceed uint64 for all data combinations
	if (math.MaxUint64-params.TxChainDataPeggingGas)/params.ChainDataPeggingGas < nByte {
		return 0, kerrors.ErrOutOfGas
	}

	return params.TxChainDataPeggingGas + params.ChainDataPeggingGas*nByte, nil
}
