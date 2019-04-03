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

package account

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/ser/rlp"
	"io"
)

// SmartContractAccount represents a smart contract account containing
// storage root and code hash.
type SmartContractAccount struct {
	*AccountCommon
	storageRoot common.Hash // merkle root of the storage trie
	codeHash    []byte
}

// smartContractAccountSerializable is an internal data structure for RLP serialization.
// This structure inherits accountCommonSerializable.
type smartContractAccountSerializable struct {
	CommonSerializable *accountCommonSerializable
	StorageRoot        common.Hash
	CodeHash           []byte
}

func newSmartContractAccount() *SmartContractAccount {
	return &SmartContractAccount{
		newAccountCommon(),
		common.Hash{},
		emptyCodeHash,
	}
}

func newSmartContractAccountWithMap(values map[AccountValueKeyType]interface{}) *SmartContractAccount {
	sca := &SmartContractAccount{
		newAccountCommonWithMap(values),
		common.Hash{},
		emptyCodeHash,
	}

	if v, ok := values[AccountValueKeyStorageRoot].(common.Hash); ok {
		sca.storageRoot = v
	}

	if v, ok := values[AccountValueKeyCodeHash].([]byte); ok {
		sca.codeHash = v
	}

	return sca
}

func newSmartContractAccountSerializable() *smartContractAccountSerializable {
	return &smartContractAccountSerializable{
		CommonSerializable: newAccountCommonSerializable(),
	}
}

func (sca *SmartContractAccount) toSerializable() *smartContractAccountSerializable {
	return &smartContractAccountSerializable{
		CommonSerializable: sca.AccountCommon.toSerializable(),
		StorageRoot:        sca.storageRoot,
		CodeHash:           sca.codeHash,
	}
}

func (sca *SmartContractAccount) fromSerializable(o *smartContractAccountSerializable) {
	sca.AccountCommon.fromSerializable(o.CommonSerializable)
	sca.storageRoot = o.StorageRoot
	sca.codeHash = o.CodeHash
}

func (sca *SmartContractAccount) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, sca.toSerializable())
}

func (sca *SmartContractAccount) DecodeRLP(s *rlp.Stream) error {
	serialized := &smartContractAccountSerializable{
		newAccountCommonSerializable(),
		common.Hash{},
		[]byte{},
	}

	if err := s.Decode(serialized); err != nil {
		return err
	}

	sca.fromSerializable(serialized)

	return nil
}

func (sca *SmartContractAccount) MarshalJSON() ([]byte, error) {
	return json.Marshal(sca.toSerializable())
}

func (sca *SmartContractAccount) UnmarshalJSON(b []byte) error {
	serialized := newSmartContractAccountSerializable()

	if err := json.Unmarshal(b, serialized); err != nil {
		return err
	}

	sca.fromSerializable(serialized)

	return nil
}

func (sca *SmartContractAccount) Type() AccountType {
	return SmartContractAccountType
}

func (sca *SmartContractAccount) GetStorageRoot() common.Hash {
	return sca.storageRoot
}

func (sca *SmartContractAccount) GetCodeHash() []byte {
	return sca.codeHash
}

func (sca *SmartContractAccount) SetStorageRoot(h common.Hash) {
	sca.storageRoot = h
}

func (sca *SmartContractAccount) SetCodeHash(h []byte) {
	sca.codeHash = h
}

func (sca *SmartContractAccount) Empty() bool {
	return sca.AccountCommon.Empty() && bytes.Equal(sca.codeHash, emptyCodeHash)
}

func (sca *SmartContractAccount) UpdateKey(key accountkey.AccountKey, currentBlockNumber uint64) error {
	return ErrAccountKeyNotModifiable
}

func (sca *SmartContractAccount) Equal(a Account) bool {
	sca2, ok := a.(*SmartContractAccount)
	if !ok {
		return false
	}

	return sca.AccountCommon.Equal(sca2.AccountCommon) &&
		sca.storageRoot == sca2.storageRoot &&
		bytes.Equal(sca.codeHash, sca2.codeHash)
}

func (sca *SmartContractAccount) DeepCopy() Account {
	return &SmartContractAccount{
		AccountCommon: sca.AccountCommon.DeepCopy(),
		storageRoot:   sca.storageRoot,
		codeHash:      common.CopyBytes(sca.codeHash),
	}
}

func (sca *SmartContractAccount) String() string {
	return fmt.Sprintf(`Common:%s
	StorageRoot: %s
	CodeHash: %s`,
		sca.AccountCommon.String(),
		sca.storageRoot.String(),
		common.Bytes2Hex(sca.codeHash))
}
