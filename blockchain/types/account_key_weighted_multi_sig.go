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
	"github.com/ground-x/klaytn/kerrors"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
)

const (
	// TODO-Klaytn-MultiSig: Need to fix the maximum number of keys allowed for an account.
	// NOTE-Klaytn-MultiSig: This value should not be reduced. If it is reduced, there is a case:
	// - the tx validation will be failed if the sender has larger keys.
	MaxNumKeysForMultiSig = uint64(10)
)

// AccountKeyWeightedMultiSig is an account key type containing a threshold and `WeightedPublicKeys`.
// `WeightedPublicKeys` contains a slice of {weight and key}.
// To be a valid tx for an account associated with `AccountKeyWeightedMultiSig`,
// the weighted sum of signed public keys should be larger than the threshold.
// Refer to AccountKeyWeightedMultiSig.Validate().
type AccountKeyWeightedMultiSig struct {
	Threshold uint
	Keys      WeightedPublicKeys
}

func NewAccountKeyWeightedMultiSig() *AccountKeyWeightedMultiSig {
	return &AccountKeyWeightedMultiSig{}
}

func NewAccountKeyWeightedMultiSigWithValues(threshold uint, keys WeightedPublicKeys) *AccountKeyWeightedMultiSig {
	return &AccountKeyWeightedMultiSig{threshold, keys}
}

func (a *AccountKeyWeightedMultiSig) Type() AccountKeyType {
	return AccountKeyTypeWeightedMultiSig
}

func (a *AccountKeyWeightedMultiSig) DeepCopy() AccountKey {
	return &AccountKeyWeightedMultiSig{
		a.Threshold, a.Keys.DeepCopy(),
	}
}

func (a *AccountKeyWeightedMultiSig) Equal(b AccountKey) bool {
	tb, ok := b.(*AccountKeyWeightedMultiSig)
	if !ok {
		return false
	}

	return a.Threshold == tb.Threshold &&
		a.Keys.Equal(tb.Keys)
}

func (a *AccountKeyWeightedMultiSig) Validate(pubkeys []*ecdsa.PublicKey) bool {
	weightedSum := uint(0)

	// To prohibit making a signature with the same key, make a map.
	// TODO-Klaytn: find another way for better performance
	pMap := make(map[string]*ecdsa.PublicKey)
	for _, bk := range pubkeys {
		b, err := rlp.EncodeToBytes((*PublicKeySerializable)(bk))
		if err != nil {
			logger.Warn("Failed to encode public keys in the tx", pubkeys)
			continue
		}
		pMap[string(b)] = bk
	}

	for _, k := range a.Keys {
		b, err := rlp.EncodeToBytes(k.Key)
		if err != nil {
			logger.Warn("Failed to encode public keys in the account", "AccountKey", a.String())
			continue
		}

		if _, ok := pMap[string(b)]; ok {
			weightedSum += k.Weight
		}
	}

	if weightedSum >= a.Threshold {
		return true
	}

	logger.Debug("AccountKeyWeightedMultiSig validation is failed", "pubkeys", pubkeys,
		"accountKeys", a.String(), "threshold", a.Threshold, "weighted sum", weightedSum)

	return false
}

func (a *AccountKeyWeightedMultiSig) String() string {
	serializer := NewAccountKeySerializerWithAccountKey(a)
	b, _ := json.Marshal(serializer)
	return string(b)
}

func (a *AccountKeyWeightedMultiSig) AccountCreationGas() (uint64, error) {
	numKeys := uint64(len(a.Keys))
	if numKeys > MaxNumKeysForMultiSig {
		return 0, kerrors.ErrMaxKeysExceed
	}
	return params.TxAccountCreationGasDefault + numKeys*params.TxAccountCreationGasPerKey, nil
}

func (a *AccountKeyWeightedMultiSig) SigValidationGas() (uint64, error) {
	numKeys := uint64(len(a.Keys))
	if numKeys > MaxNumKeysForMultiSig {
		logger.Warn("validation failed due to the number of keys in the account is larger than the limit.",
			"account", a.String())
		return 0, kerrors.ErrMaxKeysExceedInValidation
	}
	return params.TxValidationGasDefault + numKeys*params.TxValidationGasPerKey, nil
}

// WeightedPublicKey contains a public key and its weight.
// The weight is used to check whether the weighted sum of public keys are larger than
// the threshold of the AccountKeyWeightedMultiSig object.
type WeightedPublicKey struct {
	Weight uint
	Key    *PublicKeySerializable
}

func (w *WeightedPublicKey) Equal(b *WeightedPublicKey) bool {
	return w.Weight == b.Weight &&
		w.Key.Equal(b.Key)
}

func NewWeightedPublicKey(weight uint, key *PublicKeySerializable) *WeightedPublicKey {
	return &WeightedPublicKey{weight, key}
}

// WeightedPublicKeys is a slice of WeightedPublicKey objects.
type WeightedPublicKeys []*WeightedPublicKey

func (w WeightedPublicKeys) DeepCopy() WeightedPublicKeys {
	keys := make(WeightedPublicKeys, len(w))

	for i, v := range w {
		keys[i] = NewWeightedPublicKey(v.Weight, v.Key.DeepCopy())
	}

	return keys
}

func (w WeightedPublicKeys) Equal(b WeightedPublicKeys) bool {
	if len(w) != len(b) {
		return false
	}

	for i, wv := range w {
		if !wv.Equal(b[i]) {
			return false
		}
	}

	return true
}
