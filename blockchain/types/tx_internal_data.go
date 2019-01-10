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
	"errors"
	"github.com/ground-x/go-gxplatform/common"
	"math/big"
)

type TxType uint

const (
	// TxType declarations
	// There are three type declarations at each line:
	//   <base type>, <fee-delegated type>, and <fee-delegated type with a fee ratio>
	// If types other than <base type> are not useful, they are declared with underscore(_).
	// Each base type is self-descriptive.
	TxTypeLegacyTransaction, _, _ TxType = (0x10 * iota), (0x10*iota + 1), (0x10*iota + 2)
	TxTypeValueTransfer, TxTypeFeeDelegatedValueTransfer, TxTypeFeeDelegatedValueTransferWithRatio
	TxTypeAccountCreation, _, _
	TxTypeAccountUpdate, TxTypeFeeDelegatedAccountUpdate, TxTypeFeeDelegatedAccountUpdateWithRatio
	TxTypeSmartContractDeploy, TxTypeFeeDelegatedSmartContractDeploy, TxTypeFeeDelegatedSmartContractDeployWithRatio
	TxTypeSmartContractExecution, TxTypeFeeDelegatedSmartContractExecution, TxTypeFeeDelegatedSmartContractExecutionWithRatio
	TxTypeCancel, TxTypeFeeDelegatedCancel, TxTypeFeeDelegatedCancelWithRatio
	TxTypeBatch, _, _
)

type TxValueKeyType uint

const (
	TxValueKeyNonce TxValueKeyType = iota
	TxValueKeyTo
	TxValueKeyAmount
	TxValueKeyGasLimit
	TxValueKeyGasPrice
	TxValueKeyData
)

var (
	errNotTxTypeValueTransfer                 = errors.New("not value transfer transaction type")
	errNotTxTypeValueTransferWithFeeDelegator = errors.New("not a fee-delegated value transfer transaction")
	errNotTxTypeAccountCreation               = errors.New("not account creation transaction type")
	errUndefinedTxType                        = errors.New("undefined tx type")
	errCannotBeSignedByFeeDelegator           = errors.New("this transaction type cannot be signed by a fee delegator")
)

func (t TxType) String() string {
	switch t {
	case TxTypeLegacyTransaction:
		return "TxTypeLegacyTransaction"
	case TxTypeValueTransfer:
		return "TxTypeValueTransfer"
	case TxTypeFeeDelegatedValueTransfer:
		return "TxTypeFeeDelegatedValueTransfer"
	case TxTypeFeeDelegatedValueTransferWithRatio:
		return "TxTypeValueTransferWithFeeRatio"
	case TxTypeAccountCreation:
		return "TxTypeAccountCreation"
	case TxTypeAccountUpdate:
		return "TxTypeAccountUpdate"
	case TxTypeFeeDelegatedAccountUpdate:
		return "TxTypeFeeDelegatedAccountUpdate"
	case TxTypeFeeDelegatedAccountUpdateWithRatio:
		return "TxTypeFeeDelegatedAccountUpdateWithRatio"
	case TxTypeSmartContractDeploy:
		return "TxTypeSmartContractDeploy"
	case TxTypeFeeDelegatedSmartContractDeploy:
		return "TxTypeFeeDelegatedSmartContractDeploy"
	case TxTypeFeeDelegatedSmartContractDeployWithRatio:
		return "TxTypeFeeDelegatedSmartContractDeployWithRatio"
	case TxTypeSmartContractExecution:
		return "TxTypeSmartContractExecution"
	case TxTypeFeeDelegatedSmartContractExecution:
		return "TxTypeFeeDelegatedSmartContractExecution"
	case TxTypeFeeDelegatedSmartContractExecutionWithRatio:
		return "TxTypeFeeDelegatedSmartContractExecutionWithRatio"
	case TxTypeCancel:
		return "TxTypeCancel"
	case TxTypeFeeDelegatedCancel:
		return "TxTypeFeeDelegatedCancel"
	case TxTypeFeeDelegatedCancelWithRatio:
		return "TxTypeFeeDelegatedCancelWithRatio"
	case TxTypeBatch:
		return "TxTypeBatch"
	}

	return "UndefinedTxType"
}

// TxInternalData is an interface for an internal data structure of a Transaction
type TxInternalData interface {
	Type() TxType

	GetAccountNonce() uint64
	GetPrice() *big.Int
	GetGasLimit() uint64
	GetRecipient() *common.Address
	GetAmount() *big.Int
	GetPayload() []byte
	GetFrom() common.Address
	GetHash() *common.Hash
	GetVRS() (*big.Int, *big.Int, *big.Int)
	GetV() *big.Int
	GetR() *big.Int
	GetS() *big.Int

	SetAccountNonce(uint64)
	SetPrice(*big.Int)
	SetGasLimit(uint64)
	SetRecipient(common.Address)
	SetAmount(*big.Int)
	SetPayload([]byte)
	SetFrom(common.Address)
	SetHash(*common.Hash)
	SetVRS(*big.Int, *big.Int, *big.Int)
	SetV(*big.Int)
	SetR(*big.Int)
	SetS(*big.Int)

	// Equal returns true if all attributes are the same.
	Equal(t TxInternalData) bool
}

func NewTxInternalData(t TxType) (TxInternalData, error) {
	switch t {
	case TxTypeLegacyTransaction:
		return newTxdata(), nil
	}

	return nil, errUndefinedTxType
}

func NewTxInternalDataWithMap(t TxType, values map[TxValueKeyType]interface{}) (TxInternalData, error) {
	switch t {
	case TxTypeLegacyTransaction:
		return newTxdataWithMap(values), nil
	}

	return nil, errUndefinedTxType
}
