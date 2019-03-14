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
	"errors"
	"github.com/ground-x/klaytn/blockchain/types/account"
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/kerrors"
	"github.com/ground-x/klaytn/params"
	"math"
	"math/big"
)

// MaxFeeRatio is the maximum value of feeRatio. Since it is represented in percentage,
// the maximum value is 100.
const MaxFeeRatio uint8 = 100

const SubTxTypeBits uint = 3

type TxType uint8

const (
	// TxType declarations
	// There are three type declarations at each line:
	//   <base type>, <fee-delegated type>, and <fee-delegated type with a fee ratio>
	// If types other than <base type> are not useful, they are declared with underscore(_).
	// Each base type is self-descriptive.
	TxTypeLegacyTransaction, TxTypeFeeDelegatedTransactions, TxTypeFeeDelegatedWithRatioTransactions TxType = iota << SubTxTypeBits, iota<<SubTxTypeBits + 1, iota<<SubTxTypeBits + 2
	TxTypeValueTransfer, TxTypeFeeDelegatedValueTransfer, TxTypeFeeDelegatedValueTransferWithRatio
	TxTypeValueTransferMemo, TxTypeFeeDelegatedValueTransferMemo, TxTypeFeeDelegatedValueTransferMemoWithRatio
	TxTypeAccountCreation, _, _
	TxTypeAccountUpdate, TxTypeFeeDelegatedAccountUpdate, TxTypeFeeDelegatedAccountUpdateWithRatio
	TxTypeSmartContractDeploy, TxTypeFeeDelegatedSmartContractDeploy, TxTypeFeeDelegatedSmartContractDeployWithRatio
	TxTypeSmartContractExecution, TxTypeFeeDelegatedSmartContractExecution, TxTypeFeeDelegatedSmartContractExecutionWithRatio
	TxTypeCancel, TxTypeFeeDelegatedCancel, TxTypeFeeDelegatedCancelWithRatio
	TxTypeBatch, _, _
	TxTypeChainDataAnchoring, _, _
	TxTypeLast, _, _
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

	errValueKeyHumanReadableMustBool     = errors.New("HumanReadable must be a type of bool")
	errValueKeyAccountKeyMustAccountKey  = errors.New("AccountKey must be a type of AccountKey")
	errValueKeyAnchoredDataMustByteSlice = errors.New("AnchoredData must be a slice of bytes")
	errValueKeyNonceMustUint64           = errors.New("Nonce must be a type of uint64")
	errValueKeyToMustAddress             = errors.New("To must be a type of common.Address")
	errValueKeyAmountMustBigInt          = errors.New("Amount must be a type of *big.Int")
	errValueKeyGasLimitMustUint64        = errors.New("GasLimit must be a type of uint64")
	errValueKeyGasPriceMustBigInt        = errors.New("GasPrice must be a type of *big.Int")
	errValueKeyFromMustAddress           = errors.New("From must be a type of common.Address")
	errValueKeyFeePayerMustAddress       = errors.New("FeePayer must be a type of common.Address")
	errValueKeyDataMustByteSlice         = errors.New("Data must be a slice of bytes")
	errValueKeyFeeRatioMustUint8         = errors.New("FeeRatio must be a type of uint8")
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
		return "TxTypeFeeDelegatedValueTransferWithRatio"
	case TxTypeValueTransferMemo:
		return "TxTypeValueTransferMemo"
	case TxTypeFeeDelegatedValueTransferMemo:
		return "TxTypeFeeDelegatedValueTransferMemo"
	case TxTypeFeeDelegatedValueTransferMemoWithRatio:
		return "TxTypeFeeDelegatedValueTransferMemoWithRatio"
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

func (t TxType) IsAccountUpdate() bool {
	return (t &^ ((1 << SubTxTypeBits) - 1)) == TxTypeAccountUpdate
}

func (t TxType) IsContractDeploy() bool {
	return (t &^ ((1 << SubTxTypeBits) - 1)) == TxTypeSmartContractDeploy
}

func (t TxType) IsCancelTransaction() bool {
	return (t &^ ((1 << SubTxTypeBits) - 1)) == TxTypeCancel
}

func (t TxType) IsLegacyTransaction() bool {
	return t == TxTypeLegacyTransaction
}

func (t TxType) IsFeeDelegatedTransaction() bool {
	return (t & (TxTypeFeeDelegatedTransactions | TxTypeFeeDelegatedWithRatioTransactions)) != 0x0
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

	SetHash(*common.Hash)
	SetSignature(TxSignatures)

	// RawSignatureValues returns signatures as a slice of `*big.Int`.
	// Due to multi signatures, it is not good to return three values of `*big.Int`.
	// The format would be something like [v, r, s, v, r, s].
	RawSignatureValues() []*big.Int

	// ValidateSignature returns true if the signature is valid.
	ValidateSignature() bool

	// RecoverAddress returns address derived from txhash and signatures(r, s, v).
	// Since EIP155Signer modifies V value during recovering while other signers don't, it requires vfunc for the treatment.
	RecoverAddress(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) (common.Address, error)

	// RecoverPubkey returns a public key derived from txhash and signatures(r, s, v).
	// Since EIP155Signer modifies V value during recovering while other signers don't, it requires vfunc for the treatment.
	RecoverPubkey(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) ([]*ecdsa.PublicKey, error)

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

	// Validate returns nil if tx is validated with the given stateDB.
	// Otherwise, it returns an error.
	// This function is called in TxPool.validateTx() and TxInternalData.Execute().
	Validate(stateDB StateDB) error

	// IsLegacyTransaction returns true if the tx type is a legacy transaction (TxInternalDataLegacy) object.
	IsLegacyTransaction() bool

	// GetRoleTypeForValidation returns RoleType to validate this transaction.
	GetRoleTypeForValidation() accountkey.RoleType

	// String returns a string containing information about the fields of the object.
	String() string

	// Execute performs execution of the transaction according to the transaction type.
	Execute(sender ContractRef, vm VM, stateDB StateDB, gas uint64, value *big.Int) (ret []byte, usedGas uint64, err error)

	MakeRPCOutput() map[string]interface{}
}

type TxInternalDataSerializeForSignToByte interface {
	SerializeForSignToBytes() []byte
}

// TxInternalDataFeePayer has functions related to fee delegated transactions.
type TxInternalDataFeePayer interface {
	GetFeePayer() common.Address

	// GetFeePayerRawSignatureValues returns fee payer's signatures as a slice of `*big.Int`.
	// Due to multi signatures, it is not good to return three values of `*big.Int`.
	// The format would be something like [v, r, s, v, r, s].
	GetFeePayerRawSignatureValues() []*big.Int

	// RecoverFeePayerPubkey returns the fee payer's public key derived from txhash and signatures(r, s, v).
	RecoverFeePayerPubkey(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) ([]*ecdsa.PublicKey, error)

	SetFeePayerSignature(s TxSignatures)
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
// `TxInternalDataLegacy` (a legacy transaction type) does not have the field.
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

// Since we cannot access the package `blockchain/vm` directly, an interface `VM` is introduced.
// TODO-Klaytn-Refactoring: Transaction and related data structures should be a new package.
type VM interface {
	Create(caller ContractRef, code []byte, gas uint64, value *big.Int) (ret []byte, contractAddr common.Address, leftOverGas uint64, err error)
	CreateWithAddress(caller ContractRef, code []byte, gas uint64, value *big.Int, contractAddr common.Address, humanReadable bool) ([]byte, common.Address, uint64, error)
	Call(caller ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error)
}

// Since we cannot access the package `blockchain/state` directly, an interface `StateDB` is introduced.
// TODO-Klaytn-Refactoring: Transaction and related data structures should be a new package.
type StateDB interface {
	IncNonce(common.Address)
	Exist(common.Address) bool
	UpdateKey(addr common.Address, key accountkey.AccountKey) error
	CreateAccountWithMap(addr common.Address, accountType account.AccountType, values map[account.AccountValueKeyType]interface{})
	IsProgramAccount(addr common.Address) bool
}

func NewTxInternalData(t TxType) (TxInternalData, error) {
	switch t {
	case TxTypeLegacyTransaction:
		return newTxInternalDataLegacy(), nil
	case TxTypeValueTransfer:
		return newTxInternalDataValueTransfer(), nil
	case TxTypeFeeDelegatedValueTransfer:
		return newTxInternalDataFeeDelegatedValueTransfer(), nil
	case TxTypeFeeDelegatedValueTransferWithRatio:
		return NewTxInternalDataFeeDelegatedValueTransferWithRatio(), nil
	case TxTypeValueTransferMemo:
		return newTxInternalDataValueTransferMemo(), nil
	case TxTypeFeeDelegatedValueTransferMemo:
		return newTxInternalDataFeeDelegatedValueTransferMemo(), nil
	case TxTypeFeeDelegatedValueTransferMemoWithRatio:
		return newTxInternalDataFeeDelegatedValueTransferMemoWithRatio(), nil
	case TxTypeAccountCreation:
		return newTxInternalDataAccountCreation(), nil
	case TxTypeAccountUpdate:
		return newTxInternalDataAccountUpdate(), nil
	case TxTypeFeeDelegatedAccountUpdate:
		return newTxInternalDataFeeDelegatedAccountUpdate(), nil
	case TxTypeFeeDelegatedAccountUpdateWithRatio:
		return newTxInternalDataFeeDelegatedAccountUpdateWithRatio(), nil
	case TxTypeSmartContractDeploy:
		return newTxInternalDataSmartContractDeploy(), nil
	case TxTypeFeeDelegatedSmartContractDeploy:
		return newTxInternalDataFeeDelegatedSmartContractDeploy(), nil
	case TxTypeFeeDelegatedSmartContractDeployWithRatio:
		return newTxInternalDataFeeDelegatedSmartContractDeployWithRatio(), nil
	case TxTypeSmartContractExecution:
		return newTxInternalDataSmartContractExecution(), nil
	case TxTypeFeeDelegatedSmartContractExecution:
		return newTxInternalDataFeeDelegatedSmartContractExecution(), nil
	case TxTypeFeeDelegatedSmartContractExecutionWithRatio:
		return newTxInternalDataFeeDelegatedSmartContractExecutionWithRatio(), nil
	case TxTypeCancel:
		return newTxInternalDataCancel(), nil
	case TxTypeFeeDelegatedCancel:
		return newTxInternalDataFeeDelegatedCancel(), nil
	case TxTypeFeeDelegatedCancelWithRatio:
		return newTxInternalDataFeeDelegatedCancelWithRatio(), nil
	case TxTypeChainDataAnchoring:
		return newTxInternalDataChainDataAnchoring(), nil
	}

	return nil, errUndefinedTxType
}

func NewTxInternalDataWithMap(t TxType, values map[TxValueKeyType]interface{}) (TxInternalData, error) {
	switch t {
	case TxTypeLegacyTransaction:
		return newTxInternalDataLegacyWithMap(values)
	case TxTypeValueTransfer:
		return newTxInternalDataValueTransferWithMap(values)
	case TxTypeFeeDelegatedValueTransfer:
		return newTxInternalDataFeeDelegatedValueTransferWithMap(values)
	case TxTypeFeeDelegatedValueTransferWithRatio:
		return newTxInternalDataFeeDelegatedValueTransferWithRatioWithMap(values)
	case TxTypeValueTransferMemo:
		return newTxInternalDataValueTransferMemoWithMap(values)
	case TxTypeFeeDelegatedValueTransferMemo:
		return newTxInternalDataFeeDelegatedValueTransferMemoWithMap(values)
	case TxTypeFeeDelegatedValueTransferMemoWithRatio:
		return newTxInternalDataFeeDelegatedValueTransferMemoWithRatioWithMap(values)
	case TxTypeAccountCreation:
		return newTxInternalDataAccountCreationWithMap(values)
	case TxTypeAccountUpdate:
		return newTxInternalDataAccountUpdateWithMap(values)
	case TxTypeFeeDelegatedAccountUpdate:
		return newTxInternalDataFeeDelegatedAccountUpdateWithMap(values)
	case TxTypeFeeDelegatedAccountUpdateWithRatio:
		return newTxInternalDataFeeDelegatedAccountUpdateWithRatioWithMap(values)
	case TxTypeSmartContractDeploy:
		return newTxInternalDataSmartContractDeployWithMap(values)
	case TxTypeFeeDelegatedSmartContractDeploy:
		return newTxInternalDataFeeDelegatedSmartContractDeployWithMap(values)
	case TxTypeFeeDelegatedSmartContractDeployWithRatio:
		return newTxInternalDataFeeDelegatedSmartContractDeployWithRatioWithMap(values)
	case TxTypeSmartContractExecution:
		return newTxInternalDataSmartContractExecutionWithMap(values)
	case TxTypeFeeDelegatedSmartContractExecution:
		return newTxInternalDataFeeDelegatedSmartContractExecutionWithMap(values)
	case TxTypeFeeDelegatedSmartContractExecutionWithRatio:
		return newTxInternalDataFeeDelegatedSmartContractExecutionWithRatioWithMap(values)
	case TxTypeCancel:
		return newTxInternalDataCancelWithMap(values)
	case TxTypeFeeDelegatedCancel:
		return newTxInternalDataFeeDelegatedCancelWithMap(values)
	case TxTypeFeeDelegatedCancelWithRatio:
		return newTxInternalDataFeeDelegatedCancelWithRatioWithMap(values)
	case TxTypeChainDataAnchoring:
		return newTxInternalDataChainDataAnchoringWithMap(values)
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
