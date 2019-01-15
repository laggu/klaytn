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
	"encoding/json"
	"github.com/ground-x/klaytn/ser/rlp"
	"io"
)

// TxInternalDataSerializer serializes an object that implements `TxInternalData`.
type TxInternalDataSerializer struct {
	txType TxType
	tx     TxInternalData
}

// txInternalDataJSON is an internal object for JSON serialization.
type txInternalDataJSON struct {
	TxType TxType
	Tx     json.RawMessage
}

// newTxInternalDataSerializer creates a new TxInternalDataSerializer object with the given TxInternalData object.
func newTxInternalDataSerializer(t TxInternalData) *TxInternalDataSerializer {
	return &TxInternalDataSerializer{t.Type(), t}
}

// newEmptyTxInternalDataSerializer creates an empty TxInternalDataSerializer object for decoding.
func newEmptyTxInternalDataSerializer() *TxInternalDataSerializer {
	return &TxInternalDataSerializer{}
}

func (serializer *TxInternalDataSerializer) EncodeRLP(w io.Writer) error {
	// if it is the original transaction, do not encode type.
	if serializer.txType == TxTypeLegacyTransaction {
		return rlp.Encode(w, serializer.tx)
	}

	if err := rlp.Encode(w, serializer.txType); err != nil {
		return err
	}
	return rlp.Encode(w, serializer.tx)
}

func (serializer *TxInternalDataSerializer) DecodeRLP(s *rlp.Stream) error {
	if err := s.Decode(&serializer.txType); err != nil {
		// fallback to the original transaction decoding.
		txd := newEmptyTxdata()
		if err := s.Decode(txd); err != nil {
			return err
		}
		serializer.txType = TxTypeLegacyTransaction
		serializer.tx = txd
		return nil
	}

	var err error
	serializer.tx, err = NewTxInternalData(serializer.txType)
	if err != nil {
		return err
	}

	return s.Decode(serializer.tx)
}

func (serializer *TxInternalDataSerializer) MarshalJSON() ([]byte, error) {
	// if it is the original transaction, do not marshal type.
	if serializer.txType == TxTypeLegacyTransaction {
		return json.Marshal(serializer.tx)
	}
	b, err := json.Marshal(serializer.tx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&txInternalDataJSON{serializer.txType, json.RawMessage(b)})
}

func (serializer *TxInternalDataSerializer) UnmarshalJSON(b []byte) error {
	var dec txInternalDataJSON

	if err := json.Unmarshal(b, &dec); err != nil {
		return err
	}

	if len(dec.Tx) == 0 {
		// fallback to unmarshal the legacy transaction.
		txd := newEmptyTxdata()
		if err := json.Unmarshal(b, txd); err != nil {
			return err
		}
		serializer.txType = TxTypeLegacyTransaction
		serializer.tx = txd

		return nil
	}

	serializer.txType = dec.TxType

	var err error
	serializer.tx, err = NewTxInternalData(serializer.txType)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(dec.Tx), serializer.tx)
}
