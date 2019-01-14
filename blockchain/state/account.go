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

package state

import (
	"errors"
	"github.com/ground-x/go-gxplatform/common"
	"math/big"
)

type AccountType uint64

const (
	LegacyAccountType AccountType = iota
	ExternallyOwnedAccountType
	SmartContractAccountType
)

type AccountValueKeyType uint

const (
	AccountValueKeyNonce AccountValueKeyType = iota
	AccountValueKeyBalance
	AccountValueKeyStorageRoot
	AccountValueKeyCodeHash
	AccountValueKeyHumanReadable
	AccountValueKeyAccountKey
)

func (a AccountType) String() string {
	switch a {
	case LegacyAccountType:
		return "LegacyAccountType"
	case ExternallyOwnedAccountType:
		return "ExternallyOwnedAccount"
	case SmartContractAccountType:
		return "SmartContractAccount"
	}
	return "UndefinedAccountType"
}

var (
	ErrUndefinedAccountType = errors.New("undefined account type")
)

// Account is the Klaytn consensus representation of accounts.
// These objects are stored in the main account trie.
type Account interface {
	Type() AccountType

	GetNonce() uint64
	GetBalance() *big.Int
	GetHumanReadable() bool

	SetNonce(n uint64)
	SetBalance(b *big.Int)
	SetHumanReadable(b bool)

	// Empty returns whether the account is considered empty.
	// The "empty" account may be defined differently depending on the actual account type.
	// An example of an empty account could be described as the one that satisfies the following conditions:
	// - nonce is zero
	// - balance is zero
	// - codeHash is the same as emptyCodeHash
	Empty() bool

	// Equal returns true if all the attributes are exactly same. Otherwise, returns false.
	Equal(Account) bool

	// DeepCopy copies all the attributes.
	DeepCopy() Account

	// String returns all attributes of this object as a string.
	String() string
}

// ProgramAccount is an interface of an account having a program (code + storage).
// This interface is implemented by LegacyAccount and SmartContractAccount.
type ProgramAccount interface {
	Account

	GetStorageRoot() common.Hash
	GetCodeHash() []byte

	SetStorageRoot(h common.Hash)
	SetCodeHash(h []byte)
}

// NewAccountWithType creates an Account object with the given type.
func NewAccountWithType(t AccountType) (Account, error) {
	switch t {
	case LegacyAccountType:
		return newLegacyAccount(), nil
	case ExternallyOwnedAccountType:
		return newExternallyOwnedAccount(), nil
	}

	return nil, ErrUndefinedAccountType
}

// NewAccountWithType creates an Account object initialized with the given map.
func NewAccountWithMap(t AccountType, values map[AccountValueKeyType]interface{}) (Account, error) {
	switch t {
	case LegacyAccountType:
		return newLegacyAccountWithMap(values), nil
	case ExternallyOwnedAccountType:
		return newExternallyOwnedAccountWithMap(values), nil
	}

	return nil, ErrUndefinedAccountType
}
