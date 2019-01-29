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
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/math"
	"github.com/ground-x/klaytn/kerrors"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
)

// TxInternalDataChainDataAnchoring represents the transaction anchoring child chain data.
type TxInternalDataChainDataAnchoring struct {
	*TxInternalDataCommon

	AnchoredData []byte

	*TxSignature
}

func newTxInternalDataChainDataAnchoring() *TxInternalDataChainDataAnchoring {
	return &TxInternalDataChainDataAnchoring{newTxInternalDataCommon(), []byte{},
		NewTxSignature()}
}

func newTxInternalDataChainDataAnchoringWithMap(values map[TxValueKeyType]interface{}) *TxInternalDataChainDataAnchoring {
	var anchoredData []byte
	if v, ok := values[TxValueKeyAnchoredData].([]byte); ok {
		anchoredData = v
	}

	return &TxInternalDataChainDataAnchoring{newTxInternalDataCommonWithMap(values), anchoredData, NewTxSignature()}
}

func (t *TxInternalDataChainDataAnchoring) Type() TxType {
	return TxTypeChainDataAnchoring
}

func (t *TxInternalDataChainDataAnchoring) Equal(b TxInternalData) bool {
	tb, ok := b.(*TxInternalDataChainDataAnchoring)
	if !ok {
		return false
	}

	return t.TxInternalDataCommon.equal(tb.TxInternalDataCommon) &&
		t.TxSignature.equal(tb.TxSignature) &&
		bytes.Equal(t.AnchoredData, tb.AnchoredData)
}

func (t *TxInternalDataChainDataAnchoring) String() string {
	ser := newTxInternalDataSerializerWithValues(t)
	enc, _ := rlp.EncodeToBytes(ser)
	dataAnchoredRLP, _ := rlp.EncodeToBytes(t.AnchoredData)
	tx := Transaction{data: t}

	return fmt.Sprintf(`
	TX(%x)
	Type:          %s%s
	Signature:     %s
	Hex:           %x
	AnchoredData:  %s
`,
		tx.Hash(),
		t.Type().String(),
		t.TxInternalDataCommon.string(),
		t.TxSignature.string(),
		enc,
		common.Bytes2Hex(dataAnchoredRLP))
}

func (t *TxInternalDataChainDataAnchoring) SerializeForSign() []interface{} {
	infs := []interface{}{t.Type()}
	infs = append(infs, t.TxInternalDataCommon.serializeForSign()...)

	return append(infs, t.AnchoredData)
}

func (t *TxInternalDataChainDataAnchoring) SetSignature(s *TxSignature) {
	t.TxSignature = s
}

func (t *TxInternalDataChainDataAnchoring) IntrinsicGas() (uint64, error) {
	nByte := (uint64)(len(t.AnchoredData))

	// Make sure we don't exceed uint64 for all data combinations
	if (math.MaxUint64-params.TxChainDataAnchoringGas)/params.ChainDataAnchoringGas < nByte {
		return 0, kerrors.ErrOutOfGas
	}

	return params.TxChainDataAnchoringGas + params.ChainDataAnchoringGas*nByte, nil
}
