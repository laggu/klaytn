// Copyright 2018 The klaytn Authors
// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from core/types/transaction_signing.go (2018/06/04).
// Modified and improved for the klaytn development.

package types

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/params"
	"math/big"
)

var (
	ErrInvalidChainId        = errors.New("invalid chain id for signer")
	errNotTxInternalDataFrom = errors.New("not an TxInternalDataFrom")
)

// sigCache is used to cache the derived sender and contains
// the signer used to derive it.
type sigCache struct {
	signer Signer
	from   common.Address
}

// sigCachePubkey is used to cache the derived public key and contains
// the signer used to derive it.
type sigCachePubkey struct {
	signer Signer
	pubkey *ecdsa.PublicKey
}

// TODO-GX Remove the second parameter blockNumber
// MakeSigner returns a Signer based on the given chain config and block number.
func MakeSigner(config *params.ChainConfig, blockNumber *big.Int) Signer {
	return NewEIP155Signer(config.ChainID)
}

// SignTx signs the transaction using the given signer and private key
func SignTx(tx *Transaction, s Signer, prv *ecdsa.PrivateKey) (*Transaction, error) {
	h := s.Hash(tx)
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return nil, err
	}
	return tx.WithSignature(s, sig)
}

// AccountKeyPicker has a function GetKey() to retrieve an account key from statedb.
type AccountKeyPicker interface {
	GetKey(address common.Address) AccountKey
}

// ValidateSender finds a sender from both legacy and new types of transactions.
func ValidateSender(signer Signer, tx *Transaction, p AccountKeyPicker) (common.Address, error) {
	if tx.IsLegacyTransaction() {
		return Sender(signer, tx)
	}

	pubkey, err := SenderPubkey(signer, tx)
	if err != nil {
		return common.Address{}, err
	}
	txfrom, ok := tx.data.(TxInternalDataFrom)
	if !ok {
		return common.Address{}, errNotTxInternalDataFrom
	}
	from := txfrom.GetFrom()
	accKey := p.GetKey(from)

	if !accKey.Equal(NewAccountKeyPublicWithValue(pubkey)) {
		return common.Address{}, ErrInvalidSig
	}

	return from, nil
}

// Sender returns the address derived from the signature (V, R, S) using secp256k1
// elliptic curve and an error if it failed deriving or upon an incorrect
// signature.
//
// Sender may cache the address, allowing it to be used regardless of
// signing method. The cache is invalidated if the cached signer does
// not match the signer used in the current call.
func Sender(signer Signer, tx *Transaction) (common.Address, error) {
	if sc := tx.from.Load(); sc != nil {
		sigCache := sc.(sigCache)
		// If the signer used to derive from in a previous
		// call is not the same as used current, invalidate
		// the cache.
		if sigCache.signer.Equal(signer) {
			return sigCache.from, nil
		}
	}

	addr, err := signer.Sender(tx)
	if err != nil {
		return common.Address{}, err
	}
	tx.from.Store(sigCache{signer: signer, from: addr})
	return addr, nil
}

// SenderPubkey returns the public key derived from the signature (V, R, S) using secp256k1
// elliptic curve and an error if it failed deriving or upon an incorrect
// signature.
//
// SenderPubkey may cache the public key, allowing it to be used regardless of
// signing method. The cache is invalidated if the cached signer does
// not match the signer used in the current call.
func SenderPubkey(signer Signer, tx *Transaction) (*ecdsa.PublicKey, error) {
	if sc := tx.from.Load(); sc != nil {
		sigCache := sc.(sigCachePubkey)
		// If the signer used to derive from in a previous
		// call is not the same as used current, invalidate
		// the cache.
		if sigCache.signer.Equal(signer) {
			return sigCache.pubkey, nil
		}
	}

	pubkey, err := signer.SenderPubkey(tx)
	if err != nil {
		return nil, err
	}
	tx.from.Store(sigCachePubkey{signer: signer, pubkey: pubkey})
	return pubkey, nil
}

