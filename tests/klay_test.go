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
package tests

import (
	"os"
	"fmt"
	"time"
	"bytes"
	"errors"
	"testing"
	"math/big"
	"math/rand"
	"crypto/ecdsa"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/consensus"
	"github.com/ground-x/go-gxplatform/consensus/istanbul"
	"github.com/ground-x/go-gxplatform/core"
	"github.com/ground-x/go-gxplatform/core/state"
	"github.com/ground-x/go-gxplatform/core/types"
	"github.com/ground-x/go-gxplatform/core/vm"
	"github.com/ground-x/go-gxplatform/crypto"
	"github.com/ground-x/go-gxplatform/crypto/sha3"
	"github.com/ground-x/go-gxplatform/gxdb"
	"github.com/ground-x/go-gxplatform/miner"
	"github.com/ground-x/go-gxplatform/node"
	"github.com/ground-x/go-gxplatform/params"
	"github.com/ground-x/go-gxplatform/rlp"
	"github.com/ground-x/go-gxplatform/common/profile"

	istanbulBackend "github.com/ground-x/go-gxplatform/consensus/istanbul/backend"
	istanbulCore "github.com/ground-x/go-gxplatform/consensus/istanbul/core"
)

const transactionsJournalFilename = "transactions.rlp"

// If you don't want to remove 'chaindata', set removeChaindataOnExit = false
const removeChaindataOnExit = true

type BCData struct {
	bc *core.BlockChain
	addrs []*common.Address
	privKeys []*ecdsa.PrivateKey
	db gxdb.Database
	rewardBase *common.Address
	validatorAddresses []common.Address
	validatorPrivKeys  []*ecdsa.PrivateKey
	engine consensus.Istanbul
}

////////////////////////////////////////////////////////////////////////////////
// AddressBalanceMap
////////////////////////////////////////////////////////////////////////////////
type AccountInfo struct {
	balance *big.Int
	nonce uint64
}

type AccountMap map[common.Address]*AccountInfo

func (a AccountMap) Get(addr common.Address) (*AccountInfo) {
	if acc, ok := a[addr]; ok {
		return &AccountInfo{new(big.Int).Set(acc.balance), acc.nonce }
	}
	return &AccountInfo{big.NewInt(0), 0}
}

func (a AccountMap) AddBalance(addr common.Address, v *big.Int) {
	if acc, ok := a[addr]; ok {
		acc.balance.Add(acc.balance, v)
	}
}

func (a AccountMap) SubBalance(addr common.Address, v *big.Int) {
	if acc, ok := a[addr]; ok {
		acc.balance.Sub(acc.balance, v)
	}
}

func (a AccountMap) IncNonce(addr common.Address) {
	if acc, ok := a[addr]; ok {
		acc.nonce++
	}
}

func (a AccountMap) Set(addr common.Address, v *big.Int, nonce uint64) {
	a[addr] = &AccountInfo{new(big.Int).Set(v), nonce}
}

func (a AccountMap) Initialize(bcdata *BCData) (error){
	statedb, err := bcdata.bc.State()
	if err != nil {
		return err
	}

	for _, addr := range bcdata.addrs {
		a.Set(*addr, statedb.GetBalance(*addr), statedb.GetNonce(*addr))
	}

	return nil
}

func (a AccountMap) Update(txs types.Transactions, signer types.Signer) (error) {
	for _, tx := range txs {
		to := tx.To()
		v := tx.Value()

		from, err := types.Sender(signer, tx)
		if err != nil {
			return err
		}

		a.AddBalance(*to, v)
		a.SubBalance(from, v)

		a.IncNonce(from)
	}

	return nil
}

