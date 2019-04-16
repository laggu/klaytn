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
	"encoding/json"
	"fmt"
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/hexutil"
	"github.com/ground-x/klaytn/kerrors"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"io"
	"math/big"
)

// TxInternalDataAccountCreation represents a transaction creating an account.
type TxInternalDataAccountCreation struct {
	AccountNonce  uint64
	Price         *big.Int
	GasLimit      uint64
	Recipient     common.Address
	Amount        *big.Int
	From          common.Address
	HumanReadable bool
	Key           accountkey.AccountKey

	TxSignatures

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

// txInternalDataAccountCreationSerializable for RLP serialization.
type txInternalDataAccountCreationSerializable struct {
	AccountNonce  uint64
	Price         *big.Int
	GasLimit      uint64
	Recipient     common.Address
	Amount        *big.Int
	From          common.Address
	HumanReadable bool
	KeyData       []byte

	TxSignatures
}

func newTxInternalDataAccountCreation() *TxInternalDataAccountCreation {
	h := common.Hash{}
	return &TxInternalDataAccountCreation{
		Price:  new(big.Int),
		Amount: new(big.Int),
		Key:    accountkey.NewAccountKeyLegacy(),
		Hash:   &h,
	}
}

func newTxInternalDataAccountCreationWithMap(values map[TxValueKeyType]interface{}) (*TxInternalDataAccountCreation, error) {
	b := newTxInternalDataAccountCreation()

	if v, ok := values[TxValueKeyNonce].(uint64); ok {
		b.AccountNonce = v
		delete(values, TxValueKeyNonce)
	} else {
		return nil, errValueKeyNonceMustUint64
	}

	if v, ok := values[TxValueKeyGasPrice].(*big.Int); ok {
		b.Price.Set(v)
		delete(values, TxValueKeyGasPrice)
	} else {
		return nil, errValueKeyGasPriceMustBigInt
	}

	if v, ok := values[TxValueKeyGasLimit].(uint64); ok {
		b.GasLimit = v
		delete(values, TxValueKeyGasLimit)
	} else {
		return nil, errValueKeyGasLimitMustUint64
	}

	if v, ok := values[TxValueKeyTo].(common.Address); ok {
		b.Recipient = v
		delete(values, TxValueKeyTo)
	} else {
		return nil, errValueKeyToMustAddress
	}

	if v, ok := values[TxValueKeyAmount].(*big.Int); ok {
		b.Amount.Set(v)
		delete(values, TxValueKeyAmount)
	} else {
		return nil, errValueKeyAmountMustBigInt
	}

	if v, ok := values[TxValueKeyFrom].(common.Address); ok {
		b.From = v
		delete(values, TxValueKeyFrom)
	} else {
		return nil, errValueKeyFromMustAddress
	}

	if v, ok := values[TxValueKeyHumanReadable].(bool); ok {
		b.HumanReadable = v
		delete(values, TxValueKeyHumanReadable)
	} else {
		return nil, errValueKeyHumanReadableMustBool
	}

	if v, ok := values[TxValueKeyAccountKey].(accountkey.AccountKey); ok {
		b.Key = v
		delete(values, TxValueKeyAccountKey)
	} else {
		return nil, errValueKeyAccountKeyMustAccountKey
	}

	if len(values) != 0 {
		for k := range values {
			logger.Warn("unnecessary key", k.String())
		}
		return nil, errUndefinedKeyRemains
	}

	return b, nil
}

func newTxInternalDataAccountCreationSerializable() *txInternalDataAccountCreationSerializable {
	return &txInternalDataAccountCreationSerializable{}
}

func (t *TxInternalDataAccountCreation) toSerializable() *txInternalDataAccountCreationSerializable {
	serializer := accountkey.NewAccountKeySerializerWithAccountKey(t.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	return &txInternalDataAccountCreationSerializable{
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.Recipient,
		t.Amount,
		t.From,
		t.HumanReadable,
		keyEnc,
		t.TxSignatures,
	}
}

func (t *TxInternalDataAccountCreation) fromSerializable(serialized *txInternalDataAccountCreationSerializable) {
	t.AccountNonce = serialized.AccountNonce
	t.Price = serialized.Price
	t.GasLimit = serialized.GasLimit
	t.Recipient = serialized.Recipient
	t.Amount = serialized.Amount
	t.From = serialized.From
	t.HumanReadable = serialized.HumanReadable
	t.TxSignatures = serialized.TxSignatures

	serializer := accountkey.NewAccountKeySerializer()
	rlp.DecodeBytes(serialized.KeyData, serializer)
	t.Key = serializer.GetKey()
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

func (t *TxInternalDataAccountCreation) GetRoleTypeForValidation() accountkey.RoleType {
	return accountkey.RoleTransaction
}

func (t *TxInternalDataAccountCreation) Equal(a TxInternalData) bool {
	ta, ok := a.(*TxInternalDataAccountCreation)
	if !ok {
		return false
	}

	return t.AccountNonce == ta.AccountNonce &&
		t.Price.Cmp(ta.Price) == 0 &&
		t.GasLimit == ta.GasLimit &&
		t.Recipient == ta.Recipient &&
		t.Amount.Cmp(ta.Amount) == 0 &&
		t.From == ta.From &&
		t.HumanReadable == ta.HumanReadable &&
		t.Key.Equal(ta.Key) &&
		t.TxSignatures.equal(ta.TxSignatures)
}

func (t *TxInternalDataAccountCreation) String() string {
	ser := newTxInternalDataSerializerWithValues(t)
	tx := Transaction{data: t}
	enc, _ := rlp.EncodeToBytes(ser)
	return fmt.Sprintf(`
	TX(%x)
	Type:          %s
	From:          %s
	To:            %s
	Nonce:         %v
	GasPrice:      %#x
	GasLimit:      %#x
	Value:         %#x
	HumanReadable: %t
	Key:           %s
	Signature:     %s
	Hex:           %x
`,
		tx.Hash(),
		t.Type().String(),
		t.From.String(),
		t.Recipient.String(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.Amount,
		t.HumanReadable,
		t.Key.String(),
		t.TxSignatures.string(),
		enc)
}

func (t *TxInternalDataAccountCreation) IsLegacyTransaction() bool {
	return false
}

func (t *TxInternalDataAccountCreation) GetAccountNonce() uint64 {
	return t.AccountNonce
}

func (t *TxInternalDataAccountCreation) GetPrice() *big.Int {
	return new(big.Int).Set(t.Price)
}

func (t *TxInternalDataAccountCreation) GetGasLimit() uint64 {
	return t.GasLimit
}

func (t *TxInternalDataAccountCreation) GetRecipient() *common.Address {
	if t.Recipient == (common.Address{}) {
		return nil
	}

	to := common.Address(t.Recipient)
	return &to
}

func (t *TxInternalDataAccountCreation) GetAmount() *big.Int {
	return new(big.Int).Set(t.Amount)
}

func (t *TxInternalDataAccountCreation) GetFrom() common.Address {
	return t.From
}

func (t *TxInternalDataAccountCreation) GetHash() *common.Hash {
	return t.Hash
}

func (t *TxInternalDataAccountCreation) SetHash(h *common.Hash) {
	t.Hash = h
}

func (t *TxInternalDataAccountCreation) SetSignature(s TxSignatures) {
	t.TxSignatures = s
}

func (t *TxInternalDataAccountCreation) IntrinsicGas(currentBlockNumber uint64) (uint64, error) {
	gasKey, err := t.Key.AccountCreationGas(currentBlockNumber)
	if err != nil {
		return 0, err
	}

	return params.TxGasAccountCreation + gasKey, nil
}

func (t *TxInternalDataAccountCreation) SerializeForSignToBytes() []byte {
	serializer := accountkey.NewAccountKeySerializerWithAccountKey(t.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	b, _ := rlp.EncodeToBytes(struct {
		Txtype        TxType
		AccountNonce  uint64
		Price         *big.Int
		GasLimit      uint64
		Recipient     common.Address
		Amount        *big.Int
		From          common.Address
		HumanReadable bool
		Key           []byte
	}{
		t.Type(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.Recipient,
		t.Amount,
		t.From,
		t.HumanReadable,
		keyEnc,
	})

	return b
}

func (t *TxInternalDataAccountCreation) SerializeForSign() []interface{} {
	serializer := accountkey.NewAccountKeySerializerWithAccountKey(t.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	return []interface{}{
		t.Type(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.Recipient,
		t.Amount,
		t.From,
		t.HumanReadable,
		keyEnc,
	}
}

func (t *TxInternalDataAccountCreation) Validate(stateDB StateDB, currentBlockNumber uint64) error {
	to := t.Recipient
	if t.HumanReadable {
		addrString := string(bytes.TrimRightFunc(to.Bytes(), func(r rune) bool {
			if r == rune(0x0) {
				return true
			}
			return false
		}))
		if err := common.IsHumanReadableAddress(addrString); err != nil {
			return kerrors.ErrNotHumanReadableAddress
		}
	}
	// Fail if the address is already created.
	if stateDB.Exist(to) {
		return kerrors.ErrAccountAlreadyExists
	}
	if err := t.Key.Init(currentBlockNumber); err != nil {
		return err
	}

	return nil
}

func (t *TxInternalDataAccountCreation) Execute(sender ContractRef, vm VM, stateDB StateDB, currentBlockNumber uint64, gas uint64, value *big.Int) (ret []byte, usedGas uint64, err error) {
	if err := t.Validate(stateDB, currentBlockNumber); err != nil {
		stateDB.IncNonce(sender.Address())
		return nil, 0, err
	}
	to := t.Recipient
	stateDB.CreateEOA(to, t.HumanReadable, t.Key)
	stateDB.IncNonce(sender.Address())
	ret, usedGas, err = vm.Call(sender, to, []byte{}, gas, value)

	return
}

func (t *TxInternalDataAccountCreation) MakeRPCOutput() map[string]interface{} {
	serializer := accountkey.NewAccountKeySerializerWithAccountKey(t.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	return map[string]interface{}{
		"type":          t.Type().String(),
		"gas":           hexutil.Uint64(t.GasLimit),
		"gasPrice":      (*hexutil.Big)(t.Price),
		"nonce":         hexutil.Uint64(t.AccountNonce),
		"to":            t.Recipient,
		"value":         (*hexutil.Big)(t.Amount),
		"humanReadable": t.HumanReadable,
		"key":           hexutil.Bytes(keyEnc),
	}
}