// Signer encapsulates transaction signature handling. Note that this interface is not a
// stable API and may change at any time to accommodate new protocol rules.
type Signer interface {
	// Sender returns the sender address of the transaction.
	Sender(tx *Transaction) (common.Address, error)
	// SenderPubkey returns the public key derived from tx signature and txhash.
	SenderPubkey(tx *Transaction) (*ecdsa.PublicKey, error)
	// SignatureValues returns the raw R, S, V values corresponding to the
	// given signature.
	SignatureValues(tx *Transaction, sig []byte) (r, s, v *big.Int, err error)
	// Hash returns the hash to be signed.
	Hash(tx *Transaction) common.Hash
	// Equal returns true if the given signer is the same as the receiver.
	Equal(Signer) bool
}

// EIP155Transaction implements Signer using the EIP155 rules.
type EIP155Signer struct {
	chainId, chainIdMul *big.Int
}

func NewEIP155Signer(chainId *big.Int) EIP155Signer {
	if chainId == nil {
		chainId = new(big.Int)
	}
	return EIP155Signer{
		chainId:    chainId,
		chainIdMul: new(big.Int).Mul(chainId, common.Big2),
	}
}

func (s EIP155Signer) Equal(s2 Signer) bool {
	eip155, ok := s2.(EIP155Signer)
	return ok && eip155.chainId.Cmp(s.chainId) == 0
}

var big8 = big.NewInt(8)

func (s EIP155Signer) Sender(tx *Transaction) (common.Address, error) {
	if !tx.IsLegacyTransaction() {
		b, _ := json.Marshal(tx)
		logger.Warn("No need to execute Sender!", "tx", string(b))
	}

	if !tx.Protected() {
		return HomesteadSigner{}.Sender(tx)
	}
	if tx.ChainId().Cmp(s.chainId) != 0 {
		return common.Address{}, ErrInvalidChainId
	}
	txV, txR, txS := tx.data.GetVRS()
	V := new(big.Int).Sub(txV, s.chainIdMul)
	V.Sub(V, big8)
	return recoverPlain(s.Hash(tx), txR, txS, V, true)
}

func (s EIP155Signer) SenderPubkey(tx *Transaction) (*ecdsa.PublicKey, error) {
	if tx.IsLegacyTransaction() {
		b, _ := json.Marshal(tx)
		logger.Warn("No need to execute SenderPubkey!", "tx", string(b))
	}

	if !tx.Protected() {
		return HomesteadSigner{}.SenderPubkey(tx)
	}
	if tx.ChainId().Cmp(s.chainId) != 0 {
		return nil, ErrInvalidChainId
	}
	txV, txR, txS := tx.data.GetVRS()
	V := new(big.Int).Sub(txV, s.chainIdMul)
	V.Sub(V, big8)
	return recoverPlainPubkey(s.Hash(tx), txR, txS, V, true)
}

// WithSignature returns a new transaction with the given signature. This signature
// needs to be in the [R || S || V] format where V is 0 or 1.
func (s EIP155Signer) SignatureValues(tx *Transaction, sig []byte) (R, S, V *big.Int, err error) {
	R, S, V, err = HomesteadSigner{}.SignatureValues(tx, sig)
	if err != nil {
		return nil, nil, nil, err
	}
	if s.chainId.Sign() != 0 {
		V = big.NewInt(int64(sig[64] + 35))
		V.Add(V, s.chainIdMul)
	}
	return R, S, V, nil
}

// Hash returns the hash to be signed by the sender.
// It does not uniquely identify the transaction.
func (s EIP155Signer) Hash(tx *Transaction) common.Hash {
	infs := append(tx.data.SerializeForSign(),
		s.chainId, uint(0), uint(0))
	return rlpHash(infs)
}

// TODO-GX Remove HomesteadSigner
// HomesteadTransaction implements TransactionInterface using the
// homestead rules.
type HomesteadSigner struct{ FrontierSigner }

func (s HomesteadSigner) Equal(s2 Signer) bool {
	_, ok := s2.(HomesteadSigner)
	return ok
}

// SignatureValues returns signature values. This signature
// needs to be in the [R || S || V] format where V is 0 or 1.
func (hs HomesteadSigner) SignatureValues(tx *Transaction, sig []byte) (r, s, v *big.Int, err error) {
	return hs.FrontierSigner.SignatureValues(tx, sig)
}

