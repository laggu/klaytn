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
	"errors"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/kerrors"
	"github.com/ground-x/klaytn/params"
	"math"
	"math/big"
)

// MaxFeeRatio is the maximum value of feeRatio. Since it is represented in percentage,
// the maximum value is 100.
const MaxFeeRatio uint8 = 100

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
	TxTypeChainDataAnchoring, _, _
)

type TxValueKeyType uint

const (
	TxValueKeyNonce TxValueKeyType = iota
	TxValueKeyTo
	TxValueKeyAmount
	TxValueKeyGasLimit
	TxValueKeyGasPrice
	TxValueKeyData
	TxValueKeyFrom
	TxValueKeyAnchoredData
	TxValueKeyHumanReadable
	TxValueKeyAccountKey
	TxValueKeyFeePayer
	TxValueKeyFeeRatioOfFeePayer
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
	case TxTypeChainDataAnchoring:
		return "TxTypeChainDataAnchoring"
	}

	return "UndefinedTxType"
}

func (t TxType) IsAccountCreation() bool {
	return t == TxTypeAccountCreation
}

func (t TxType) IsContractDeploy() bool {
	return t == TxTypeSmartContractDeploy ||
		t == TxTypeFeeDelegatedSmartContractDeploy ||
		t == TxTypeFeeDelegatedSmartContractDeployWithRatio
}

func (t TxType) IsLegacyTransaction() bool {
	return t == TxTypeLegacyTransaction
}

// TxInternalData is an interface for an internal data structure of a Transaction
type TxInternalData interface {
	Type() TxType

	GetAccountNonce() uint64
	GetPrice() *big.Int
	GetGasLimit() uint64
	GetRecipient() *common.Address
	GetAmount() *big.Int
	GetHash() *common.Hash
	GetVRS() (*big.Int, *big.Int, *big.Int)

	SetHash(*common.Hash)
	SetSignature(*TxSignature)
	SetVRS(*big.Int, *big.Int, *big.Int)

	// ChainId returns which chain id this transaction was signed for (if at all)
	ChainId() *big.Int

	// Protected returns whether the transaction is protected from replay protection.
	Protected() bool

	// Equal returns true if all attributes are the same.
	Equal(t TxInternalData) bool

	// IntrinsicGas computes additional 'intrinsic gas' based on tx types.
	IntrinsicGas() (uint64, error)

	// SerializeForSign returns a slice containing attributes to make its tx signature.
	SerializeForSign() []interface{}

	// IsLegacyTransaction returns true if the tx type is a legacy transaction (txdata) object.
	IsLegacyTransaction() bool

	// String returns a string containing information about the fields of the object.
	String() string
}

// TxInternalDataFeePayer has functions related to fee delegated transactions.
type TxInternalDataFeePayer interface {
	GetFeePayer() common.Address
	GetFeePayerVRS() (*big.Int, *big.Int, *big.Int)

	SetFeePayerSignature(s *TxSignature)
}

// TxInternalDataFeeRatio has a function `GetFeeRatio`.
type TxInternalDataFeeRatio interface {
	// GetFeeRatio returns a ratio of tx fee paid by the fee payer in percentage.
	// For example, if it is 30, 30% of tx fee will be paid by the fee payer.
	// 70% will be paid by the sender.
	GetFeeRatio() uint8
}

// TxInternalDataFrom has a function `GetFrom()`.
// All other transactions to be implemented will have `from` field, but
// `txdata` (a legacy transaction type) does not have the field.
// Hence, this function is defined in another interface TxInternalDataFrom.
type TxInternalDataFrom interface {
	GetFrom() common.Address
}

// TxInternalDataPayload has a function `GetPayload()`.
// Since the payload field is not a common field for all tx types, we provide
// an interface `TxInternalDataPayload` to obtain the payload.
type TxInternalDataPayload interface {
	GetPayload() []byte
}

func NewTxInternalData(t TxType) (TxInternalData, error) {
	switch t {
	case TxTypeLegacyTransaction:
		return newTxdata(), nil
	case TxTypeValueTransfer:
		return newTxInternalDataValueTransfer(), nil
	case TxTypeFeeDelegatedValueTransfer:
		return NewTxInternalDataFeeDelegatedValueTransfer(), nil
	case TxTypeFeeDelegatedValueTransferWithRatio:
		return NewTxInternalDataFeeDelegatedValueTransferWithRatio(), nil
	case TxTypeAccountCreation:
		return newTxInternalDataAccountCreation(), nil
	case TxTypeSmartContractDeploy:
		return newTxInternalDataSmartContractDeploy(), nil
	case TxTypeSmartContractExecution:
		return newTxInternalDataSmartContractExecution(), nil
	case TxTypeChainDataAnchoring:
		return newTxInternalDataChainDataAnchoring(), nil
	}

	return nil, errUndefinedTxType
}

func NewTxInternalDataWithMap(t TxType, values map[TxValueKeyType]interface{}) (TxInternalData, error) {
	switch t {
	case TxTypeLegacyTransaction:
		return newTxdataWithMap(values), nil
	case TxTypeValueTransfer:
		return newTxInternalDataValueTransferWithMap(values), nil
	case TxTypeFeeDelegatedValueTransfer:
		return NewTxInternalDataFeeDelegatedValueTransferWithMap(values), nil
	case TxTypeFeeDelegatedValueTransferWithRatio:
		return NewTxInternalDataFeeDelegatedValueTransferWithRatioWithMap(values), nil
	case TxTypeAccountCreation:
		return newTxInternalDataAccountCreationWithMap(values), nil
	case TxTypeSmartContractDeploy:
		return newTxInternalDataSmartContractDeployWithMap(values), nil
	case TxTypeSmartContractExecution:
		return newTxInternalDataSmartContractExecutionWithMap(values), nil
	case TxTypeChainDataAnchoring:
		return newTxInternalDataChainDataAnchoringWithMap(values), nil
	}

	return nil, errUndefinedTxType
}

func intrinsicGasPayload(data []byte) (uint64, error) {
	gas := uint64(0)
	// Bump the required gas by the amount of transactional data
	if len(data) > 0 {
		// Zero and non-zero bytes are priced differently
		var nz uint64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		// Make sure we don't exceed uint64 for all data combinations
		if (math.MaxUint64-gas)/params.TxDataNonZeroGas < nz {
			return 0, kerrors.ErrOutOfGas
		}
		gas += nz * params.TxDataNonZeroGas

		z := uint64(len(data)) - nz
		if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
			return 0, kerrors.ErrOutOfGas
		}
		gas += z * params.TxDataZeroGas
	}

	return gas, nil
}

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func IntrinsicGas(data []byte, contractCreation, homestead bool) (uint64, error) {
	// Set the starting gas for the raw transaction
	var gas uint64
	if contractCreation && homestead {
		gas = params.TxGasContractCreation
	} else {
		gas = params.TxGas
	}
	gasPayload, err := intrinsicGasPayload(data)
	if err != nil {
		return 0, err
	}

	return gas + gasPayload, nil
}

// CalcFeeWithRatio returns feePayer's fee and sender's fee based on feeRatio.
// For example, if fee = 100 and feeRatio = 30, feePayer = 30 and feeSender = 70.
func CalcFeeWithRatio(feeRatio uint8, fee *big.Int) (*big.Int, *big.Int) {
	// feePayer = fee * ratio / 100
	feePayer := new(big.Int).Div(new(big.Int).Mul(fee, new(big.Int).SetUint64(uint64(feeRatio))), common.Big100)
	// feeSender = fee - feePayer
	feeSender := new(big.Int).Sub(fee, feePayer)

	return feePayer, feeSender
}
