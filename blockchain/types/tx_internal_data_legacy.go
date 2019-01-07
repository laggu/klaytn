// Copyright 2018 The go-klaytn Authors
// This file is part of the go-klaytn library.
//
// The go-klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-klaytn library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"bytes"
	"github.com/ground-x/go-gxplatform/common"
	"math/big"
)

func newEmptyTxdata() *txdata {
	return &txdata{}
}

func newTxdata(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *txdata {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := txdata{
		AccountNonce: nonce,
		Recipient:    to,
		Payload:      data,
		Amount:       new(big.Int),
		GasLimit:     gasLimit,
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}

	return &d
}

func (t *txdata) Type() TxType {
	return TxTypeLegacyTransaction
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

func (t *txdata) GetFrom() common.Address {
	logger.Crit("Should not be called!")
	return common.Address{}
}

func (t *txdata) GetHash() *common.Hash {
	return t.Hash
}

func (t *txdata) GetPayload() []byte {
	return t.Payload
}

func (t *txdata) GetVRS() (*big.Int, *big.Int, *big.Int) {
	return t.V, t.R, t.S
}

func (t *txdata) GetV() *big.Int {
	return t.V
}

func (t *txdata) GetR() *big.Int {
	return t.R
}

func (t *txdata) GetS() *big.Int {
	return t.S
}

func (t *txdata) SetAccountNonce(n uint64) {
	t.AccountNonce = n
}

func (t *txdata) SetPrice(p *big.Int) {
	t.Price.Set(p)
}

func (t *txdata) SetGasLimit(g uint64) {
	t.GasLimit = g
}

func (t *txdata) SetRecipient(r common.Address) {
	t.Recipient = &r
}

func (t *txdata) SetAmount(a *big.Int) {
	t.Amount.Set(a)
}

func (t *txdata) SetFrom(f common.Address) {
	// DO NOTHING
}

func (t *txdata) SetHash(h *common.Hash) {
	t.Hash = h
}

func (t *txdata) SetPayload(b []byte) {
	t.Payload = b
}

func (t *txdata) SetVRS(v *big.Int, r *big.Int, s *big.Int) {
	t.V.Set(v)
	t.R.Set(r)
	t.S.Set(s)
}

func (t *txdata) SetV(v *big.Int) {
	t.V.Set(v)
}

func (t *txdata) SetR(r *big.Int) {
	t.R.Set(r)
}

func (t *txdata) SetS(s *big.Int) {
	t.S.Set(s)
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
