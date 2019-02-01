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
	"fmt"
	"github.com/ground-x/klaytn/common"
	"math/big"
)

// TxInternalDataCommon is a common data structure for new types of transactions.
// Its fields are used for all new transaction types.
type TxInternalDataCommon struct {
	AccountNonce uint64
	Price        *big.Int
	GasLimit     uint64
	Recipient    common.Address
	Amount       *big.Int
	From         common.Address

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

// newTxInternalDataCommon creates an empty TxInternalDataCommon object with initializing *big.Int variables.
func newTxInternalDataCommon() *TxInternalDataCommon {
	return &TxInternalDataCommon{
		Price:  new(big.Int),
		Amount: new(big.Int),
	}
}

// newTxInternalDataCommonWithMap creates an TxInternalDataCommon object and initializes it with given attributes in the map.
func newTxInternalDataCommonWithMap(values map[TxValueKeyType]interface{}) (*TxInternalDataCommon, error) {
	d := newTxInternalDataCommon()

	if v, ok := values[TxValueKeyNonce].(uint64); ok {
		d.AccountNonce = v
	} else {
		return nil, errValueKeyNonceMustUint64
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

	h := common.Hash{}
	d.Hash = &h

	return d, nil
}

func (t *TxInternalDataCommon) GetAccountNonce() uint64 {
	return t.AccountNonce
}

func (t *TxInternalDataCommon) GetPrice() *big.Int {
	return new(big.Int).Set(t.Price)
}

func (t *TxInternalDataCommon) GetGasLimit() uint64 {
	return t.GasLimit
}

func (t *TxInternalDataCommon) GetRecipient() *common.Address {
	if t.Recipient == (common.Address{}) {
		return nil
	}

	to := common.Address(t.Recipient)
	return &to
}

func (t *TxInternalDataCommon) GetAmount() *big.Int {
	return new(big.Int).Set(t.Amount)
}

func (t *TxInternalDataCommon) GetFrom() common.Address {
	return t.From
}

func (t *TxInternalDataCommon) GetHash() *common.Hash {
	return t.Hash
}

func (t *TxInternalDataCommon) SetHash(h *common.Hash) {
	t.Hash = h
}

func (t *TxInternalDataCommon) serializeForSign() []interface{} {
	return []interface{}{
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.Recipient,
		t.Amount,
		t.From,
	}
}

func (t *TxInternalDataCommon) equal(b *TxInternalDataCommon) bool {
	return t.AccountNonce == b.AccountNonce &&
		t.Price.Cmp(b.Price) == 0 &&
		t.GasLimit == b.GasLimit &&
		t.Recipient == b.Recipient &&
		t.Amount.Cmp(b.Amount) == 0 &&
		t.From == b.From
}

func (t *TxInternalDataCommon) intrinsicGas() (uint64, error) {
	return IntrinsicGas([]byte{}, false, true)
}

func (t *TxInternalDataCommon) IsLegacyTransaction() bool {
	return false
}

func (t *TxInternalDataCommon) string() string {
	return fmt.Sprintf(`
	From:          %s
	To:            %s
	Nonce:         %v
	GasPrice:      %#x
	GasLimit:      %#x
	Value:         %#x`,
		t.From.String(),
		t.Recipient.String(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.Amount,
	)
}
