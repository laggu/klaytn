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
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"io"
	"math/big"
)

// TxInternalDataAccountUpdate represents a transaction updating a key of an account.
type TxInternalDataAccountUpdate struct {
	AccountNonce uint64
	Price        *big.Int
	GasLimit     uint64
	From         common.Address
	Key          AccountKey

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`

	*TxSignature
}

type txInternalDataAccountUpdateSerializable struct {
	AccountNonce uint64
	Price        *big.Int
	GasLimit     uint64
	From         common.Address
	Key          []byte

	*TxSignature
}

func newTxInternalDataAccountUpdate() *TxInternalDataAccountUpdate {
	return &TxInternalDataAccountUpdate{
		Price:       new(big.Int),
		Key:         NewAccountKeyNil(),
		TxSignature: NewTxSignature(),
	}
}

func newTxInternalDataAccountUpdateWithMap(values map[TxValueKeyType]interface{}) (*TxInternalDataAccountUpdate, error) {
	d := newTxInternalDataAccountUpdate()

	if v, ok := values[TxValueKeyNonce].(uint64); ok {
		d.AccountNonce = v
	} else {
		return nil, errValueKeyNonceMustUint64
	}

	if v, ok := values[TxValueKeyGasLimit].(uint64); ok {
		d.GasLimit = v
	} else {
		return nil, errValueKeyGasLimitMustUint64
	}

	if v, ok := values[TxValueKeyGasPrice].(*big.Int); ok {
		d.Price.Set(v)
	} else {
		return nil, errValueKeyGasPriceMustBigInt
	}

	if v, ok := values[TxValueKeyFrom].(common.Address); ok {
		d.From = v
	} else {
		return nil, errValueKeyFromMustAddress
	}

	if v, ok := values[TxValueKeyAccountKey].(AccountKey); ok {
		d.Key = v
	} else {
		return nil, errValueKeyAccountKeyMustAccountKey
	}

	return d, nil
}

func newTxInternalDataAccountUpdateSerializable() *txInternalDataAccountUpdateSerializable {
	return &txInternalDataAccountUpdateSerializable{}
}

func (t *TxInternalDataAccountUpdate) toSerializable() *txInternalDataAccountUpdateSerializable {
	serializer := NewAccountKeySerializerWithAccountKey(t.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	return &txInternalDataAccountUpdateSerializable{
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.From,
		keyEnc,
		t.TxSignature,
	}
}

func (t *TxInternalDataAccountUpdate) fromSerializable(serialized *txInternalDataAccountUpdateSerializable) {
	t.AccountNonce = serialized.AccountNonce
	t.Price = serialized.Price
	t.GasLimit = serialized.GasLimit
	t.From = serialized.From
	t.TxSignature = serialized.TxSignature

	serializer := NewAccountKeySerializer()
	rlp.DecodeBytes(serialized.Key, serializer)
	t.Key = serializer.key
}

func (t *TxInternalDataAccountUpdate) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, t.toSerializable())
}

func (t *TxInternalDataAccountUpdate) DecodeRLP(s *rlp.Stream) error {
	dec := newTxInternalDataAccountUpdateSerializable()

	if err := s.Decode(dec); err != nil {
		return err
	}
	t.fromSerializable(dec)

	return nil
}

func (t *TxInternalDataAccountUpdate) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.toSerializable())
}

func (t *TxInternalDataAccountUpdate) UnmarshalJSON(b []byte) error {
	dec := newTxInternalDataAccountUpdateSerializable()

	if err := json.Unmarshal(b, dec); err != nil {
		return err
	}

	t.fromSerializable(dec)

	return nil
}

func (t *TxInternalDataAccountUpdate) Type() TxType {
	return TxTypeAccountUpdate
}

func (t *TxInternalDataAccountUpdate) GetAccountNonce() uint64 {
	return t.AccountNonce
}

func (t *TxInternalDataAccountUpdate) GetPrice() *big.Int {
	return t.Price
}

func (t *TxInternalDataAccountUpdate) GetGasLimit() uint64 {
	return t.GasLimit
}

func (t *TxInternalDataAccountUpdate) GetRecipient() *common.Address {
	return &common.Address{}
}

func (t *TxInternalDataAccountUpdate) GetAmount() *big.Int {
	return common.Big0
}

func (t *TxInternalDataAccountUpdate) GetFrom() common.Address {
	return t.From
}

func (t *TxInternalDataAccountUpdate) GetHash() *common.Hash {
	return t.Hash
}

func (t *TxInternalDataAccountUpdate) SetHash(h *common.Hash) {
	t.Hash = h
}

func (t *TxInternalDataAccountUpdate) IsLegacyTransaction() bool {
	return false
}

func (t *TxInternalDataAccountUpdate) Equal(a TxInternalData) bool {
	ta, ok := a.(*TxInternalDataAccountUpdate)
	if !ok {
		return false
	}

	return t.AccountNonce == ta.AccountNonce &&
		t.Price.Cmp(ta.Price) == 0 &&
		t.GasLimit == ta.GasLimit &&
		t.From == ta.From &&
		t.Key.Equal(ta.Key) &&
		t.TxSignature.equal(ta.TxSignature)
}

func (t *TxInternalDataAccountUpdate) String() string {
	ser := newTxInternalDataSerializerWithValues(t)
	tx := Transaction{data: t}
	enc, _ := rlp.EncodeToBytes(ser)
	return fmt.Sprintf(`
	TX(%x)
	Type:          %s
	From:          %s
	Nonce:         %v
	GasPrice:      %#x
	GasLimit:      %#x
	Key:           %s
	Signature:     %s
	Hex:           %x
`,
		tx.Hash(),
		t.Type().String(),
		t.From.String(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.Key.String(),
		t.TxSignature.string(),
		enc)
}

func (t *TxInternalDataAccountUpdate) SetSignature(s *TxSignature) {
	t.TxSignature = s
}

func (t *TxInternalDataAccountUpdate) IntrinsicGas() (uint64, error) {
	gasKey, err := t.Key.AccountCreationGas()
	if err != nil {
		return 0, err
	}

	return params.TxGasAccountUpdate + gasKey, nil
}

func (t *TxInternalDataAccountUpdate) SerializeForSign() []interface{} {
	serializer := NewAccountKeySerializerWithAccountKey(t.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	return []interface{}{
		t.Type(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.From,
		keyEnc,
	}
}
