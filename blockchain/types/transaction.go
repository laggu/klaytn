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
// This file is derived from core/types/transaction.go (2018/06/04).
// Modified and improved for the klaytn development.

package types

import (
	"container/heap"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/ser/rlp"
	"io"
	"math/big"
	"sync/atomic"
)

var (
	ErrInvalidSig                     = errors.New("invalid transaction v, r, s values")
	errNoSigner                       = errors.New("missing signing methods")
	ErrInvalidTxTypeForPeggedData     = errors.New("invalid transaction type for pegged data")
	errLegacyTransaction              = errors.New("should not be called by a legacy transaction")
	errNotImplementTxInternalDataFrom = errors.New("not implement TxInternalDataFrom")
	errNotFeePayer                    = errors.New("not implement fee payer interface")
)

// deriveSigner makes a *best* guess about which signer to use.
func deriveSigner(V *big.Int) Signer {
	if V.Sign() != 0 && isProtectedV(V) {
		return NewEIP155Signer(deriveChainId(V))
	} else {
		return HomesteadSigner{}
	}
}

type Transaction struct {
	data TxInternalData
	// caches
	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

func NewTransactionWithMap(t TxType, values map[TxValueKeyType]interface{}) (*Transaction, error) {
	txdata, err := NewTxInternalDataWithMap(t, values)
	if err != nil {
		return nil, err
	}
	return &Transaction{data: txdata}, nil
}

func NewTransaction(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	return newTransaction(nonce, &to, amount, gasLimit, gasPrice, data)
}

func NewContractCreation(nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	return newTransaction(nonce, nil, amount, gasLimit, gasPrice, data)
}

func newTransaction(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := txdata{
		AccountNonce: nonce,
		Recipient:    to,
		Payload:      data,
		Amount:       new(big.Int),
		GasLimit:     gasLimit,
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}

	return &Transaction{data: &d}
}

// ChainId returns which chain id this transaction was signed for (if at all)
func (tx *Transaction) ChainId() *big.Int {
	return tx.data.ChainId()
}

// Protected returns whether the transaction is protected from replay protection.
func (tx *Transaction) Protected() bool {
	return tx.data.Protected()
}

func isProtectedV(V *big.Int) bool {
	if V.BitLen() <= 8 {
		v := V.Uint64()
		return v != 27 && v != 28
	}
	// anything not 27 or 28 are considered unprotected
	return true
}

// EncodeRLP implements rlp.Encoder
func (tx *Transaction) EncodeRLP(w io.Writer) error {
	serializer := newTxInternalDataSerializerWithValues(tx.data)
	return rlp.Encode(w, serializer)
}

// DecodeRLP implements rlp.Decoder
func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	serializer := newTxInternalDataSerializer()
	err := s.Decode(serializer)
	tx.data = serializer.tx
	if err == nil {
		tx.size.Store(common.StorageSize(rlp.ListSize(size)))
	}

	return err
}

// MarshalJSON encodes the web3 RPC transaction format.
func (tx *Transaction) MarshalJSON() ([]byte, error) {
	hash := tx.Hash()
	data := tx.data
	data.SetHash(&hash)
	serializer := newTxInternalDataSerializerWithValues(tx.data)
	return json.Marshal(serializer)
}

// UnmarshalJSON decodes the web3 RPC transaction format.
func (tx *Transaction) UnmarshalJSON(input []byte) error {
	serializer := newTxInternalDataSerializer()
	if err := json.Unmarshal(input, serializer); err != nil {
		return err
	}
	var V byte
	v, r, s := serializer.tx.GetVRS()
	if isProtectedV(v) {
		chainID := deriveChainId(v).Uint64()
		V = byte(v.Uint64() - 35 - 2*chainID)
	} else {
		V = byte(v.Uint64() - 27)
	}
	if !crypto.ValidateSignatureValues(V, r, s, false) {
		return ErrInvalidSig
	}
	*tx = Transaction{data: serializer.tx}
	return nil
}