func (hs HomesteadSigner) Sender(tx *Transaction) (common.Address, error) {
	if !tx.IsLegacyTransaction() {
		b, _ := json.Marshal(tx)
		logger.Warn("No need to execute Sender!", "tx", string(b))
	}

	v, r, s := tx.data.GetVRS()
	return recoverPlain(hs.Hash(tx), r, s, v, true)
}

func (hs HomesteadSigner) SenderPubkey(tx *Transaction) (*ecdsa.PublicKey, error) {
	if tx.IsLegacyTransaction() {
		b, _ := json.Marshal(tx)
		logger.Warn("No need to execute SenderPubkey!", "tx", string(b))
	}

	v, r, s := tx.data.GetVRS()
	return recoverPlainPubkey(hs.Hash(tx), r, s, v, true)
}

type FrontierSigner struct{}

func (s FrontierSigner) Equal(s2 Signer) bool {
	_, ok := s2.(FrontierSigner)
	return ok
}

// SignatureValues returns signature values. This signature
// needs to be in the [R || S || V] format where V is 0 or 1.
func (fs FrontierSigner) SignatureValues(tx *Transaction, sig []byte) (r, s, v *big.Int, err error) {
	if len(sig) != 65 {
		panic(fmt.Sprintf("wrong size for signature: got %d, want 65", len(sig)))
	}
	r = new(big.Int).SetBytes(sig[:32])
	s = new(big.Int).SetBytes(sig[32:64])
	v = new(big.Int).SetBytes([]byte{sig[64] + 27})
	return r, s, v, nil
}

// Hash returns the hash to be signed by the sender.
// It does not uniquely identify the transaction.
func (fs FrontierSigner) Hash(tx *Transaction) common.Hash {
	return rlpHash(tx.data.SerializeForSign())
}

func (fs FrontierSigner) Sender(tx *Transaction) (common.Address, error) {
	if !tx.IsLegacyTransaction() {
		b, _ := json.Marshal(tx)
		logger.Warn("No need to execute Sender!", "tx", string(b))
	}

	v, r, s := tx.data.GetVRS()
	return recoverPlain(fs.Hash(tx), r, s, v, false)
}

func (fs FrontierSigner) SenderPubkey(tx *Transaction) (*ecdsa.PublicKey, error) {
	if tx.IsLegacyTransaction() {
		b, _ := json.Marshal(tx)
		logger.Warn("No need to execute SenderPubkey!", "tx", string(b))
	}

	v, r, s := tx.data.GetVRS()
	return recoverPlainPubkey(fs.Hash(tx), r, s, v, false)
}

func recoverPlainCommon(sighash common.Hash, R, S, Vb *big.Int, homestead bool) ([]byte, error) {
	if Vb.BitLen() > 8 {
		return []byte{}, ErrInvalidSig
	}
	V := byte(Vb.Uint64() - 27)
	if !crypto.ValidateSignatureValues(V, R, S, homestead) {
		return []byte{}, ErrInvalidSig
	}
	// encode the snature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V
	// recover the public key from the snature
	pub, err := crypto.Ecrecover(sighash[:], sig)
	if err != nil {
		return []byte{}, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return []byte{}, errors.New("invalid public key")
	}
	return pub, nil
}

func recoverPlain(sighash common.Hash, R, S, Vb *big.Int, homestead bool) (common.Address, error) {
	pub, err := recoverPlainCommon(sighash, R, S, Vb, homestead)
	if err != nil {
		return common.Address{}, err
	}

	var addr common.Address
	copy(addr[:], crypto.Keccak256(pub[1:])[12:])
	return addr, nil
}

func recoverPlainPubkey(sighash common.Hash, R, S, Vb *big.Int, homestead bool) (*ecdsa.PublicKey, error) {
	pub, err := recoverPlainCommon(sighash, R, S, Vb, homestead)
	if err != nil {
		return nil, err
	}

	pubkey, err := crypto.UnmarshalPubkey(pub)
	if err != nil {
		return nil, err
	}

	return pubkey, nil
}

// deriveChainId derives the chain id from the given v parameter
func deriveChainId(v *big.Int) *big.Int {
	if v.BitLen() <= 64 {
		v := v.Uint64()
		if v == 27 || v == 28 {
			return new(big.Int)
		}
		return new(big.Int).SetUint64((v - 35) / 2)
	}
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, common.Big2)
}
