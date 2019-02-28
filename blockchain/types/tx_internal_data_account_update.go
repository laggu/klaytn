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
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/hexutil"
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
	Key          accountkey.AccountKey

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`

	TxSignatures
}

type txInternalDataAccountUpdateSerializable struct {
	AccountNonce uint64
	Price        *big.Int
	GasLimit     uint64
	From         common.Address
	Key          []byte

	TxSignatures
}

func newTxInternalDataAccountUpdate() *TxInternalDataAccountUpdate {
	return &TxInternalDataAccountUpdate{
		Price:        new(big.Int),
		Key:          accountkey.NewAccountKeyLegacy(),
		TxSignatures: NewTxSignatures(),
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

	if v, ok := values[TxValueKeyAccountKey].(accountkey.AccountKey); ok {
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
	serializer := accountkey.NewAccountKeySerializerWithAccountKey(t.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	return &txInternalDataAccountUpdateSerializable{
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.From,
		keyEnc,
		t.TxSignatures,
	}
}

func (t *TxInternalDataAccountUpdate) fromSerializable(serialized *txInternalDataAccountUpdateSerializable) {
	t.AccountNonce = serialized.AccountNonce
	t.Price = serialized.Price
	t.GasLimit = serialized.GasLimit
	t.From = serialized.From
	t.TxSignatures = serialized.TxSignatures

	serializer := accountkey.NewAccountKeySerializer()
	rlp.DecodeBytes(serialized.Key, serializer)
	t.Key = serializer.GetKey()
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

func (t *TxInternalDataAccountUpdate) GetRoleTypeForValidation() accountkey.RoleType {
	return accountkey.RoleAccountUpdate
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
		t.TxSignatures.equal(ta.TxSignatures)
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
		t.TxSignatures.string(),
		enc)
}

func (t *TxInternalDataAccountUpdate) SetSignature(s TxSignatures) {
	t.TxSignatures = s
}

func (t *TxInternalDataAccountUpdate) IntrinsicGas() (uint64, error) {
	gasKey, err := t.Key.AccountCreationGas()
	if err != nil {
		return 0, err
	}

	return params.TxGasAccountUpdate + gasKey, nil
}

func (t *TxInternalDataAccountUpdate) SerializeForSignToBytes() []byte {
	serializer := accountkey.NewAccountKeySerializerWithAccountKey(t.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	b, _ := rlp.EncodeToBytes(struct {
		Txtype       TxType
		AccountNonce uint64
		Price        *big.Int
		GasLimit     uint64
		From         common.Address
		Key          []byte
	}{
		t.Type(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.From,
		keyEnc,
	})

	return b
}

func (t *TxInternalDataAccountUpdate) SerializeForSign() []interface{} {
	serializer := accountkey.NewAccountKeySerializerWithAccountKey(t.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	b, _ := rlp.EncodeToBytes(struct {
		Txtype       TxType
		AccountNonce uint64
		Price        *big.Int
		GasLimit     uint64
		From         common.Address
		Key          []byte
	}{
		t.Type(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.From,
		keyEnc,
	})

	return []interface{}{b}
}

func (t *TxInternalDataAccountUpdate) Execute(sender ContractRef, vm VM, stateDB StateDB, gas uint64, value *big.Int) (ret []byte, usedGas uint64, err, vmerr error) {
	stateDB.IncNonce(sender.Address())
	err = stateDB.UpdateKey(sender.Address(), t.Key)
	return nil, gas, err, nil
}

func (t *TxInternalDataAccountUpdate) MakeRPCOutput() map[string]interface{} {
	serializer := accountkey.NewAccountKeySerializerWithAccountKey(t.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	return map[string]interface{}{
		"type":     t.Type().String(),
		"gas":      hexutil.Uint64(t.GasLimit),
		"gasPrice": (*hexutil.Big)(t.Price),
		"nonce":    hexutil.Uint64(t.AccountNonce),
		"key":      hexutil.Bytes(keyEnc),
	}
}
