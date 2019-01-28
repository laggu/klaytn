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
	"encoding/json"
	"fmt"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"io"
)

// TxInternalDataAccountCreation represents a transaction creating an account.
type TxInternalDataAccountCreation struct {
	*TxInternalDataCommon

	HumanReadable bool
	Key           AccountKey

	*TxSignature
}

// txInternalDataAccountCreationSerializable for RLP serialization.
type txInternalDataAccountCreationSerializable struct {
	*TxInternalDataCommon

	HumanReadable bool
	KeyData       []byte

	*TxSignature
}

func newTxInternalDataAccountCreation() *TxInternalDataAccountCreation {
	return &TxInternalDataAccountCreation{
		TxInternalDataCommon: newTxInternalDataCommon(),
		HumanReadable:        false,
		Key:                  NewAccountKeyNil(),
	}
}

func newTxInternalDataAccountCreationWithMap(values map[TxValueKeyType]interface{}) *TxInternalDataAccountCreation {
	b := &TxInternalDataAccountCreation{
		TxInternalDataCommon: newTxInternalDataCommonWithMap(values),
		HumanReadable:        false,
		Key:                  NewAccountKeyNil(),
		TxSignature:          NewTxSignature(),
	}
	if v, ok := values[TxValueKeyHumanReadable].(bool); ok {
		b.HumanReadable = v
	}

	if v, ok := values[TxValueKeyAccountKey].(AccountKey); ok {
		b.Key = v
	}

	return b
}

func newTxInternalDataAccountCreationSerializable() *txInternalDataAccountCreationSerializable {
	return &txInternalDataAccountCreationSerializable{}
}

func (t *TxInternalDataAccountCreation) toSerializable() *txInternalDataAccountCreationSerializable {
	serializer := NewAccountKeySerializerWithAccountKey(t.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	return &txInternalDataAccountCreationSerializable{
		t.TxInternalDataCommon,
		t.HumanReadable,
		keyEnc,
		t.TxSignature,
	}
}

func (t *TxInternalDataAccountCreation) fromSerializable(serialized *txInternalDataAccountCreationSerializable) {
	t.TxInternalDataCommon = serialized.TxInternalDataCommon
	t.HumanReadable = serialized.HumanReadable
	t.TxSignature = serialized.TxSignature

	serializer := NewAccountKeySerializer()
	rlp.DecodeBytes(serialized.KeyData, serializer)
	t.Key = serializer.key
}

func (t *TxInternalDataAccountCreation) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, t.toSerializable())
}

func (t *TxInternalDataAccountCreation) DecodeRLP(s *rlp.Stream) error {
	dec := newTxInternalDataAccountCreationSerializable()

	if err := s.Decode(dec); err != nil {
		return err
	}
	t.fromSerializable(dec)

	return nil
}

func (t *TxInternalDataAccountCreation) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.toSerializable())
}

func (t *TxInternalDataAccountCreation) UnmarshalJSON(b []byte) error {
	dec := newTxInternalDataAccountCreationSerializable()

	if err := json.Unmarshal(b, dec); err != nil {
		return err
	}

	t.fromSerializable(dec)

	return nil
}

func (t *TxInternalDataAccountCreation) Type() TxType {
	return TxTypeAccountCreation
}

func (t *TxInternalDataAccountCreation) Equal(a TxInternalData) bool {
	ta, ok := a.(*TxInternalDataAccountCreation)
	if !ok {
		return false
	}

	return t.TxInternalDataCommon.equal(ta.TxInternalDataCommon) &&
		t.HumanReadable == ta.HumanReadable &&
		t.Key.Equal(ta.Key) &&
		t.TxSignature.equal(ta.TxSignature)
}

func (t *TxInternalDataAccountCreation) String() string {
	ser := newTxInternalDataSerializerWithValues(t)
	tx := Transaction{data: t}
	enc, _ := rlp.EncodeToBytes(ser)
	return fmt.Sprintf(`
	TX(%x)
	Type:          %s%s
	HumanReadable: %t
	Key:           %s
	Signature:     %s
	Hex:           %x
`,
		tx.Hash(),
		t.Type().String(),
		t.TxInternalDataCommon.string(),
		t.HumanReadable,
		t.Key.String(),
		t.TxSignature.string(),
		enc)
}

func (t *TxInternalDataAccountCreation) SetSignature(s *TxSignature) {
	t.TxSignature = s
}

func (t *TxInternalDataAccountCreation) IntrinsicGas() (uint64, error) {
	gas := params.TxGasAccountCreation + t.Key.AccountCreationGas()

	return gas, nil
}

func (t *TxInternalDataAccountCreation) SerializeForSign() []interface{} {
	infs := []interface{}{t.Type()}
	serializer := NewAccountKeySerializerWithAccountKey(t.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	return append(infs,
		t.TxInternalDataCommon.serializeForSign(),
		t.HumanReadable,
		keyEnc,
	)
}