func (a AccountMap) Verify(statedb *state.StateDB) (error) {
	for addr, acc := range a {
		if acc.nonce != statedb.GetNonce(addr) {
			return errors.New(fmt.Sprintf("[%s] nonce is different!! statedb(%d) != accountMap(%d).\n",
				addr.Hex(), statedb.GetNonce(addr), acc.nonce))
		}

		if acc.balance.Cmp(statedb.GetBalance(addr)) != 0 {
			return errors.New(fmt.Sprintf("[%s] balance is different!! statedb(%s) != accountMap(%s).\n",
				addr.Hex(), statedb.GetBalance(addr).String(), acc.balance.String()))
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
func NewDatabase(dbtype string) (db gxdb.Database, err error) {
	switch dbtype {
	case gxdb.LEVELDB:
		db, err = gxdb.NewLDBDatabase("chaindata", 16, 16)
		if err != nil {
			return nil, err
		}

	default:
		db = gxdb.NewMemDatabase()
	}

	return
}

// Copied from consensus/istanbul/backend/engine.go
func prepareIstanbulExtra(validators []common.Address) ([]byte, error) {
	var buf bytes.Buffer

	buf.Write(bytes.Repeat([]byte{0x0}, types.IstanbulExtraVanity))

	ist := &types.IstanbulExtra{
		Validators:    validators,
		Seal:          []byte{},
		CommittedSeal: [][]byte{},
	}

	payload, err := rlp.EncodeToBytes(&ist)
	if err != nil {
		return nil, err
	}
	return append(buf.Bytes(), payload...), nil
}

func initBlockchain(conf *node.Config, db gxdb.Database, coinbaseAddrs []*common.Address, validators []common.Address,
	engine consensus.Engine) (*core.BlockChain, error) {

	extraData, err := prepareIstanbulExtra(validators)

	genesis := core.DefaultGenesisBlock()
	genesis.Coinbase = *coinbaseAddrs[0]
	genesis.Config = Forks["Byzantium"]
	genesis.GasLimit = 100000000
	genesis.ExtraData = extraData
	genesis.Nonce = 0
	genesis.Mixhash = types.IstanbulDigest
	genesis.Difficulty = big.NewInt(1)

	alloc := make(core.GenesisAlloc)
	for _, a := range coinbaseAddrs {
		alloc[*a] = core.GenesisAccount{Balance: big.NewInt(1000000000000000000)}
	}

	genesis.Alloc = alloc

	chainConfig, _, err := core.SetupGenesisBlock(db, genesis)
	if _, ok := err.(*params.ConfigCompatError); err != nil && !ok {
		return nil, err
	}

	chain, err := core.NewBlockChain(db, nil, chainConfig, engine, vm.Config{})
	if err != nil {
		return nil, err
	}

	return chain, nil
}

func createAccounts(numAccounts int) ([]*common.Address, []*ecdsa.PrivateKey, error) {
	accs := make([]*common.Address, numAccounts)
	privKeys := make([]*ecdsa.PrivateKey, numAccounts)

	for i := 0; i < numAccounts; i++ {
		k, err := crypto.GenerateKey()
		if err != nil {
			return nil, nil, err
		}
		keyAddr := crypto.PubkeyToAddress(k.PublicKey)

		accs[i] = &keyAddr
		privKeys[i] = k
	}

	return accs, privKeys, nil
}

func prepareHeader(bcdata *BCData) (*types.Header, error) {
	tstart := time.Now()
	parent := bcdata.bc.CurrentBlock()

	tstamp := tstart.Unix()
	if parent.Time().Cmp(new(big.Int).SetInt64(tstamp)) >= 0 {
		tstamp = parent.Time().Int64() + 1
	}
	// this will ensure we're not going off too far in the future
	if now := time.Now().Unix(); tstamp > now {
		wait := time.Duration(tstamp-now) * time.Second
		time.Sleep(wait)
	}

	num := parent.Number()
	header := &types.Header{
		ParentHash: parent.Hash(),
		Coinbase:   common.Address{},
		Number:     num.Add(num, common.Big1),
		GasLimit:   core.CalcGasLimit(parent),
		Time:       big.NewInt(tstamp),
	}

	if err := bcdata.engine.Prepare(bcdata.bc, header); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to prepare header for mining %s.\n", err))
	}

	return header, nil
}

func makeTransactionsFrom(accountMap *AccountMap, from common.Address, privKey *ecdsa.PrivateKey,
	signer types.Signer, toAddrs []*common.Address, amount *big.Int) (types.Transactions, error) {

	txs := make(types.Transactions, 0, len(toAddrs))
	nonce := (*accountMap)[from].nonce
	for _, a := range toAddrs {
		txamount := amount
		if txamount == nil {
			txamount = big.NewInt(rand.Int63n(10))
			txamount = txamount.Add(txamount, big.NewInt(1))
		}
		var gasLimit uint64 = 1000000
		gasPrice := new(big.Int).SetInt64(0)
		data := []byte{}

		tx := types.NewTransaction(nonce, *a, txamount, gasLimit, gasPrice, data)
		signedTx, err := types.SignTx(tx, signer, privKey)
		if err != nil {
			return nil, err
		}

		txs = append(txs, signedTx)

		nonce++
	}

	return txs, nil
}

func mineABlock(bcdata *BCData, transactions types.Transactions, signer types.Signer) (*types.Block, error) {
	// Set the block header
	start := time.Now()
	header, err := prepareHeader(bcdata)
	if err != nil {
		return nil, err
	}
	profile.Prof.Profile("mine_prepareHeader", time.Now().Sub(start))

	statedb, err := bcdata.bc.State()
	if err != nil {
		return nil, err
	}

	// Group transactions by the sender address
	start = time.Now()
	txs := make(map[common.Address]types.Transactions)
	for _, tx := range transactions {
		acc, err := types.Sender(signer, tx)
		if err != nil {
			return nil, err
		}
		txs[acc] = append(txs[acc], tx)
	}
	profile.Prof.Profile("mine_groupTransactions", time.Now().Sub(start))

	// Create a transaction set where transactions are sorted by price and nonce
	start = time.Now()
	txset := types.NewTransactionsByPriceAndNonce(signer, txs) // TODO-GX-issue136 gasPrice
	profile.Prof.Profile("mine_NewTransactionsByPriceAndNonce", time.Now().Sub(start))

	// Apply the set of transactions
	start = time.Now()
	gp := new(core.GasPool)
	gp = gp.AddGas(1000000000000000000)
	task := miner.NewTask(bcdata.bc.Config(), signer, statedb, gp, header)
	task.ApplyTransactions(txset, bcdata.bc, *bcdata.rewardBase)
	newtxs := task.Transactions()
	receipts := task.Receipts()
	profile.Prof.Profile("mine_ApplyTransactions", time.Now().Sub(start))

	// Finalize the block
	start = time.Now()
	b, err := bcdata.engine.Finalize(bcdata.bc, header, statedb, newtxs, []*types.Header{}, receipts)
	if err != nil {
		return nil, err
	}
	profile.Prof.Profile("mine_finalize_block", time.Now().Sub(start))

	////////////////////////////////////////////////////////////////////////////////

	start = time.Now()
	b, err = sealBlock(b, bcdata.validatorPrivKeys)
	if err != nil {
		return nil, err
	}
	profile.Prof.Profile("mine_seal_block", time.Now().Sub(start))

	return b, nil
}

// Copied from consensus/istanbul/backend/engine.go
func sigHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewKeccak256()

	// Clean seal is required for calculating proposer seal.
	rlp.Encode(hasher, types.IstanbulFilteredHeader(header, false))
	hasher.Sum(hash[:0])
	return hash
}

// writeSeal writes the extra-data field of the given header with the given seals.
// Copied from consensus/istanbul/backend/engine.go
func writeSeal(h *types.Header, seal []byte) error {
	if len(seal)%types.IstanbulExtraSeal != 0 {
		return errors.New("invalid signature")
	}

	istanbulExtra, err := types.ExtractIstanbulExtra(h)
	if err != nil {
		return err
	}

	istanbulExtra.Seal = seal
	payload, err := rlp.EncodeToBytes(&istanbulExtra)
	if err != nil {
		return err
	}

	h.Extra = append(h.Extra[:types.IstanbulExtraVanity], payload...)
	return nil
}

// writeCommittedSeals writes the extra-data field of a block header with given committed seals.
// Copied from consensus/istanbul/backend/engine.go
func writeCommittedSeals(h *types.Header, committedSeals [][]byte) error {
	errInvalidCommittedSeals := errors.New("invalid committed seals")

	if len(committedSeals) == 0 {
		return errInvalidCommittedSeals
	}

	for _, seal := range committedSeals {
		if len(seal) != types.IstanbulExtraSeal {
			return errInvalidCommittedSeals
		}
	}

	istanbulExtra, err := types.ExtractIstanbulExtra(h)
	if err != nil {
		return err
	}

	istanbulExtra.CommittedSeal = make([][]byte, len(committedSeals))
	copy(istanbulExtra.CommittedSeal, committedSeals)

	payload, err := rlp.EncodeToBytes(&istanbulExtra)
	if err != nil {
		return err
	}

	h.Extra = append(h.Extra[:types.IstanbulExtraVanity], payload...)
	return nil
}

// sign implements istanbul.backend.Sign
// Copied from consensus/istanbul/backend/backend.go
func sign(data []byte, privkey *ecdsa.PrivateKey) ([]byte, error) {
	hashData := crypto.Keccak256([]byte(data))
	return crypto.Sign(hashData, privkey)
}

func makeCommittedSeal(h *types.Header, privKeys []*ecdsa.PrivateKey) ([][]byte, error) {
	committedSeals := make([][]byte, 0, 3)

	for i := 1; i < 4; i++ {
		seal := istanbulCore.PrepareCommittedSeal(h.Hash())
		committedSeal, err := sign(seal, privKeys[i])
		if err != nil {
			return nil, err
		}
		committedSeals = append(committedSeals, committedSeal)
	}

	return committedSeals, nil
}

func sealBlock(b *types.Block, privKeys []*ecdsa.PrivateKey) (*types.Block, error) {
	header := b.Header()

	seal, err := sign(sigHash(header).Bytes(), privKeys[0])
	if err != nil {
		return nil, err
	}

	err = writeSeal(header, seal)
	if err != nil {
		return nil, err
	}

	committedSeals, err := makeCommittedSeal(header, privKeys)
	if err != nil {
		return nil, err
	}

	err = writeCommittedSeals(header, committedSeals)
	if err != nil {
		return nil, err
	}

	return b.WithSeal(header), nil
}

func initializeBC(maxAccounts, numValidators int) (*BCData, error) {
	conf := node.DefaultConfig

	// Remove leveldb dir if exists
	if _, err := os.Stat("chaindata"); err == nil {
		os.RemoveAll("chaindata")
	}

	// Remove transactions.rlp if exists
	if _, err := os.Stat(transactionsJournalFilename); err == nil {
		os.RemoveAll(transactionsJournalFilename)
	}

	////////////////////////////////////////////////////////////////////////////////
	// Create a database
	chainDb, err := NewDatabase(gxdb.LEVELDB)
	if err != nil {
		return nil, err
	}
	////////////////////////////////////////////////////////////////////////////////
	// Create accounts as many as maxAccounts
	if numValidators > maxAccounts {
		return nil, errors.New("maxAccounts should be bigger numValidators!!")
	}
	addrs, privKeys, err := createAccounts(maxAccounts)
	if err != nil {
		return nil, err
	}

	////////////////////////////////////////////////////////////////////////////////
	// Set the genesis address
	genesisAddr := *addrs[0]

	////////////////////////////////////////////////////////////////////////////////
	// Use first 4 accounts as vaildators
	validatorPrivKeys := make([]*ecdsa.PrivateKey, numValidators)
	validatorAddresses := make([]common.Address, numValidators)
	for i := 0; i < numValidators; i++ {
		validatorPrivKeys[i] = privKeys[i]
		validatorAddresses[i] = *addrs[i]
	}

	////////////////////////////////////////////////////////////////////////////////
	// Setup istanbul consensus backend
	engine := istanbulBackend.New(genesisAddr, genesisAddr, istanbul.DefaultConfig, validatorPrivKeys[0], chainDb)

	////////////////////////////////////////////////////////////////////////////////
	// Make a blockchain
	bc, err := initBlockchain(&conf, chainDb, addrs, validatorAddresses, engine)
	if err != nil {
		return nil, err
	}

	return &BCData{bc, addrs, privKeys, chainDb,
	&genesisAddr, validatorAddresses,
	validatorPrivKeys, engine }, nil
}

func shutdown(bcdata *BCData) {
	bcdata.bc.Stop()

	bcdata.db.Close()
	// Remove leveldb dir which was created for this test.
	if removeChaindataOnExit {
		os.RemoveAll("chaindata")
		os.RemoveAll(transactionsJournalFilename)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TestValueTransfer
////////////////////////////////////////////////////////////////////////////////
func TestValueTransfer(t *testing.T) {
	testValueTransfer(t, 1000, 4, 3)
}

func testValueTransfer(t *testing.T, numMaxAccounts, numValidators, numGeneratedBlocks int) {
	// Initialize blockchain
	start := time.Now()
	bcdata, err := initializeBC(numMaxAccounts, numValidators)
	if err != nil {
		t.Fatal(err)
	}
	profile.Prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer shutdown(bcdata)

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := make(AccountMap)
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}
	profile.Prof.Profile("main_init_accountMap", time.Now().Sub(start))

	for i := 0; i < numGeneratedBlocks; i++ {
		//fmt.Printf("iteration %d\n", i)

		// Make a set of transactions
		start = time.Now()
		signer := types.MakeSigner(bcdata.bc.Config(), bcdata.bc.CurrentHeader().Number)
		from := bcdata.addrs[0]
		privKey := bcdata.privKeys[0]
		transactions, err := makeTransactionsFrom(&accountMap, *from, privKey, signer, bcdata.addrs, nil)
		if err != nil {
			t.Fatal(err)
		}
		profile.Prof.Profile("main_makeTransactions", time.Now().Sub(start))

		// Update accountMap
		start = time.Now()
		if err := accountMap.Update(transactions, signer); err != nil {
			t.Fatal(err)
		}
		profile.Prof.Profile("main_update_accountMap", time.Now().Sub(start))

		// Mine a block!
		start = time.Now()
		b, err := mineABlock(bcdata, transactions, signer)
		if err != nil {
			t.Fatal(err)
		}
		profile.Prof.Profile("main_mineABlock", time.Now().Sub(start))

		// Insert the block into the blockchain
		start := time.Now()
		if n, err := bcdata.bc.InsertChain(types.Blocks{b}); err != nil{
			t.Fatal(fmt.Sprintf("err = %s, n = %d\n", err, n))
		}
		profile.Prof.Profile("main_insert_blockchain", time.Now().Sub(start))

		// Apply reward
		start = time.Now()
		rewardAddr := *bcdata.rewardBase
		accountMap.AddBalance(rewardAddr, big.NewInt(1000000000000000000))
		profile.Prof.Profile("main_apply_reward", time.Now().Sub(start))

		// Verification with accountMap
		start = time.Now()
		statedb, err := bcdata.bc.State()
		if err != nil {
			t.Fatal(err)
		}
		if err := accountMap.Verify(statedb); err != nil {
			t.Fatal(err)
		}
		profile.Prof.Profile("main_verification", time.Now().Sub(start))
	}

	if testing.Verbose() {
		profile.Prof.PrintProfileInfo()
	}
}
