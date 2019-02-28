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
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
)

// TxInternalDataFeeDelegatedSmartContractDeploy represents a fee-delegated transaction creating a smart contract.
type TxInternalDataFeeDelegatedSmartContractDeploy struct {
	AccountNonce  uint64
	Price         *big.Int
	GasLimit      uint64
	Recipient     common.Address
	Amount        *big.Int
	From          common.Address
	Payload       []byte
	HumanReadable bool

	TxSignatures

	FeePayer          common.Address
	FeePayerSignature TxSignatures

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

func newTxInternalDataFeeDelegatedSmartContractDeploy() *TxInternalDataFeeDelegatedSmartContractDeploy {
	h := common.Hash{}
	return &TxInternalDataFeeDelegatedSmartContractDeploy{
		Price:  new(big.Int),
		Amount: new(big.Int),
		Hash:   &h,
	}
}

func newTxInternalDataFeeDelegatedSmartContractDeployWithMap(values map[TxValueKeyType]interface{}) (*TxInternalDataFeeDelegatedSmartContractDeploy, error) {
	t := newTxInternalDataFeeDelegatedSmartContractDeploy()

	if v, ok := values[TxValueKeyNonce].(uint64); ok {
		t.AccountNonce = v
	} else {
		return nil, errValueKeyNonceMustUint64
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

	if v, ok := values[TxValueKeyGasLimit].(uint64); ok {
		t.GasLimit = v
	} else {
		return nil, errValueKeyGasLimitMustUint64
	}

	if v, ok := values[TxValueKeyGasPrice].(*big.Int); ok {
		t.Price.Set(v)
	} else {
		return nil, errValueKeyGasPriceMustBigInt
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

	if v, ok := values[TxValueKeyHumanReadable].(bool); ok {
		t.HumanReadable = v
	} else {
		return nil, errValueKeyHumanReadableMustBool
	}

	if v, ok := values[TxValueKeyFeePayer].(common.Address); ok {
		t.FeePayer = v
	} else {
		return nil, errValueKeyFeePayerMustAddress
	}

	return t, nil
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) Type() TxType {
	return TxTypeFeeDelegatedSmartContractDeploy
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) GetRoleTypeForValidation() accountkey.RoleType {
	return accountkey.RoleTransaction
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) GetPayload() []byte {
	return t.Payload
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) Equal(a TxInternalData) bool {
	ta, ok := a.(*TxInternalDataFeeDelegatedSmartContractDeploy)
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
		t.HumanReadable == ta.HumanReadable &&
		t.TxSignatures.equal(ta.TxSignatures) &&
		t.FeePayer == ta.FeePayer &&
		t.FeePayerSignature.equal(ta.FeePayerSignature)
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) IsLegacyTransaction() bool {
	return false
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) GetAccountNonce() uint64 {
	return t.AccountNonce
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) GetPrice() *big.Int {
	return new(big.Int).Set(t.Price)
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) GetGasLimit() uint64 {
	return t.GasLimit
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) GetRecipient() *common.Address {
	if t.Recipient == (common.Address{}) {
		return nil
	}

	to := common.Address(t.Recipient)
	return &to
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) GetAmount() *big.Int {
	return new(big.Int).Set(t.Amount)
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) GetFrom() common.Address {
	return t.From
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) GetHash() *common.Hash {
	return t.Hash
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) GetFeePayer() common.Address {
	return t.FeePayer
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) GetFeePayerRawSignatureValues() []*big.Int {
	return t.FeePayerSignature.RawSignatureValues()
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) SetHash(h *common.Hash) {
	t.Hash = h
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) SetSignature(s TxSignatures) {
	t.TxSignatures = s
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) SetFeePayerSignature(s TxSignatures) {
	t.FeePayerSignature = s
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) RecoverFeePayerPubkey(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) ([]*ecdsa.PublicKey, error) {
	return t.FeePayerSignature.RecoverPubkey(txhash, homestead, vfunc)
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) String() string {
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
	Paylod:        %x
	HumanReadable: %v
	Signature:     %s
	FeePayer:      %s
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
		common.Bytes2Hex(t.Payload),
		t.HumanReadable,
		t.TxSignatures.string(),
		t.FeePayer.String(),
		t.FeePayerSignature.string(),
		enc)

}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) IntrinsicGas() (uint64, error) {
	gas := params.TxGasContractCreation + params.TxGasFeeDelegated

	gasPayload, err := intrinsicGasPayload(t.Payload)
	if err != nil {
		return 0, err
	}

	return gas + gasPayload, nil
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) SerializeForSignToBytes() []byte {
	b, _ := rlp.EncodeToBytes(struct {
		Txtype        TxType
		AccountNonce  uint64
		Price         *big.Int
		GasLimit      uint64
		Recipient     common.Address
		Amount        *big.Int
		From          common.Address
		Payload       []byte
		HumanReadable bool
	}{
		t.Type(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.Recipient,
		t.Amount,
		t.From,
		t.Payload,
		t.HumanReadable,
	})

	return b
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) SerializeForSign() []interface{} {
	return []interface{}{
		t.Type(),
		t.AccountNonce,
		t.Price,
		t.GasLimit,
		t.Recipient,
		t.Amount,
		t.From,
		t.Payload,
		t.HumanReadable,
	}
}

func (t *TxInternalDataFeeDelegatedSmartContractDeploy) Execute(sender ContractRef, vm VM, stateDB StateDB, gas uint64, value *big.Int) (ret []byte, usedGas uint64, err, vmerr error) {
	ret, _, usedGas, vmerr = vm.CreateWithAddress(sender, t.Payload, gas, value, t.Recipient, t.HumanReadable)

	return
}
