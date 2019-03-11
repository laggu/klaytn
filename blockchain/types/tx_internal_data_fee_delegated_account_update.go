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

// TxInternalDataFeeDelegatedAccountUpdate represents a fee-delegated transaction updating a key of an account.
type TxInternalDataFeeDelegatedAccountUpdate struct {
	AccountNonce uint64
	Price        *big.Int
	GasLimit     uint64
	From         common.Address
	Key          accountkey.AccountKey

	TxSignatures

	FeePayer          common.Address
	FeePayerSignature TxSignatures

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

type txInternalDataFeeDelegatedAccountUpdateSerializable struct {
	AccountNonce uint64
	Price        *big.Int
	GasLimit     uint64
	From         common.Address
	Key          []byte

	TxSignatures

	FeePayer          common.Address
	FeePayerSignature TxSignatures
}

func newTxInternalDataFeeDelegatedAccountUpdate() *TxInternalDataFeeDelegatedAccountUpdate {
	return &TxInternalDataFeeDelegatedAccountUpdate{
		Price:        new(big.Int),
		Key:          accountkey.NewAccountKeyLegacy(),
		TxSignatures: NewTxSignatures(),
	}
}

func newTxInternalDataFeeDelegatedAccountUpdateWithMap(values map[TxValueKeyType]interface{}) (*TxInternalDataFeeDelegatedAccountUpdate, error) {
	d := newTxInternalDataFeeDelegatedAccountUpdate()

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

	if v, ok := values[TxValueKeyFeePayer].(common.Address); ok {
		d.FeePayer = v
	} else {
		return nil, errValueKeyFeePayerMustAddress
	}

	return d, nil
}

func newTxInternalDataFeeDelegatedAccountUpdateSerializable() *txInternalDataFeeDelegatedAccountUpdateSerializable {
	return &txInternalDataFeeDelegatedAccountUpdateSerializable{}
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) toSerializable() *txInternalDataFeeDelegatedAccountUpdateSerializable {
	serializer := accountkey.NewAccountKeySerializerWithAccountKey(t.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	return &txInternalDataFeeDelegatedAccountUpdateSerializable{
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.From,
		keyEnc,
		t.TxSignatures,
		t.FeePayer,
		t.FeePayerSignature,
	}
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) fromSerializable(serialized *txInternalDataFeeDelegatedAccountUpdateSerializable) {
	t.AccountNonce = serialized.AccountNonce
	t.Price = serialized.Price
	t.GasLimit = serialized.GasLimit
	t.From = serialized.From
	t.TxSignatures = serialized.TxSignatures
	t.FeePayer = serialized.FeePayer
	t.FeePayerSignature = serialized.FeePayerSignature

	serializer := accountkey.NewAccountKeySerializer()
	rlp.DecodeBytes(serialized.Key, serializer)
	t.Key = serializer.GetKey()
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, t.toSerializable())
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) DecodeRLP(s *rlp.Stream) error {
	dec := newTxInternalDataFeeDelegatedAccountUpdateSerializable()

	if err := s.Decode(dec); err != nil {
		return err
	}
	t.fromSerializable(dec)

	return nil
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.toSerializable())
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) UnmarshalJSON(b []byte) error {
	dec := newTxInternalDataFeeDelegatedAccountUpdateSerializable()

	if err := json.Unmarshal(b, dec); err != nil {
		return err
	}

	t.fromSerializable(dec)

	return nil
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) Type() TxType {
	return TxTypeFeeDelegatedAccountUpdate
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) GetRoleTypeForValidation() accountkey.RoleType {
	return accountkey.RoleAccountUpdate
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) GetAccountNonce() uint64 {
	return t.AccountNonce
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) GetPrice() *big.Int {
	return t.Price
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) GetGasLimit() uint64 {
	return t.GasLimit
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) GetRecipient() *common.Address {
	return &common.Address{}
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) GetAmount() *big.Int {
	return common.Big0
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) GetFrom() common.Address {
	return t.From
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) GetHash() *common.Hash {
	return t.Hash
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) GetFeePayer() common.Address {
	return t.FeePayer
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) GetFeePayerRawSignatureValues() []*big.Int {
	return t.FeePayerSignature.RawSignatureValues()
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) SetHash(h *common.Hash) {
	t.Hash = h
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) IsLegacyTransaction() bool {
	return false
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) Equal(a TxInternalData) bool {
	ta, ok := a.(*TxInternalDataFeeDelegatedAccountUpdate)
	if !ok {
		return false
	}

	return t.AccountNonce == ta.AccountNonce &&
		t.Price.Cmp(ta.Price) == 0 &&
		t.GasLimit == ta.GasLimit &&
		t.From == ta.From &&
		t.Key.Equal(ta.Key) &&
		t.TxSignatures.equal(ta.TxSignatures) &&
		t.FeePayer == ta.FeePayer &&
		t.FeePayerSignature.equal(ta.FeePayerSignature)
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) String() string {
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
	FeePayer:      %s
	FeePayerSig:   %s
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
		t.FeePayer.String(),
		t.FeePayerSignature.string(),
		enc)
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) SetSignature(s TxSignatures) {
	t.TxSignatures = s
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) SetFeePayerSignature(s TxSignatures) {
	t.FeePayerSignature = s
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) RecoverFeePayerPubkey(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) ([]*ecdsa.PublicKey, error) {
	return t.FeePayerSignature.RecoverPubkey(txhash, homestead, vfunc)
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) IntrinsicGas() (uint64, error) {
	gasKey, err := t.Key.AccountCreationGas()
	if err != nil {
		return 0, err
	}

	return params.TxGasAccountUpdate + gasKey + params.TxGasFeeDelegated, nil
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) SerializeForSignToBytes() []byte {
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

func (t *TxInternalDataFeeDelegatedAccountUpdate) SerializeForSign() []interface{} {
	serializer := accountkey.NewAccountKeySerializerWithAccountKey(t.Key)
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

func (t *TxInternalDataFeeDelegatedAccountUpdate) Validate(stateDB StateDB) error {
	// TODO-Klaytn-Accounts: need validation of t.key?
	return nil
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) Execute(sender ContractRef, vm VM, stateDB StateDB, gas uint64, value *big.Int) (ret []byte, usedGas uint64, err, vmerr error) {
	stateDB.IncNonce(sender.Address())
	err = stateDB.UpdateKey(sender.Address(), t.Key)

	return nil, gas, err, nil
}

func (t *TxInternalDataFeeDelegatedAccountUpdate) MakeRPCOutput() map[string]interface{} {
	serializer := accountkey.NewAccountKeySerializerWithAccountKey(t.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	return map[string]interface{}{
		"type":     t.Type().String(),
		"gas":      hexutil.Uint64(t.GasLimit),
		"gasPrice": (*hexutil.Big)(t.Price),
		"nonce":    hexutil.Uint64(t.AccountNonce),
		"key":      hexutil.Bytes(keyEnc),
		"feePayer": t.FeePayer,
	}
}