func (tx *Transaction) Gas() uint64                   { return tx.data.GetGasLimit() }
func (tx *Transaction) GasPrice() *big.Int            { return new(big.Int).Set(tx.data.GetPrice()) }
func (tx *Transaction) Value() *big.Int               { return new(big.Int).Set(tx.data.GetAmount()) }
func (tx *Transaction) Nonce() uint64                 { return tx.data.GetAccountNonce() }
func (tx *Transaction) CheckNonce() bool              { return true }
func (tx *Transaction) Type() TxType                  { return tx.data.Type() }
func (tx *Transaction) IntrinsicGas() (uint64, error) { return tx.data.IntrinsicGas() }
func (tx *Transaction) IsLegacyTransaction() bool     { return tx.data.IsLegacyTransaction() }

func (tx *Transaction) Data() []byte {
	tp, ok := tx.data.(TxInternalDataPayload)
	if !ok {
		return []byte{}
	}

	return common.CopyBytes(tp.GetPayload())
}

// PeggedData returns the pegged data of the chain data pegging transaction.
// if the tx is not chain data pegging transaction, it will return error.
func (tx *Transaction) PeggedData() ([]byte, error) {
	txData, ok := tx.data.(*TxInternalDataChainDataPegging)
	if ok {
		return txData.PeggedData, nil
	}
	return []byte{}, ErrInvalidTxTypeForPeggedData
}

// To returns the recipient address of the transaction.
// It returns nil if the transaction is a contract creation.
func (tx *Transaction) To() *common.Address {
	if tx.data.GetRecipient() == nil {
		return nil
	}
	to := *tx.data.GetRecipient()
	return &to
}

// From returns the from address of the transaction.
// Since a legacy transaction (txdata) does not have the field `from`,
// calling From() is failed for `txdata`.
func (tx *Transaction) From() (common.Address, error) {
	if tx.IsLegacyTransaction() {
		return common.Address{}, errLegacyTransaction
	}

	tf, ok := tx.data.(TxInternalDataFrom)
	if !ok {
		return common.Address{}, errNotImplementTxInternalDataFrom
	}

	return tf.GetFrom(), nil
}

// FeePayer returns the fee payer address.
// If the tx is a fee-delegated transaction, it returns the specified fee payer.
// Otherwise, it returns `from` of the tx.
func (tx *Transaction) FeePayer() (common.Address, error) {
	tf, ok := tx.data.(TxInternalDataFeePayer)
	if !ok {
		// if the tx is not a fee-delegated transaction, the fee payer is `from` of the tx.
		return tx.From()
	}

	return tf.GetFeePayer(), nil
}

// Hash hashes the RLP encoding of tx.
// It uniquely identifies the transaction.
func (tx *Transaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := rlpHash(tx)
	tx.hash.Store(v)
	return v
}

// Size returns the true RLP encoded storage size of the transaction, either by
// encoding and returning it, or returning a previsouly cached value.
func (tx *Transaction) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	rlp.Encode(&c, &tx.data)
	tx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

// AsMessageWithAccountKeyPicker returns the transaction as a blockchain.Message.
//
// AsMessageWithAccountKeyPicker requires a signer to derive the sender and AccountKeyPicker.
//
// XXX Rename message to something less arbitrary?
func (tx *Transaction) AsMessageWithAccountKeyPicker(s Signer, picker AccountKeyPicker) (Message, error) {
	intrinsicGas, err := tx.IntrinsicGas()
	if err != nil {
		return Message{}, err
	}
	msg := Message{
		nonce:         tx.data.GetAccountNonce(),
		gasLimit:      tx.data.GetGasLimit(),
		gasPrice:      new(big.Int).Set(tx.data.GetPrice()),
		to:            tx.data.GetRecipient(),
		amount:        tx.data.GetAmount(),
		data:          tx.Data(),
		checkNonce:    true,
		intrinsicGas:  intrinsicGas,
		txType:        tx.data.Type(),
		accountKey:    NewAccountKeyNil(),
		humanReadable: false,
	}

	if ta, ok := tx.data.(*TxInternalDataAccountCreation); ok {
		msg.accountKey = ta.Key
		msg.humanReadable = ta.HumanReadable
	}

	msg.from, err = ValidateSender(s, tx, picker)

	// TODO-Klaytn-FeePayer: set feePayer appropriately after feePayer feature is implemented.
	msg.feePayer = msg.from
	return msg, err
}

