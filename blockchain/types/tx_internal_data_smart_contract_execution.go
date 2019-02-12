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
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
)

// TxInternalDataSmartContractExecution represents a transaction executing a smart contract.
type TxInternalDataSmartContractExecution struct {
	*TxInternalDataCommon

	Payload []byte

	TxSignatures
}

func newTxInternalDataSmartContractExecution() *TxInternalDataSmartContractExecution {
	return &TxInternalDataSmartContractExecution{
		newTxInternalDataCommon(),
		[]byte{},
		NewTxSignatures(),
	}
}

func newTxInternalDataSmartContractExecutionWithMap(values map[TxValueKeyType]interface{}) (*TxInternalDataSmartContractExecution, error) {
	c, err := newTxInternalDataCommonWithMap(values)
	if err != nil {
		return nil, err
	}

	t := &TxInternalDataSmartContractExecution{
		c, []byte{}, NewTxSignatures(),
	}

	if v, ok := values[TxValueKeyData].([]byte); ok {
		t.Payload = v
	} else {
		return nil, errValueKeyDataMustByteSlice
	}

	return t, nil
}

func (t *TxInternalDataSmartContractExecution) Type() TxType {
	return TxTypeSmartContractExecution
}

func (t *TxInternalDataSmartContractExecution) GetPayload() []byte {
	return t.Payload
}

func (t *TxInternalDataSmartContractExecution) Equal(a TxInternalData) bool {
	ta, ok := a.(*TxInternalDataSmartContractExecution)
	if !ok {
		return false
	}

	return t.TxInternalDataCommon.equal(ta.TxInternalDataCommon) &&
		bytes.Equal(t.Payload, ta.Payload) &&
		t.TxSignatures.equal(ta.TxSignatures)
}

func (t *TxInternalDataSmartContractExecution) SetSignature(s TxSignatures) {
	t.TxSignatures = s
}

func (t *TxInternalDataSmartContractExecution) String() string {
	ser := newTxInternalDataSerializerWithValues(t)
	tx := Transaction{data: t}
	enc, _ := rlp.EncodeToBytes(ser)
	return fmt.Sprintf(`
	TX(%x)
	Type:          %s%s
	Signature:     %s
	Paylod:        %x
	Hex:           %x
`,
		tx.Hash(),
		t.Type().String(),
		t.TxInternalDataCommon.string(),
		t.TxSignatures.string(),
		common.Bytes2Hex(t.Payload),
		enc)
}

func (t *TxInternalDataSmartContractExecution) IntrinsicGas() (uint64, error) {
	gas := params.TxGas

	gasPayload, err := intrinsicGasPayload(t.Payload)
	if err != nil {
		return 0, err
	}

	return gas + gasPayload, nil
}

func (t *TxInternalDataSmartContractExecution) SerializeForSign() []interface{} {
	infs := []interface{}{t.Type()}
	infs = append(infs, t.TxInternalDataCommon.serializeForSign()...)

	return append(infs, t.Payload)
}
