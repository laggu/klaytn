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
	"github.com/ground-x/klaytn/kerrors"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
)

// TxInternalDataFeeDelegatedCancelWithRatio is a fee-delegated transaction that cancels a transaction previously submitted into txpool by replacement.
// Since Klaytn defines fixed gas price for all transactions, a transaction cannot be replaced with
// another transaction with higher gas price. To provide tx replacement, TxInternalDataFeeDelegatedCancelWithRatio is introduced.
// To replace a previously added tx, send a TxInternalFeeDelegatedCancelWithRatio transaction with the same nonce.
// TxInternalDataFeeDelegatedCancelWithRatio has a specified fee ratio between the sender and the fee payer.
// The ratio is a fee payer's ratio in percentage.
// For example, if it is 20, 20% of tx fee will be paid by the fee payer.
// 80% of tx fee will be paid by the sender.
type TxInternalDataFeeDelegatedCancelWithRatio struct {
	AccountNonce uint64
	Price        *big.Int
	GasLimit     uint64
	From         common.Address
	FeeRatio     FeeRatio

	TxSignatures

	FeePayer          common.Address
	FeePayerSignature TxSignatures

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

func newTxInternalDataFeeDelegatedCancelWithRatio() *TxInternalDataFeeDelegatedCancelWithRatio {
	return &TxInternalDataFeeDelegatedCancelWithRatio{
		Price:        new(big.Int),
		TxSignatures: NewTxSignatures(),
	}
}

func newTxInternalDataFeeDelegatedCancelWithRatioWithMap(values map[TxValueKeyType]interface{}) (*TxInternalDataFeeDelegatedCancelWithRatio, error) {
	d := newTxInternalDataFeeDelegatedCancelWithRatio()

	if v, ok := values[TxValueKeyNonce].(uint64); ok {
		d.AccountNonce = v
		delete(values, TxValueKeyNonce)
	} else {
		return nil, errValueKeyNonceMustUint64
	}

	if v, ok := values[TxValueKeyGasLimit].(uint64); ok {
		d.GasLimit = v
		delete(values, TxValueKeyGasLimit)
	} else {
		return nil, errValueKeyGasLimitMustUint64
	}

	if v, ok := values[TxValueKeyGasPrice].(*big.Int); ok {
		d.Price.Set(v)
		delete(values, TxValueKeyGasPrice)
	} else {
		return nil, errValueKeyGasPriceMustBigInt
	}

	if v, ok := values[TxValueKeyFrom].(common.Address); ok {
		d.From = v
		delete(values, TxValueKeyFrom)
	} else {
		return nil, errValueKeyFromMustAddress
	}

	if v, ok := values[TxValueKeyFeePayer].(common.Address); ok {
		d.FeePayer = v
		delete(values, TxValueKeyFeePayer)
	} else {
		return nil, errValueKeyFeePayerMustAddress
	}

	if v, ok := values[TxValueKeyFeeRatioOfFeePayer].(FeeRatio); ok {
		d.FeeRatio = v
		delete(values, TxValueKeyFeeRatioOfFeePayer)
	} else {
		return nil, errValueKeyFeeRatioMustUint8
	}

	if len(values) != 0 {
		for k := range values {
			fmt.Println("unnecessary key", k.String())
		}
		return nil, errUndefinedKeyRemains
	}

	return d, nil
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) Type() TxType {
	return TxTypeFeeDelegatedCancelWithRatio
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) GetRoleTypeForValidation() accountkey.RoleType {
	return accountkey.RoleTransaction
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) GetAccountNonce() uint64 {
	return t.AccountNonce
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) GetPrice() *big.Int {
	return t.Price
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) GetGasLimit() uint64 {
	return t.GasLimit
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) GetRecipient() *common.Address {
	return &common.Address{}
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) GetAmount() *big.Int {
	return common.Big0
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) GetFrom() common.Address {
	return t.From
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) GetHash() *common.Hash {
	return t.Hash
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) GetFeePayer() common.Address {
	return t.FeePayer
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) GetFeePayerRawSignatureValues() []*big.Int {
	return t.FeePayerSignature.RawSignatureValues()
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) GetFeeRatio() FeeRatio {
	return t.FeeRatio
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) SetHash(h *common.Hash) {
	t.Hash = h
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) SetFeePayerSignature(s TxSignatures) {
	t.FeePayerSignature = s
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) RecoverFeePayerPubkey(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) ([]*ecdsa.PublicKey, error) {
	return t.FeePayerSignature.RecoverPubkey(txhash, homestead, vfunc)
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) IsLegacyTransaction() bool {
	return false
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) Equal(b TxInternalData) bool {
	ta, ok := b.(*TxInternalDataFeeDelegatedCancelWithRatio)
	if !ok {
		return false
	}

	return t.AccountNonce == ta.AccountNonce &&
		t.Price.Cmp(ta.Price) == 0 &&
		t.GasLimit == ta.GasLimit &&
		t.From == ta.From &&
		t.FeeRatio == ta.FeeRatio &&
		t.TxSignatures.equal(ta.TxSignatures) &&
		t.FeePayer == ta.FeePayer &&
		t.FeePayerSignature.equal(ta.FeePayerSignature)
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) String() string {
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
	Signature:     %s
	FeePayer:      %s
	FeeRatio:      %d
	FeePayerSig:   %s
	Hex:           %x
`,
		tx.Hash(),
		t.Type().String(),
		t.From.String(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.TxSignatures.string(),
		t.FeePayer.String(),
		t.FeeRatio,
		t.FeePayerSignature.string(),
		enc)
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) SetSignature(s TxSignatures) {
	t.TxSignatures = s
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) IntrinsicGas() (uint64, error) {
	return params.TxGasCancel + params.TxGasFeeDelegatedWithRatio, nil
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) SerializeForSignToBytes() []byte {
	b, _ := rlp.EncodeToBytes(struct {
		Txtype       TxType
		AccountNonce uint64
		Price        *big.Int
		GasLimit     uint64
		From         common.Address
		FeeRatio     FeeRatio
	}{
		t.Type(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.From,
		t.FeeRatio,
	})

	return b
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) SerializeForSign() []interface{} {
	return []interface{}{
		t.Type(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.From,
		t.FeeRatio,
	}
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) Validate(stateDB StateDB, currentBlockNumber uint64) error {
	if t.FeeRatio.IsValid() == false {
		return kerrors.ErrFeeRatioOutOfRange
	}

	return nil
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) Execute(sender ContractRef, vm VM, stateDB StateDB, currentBlockNumber uint64, gas uint64, value *big.Int) (ret []byte, usedGas uint64, err error) {
	stateDB.IncNonce(sender.Address())
	return nil, gas, nil
}

func (t *TxInternalDataFeeDelegatedCancelWithRatio) MakeRPCOutput() map[string]interface{} {
	return map[string]interface{}{
		"type":     t.Type().String(),
		"gas":      hexutil.Uint64(t.GasLimit),
		"gasPrice": (*hexutil.Big)(t.Price),
		"nonce":    hexutil.Uint64(t.AccountNonce),
		"feePayer": t.FeePayer,
		"feeRatio": hexutil.Uint(t.FeeRatio),
	}
}
