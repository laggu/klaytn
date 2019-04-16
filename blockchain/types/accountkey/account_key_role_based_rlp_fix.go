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
	"github.com/ground-x/klaytn/fork"
	"github.com/ground-x/klaytn/kerrors"
	"github.com/ground-x/klaytn/ser/rlp"
	"io"
)

// AccountKeyRoleBasedRlpFix represents a role-based key.
// The difference between AccountKeyRoleBased and AccountKeyRoleBasedRlpFix is RLP encoding.
// Since AccountKeyRoleBased does not support caver-js, AccountKeyRoleBasedRlpFix is introduced.
type AccountKeyRoleBasedRlpFix []AccountKey

func NewAccountKeyRoleBasedRlpFix() *AccountKeyRoleBasedRlpFix {
	return &AccountKeyRoleBasedRlpFix{}
}

func NewAccountKeyRoleBasedRlpFixWithValues(keys []AccountKey) *AccountKeyRoleBasedRlpFix {
	return (*AccountKeyRoleBasedRlpFix)(&keys)
}

func (a *AccountKeyRoleBasedRlpFix) Type() AccountKeyType {
	return AccountKeyTypeRoleBasedRlpFix
}

func (a *AccountKeyRoleBasedRlpFix) IsCompositeType() bool {
	return true
}

func (a *AccountKeyRoleBasedRlpFix) DeepCopy() AccountKey {
	n := make(AccountKeyRoleBasedRlpFix, len(*a))

	for i, k := range *a {
		n[i] = k.DeepCopy()
	}

	return &n
}

func (a *AccountKeyRoleBasedRlpFix) Equal(b AccountKey) bool {
	tb, ok := b.(*AccountKeyRoleBasedRlpFix)
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

func (a *AccountKeyRoleBasedRlpFix) EncodeRLP(w io.Writer) error {
	enc := make([][]byte, len(*a))

	for i, k := range *a {
		enc[i], _ = rlp.EncodeToBytes(NewAccountKeySerializerWithAccountKey(k))
	}

	return rlp.Encode(w, enc)
}

func (a *AccountKeyRoleBasedRlpFix) DecodeRLP(s *rlp.Stream) error {
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

	*a = (AccountKeyRoleBasedRlpFix)(keys)

	return nil
}

func (a *AccountKeyRoleBasedRlpFix) MarshalJSON() ([]byte, error) {
	serializers := make([]*AccountKeySerializer, len(*a))

	for i, k := range *a {
		serializers[i] = NewAccountKeySerializerWithAccountKey(k)
	}

	return json.Marshal(serializers)
}

func (a *AccountKeyRoleBasedRlpFix) UnmarshalJSON(b []byte) error {
	var serializers []*AccountKeySerializer
	if err := json.Unmarshal(b, &serializers); err != nil {
		return err
	}

	*a = make(AccountKeyRoleBasedRlpFix, len(serializers))
	for i, s := range serializers {
		(*a)[i] = s.key
	}

	return nil
}

func (a *AccountKeyRoleBasedRlpFix) Validate(r RoleType, pubkeys []*ecdsa.PublicKey) bool {
	if len(*a) > int(r) {
		return (*a)[r].Validate(r, pubkeys)
	}
	return a.getDefaultKey().Validate(r, pubkeys)
}

// ValidateAccountCreation validates keys when creating an account with this key.
func (a *AccountKeyRoleBasedRlpFix) ValidateAccountCreation() error {
	// 1. RoleTransaction should exist at least.
	if len(*a) < 1 {
		return errKeyLengthZero
	}

	// 2. Prohibited key types are: Nil and RoleBased.
	for _, k := range *a {
		if k.Type() == AccountKeyTypeNil ||
			k.IsCompositeType() {
			return errKeyShouldNotBeNilOrCompositeType
		}
	}

	return nil
}

func (a *AccountKeyRoleBasedRlpFix) getDefaultKey() AccountKey {
	return (*a)[RoleTransaction]
}

func (a *AccountKeyRoleBasedRlpFix) String() string {
	serializer := NewAccountKeySerializerWithAccountKey(a)
	b, _ := json.Marshal(serializer)
	return string(b)
}

func (a *AccountKeyRoleBasedRlpFix) AccountCreationGas(currentBlockNumber uint64) (uint64, error) {
	gas := uint64(0)
	for _, k := range *a {
		gasK, err := k.AccountCreationGas(currentBlockNumber)
		if err != nil {
			return 0, err
		}
		gas += gasK
	}

	return gas, nil
}

func (a *AccountKeyRoleBasedRlpFix) SigValidationGas(currentBlockNumber uint64, r RoleType) (uint64, error) {
	var key AccountKey
	// Set the key used to sign for validation.
	if len(*a) > int(r) {
		key = (*a)[r]
	} else {
		key = a.getDefaultKey()
	}

	gas, err := key.SigValidationGas(currentBlockNumber, r)
	if err != nil {
		return 0, err
	}

	return gas, nil
}

func (a *AccountKeyRoleBasedRlpFix) Init(currentBlockNumber uint64) error {
	if fork.IsRoleBasedRLPFixEnabled(currentBlockNumber) == false {
		return kerrors.ErrNotSupported
	}
	// A zero-role key is not allowed.
	if len(*a) == 0 {
		return kerrors.ErrZeroLength
	}
	// Do not allow undefined roles.
	if len(*a) > (int)(RoleLast) {
		return kerrors.ErrLengthTooLong
	}
	for i := 0; i < len(*a); i++ {
		// A composite key is not allowed.
		if (*a)[i].IsCompositeType() {
			return kerrors.ErrNestedCompositeType
		}

		// If any key in the role cannot be initialized, return an error.
		if err := (*a)[i].Init(currentBlockNumber); err != nil {
			return err
		}
	}

	return nil
}

func (a *AccountKeyRoleBasedRlpFix) Update(key AccountKey, currentBlockNumber uint64) error {
	if fork.IsRoleBasedRLPFixEnabled(currentBlockNumber) == false {
		return kerrors.ErrNotSupported
	}
	if ak, ok := key.(*AccountKeyRoleBasedRlpFix); ok {
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
			// A composite key is not allowed.
			if (*a)[i].IsCompositeType() {
				return kerrors.ErrNestedCompositeType
			}
			// Skip if AccountKeyNil.
			if (*ak)[i].Type() == AccountKeyTypeNil {
				continue
			}
			if err := (*ak)[i].Init(currentBlockNumber); err != nil {
				return err
			}
			(*a)[i] = (*ak)[i]
		}

		return nil
	}

	// Update is not possible if the type is different.
	return kerrors.ErrDifferentAccountKeyType
}
