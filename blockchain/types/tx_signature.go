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
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/crypto"
	"math/big"
)

// TxSignature contains a signature of tx (V, R, S).
type TxSignature struct {
	V *big.Int
	R *big.Int
	S *big.Int
}

func NewTxSignature() *TxSignature {
	return &TxSignature{
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
	}
}

func NewTxSignatureWithValues(signer Signer, txhash common.Hash, prv *ecdsa.PrivateKey) (*TxSignature, error) {
	sig, err := crypto.Sign(txhash[:], prv)
	if err != nil {
		return nil, err
	}

	txsig := &TxSignature{}

	txsig.R, txsig.S, txsig.V, err = signer.SignatureValues(sig)
	if err != nil {
		return nil, err
	}

	return txsig, nil
}

func (t *TxSignature) GetVRS() (*big.Int, *big.Int, *big.Int) {
	return t.V, t.R, t.S
}

func (t *TxSignature) SetVRS(v *big.Int, r *big.Int, s *big.Int) {
	t.V = v
	t.R = r
	t.S = s
}
