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
	"crypto/ecdsa"
	"fmt"
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/hexutil"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
)

// TxInternalDataFeeDelegatedValueTransferMemo represents a fee-delegated transaction transferring KLAY.
type TxInternalDataFeeDelegatedValueTransferMemo struct {
	AccountNonce uint64
	Price        *big.Int
	GasLimit     uint64
	Recipient    common.Address
	Amount       *big.Int
	From         common.Address
	Payload      []byte

	TxSignatures

	FeePayer          common.Address
	FeePayerSignature TxSignatures

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

func newTxInternalDataFeeDelegatedValueTransferMemo() *TxInternalDataFeeDelegatedValueTransferMemo {
	h := common.Hash{}

	return &TxInternalDataFeeDelegatedValueTransferMemo{
		Price:  new(big.Int),
		Amount: new(big.Int),
		Hash:   &h,
	}
}

func newTxInternalDataFeeDelegatedValueTransferMemoWithMap(values map[TxValueKeyType]interface{}) (*TxInternalDataFeeDelegatedValueTransferMemo, error) {
	d := newTxInternalDataFeeDelegatedValueTransferMemo()

	if v, ok := values[TxValueKeyNonce].(uint64); ok {
		d.AccountNonce = v
	} else {
		return nil, errValueKeyNonceMustUint64
	}

	if v, ok := values[TxValueKeyGasPrice].(*big.Int); ok {
		d.Price.Set(v)
	} else {
		return nil, errValueKeyGasPriceMustBigInt
	}

	if v, ok := values[TxValueKeyGasLimit].(uint64); ok {
		d.GasLimit = v
	} else {
		return nil, errValueKeyGasLimitMustUint64
	}

	if v, ok := values[TxValueKeyTo].(common.Address); ok {
		d.Recipient = v
	} else {
		return nil, errValueKeyToMustAddress
	}

	if v, ok := values[TxValueKeyAmount].(*big.Int); ok {
		d.Amount.Set(v)
	} else {
		return nil, errValueKeyAmountMustBigInt
	}

	if v, ok := values[TxValueKeyFrom].(common.Address); ok {
		d.From = v
	} else {
		return nil, errValueKeyFromMustAddress
	}

	if v, ok := values[TxValueKeyData].([]byte); ok {
		d.Payload = v
	} else {
		return nil, errValueKeyDataMustByteSlice
	}

	if v, ok := values[TxValueKeyFeePayer].(common.Address); ok {
		d.FeePayer = v
	} else {
		return nil, errValueKeyFeePayerMustAddress
	}

	return d, nil
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) Type() TxType {
	return TxTypeFeeDelegatedValueTransferMemo
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) GetRoleTypeForValidation() accountkey.RoleType {
	return accountkey.RoleTransaction
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) Equal(b TxInternalData) bool {
	tb, ok := b.(*TxInternalDataFeeDelegatedValueTransferMemo)
	if !ok {
		return false
	}

	return t.AccountNonce == tb.AccountNonce &&
		t.Price.Cmp(tb.Price) == 0 &&
		t.GasLimit == tb.GasLimit &&
		t.Recipient == tb.Recipient &&
		t.Amount.Cmp(tb.Amount) == 0 &&
		t.From == tb.From &&
		bytes.Equal(t.Payload, tb.Payload) &&
		t.TxSignatures.equal(tb.TxSignatures) &&
		t.FeePayer == tb.FeePayer &&
		t.FeePayerSignature.equal(tb.FeePayerSignature)
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) String() string {
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
	Signature:     %s
	FeePayer:      %s
	FeePayerSig:   %s
	Paylod:        %x
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
		t.TxSignatures.string(),
		t.FeePayer.String(),
		t.FeePayerSignature.string(),
		common.Bytes2Hex(t.Payload),
		enc)
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) IsLegacyTransaction() bool {
	return false
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) GetAccountNonce() uint64 {
	return t.AccountNonce
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) GetPrice() *big.Int {
	return new(big.Int).Set(t.Price)
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) GetGasLimit() uint64 {
	return t.GasLimit
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) GetRecipient() *common.Address {
	if t.Recipient == (common.Address{}) {
		return nil
	}

	to := common.Address(t.Recipient)
	return &to
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) GetAmount() *big.Int {
	return new(big.Int).Set(t.Amount)
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) GetFrom() common.Address {
	return t.From
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) GetPayload() []byte {
	return t.Payload
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) GetHash() *common.Hash {
	return t.Hash
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) GetFeePayer() common.Address {
	return t.FeePayer
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) GetFeePayerRawSignatureValues() []*big.Int {
	return t.FeePayerSignature.RawSignatureValues()
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) SetHash(h *common.Hash) {
	t.Hash = h
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) SetSignature(s TxSignatures) {
	t.TxSignatures = s
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) SetFeePayerSignature(s TxSignatures) {
	t.FeePayerSignature = s
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) RecoverFeePayerPubkey(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) ([]*ecdsa.PublicKey, error) {
	return t.FeePayerSignature.RecoverPubkey(txhash, homestead, vfunc)
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) IntrinsicGas() (uint64, error) {
	gasPayload, err := intrinsicGasPayload(t.Payload)
	if err != nil {
		return 0, err
	}

	return params.TxGas + gasPayload + params.TxGasFeeDelegated, nil
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) SerializeForSignToBytes() []byte {
	b, _ := rlp.EncodeToBytes(struct {
		Txtype       TxType
		AccountNonce uint64
		Price        *big.Int
		GasLimit     uint64
		Recipient    common.Address
		Amount       *big.Int
		From         common.Address
		Payload      []byte
	}{
		t.Type(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.Recipient,
		t.Amount,
		t.From,
		t.Payload,
	})

	return b
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) SerializeForSign() []interface{} {
	return []interface{}{
		t.Type(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.Recipient,
		t.Amount,
		t.From,
		t.Payload,
	}
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) Validate(stateDB StateDB) error {
	// No more validation required for TxInternalDataFeeDelegatedValueTransferMemo.
	return nil
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) Execute(sender ContractRef, vm VM, stateDB StateDB, gas uint64, value *big.Int) (ret []byte, usedGas uint64, err error) {
	stateDB.IncNonce(sender.Address())
	return vm.Call(sender, t.Recipient, t.Payload, gas, value)
}

func (t *TxInternalDataFeeDelegatedValueTransferMemo) MakeRPCOutput() map[string]interface{} {
	return map[string]interface{}{
		"type":     t.Type().String(),
		"gas":      hexutil.Uint64(t.GasLimit),
		"gasPrice": (*hexutil.Big)(t.Price),
		"input":    hexutil.Bytes(t.Payload),
		"nonce":    hexutil.Uint64(t.AccountNonce),
		"to":       t.Recipient,
		"value":    (*hexutil.Big)(t.Amount),
		"feePayer": t.FeePayer,
	}
}
