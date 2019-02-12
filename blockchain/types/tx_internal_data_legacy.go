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
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/hexutil"
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
)

//go:generate gencodec -type txdata -field-override txdataMarshaling -out gen_tx_json.go

type txdata struct {
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

func newEmptyTxdata() *txdata {
	return &txdata{}
}

func newTxdata() *txdata {
	return &txdata{
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

func newTxdataWithValues(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *txdata {
	d := newTxdata()

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

func newTxdataWithMap(values map[TxValueKeyType]interface{}) (*txdata, error) {
	d := newTxdata()

	if v, ok := values[TxValueKeyNonce].(uint64); ok {
		d.AccountNonce = v
	} else {
		return nil, errValueKeyNonceMustUint64
	}

	if v, ok := values[TxValueKeyTo].(common.Address); ok {
		d.Recipient = &v
	} else {
		return nil, errValueKeyToMustAddress
	}

	if v, ok := values[TxValueKeyAmount].(*big.Int); ok {
		d.Amount.Set(v)
	} else {
		return nil, errValueKeyAmountMustBigInt
	}

	if v, ok := values[TxValueKeyData].([]byte); ok {
		d.Payload = common.CopyBytes(v)
	} else {
		return nil, errValueKeyDataMustByteSlice
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

	return d, nil
}

func (t *txdata) Type() TxType {
	return TxTypeLegacyTransaction
}

func (t *txdata) ChainId() *big.Int {
	return deriveChainId(t.V)
}

func (t *txdata) Protected() bool {
	return isProtectedV(t.V)
}

func (t *txdata) GetAccountNonce() uint64 {
	return t.AccountNonce
}

func (t *txdata) GetPrice() *big.Int {
	return new(big.Int).Set(t.Price)
}

func (t *txdata) GetGasLimit() uint64 {
	return t.GasLimit
}

func (t *txdata) GetRecipient() *common.Address {
	return t.Recipient
}

func (t *txdata) GetAmount() *big.Int {
	return new(big.Int).Set(t.Amount)
}

func (t *txdata) GetHash() *common.Hash {
	return t.Hash
}

func (t *txdata) GetPayload() []byte {
	return t.Payload
}

func (t *txdata) SetHash(h *common.Hash) {
	t.Hash = h
}

func (t *txdata) SetSignature(s TxSignatures) {
	if len(s) != 1 {
		logger.Crit("LegacyTransaction receives a single signature only!")
	}

	t.V = s[0].V
	t.R = s[0].R
	t.S = s[0].S
}

func (t *txdata) RawSignatureValues() []*big.Int {
	return []*big.Int{t.V, t.R, t.S}
}

func (t *txdata) ValidateSignature() bool {
	return validateSignature(t.V, t.R, t.S)
}

func (t *txdata) RecoverAddress(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) (common.Address, error) {
	V := vfunc(t.V)
	return recoverPlain(txhash, t.R, t.S, V, homestead)
}

func (t *txdata) RecoverPubkey(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) ([]*ecdsa.PublicKey, error) {
	V := vfunc(t.V)

	pk, err := recoverPlainPubkey(txhash, t.R, t.S, V, homestead)
	if err != nil {
		return nil, err
	}

	return []*ecdsa.PublicKey{pk}, nil
}

func (t *txdata) IntrinsicGas() (uint64, error) {
	return IntrinsicGas(t.Payload, t.Recipient == nil, true)
}

func (t *txdata) SerializeForSign() []interface{} {
	return []interface{}{
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.Recipient,
		t.Amount,
		t.Payload,
	}
}

func (t *txdata) IsLegacyTransaction() bool {
	return true
}

func (t *txdata) equalHash(a *txdata) bool {
	if t.GetHash() == nil && a.GetHash() == nil {
		return true
	}

	if t.GetHash() != nil && a.GetHash() != nil &&
		bytes.Equal(t.GetHash().Bytes(), a.GetHash().Bytes()) {
		return true
	}

	return false
}

func (t *txdata) equalRecipient(a *txdata) bool {
	if t.Recipient == nil && a.Recipient == nil {
		return true
	}

	if t.Recipient != nil && a.Recipient != nil && bytes.Equal(t.Recipient.Bytes(), a.Recipient.Bytes()) {
		return true
	}

	return false
}

func (t *txdata) Equal(a TxInternalData) bool {
	ta, ok := a.(*txdata)
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

func (t *txdata) String() string {
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
