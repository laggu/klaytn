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
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/hexutil"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
)

// TxInternalDataFeeDelegatedValueTransferWithRatio represents a fee-delegated value transfer transaction with a
// specified fee ratio between the sender and the fee payer.
// The ratio is a fee payer's ratio in percentage.
// For example, if it is 20, 20% of tx fee will be paid by the fee payer.
// 80% of tx fee will be paid by the sender.
type TxInternalDataFeeDelegatedValueTransferWithRatio struct {
	AccountNonce uint64
	Price        *big.Int
	GasLimit     uint64
	Recipient    common.Address
	Amount       *big.Int
	From         common.Address
	FeeRatio     uint8

	TxSignatures

	FeePayer          common.Address
	FeePayerSignature TxSignatures

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

func NewTxInternalDataFeeDelegatedValueTransferWithRatio() *TxInternalDataFeeDelegatedValueTransferWithRatio {
	h := common.Hash{}
	return &TxInternalDataFeeDelegatedValueTransferWithRatio{
		Price:  new(big.Int),
		Amount: new(big.Int),
		Hash:   &h,
	}
}

func newTxInternalDataFeeDelegatedValueTransferWithRatioWithMap(values map[TxValueKeyType]interface{}) (*TxInternalDataFeeDelegatedValueTransferWithRatio, error) {
	d := NewTxInternalDataFeeDelegatedValueTransferWithRatio()

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

	if v, ok := values[TxValueKeyFeeRatioOfFeePayer].(uint8); ok {
		d.FeeRatio = v
	} else {
		return nil, errValueKeyFeeRatioMustUint8
	}

	if v, ok := values[TxValueKeyFeePayer].(common.Address); ok {
		d.FeePayer = v
	} else {
		return nil, errValueKeyFeePayerMustAddress
	}

	return d, nil
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) Type() TxType {
	return TxTypeFeeDelegatedValueTransferWithRatio
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) GetRoleTypeForValidation() accountkey.RoleType {
	return accountkey.RoleTransaction
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) Equal(a TxInternalData) bool {
	ta, ok := a.(*TxInternalDataFeeDelegatedValueTransferWithRatio)
	if !ok {
		return false
	}

	return t.AccountNonce == ta.AccountNonce &&
		t.Price.Cmp(ta.Price) == 0 &&
		t.GasLimit == ta.GasLimit &&
		t.Recipient == ta.Recipient &&
		t.Amount.Cmp(ta.Amount) == 0 &&
		t.From == ta.From &&
		t.FeeRatio == ta.FeeRatio &&
		t.TxSignatures.equal(ta.TxSignatures) &&
		t.FeePayer == ta.FeePayer &&
		t.FeePayerSignature.equal(ta.FeePayerSignature)
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) IsLegacyTransaction() bool {
	return false
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) GetFeeRatio() uint8 {
	return t.FeeRatio
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) GetAccountNonce() uint64 {
	return t.AccountNonce
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) GetPrice() *big.Int {
	return new(big.Int).Set(t.Price)
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) GetGasLimit() uint64 {
	return t.GasLimit
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) GetRecipient() *common.Address {
	if t.Recipient == (common.Address{}) {
		return nil
	}

	to := common.Address(t.Recipient)
	return &to
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) GetAmount() *big.Int {
	return new(big.Int).Set(t.Amount)
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) GetFrom() common.Address {
	return t.From
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) GetHash() *common.Hash {
	return t.Hash
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) GetFeePayer() common.Address {
	return t.FeePayer
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) GetFeePayerRawSignatureValues() []*big.Int {
	return t.FeePayerSignature.RawSignatureValues()
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) SetHash(h *common.Hash) {
	t.Hash = h
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) SetSignature(s TxSignatures) {
	t.TxSignatures = s
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) SetFeePayerSignature(s TxSignatures) {
	t.FeePayerSignature = s
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) RecoverFeePayerPubkey(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) ([]*ecdsa.PublicKey, error) {
	return t.FeePayerSignature.RecoverPubkey(txhash, homestead, vfunc)
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) String() string {
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
	FeeRatio:      %d
	FeePayerSig:   %s
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
		t.FeeRatio,
		t.FeePayerSignature.string(),
		enc)
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) IntrinsicGas() (uint64, error) {
	return params.TxGas + params.TxGasFeeDelegatedWithRatio, nil
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) SerializeForSignToBytes() []byte {
	b, _ := rlp.EncodeToBytes(struct {
		Txtype       TxType
		AccountNonce uint64
		Price        *big.Int
		GasLimit     uint64
		Recipient    common.Address
		Amount       *big.Int
		From         common.Address
		FeeRatio     uint8
	}{
		t.Type(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.Recipient,
		t.Amount,
		t.From,
		t.FeeRatio,
	})

	return b
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) SerializeForSign() []interface{} {
	return []interface{}{
		t.Type(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.Recipient,
		t.Amount,
		t.From,
		t.FeeRatio,
	}
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) Execute(sender ContractRef, vm VM, stateDB StateDB, gas uint64, value *big.Int) (ret []byte, usedGas uint64, err, vmerr error) {
	stateDB.IncNonce(sender.Address())
	ret, usedGas, vmerr = vm.Call(sender, t.Recipient, nil, gas, value)

	return
}

func (t *TxInternalDataFeeDelegatedValueTransferWithRatio) MakeRPCOutput() map[string]interface{} {
	return map[string]interface{}{
		"type":     t.Type().String(),
		"gas":      hexutil.Uint64(t.GasLimit),
		"gasPrice": (*hexutil.Big)(t.Price),
		"nonce":    hexutil.Uint64(t.AccountNonce),
		"to":       t.Recipient,
		"value":    (*hexutil.Big)(t.Amount),
		"feeRatio": hexutil.Uint(t.FeeRatio),
		"feePayer": t.FeePayer,
	}
}
