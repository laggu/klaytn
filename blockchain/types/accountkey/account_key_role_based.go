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
	"encoding/json"
	"errors"
	"github.com/ground-x/klaytn/kerrors"
	"github.com/ground-x/klaytn/ser/rlp"
	"io"
)

type RoleType int

const (
	RoleTransaction RoleType = iota
	RoleAccountUpdate
	RoleFeePayer
	// TODO-Klaytn-Accounts: more roles can be listed here.
	RoleLast
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
	enc := make([][]byte, len(*a))

	for i, k := range *a {
		enc[i], _ = rlp.EncodeToBytes(NewAccountKeySerializerWithAccountKey(k))
	}

	return rlp.Encode(w, enc)
}

func (a *AccountKeyRoleBased) DecodeRLP(s *rlp.Stream) error {
	enc := [][]byte{}
	if err := s.Decode(&enc); err != nil {
		return err
	}

	keys := make([]AccountKey, len(enc))
	for i, b := range enc {
		serializer := NewAccountKeySerializer()
		if err := rlp.DecodeBytes(b, &serializer); err != nil {
			return err
		}
		keys[i] = serializer.key
	}

	*a = (AccountKeyRoleBased)(keys)

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

func (a *AccountKeyRoleBased) Validate(r RoleType, pubkeys []*ecdsa.PublicKey) bool {
	if len(*a) > int(RoleAccountUpdate) {
		return (*a)[r].Validate(r, pubkeys)
	}
	return a.getDefaultKey().Validate(r, pubkeys)
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

func (a *AccountKeyRoleBased) Init() error {
	// A zero-role key is not allowed.
	if len(*a) == 0 {
		return kerrors.ErrZeroLength
	}
	// Do not allow undefined roles.
	if len(*a) > (int)(RoleLast) {
		return kerrors.ErrLengthTooLong
	}
	for i := 0; i < len(*a); i++ {
		// A nested role-based key is not allowed.
		if _, ok := (*a)[i].(*AccountKeyRoleBased); ok {
			return kerrors.ErrNestedRoleBasedKey
		}

		// If any key in the role cannot be initialized, return an error.
		if err := (*a)[i].Init(); err != nil {
			return err
		}
	}

	return nil
}

func (a *AccountKeyRoleBased) Update(key AccountKey) error {
	if ak, ok := key.(*AccountKeyRoleBased); ok {
		lenAk := len(*ak)
		lenA := len(*a)
		// If no key is to be replaced, it is regarded as a fail.
		if lenAk == 0 {
			return kerrors.ErrZeroLength
		}
		// Do not allow undefined roles.
		if lenAk > (int)(RoleLast) {
			return kerrors.ErrLengthTooLong
		}
		if lenA < lenAk {
			*a = append(*a, (*ak)[lenA:]...)
		}
		for i := 0; i < lenAk; i++ {
			if _, ok := (*ak)[i].(*AccountKeyRoleBased); ok {
				// A nested role-based key is not allowed.
				return kerrors.ErrNestedRoleBasedKey
			}
			// Skip if AccountKeyNil.
			if (*ak)[i].Type() == AccountKeyTypeNil {
				continue
			}
			if err := (*ak)[i].Init(); err != nil {
				return err
			}
			(*a)[i] = (*ak)[i]
		}

		return nil
	}

	// Update is not possible if the type is different.
	return kerrors.ErrDifferentAccountKeyType
}
