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
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
)

//go:generate gencodec -type TxInternalDataLegacy -field-override txdataMarshaling -out gen_tx_json.go

type TxInternalDataLegacy struct {
	AccountNonce uint64          `json:"nonce"    gencodec:"required"`
	Price        *big.Int        `json:"gasPrice" gencodec:"required"`
	GasLimit     uint64          `json:"gas"      gencodec:"required"`
	Recipient    *common.Address `json:"to"       rlp:"nil"` // nil means contract creation
	Amount       *big.Int        `json:"value"    gencodec:"required"`
	Payload      []byte          `json:"input"    gencodec:"required"`

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

type txdataMarshaling struct {
	AccountNonce hexutil.Uint64
	Price        *hexutil.Big
	GasLimit     hexutil.Uint64
	Amount       *hexutil.Big
	Payload      hexutil.Bytes
	V            *hexutil.Big
	R            *hexutil.Big
	S            *hexutil.Big
}

func newEmptyTxInternalDataLegacy() *TxInternalDataLegacy {
	return &TxInternalDataLegacy{}
}

func newTxInternalDataLegacy() *TxInternalDataLegacy {
	return &TxInternalDataLegacy{
		AccountNonce: 0,
		Recipient:    nil,
		Payload:      []byte{},
		Amount:       new(big.Int),
		GasLimit:     0,
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
	}
}

func newTxInternalDataLegacyWithValues(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *TxInternalDataLegacy {
	d := newTxInternalDataLegacy()

	d.AccountNonce = nonce
	d.Recipient = to
	d.GasLimit = gasLimit

	if len(data) > 0 {
		d.Payload = common.CopyBytes(data)
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}

	return d
}

func newTxInternalDataLegacyWithMap(values map[TxValueKeyType]interface{}) (*TxInternalDataLegacy, error) {
	d := newTxInternalDataLegacy()

	if v, ok := values[TxValueKeyNonce].(uint64); ok {
		d.AccountNonce = v
		delete(values, TxValueKeyNonce)
	} else {
		return nil, errValueKeyNonceMustUint64
	}

	if v, ok := values[TxValueKeyTo].(common.Address); ok {
		d.Recipient = &v
		delete(values, TxValueKeyTo)
	} else {
		return nil, errValueKeyToMustAddress
	}

	if v, ok := values[TxValueKeyAmount].(*big.Int); ok {
		d.Amount.Set(v)
		delete(values, TxValueKeyAmount)
	} else {
		return nil, errValueKeyAmountMustBigInt
	}

	if v, ok := values[TxValueKeyData].([]byte); ok {
		d.Payload = common.CopyBytes(v)
		delete(values, TxValueKeyData)
	} else {
		return nil, errValueKeyDataMustByteSlice
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

	if len(values) != 0 {
		for k := range values {
			fmt.Println("unnecessary key", k.String())
		}
		return nil, errUndefinedKeyRemains
	}

	return d, nil
}

func (t *TxInternalDataLegacy) Type() TxType {
	return TxTypeLegacyTransaction
}

func (t *TxInternalDataLegacy) GetRoleTypeForValidation() accountkey.RoleType {
	return accountkey.RoleTransaction
}

func (t *TxInternalDataLegacy) ChainId() *big.Int {
	return deriveChainId(t.V)
}

func (t *TxInternalDataLegacy) Protected() bool {
	return isProtectedV(t.V)
}

func (t *TxInternalDataLegacy) GetAccountNonce() uint64 {
	return t.AccountNonce
}

func (t *TxInternalDataLegacy) GetPrice() *big.Int {
	return new(big.Int).Set(t.Price)
}

func (t *TxInternalDataLegacy) GetGasLimit() uint64 {
	return t.GasLimit
}

func (t *TxInternalDataLegacy) GetRecipient() *common.Address {
	return t.Recipient
}

func (t *TxInternalDataLegacy) GetAmount() *big.Int {
	return new(big.Int).Set(t.Amount)
}

func (t *TxInternalDataLegacy) GetHash() *common.Hash {
	return t.Hash
}

func (t *TxInternalDataLegacy) GetPayload() []byte {
	return t.Payload
}

func (t *TxInternalDataLegacy) SetHash(h *common.Hash) {
	t.Hash = h
}

func (t *TxInternalDataLegacy) SetSignature(s TxSignatures) {
	if len(s) != 1 {
		logger.Crit("LegacyTransaction receives a single signature only!")
	}

	t.V = s[0].V
	t.R = s[0].R
	t.S = s[0].S
}

func (t *TxInternalDataLegacy) RawSignatureValues() []*big.Int {
	return []*big.Int{t.V, t.R, t.S}
}

func (t *TxInternalDataLegacy) ValidateSignature() bool {
	return validateSignature(t.V, t.R, t.S)
}

func (t *TxInternalDataLegacy) RecoverAddress(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) (common.Address, error) {
	V := vfunc(t.V)
	return recoverPlain(txhash, t.R, t.S, V, homestead)
}

func (t *TxInternalDataLegacy) RecoverPubkey(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) ([]*ecdsa.PublicKey, error) {
	V := vfunc(t.V)

	pk, err := recoverPlainPubkey(txhash, t.R, t.S, V, homestead)
	if err != nil {
		return nil, err
	}

	return []*ecdsa.PublicKey{pk}, nil
}

func (t *TxInternalDataLegacy) IntrinsicGas(currentBlockNumber uint64) (uint64, error) {
	return IntrinsicGas(t.Payload, t.Recipient == nil, true)
}

func (t *TxInternalDataLegacy) SerializeForSign() []interface{} {
	return []interface{}{
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.Recipient,
		t.Amount,
		t.Payload,
	}
}

func (t *TxInternalDataLegacy) IsLegacyTransaction() bool {
	return true
}

func (t *TxInternalDataLegacy) equalHash(a *TxInternalDataLegacy) bool {
	if t.GetHash() == nil && a.GetHash() == nil {
		return true
	}

	if t.GetHash() != nil && a.GetHash() != nil &&
		bytes.Equal(t.GetHash().Bytes(), a.GetHash().Bytes()) {
		return true
	}

	return false
}

func (t *TxInternalDataLegacy) equalRecipient(a *TxInternalDataLegacy) bool {
	if t.Recipient == nil && a.Recipient == nil {
		return true
	}

	if t.Recipient != nil && a.Recipient != nil && bytes.Equal(t.Recipient.Bytes(), a.Recipient.Bytes()) {
		return true
	}

	return false
}

func (t *TxInternalDataLegacy) Equal(a TxInternalData) bool {
	ta, ok := a.(*TxInternalDataLegacy)
	if !ok {
		return false
	}

	return t.AccountNonce == ta.AccountNonce &&
		t.Price.Cmp(ta.Price) == 0 &&
		t.GasLimit == ta.GasLimit &&
		t.equalRecipient(ta) &&
		t.Amount.Cmp(ta.Amount) == 0 &&
		t.V.Cmp(ta.V) == 0 &&
		t.R.Cmp(ta.R) == 0 &&
		t.S.Cmp(ta.S) == 0
}

func (t *TxInternalDataLegacy) String() string {
	var from, to string
	tx := &Transaction{data: t}

	v, r, s := t.V, t.R, t.S
	if v != nil {
		// make a best guess about the signer and use that to derive
		// the sender.
		signer := deriveSigner(v)
		if f, err := Sender(signer, tx); err != nil { // derive but don't cache
			from = "[invalid sender: invalid sig]"
		} else {
			from = fmt.Sprintf("%x", f[:])
		}
	} else {
		from = "[invalid sender: nil V field]"
	}

	if t.GetRecipient() == nil {
		to = "[contract creation]"
	} else {
		to = fmt.Sprintf("%x", t.GetRecipient().Bytes())
	}
	enc, _ := rlp.EncodeToBytes(t)
	return fmt.Sprintf(`
	TX(%x)
	Contract: %v
	From:     %s
	To:       %s
	Nonce:    %v
	GasPrice: %#x
	GasLimit  %#x
	Value:    %#x
	Data:     0x%x
	V:        %#x
	R:        %#x
	S:        %#x
	Hex:      %x
`,
		tx.Hash(),
		t.GetRecipient() == nil,
		from,
		to,
		t.GetAccountNonce(),
		t.GetPrice(),
		t.GetGasLimit(),
		t.GetAmount(),
		t.GetPayload(),
		v,
		r,
		s,
		enc,
	)
}

func (t *TxInternalDataLegacy) Validate(stateDB StateDB, currentBlockNumber uint64) error {
	// No more validation required for TxInternalDataLegacy.
	return nil
}

func (t *TxInternalDataLegacy) FillContractAddress(from common.Address, r *Receipt) {
	if t.Recipient == nil {
		codeHash := crypto.Keccak256Hash(t.Payload)
		r.ContractAddress = crypto.CreateAddress(from, t.AccountNonce, codeHash)
	}
}

func (t *TxInternalDataLegacy) Execute(sender ContractRef, vm VM, stateDB StateDB, currentBlockNumber uint64, gas uint64, value *big.Int) (ret []byte, usedGas uint64, err error) {
	if t.Recipient == nil {
		ret, _, usedGas, err = vm.Create(sender, t.Payload, gas, value)
	} else {
		stateDB.IncNonce(sender.Address())
		ret, usedGas, err = vm.Call(sender, *t.Recipient, t.Payload, gas, value)
	}

	return ret, usedGas, err
}

func (t *TxInternalDataLegacy) MakeRPCOutput() map[string]interface{} {
	return map[string]interface{}{
		"type":     t.Type().String(),
		"gas":      hexutil.Uint64(t.GasLimit),
		"gasPrice": (*hexutil.Big)(t.Price),
		"input":    hexutil.Bytes(t.Payload),
		"nonce":    hexutil.Uint64(t.AccountNonce),
		"to":       t.Recipient,
		"value":    (*hexutil.Big)(t.Amount),
	}
}