// WithSignature returns a new transaction with the given signature.
// This signature needs to be formatted as described in the yellow paper (v+27).
func (tx *Transaction) WithSignature(signer Signer, sig []byte) (*Transaction, error) {
	r, s, v, err := signer.SignatureValues(sig)
	if err != nil {
		return nil, err
	}
	cpy := &Transaction{data: tx.data}
	cpy.data.SetVRS(v, r, s)
	return cpy, nil
}

// Cost returns amount + gasprice * gaslimit.
func (tx *Transaction) Cost() *big.Int {
	total := new(big.Int).Mul(tx.data.GetPrice(), new(big.Int).SetUint64(tx.data.GetGasLimit()))
	total.Add(total, tx.data.GetAmount())
	return total
}

func (tx *Transaction) Sign(s Signer, prv *ecdsa.PrivateKey) error {
	h := s.Hash(tx)
	sig, err := NewTxSignatureWithValues(s, h, prv)
	if err != nil {
		return err
	}

	tx.SetSignature(sig)
	return nil
}

// SignFeePayer signs the tx with the given signer and private key as a fee payer.
func (tx *Transaction) SignFeePayer(s Signer, prv *ecdsa.PrivateKey) error {
	h := s.Hash(tx)
	sig, err := NewTxSignatureWithValues(s, h, prv)
	if err != nil {
		return err
	}

	if err := tx.SetFeePayerSignature(sig); err != nil {
		return err
	}

	return nil
}

func (tx *Transaction) SetFeePayerSignature(s *TxSignature) error {
	tf, ok := tx.data.(TxInternalDataFeePayer)
	if !ok {
		return errNotFeePayer
	}

	tf.SetFeePayerSignature(s)

	return nil
}

func (tx *Transaction) SetSignature(signature *TxSignature) {
	tx.data.SetSignature(signature)
}

func (tx *Transaction) RawSignatureValues() (*big.Int, *big.Int, *big.Int) {
	return tx.data.GetVRS()
}

func (tx *Transaction) String() string {
	return tx.data.String()
}

// GetChildChainAddr returns the pointer of sender address if a tx is a
// data pegging tx from child chain. If not, it returns nil.
func (tx *Transaction) GetChildChainAddr(signer Signer) *common.Address {
	// TODO-Klaytn-ServiceChain This function will be removed once new transaction type is introduced.
	from, err := Sender(signer, tx)
	if err != nil {
		logger.Error("failed to decode the address of the sender", "tx", tx.hash)
		return nil
	}
	if tx.To() == nil || from != *tx.To() {
		return nil
	}
	return &from
}

// Transactions is a Transaction slice type for basic sorting.
type Transactions []*Transaction

// Len returns the length of s.
func (s Transactions) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s Transactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetRlp implements Rlpable and returns the i'th element of s in rlp.
func (s Transactions) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}

// TxDifference returns a new set t which is the difference between a to b.
func TxDifference(a, b Transactions) (keep Transactions) {
	keep = make(Transactions, 0, len(a))

	remove := make(map[common.Hash]struct{})
	for _, tx := range b {
		remove[tx.Hash()] = struct{}{}
	}

	for _, tx := range a {
		if _, ok := remove[tx.Hash()]; !ok {
			keep = append(keep, tx)
		}
	}

	return keep
}

// TxByNonce implements the sort interface to allow sorting a list of transactions
// by their nonces. This is usually only useful for sorting transactions from a
// single account, otherwise a nonce comparison doesn't make much sense.
type TxByNonce Transactions

func (s TxByNonce) Len() int { return len(s) }
func (s TxByNonce) Less(i, j int) bool {
	return s[i].data.GetAccountNonce() < s[j].data.GetAccountNonce()
}
func (s TxByNonce) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// TxByPrice implements both the sort and the heap interface, making it useful
// for all at once sorting as well as individually adding and removing elements.
type TxByPrice Transactions

func (s TxByPrice) Len() int           { return len(s) }
func (s TxByPrice) Less(i, j int) bool { return s[i].data.GetPrice().Cmp(s[j].data.GetPrice()) > 0 }
func (s TxByPrice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s *TxByPrice) Push(x interface{}) {
	*s = append(*s, x.(*Transaction))
}

func (s *TxByPrice) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

