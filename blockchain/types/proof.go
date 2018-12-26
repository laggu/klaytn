// Copyright 2018 The go-klaytn Authors
//
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

import (
	"errors"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/ser/rlp"
	"io"
	"math/big"
	"strings"
)

var (
	ProofDigest     = common.HexToHash("0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365")
	ErrInvalidProof = errors.New("invalid proof message data")
)

type Proof struct {
	Solver      common.Address
	BlockNumber *big.Int
	Nonce       uint64
}

// EncodeRLP serializes ist into the Ethereum RLP format.
func (pr *Proof) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		pr.Solver,
		pr.BlockNumber,
		pr.Nonce,
	})
}

// DecodeRLP implements rlp.Decoder, and load the istanbul fields from a RLP stream.
func (pr *Proof) DecodeRLP(s *rlp.Stream) error {
	var proof struct {
		Solver      common.Address
		BlockNumber *big.Int
		Nonce       uint64
	}
	if err := s.Decode(&proof); err != nil {
		return err
	}
	pr.Solver, pr.BlockNumber, pr.Nonce = proof.Solver, proof.BlockNumber, proof.Nonce
	return nil
}

func (pr *Proof) Compare(v Proof) bool {

	comp := strings.Compare(pr.Solver.String(), v.Solver.String())
	if comp != 0 {
		return false
	}

	if pr.BlockNumber.Cmp(v.BlockNumber) != 0 {
		return false
	}
	if (pr.Nonce - pr.Nonce) != 0 {
		return false
	}
	return true
}
