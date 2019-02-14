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
	"fmt"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
)

// TxInternalDataSmartContractExecution represents a transaction executing a smart contract.
type TxInternalDataSmartContractExecution struct {
	AccountNonce uint64
	Price        *big.Int
	GasLimit     uint64
	Recipient    common.Address
	Amount       *big.Int
	From         common.Address
	Payload      []byte

	TxSignatures

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

func newTxInternalDataSmartContractExecution() *TxInternalDataSmartContractExecution {
	h := common.Hash{}
	return &TxInternalDataSmartContractExecution{
		Price:  new(big.Int),
		Amount: new(big.Int),
		Hash:   &h,
	}
}

func newTxInternalDataSmartContractExecutionWithMap(values map[TxValueKeyType]interface{}) (*TxInternalDataSmartContractExecution, error) {
	t := newTxInternalDataSmartContractExecution()

	if v, ok := values[TxValueKeyNonce].(uint64); ok {
		t.AccountNonce = v
	} else {
		return nil, errValueKeyNonceMustUint64
	}

	if v, ok := values[TxValueKeyGasPrice].(*big.Int); ok {
		t.Price.Set(v)
	} else {
		return nil, errValueKeyGasPriceMustBigInt
	}

	if v, ok := values[TxValueKeyGasLimit].(uint64); ok {
		t.GasLimit = v
	} else {
		return nil, errValueKeyGasLimitMustUint64
	}

	if v, ok := values[TxValueKeyTo].(common.Address); ok {
		t.Recipient = v
	} else {
		return nil, errValueKeyToMustAddress
	}

	if v, ok := values[TxValueKeyAmount].(*big.Int); ok {
		t.Amount.Set(v)
	} else {
		return nil, errValueKeyAmountMustBigInt
	}

	if v, ok := values[TxValueKeyFrom].(common.Address); ok {
		t.From = v
	} else {
		return nil, errValueKeyFromMustAddress
	}

	if v, ok := values[TxValueKeyData].([]byte); ok {
		t.Payload = v
	} else {
		return nil, errValueKeyDataMustByteSlice
	}

	return t, nil
}

func (t *TxInternalDataSmartContractExecution) Type() TxType {
	return TxTypeSmartContractExecution
}

func (t *TxInternalDataSmartContractExecution) Equal(a TxInternalData) bool {
	ta, ok := a.(*TxInternalDataSmartContractExecution)
	if !ok {
		return false
	}

	return t.AccountNonce == ta.AccountNonce &&
		t.Price.Cmp(ta.Price) == 0 &&
		t.GasLimit == ta.GasLimit &&
		t.Recipient == ta.Recipient &&
		t.Amount.Cmp(ta.Amount) == 0 &&
		t.From == ta.From &&
		bytes.Equal(t.Payload, ta.Payload) &&
		t.TxSignatures.equal(ta.TxSignatures)
}

func (t *TxInternalDataSmartContractExecution) IsLegacyTransaction() bool {
	return false
}

func (t *TxInternalDataSmartContractExecution) GetPayload() []byte {
	return t.Payload
}

func (t *TxInternalDataSmartContractExecution) GetAccountNonce() uint64 {
	return t.AccountNonce
}

func (t *TxInternalDataSmartContractExecution) GetPrice() *big.Int {
	return new(big.Int).Set(t.Price)
}

func (t *TxInternalDataSmartContractExecution) GetGasLimit() uint64 {
	return t.GasLimit
}

func (t *TxInternalDataSmartContractExecution) GetRecipient() *common.Address {
	if t.Recipient == (common.Address{}) {
		return nil
	}

	to := common.Address(t.Recipient)
	return &to
}

func (t *TxInternalDataSmartContractExecution) GetAmount() *big.Int {
	return new(big.Int).Set(t.Amount)
}

func (t *TxInternalDataSmartContractExecution) GetFrom() common.Address {
	return t.From
}

func (t *TxInternalDataSmartContractExecution) GetHash() *common.Hash {
	return t.Hash
}

func (t *TxInternalDataSmartContractExecution) SetHash(h *common.Hash) {
	t.Hash = h
}

func (t *TxInternalDataSmartContractExecution) SetSignature(s TxSignatures) {
	t.TxSignatures = s
}

func (t *TxInternalDataSmartContractExecution) String() string {
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
		common.Bytes2Hex(t.Payload),
		enc)
}

func (t *TxInternalDataSmartContractExecution) IntrinsicGas() (uint64, error) {
	gas := params.TxGas

	gasPayload, err := intrinsicGasPayload(t.Payload)
	if err != nil {
		return 0, err
	}

	return gas + gasPayload, nil
}

func (t *TxInternalDataSmartContractExecution) SerializeForSign() []interface{} {
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
