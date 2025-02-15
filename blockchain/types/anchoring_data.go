// Copyright 2019 The klaytn Authors
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
	"github.com/klaytn/klaytn/common"
	"github.com/klaytn/klaytn/ser/rlp"
	"math/big"
)

const (
	AnchoringDataType0 uint8 = 0
)

type AnchoringData struct {
	Type uint8
	Data []byte
}

// AnchoringDataLegacy is an old anchoring type that does not support an data type.
type AnchoringDataLegacy struct {
	BlockHash     common.Hash
	TxHash        common.Hash
	ParentHash    common.Hash
	ReceiptHash   common.Hash
	StateRootHash common.Hash
	BlockNumber   *big.Int
}

type AnchoringDataInternalType0 struct {
	BlockHash     common.Hash
	TxHash        common.Hash
	ParentHash    common.Hash
	ReceiptHash   common.Hash
	StateRootHash common.Hash
	BlockNumber   *big.Int
	Period        *big.Int
	TxCount       *big.Int
}

func NewAnchoringDataType0(block *Block, period *big.Int, txCount *big.Int) (*AnchoringData, error) {
	data := &AnchoringDataInternalType0{block.Hash(), block.Header().TxHash,
		block.Header().ParentHash, block.Header().ReceiptHash,
		block.Header().Root, block.Header().Number, period, txCount}
	encodedCCTxData, err := rlp.EncodeToBytes(data)
	if err != nil {
		return nil, err
	}
	return &AnchoringData{AnchoringDataType0, encodedCCTxData}, nil
}
