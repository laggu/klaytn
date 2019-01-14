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

// AccountKeyFail is used to prevent smart contract accounts from withdrawing tokens
// from themselves with a public key recovery mechanism.
// Klaytn assumes that the only way to take tokens from smart contract account is using
// `transfer()` in the smart contract code.
type AccountKeyFail struct {
}

var globalFailKey = &AccountKeyFail{}

// NewAccountKeyFail creates a new AccountKeyFail object.
// Since AccountKeyFail has no attributes, use one global variable for all allocations.
func NewAccountKeyFail() *AccountKeyFail { return globalFailKey }

func (a *AccountKeyFail) Type() AccountKeyType {
	return AccountKeyTypeFail
}

func (a *AccountKeyFail) Equal(b AccountKey) bool {
	// This type of account key always returns false.
	return false
}

func (a *AccountKeyFail) String() string {
	return "AccountKeyFail"
}

func (a *AccountKeyFail) DeepCopy() AccountKey {
	return NewAccountKeyFail()
}