// TransactionsByPriceAndNonce represents a set of transactions that can return
// transactions in a profit-maximizing sorted order, while supporting removing
// entire batches of transactions for non-executable accounts.
type TransactionsByPriceAndNonce struct {
	txs    map[common.Address]Transactions // Per account nonce-sorted list of transactions
	heads  TxByPrice                       // Next transaction for each unique account (price heap)
	signer Signer                          // Signer for the set of transactions
}

// ############ method for debug
func (t *TransactionsByPriceAndNonce) Count() (int, int) {
	var count int

	for _, tx := range t.txs {
		count += tx.Len()
	}

	return len(t.txs), count
}

func (t *TransactionsByPriceAndNonce) Txs() map[common.Address]Transactions {
	return t.txs
}

// TODO-Klaytn-Issue136 gasprice
// NewTransactionsByPriceAndNonce creates a transaction set that can retrieve
// price sorted transactions in a nonce-honouring way.
//
// Note, the input map is reowned so the caller should not interact any more with
// if after providing it to the constructor.
func NewTransactionsByPriceAndNonce(signer Signer, txs map[common.Address]Transactions) *TransactionsByPriceAndNonce {
	// Initialize a price based heap with the head transactions
	heads := make(TxByPrice, 0, len(txs))
	for _, accTxs := range txs {
		heads = append(heads, accTxs[0])
		// Ensure the sender address is from the signer
		acc, _ := Sender(signer, accTxs[0])
		txs[acc] = accTxs[1:]
	}
	heap.Init(&heads)

	// Assemble and return the transaction set
	return &TransactionsByPriceAndNonce{
		txs:    txs,
		heads:  heads,
		signer: signer,
	}
}

// Peek returns the next transaction by price.
func (t *TransactionsByPriceAndNonce) Peek() *Transaction {
	if len(t.heads) == 0 {
		return nil
	}
	return t.heads[0]
}

// Shift replaces the current best head with the next one from the same account.
func (t *TransactionsByPriceAndNonce) Shift() {
	acc, _ := Sender(t.signer, t.heads[0])
	if txs, ok := t.txs[acc]; ok && len(txs) > 0 {
		t.heads[0], t.txs[acc] = txs[0], txs[1:]
		heap.Fix(&t.heads, 0)
	} else {
		heap.Pop(&t.heads)
	}
}

// Pop removes the best transaction, *not* replacing it with the next one from
// the same account. This should be used when a transaction cannot be executed
// and hence all subsequent ones should be discarded from the same account.
func (t *TransactionsByPriceAndNonce) Pop() {
	heap.Pop(&t.heads)
}

// Message is a fully derived transaction and implements blockchain.Message
//
// NOTE: In a future PR this will be removed.
type Message struct {
	to            *common.Address
	from          common.Address
	feePayer      common.Address
	nonce         uint64
	amount        *big.Int
	gasLimit      uint64
	gasPrice      *big.Int
	data          []byte
	checkNonce    bool
	intrinsicGas  uint64
	txType        TxType
	accountKey    AccountKey
	humanReadable bool
}

func NewMessage(from common.Address, to *common.Address, nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, checkNonce bool, intrinsicGas uint64) Message {
	return Message{
		from:          from,
		feePayer:      from,
		to:            to,
		nonce:         nonce,
		amount:        amount,
		gasLimit:      gasLimit,
		gasPrice:      gasPrice,
		data:          data,
		checkNonce:    checkNonce,
		intrinsicGas:  intrinsicGas,
		txType:        TxTypeLegacyTransaction,
		accountKey:    NewAccountKeyNil(),
		humanReadable: false,
	}
}

func (m Message) From() common.Address          { return m.from }
func (m Message) FeePayer() common.Address      { return m.feePayer }
func (m Message) To() *common.Address           { return m.to }
func (m Message) GasPrice() *big.Int            { return m.gasPrice }
func (m Message) Value() *big.Int               { return m.amount }
func (m Message) Gas() uint64                   { return m.gasLimit }
func (m Message) Nonce() uint64                 { return m.nonce }
func (m Message) Data() []byte                  { return m.data }
func (m Message) CheckNonce() bool              { return m.checkNonce }
func (m Message) IntrinsicGas() (uint64, error) { return m.intrinsicGas, nil }
func (m Message) TxType() TxType                { return m.txType }
func (m Message) AccountKey() AccountKey        { return m.accountKey }
func (m Message) HumanReadable() bool           { return m.humanReadable }
