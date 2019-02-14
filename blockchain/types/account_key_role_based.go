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
	"encoding/json"
	"errors"
	"github.com/ground-x/klaytn/ser/rlp"
	"io"
)

type RoleType int

const (
	RoleTransaction RoleType = iota
	RoleAccountUpdate
	RoleFeePayer
	// TODO-Klaytn-Accounts: more roles can be listed here.
)

var (
	errKeyLengthZero                = errors.New("key length is zero")
	errKeyShouldNotBeNilOrRoleBased = errors.New("key should not be nil or rolebased")
)

// AccountKeyRoleBased represents a role-based key.
// The roles are defined like below:
// RoleTransaction   - this key is used to verify transactions transferring values.
// RoleAccountUpdate - this key is used to update keys in the account when using TxTypeAccountUpdate.
// RoleFeePayer      - this key is used to pay tx fee when using fee-delegated transactions.
//                     If an account has a key of this role and wants to pay tx fee,
//                     fee-delegated transactions should be signed by this key.
//
// If RoleAccountUpdate or RoleFeePayer is not set, RoleTransaction will be used instead by default.
type AccountKeyRoleBased []AccountKey

func NewAccountKeyRoleBased() *AccountKeyRoleBased {
	return &AccountKeyRoleBased{}
}

func NewAccountKeyRoleBasedWithValues(keys []AccountKey) *AccountKeyRoleBased {
	return (*AccountKeyRoleBased)(&keys)
}

func (a *AccountKeyRoleBased) Type() AccountKeyType {
	return AccountKeyTypeRoleBased
}

func (a *AccountKeyRoleBased) DeepCopy() AccountKey {
	n := make(AccountKeyRoleBased, len(*a))

	for i, k := range *a {
		n[i] = k.DeepCopy()
	}

	return &n
}

func (a *AccountKeyRoleBased) Equal(b AccountKey) bool {
	tb, ok := b.(*AccountKeyRoleBased)
	if !ok {
		return false
	}

	if len(*a) != len(*tb) {
		return false
	}

	for i, tbi := range *tb {
		if (*a)[i].Equal(tbi) == false {
			return false
		}
	}

	return true
}

func (a *AccountKeyRoleBased) EncodeRLP(w io.Writer) error {
	serializers := make([]*AccountKeySerializer, len(*a))

	for i, k := range *a {
		serializers[i] = NewAccountKeySerializerWithAccountKey(k)
	}

	return rlp.Encode(w, serializers)
}

func (a *AccountKeyRoleBased) DecodeRLP(s *rlp.Stream) error {
	serializers := []*AccountKeySerializer{}
	if err := s.Decode(&serializers); err != nil {
		return err
	}
	*a = make(AccountKeyRoleBased, len(serializers))
	for i, s := range serializers {
		(*a)[i] = s.key
	}

	return nil
}

func (a *AccountKeyRoleBased) MarshalJSON() ([]byte, error) {
	serializers := make([]*AccountKeySerializer, len(*a))

	for i, k := range *a {
		serializers[i] = NewAccountKeySerializerWithAccountKey(k)
	}

	return json.Marshal(serializers)
}

func (a *AccountKeyRoleBased) UnmarshalJSON(b []byte) error {
	var serializers []*AccountKeySerializer
	if err := json.Unmarshal(b, &serializers); err != nil {
		return err
	}

	*a = make(AccountKeyRoleBased, len(serializers))
	for i, s := range serializers {
		(*a)[i] = s.key
	}

	return nil
}

func (a *AccountKeyRoleBased) Validate(pubkeys []*ecdsa.PublicKey) bool {
	return a.getDefaultKey().Validate(pubkeys)
}

func (a *AccountKeyRoleBased) ValidateWithTxType(txtype TxType, pubkeys []*ecdsa.PublicKey) bool {
	if txtype == TxTypeAccountUpdate {
		if len(*a) > int(RoleAccountUpdate) {
			return (*a)[RoleAccountUpdate].Validate(pubkeys)
		}
	}
	// Fallback to the default key if the above conditions are not met.
	return a.getDefaultKey().Validate(pubkeys)
}

func (a *AccountKeyRoleBased) ValidateFeePayer(pubkeys []*ecdsa.PublicKey) bool {
	return a.getDefaultKey().Validate(pubkeys)
}

func (a *AccountKeyRoleBased) ValidateFeePayerWithTxType(txtype TxType, pubkeys []*ecdsa.PublicKey) bool {
	if txtype.IsFeeDelegatedTransaction() {
		// If the above condition is passed, it means the tx is a fee-delegated tx type.
		if len(*a) > int(RoleFeePayer) {
			return (*a)[RoleFeePayer].Validate(pubkeys)
		}
	}

	// Fallback to the default key if the above conditions are not met.
	return a.getDefaultKey().Validate(pubkeys)
}

// ValidateAccountCreation validates keys when creating an account with this key.
func (a *AccountKeyRoleBased) ValidateAccountCreation() error {
	// 1. RoleTransaction should exist at least.
	if len(*a) < 1 {
		return errKeyLengthZero
	}

	// 2. Prohibited key types are: Nil and RoleBased.
	for _, k := range *a {
		if k.Type() == AccountKeyTypeNil ||
			k.Type() == AccountKeyTypeRoleBased {
			return errKeyShouldNotBeNilOrRoleBased
		}
	}

	return nil
}

func (a *AccountKeyRoleBased) getDefaultKey() AccountKey {
	return (*a)[RoleTransaction]
}

func (a *AccountKeyRoleBased) String() string {
	serializer := NewAccountKeySerializerWithAccountKey(a)
	b, _ := json.Marshal(serializer)
	return string(b)
}

func (a *AccountKeyRoleBased) Update(key *AccountKeyRoleBased) error {
	for i, k := range *key {
		// Update only if the key is not an AccountKeyNil object.
		if kk, ok := k.(*AccountKeyNil); !ok {
			(*a)[i] = kk
		}
	}

	return nil
}

func (a *AccountKeyRoleBased) AccountCreationGas() (uint64, error) {
	gas := uint64(0)
	for _, k := range *a {
		gasK, err := k.AccountCreationGas()
		if err != nil {
			return 0, err
		}
		gas += gasK
	}

	return gas, nil
}

func (a *AccountKeyRoleBased) SigValidationGas() (uint64, error) {
	gas := uint64(0)
	for _, k := range *a {
		gasK, err := k.SigValidationGas()
		if err != nil {
			return 0, err
		}
		gas += gasK
	}

	return gas, nil
}
