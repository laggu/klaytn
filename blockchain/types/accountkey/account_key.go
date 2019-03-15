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

package accountkey

import (
	"crypto/ecdsa"
	"errors"
	"github.com/ground-x/klaytn/log"
)

type AccountKeyType uint8

const (
	AccountKeyTypeNil AccountKeyType = iota
	AccountKeyTypeLegacy
	AccountKeyTypePublic
	AccountKeyTypeFail
	AccountKeyTypeWeightedMultiSig
	AccountKeyTypeRoleBased
)

var (
	errUndefinedAccountKeyType = errors.New("undefined account key type")
)

var logger = log.NewModuleLogger(log.BlockchainTypesAccountKey)

func (a AccountKeyType) IsLegacyAccountKey() bool {
	return a == AccountKeyTypeLegacy
}

// AccountKey is a common interface to exploit polymorphism of AccountKey.
// Currently, we have the following implementations of AccountKey:
// - AccountKeyLegacy
// - AccountKeyPublic
type AccountKey interface {
	// Type returns the type of account key.
	Type() AccountKeyType

	// String returns a string containing all the attributes of the object.
	String() string

	// Equal returns true if all the attributes are the same. Otherwise, it returns false.
	Equal(AccountKey) bool

	// Validate returns true if the given public keys are verifiable with the AccountKey.
	Validate(RoleType, []*ecdsa.PublicKey) bool

	// DeepCopy creates a new object and copies all the attributes to the new object.
	DeepCopy() AccountKey

	// AccountCreationGas returns gas required to create an account with the corresponding key.
	AccountCreationGas() (uint64, error)

	// SigValidationGas returns gas required to validate a tx with the account.
	SigValidationGas() (uint64, error)

	// Init returns an error if all data in the key is invalid.
	// This checks that the key is ready to be assigned to an account.
	Init() error

	// Update returns an error if `key` cannot be assigned to itself.
	Update(key AccountKey) error
}

func NewAccountKey(t AccountKeyType) (AccountKey, error) {
	switch t {
	case AccountKeyTypeNil:
		return NewAccountKeyNil(), nil
	case AccountKeyTypeLegacy:
		return NewAccountKeyLegacy(), nil
	case AccountKeyTypePublic:
		return NewAccountKeyPublic(), nil
	case AccountKeyTypeFail:
		return NewAccountKeyFail(), nil
	case AccountKeyTypeWeightedMultiSig:
		return NewAccountKeyWeightedMultiSig(), nil
	case AccountKeyTypeRoleBased:
		return NewAccountKeyRoleBased(), nil
	}

	return nil, errUndefinedAccountKeyType
}
