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
	"io/ioutil"
	"crypto/ecdsa"
	"github.com/ground-x/go-gxplatform/accounts"
	"github.com/ground-x/go-gxplatform/accounts/keystore"
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
	"github.com/ground-x/go-gxplatform/node"
	"github.com/ground-x/go-gxplatform/params"
	"github.com/ground-x/go-gxplatform/rlp"
	"github.com/ground-x/go-gxplatform/common/profile"

	istanbulBackend "github.com/ground-x/go-gxplatform/consensus/istanbul/backend"
	istanbulCore "github.com/ground-x/go-gxplatform/consensus/istanbul/core"
)

////////////////////////////////////////////////////////////////////////////////
// AddressBalanceMap
////////////////////////////////////////////////////////////////////////////////
type AddressBalanceMap struct {
	balanceMap map[common.Address]*big.Int
}

func (a *AddressBalanceMap) Get(addr *common.Address) (*big.Int) {
	return new(big.Int).Set(a.balanceMap[*addr])
}

func (a *AddressBalanceMap) Add(addr *common.Address, v *big.Int) {
	if b, ok := a.balanceMap[*addr]; ok {
		b.Add(b, v)
	}
}

func (a *AddressBalanceMap) Sub(addr *common.Address, v *big.Int) {
	if b, ok := a.balanceMap[*addr]; ok {
		b.Sub(b, v)
	}
}

func (a *AddressBalanceMap) Set(addr *common.Address, v *big.Int) {
	a.balanceMap[*addr] = new(big.Int).Set(v)
}

func NewAddressBalanceMap() (*AddressBalanceMap) {
	return &AddressBalanceMap{
		balanceMap: make(map[common.Address]*big.Int),
	}
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

// Copied from node/config.go
func makeAccountManager(conf *node.Config) (*accounts.Manager, string, error) {
	scryptN, scryptP, keydir, err := conf.AccountConfig()
	var ephemeral string
	if keydir == "" {
		// There is no datadir.
		keydir, err = ioutil.TempDir("", "go-klaytn-keystore")
		ephemeral = keydir
	}

	if err != nil {
		return nil, "", err
	}
	if err := os.MkdirAll(keydir, 0700); err != nil {
		return nil, "", err
	}
	// Assemble the account manager and supported backends
	backends := []accounts.Backend{
		keystore.NewKeyStore(keydir, scryptN, scryptP),
	}
	return accounts.NewManager(backends...), ephemeral, nil
}

// fetchKeystore retrieves the encrypted keystore from the account manager.
// Copied from internal/gxapi/api.go
func fetchKeystore(am *accounts.Manager) *keystore.KeyStore {
	return am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
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

func initBlockchain(conf *node.Config, db gxdb.Database, coinbase *common.Address, validators []common.Address,
	engine consensus.Engine) (*core.BlockChain, error) {

	extraData, err := prepareIstanbulExtra(validators)

	genesis := core.DefaultGenesisBlock()
	genesis.Coinbase = *coinbase
	genesis.Config = Forks["Byzantium"]
	genesis.GasLimit = 100000000
	genesis.ExtraData = extraData
	genesis.Nonce = 0
	genesis.Mixhash = types.IstanbulDigest
	genesis.Difficulty = big.NewInt(1)

	genesis.Alloc = core.GenesisAlloc{*coinbase: {Balance: big.NewInt(1000000000)}}

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

func createAccounts(max_accounts int, ks *keystore.KeyStore, tb testing.TB) {
	num_accounts := len(ks.Accounts())
	for i := 0; i < max_accounts-num_accounts; i++ {
		acc, err := ks.NewAccount("")
		if err != nil {
			tb.Fatal(err)
		}

		tb.Logf("creating account %s, remaining %d accounts...\n",
			acc.Address.Hex(), max_accounts-num_accounts-i)
	}

}

func getPrivateKey(ks *keystore.KeyStore, account accounts.Account) (*keystore.Key, error) {
	keyJSON, err := ks.Export(account, "", "")
	if err != nil {
		return nil, err
	}

	return keystore.DecryptKey(keyJSON, "")
}

func prepareHeader(bc *core.BlockChain, genesis_addr *common.Address, validators []common.Address) (*types.Header, error) {
	tstart := time.Now()
	parent := bc.CurrentBlock()

	tstamp := tstart.Unix()
	if parent.Time().Cmp(new(big.Int).SetInt64(tstamp)) >= 0 {
		tstamp = parent.Time().Int64() + 1
	}

	num := parent.Number()
	return &types.Header{
		ParentHash: parent.Hash(),
		Coinbase:   *genesis_addr,
		Number:     num.Add(num, common.Big1),
		GasLimit:   core.CalcGasLimit(parent),
		Time:       big.NewInt(tstamp),
	}, nil
}

func makeTransactions(from *accounts.Account, startNonce uint64, ks *keystore.KeyStore,
	chainID *big.Int, bc *core.BlockChain, privKey *ecdsa.PrivateKey,
	header *types.Header, addressBalanceMap *AddressBalanceMap,
	tb testing.TB) (types.Transactions, error) {

	txs := make(types.Transactions, 0, len(addressBalanceMap.balanceMap))
	nonce := startNonce
	signer := types.MakeSigner(bc.Config(), header.Number)
	for a, _ := range addressBalanceMap.balanceMap {
		amount := big.NewInt(rand.Int63n(10))
		amount = amount.Add(amount, big.NewInt(1))
		var gasLimit uint64 = 1000000
		gasPrice := new(big.Int).SetInt64(0)
		data := []byte{}

		tx := types.NewTransaction(nonce, a, amount, gasLimit, gasPrice, data)
		signedTx, err := types.SignTx(tx, signer, privKey)
		if err != nil {
			return nil, err
		}

		tb.Logf("transferring (%d) %s -> %s (%s)\n", nonce, from.Address.Hex(), a.Hex(), amount.String())

		txs = append(txs, signedTx)

		nonce++
	}

	return txs, nil
}

// reference: miner/worker.go
func commitTransactions(bc *core.BlockChain, txs types.Transactions, coinbase *common.Address,
	statedb *state.StateDB, header *types.Header,
	addressBalanceMap *AddressBalanceMap) (types.Receipts, error){

	gp := new(core.GasPool)
	gp = gp.AddGas(10000000000)

	receipts := make(types.Receipts, len(txs))

	for i, tx := range txs {
		snap := statedb.Snapshot()

		// update addressBalance map
		to := tx.To()
		if *to != *coinbase {
			value := tx.Value()
			addressBalanceMap.Sub(coinbase, value)
			addressBalanceMap.Add(to, value)
		}

		receipt, _, err := core.ApplyTransaction(bc.Config(), bc, coinbase, gp, statedb, header,
			tx, &header.GasUsed, vm.Config{})
		if err != nil {
			statedb.RevertToSnapshot(snap)
			return nil, err
		}

		receipts[i] = receipt
	}

	return receipts, nil
}


func VerifyValueTransfer(addressBalanceMap *AddressBalanceMap, statedb *state.StateDB, tb testing.TB) {
	for a, b := range addressBalanceMap.balanceMap {
		if b.Cmp(statedb.GetBalance(a)) != 0 {
			tb.Errorf("[%s] b = %s, bc = %s\n", a.Hex(), b.String(), statedb.GetBalance(a).String())
		}
	}
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

////////////////////////////////////////////////////////////////////////////////
// TestValueTransfer
////////////////////////////////////////////////////////////////////////////////
func TestValueTransfer(t *testing.T) {
	conf := node.DefaultConfig

	// Remove leveldb dir if exists
	if _, err := os.Stat("chaindata"); err == nil {
		os.RemoveAll("chaindata")
	}

	////////////////////////////////////////////////////////////////////////////////
	// 1. Create the account manager
	am, _, err := makeAccountManager(&conf)
	if err != nil {
		t.Fatal(err)
	}

	////////////////////////////////////////////////////////////////////////////////
	// 2. Create a database
	chainDb, err := NewDatabase(gxdb.LEVELDB)
	if err != nil {
		t.Fatal(err)
	}
	defer chainDb.Close()
	// Remove leveldb dir which was created for this test.
	defer os.RemoveAll("chaindata")

	////////////////////////////////////////////////////////////////////////////////
	// 3. Create accounts as many as max_accounts
	max_accounts := 10
	num_validators := 4

	// TODO: make num_validator and max_accounts as arguments
	if num_validators > max_accounts {
		t.Fatalf("max_accounts should be bigger num_validators!!")
	}
	ks := fetchKeystore(am)
	createAccounts(max_accounts, ks, t)

	accounts := ks.Accounts()

	////////////////////////////////////////////////////////////////////////////////
	// 4. Set the genesis account
	genesis_acc := accounts[0]
	genesis_addr := genesis_acc.Address

	////////////////////////////////////////////////////////////////////////////////
	// 5. Use first 4 accounts as vaildators
	validatorPrivKeys := make([]*ecdsa.PrivateKey, num_validators)
	validatorAddresses := make([]common.Address, num_validators)
	for i := 0; i < num_validators; i++ {
		k, _ := getPrivateKey(ks, accounts[i])
		validatorPrivKeys[i] = k.PrivateKey
		validatorAddresses[i] = accounts[i].Address
	}

	////////////////////////////////////////////////////////////////////////////////
	// 6. Setup istanbul consensus backend
	start := time.Now()
	engine := istanbulBackend.New(genesis_addr, genesis_addr, istanbul.DefaultConfig, validatorPrivKeys[0], chainDb)
	profile.Prof.Profile("main_istanbul_engine_creation", time.Now().Sub(start))

	////////////////////////////////////////////////////////////////////////////////
	// 7. Make a blockchain
	start = time.Now()
	bc, err := initBlockchain(&conf, chainDb, &genesis_addr, validatorAddresses, engine)
	if err != nil {
		t.Fatal(err)
	}
	defer bc.Stop()
	profile.Prof.Profile("main_init_blockchain", time.Now().Sub(start))

	////////////////////////////////////////////////////////////////////////////////
	// 8. Get block state just after the genesis block
	statedb, err := bc.State()
	if err != nil {
		t.Fatal(err)
	}

	////////////////////////////////////////////////////////////////////////////////
	// 9. Initialize address-balance map for verification
	addressBalanceMap := NewAddressBalanceMap()
	for i := 0; i < max_accounts; i++ {
		addressBalanceMap.Set(&accounts[i].Address, statedb.GetBalance(accounts[i].Address))
	}

	chainID := bc.Config().ChainID

	////////////////////////////////////////////////////////////////////////////////
	// 10. Set the block header
	start = time.Now()
	header, err := prepareHeader(bc, &genesis_addr, validatorAddresses)
	if err != nil {
		t.Fatal(err)
	}

	////////////////////////////////////////////////////////////////////////////////
	// 11. Prepare istanbul block header
	if err := engine.Prepare(bc, header); err != nil {
		fmt.Errorf("Failed to prepare header for mining.\n")
		t.Fatal(err)
	}
	profile.Prof.Profile("main_prepareHeader", time.Now().Sub(start))

	////////////////////////////////////////////////////////////////////////////////
	// 12. Make a set of transactions
	transactions, err := makeTransactions(&genesis_acc, 0, ks, chainID,
		bc, validatorPrivKeys[0], header, addressBalanceMap, t)
	if err != nil {
		t.Fatal(err)
	}

	////////////////////////////////////////////////////////////////////////////////
	// 12. Apply the set of transactions
	start = time.Now()
	receipts, err := commitTransactions(bc, transactions, &genesis_addr, statedb, header,
		addressBalanceMap)
	if err != nil {
		t.Fatal(err)
	}
	profile.Prof.Profile("main_commitTransactions", time.Now().Sub(start))

	////////////////////////////////////////////////////////////////////////////////
	// 13. Finalize the block
	start = time.Now()
	b, err := engine.Finalize(bc, header, statedb, transactions, []*types.Header{}, receipts)
	if err != nil {
		fmt.Println("Failed to finalize block for sealing.")
		t.Fatal(err)
	}
	profile.Prof.Profile("main_finalize_block", time.Now().Sub(start))

	////////////////////////////////////////////////////////////////////////////////
	// 14. Seal the block to pass istanbul consensus verification
	start = time.Now()
	b, err = sealBlock(b, validatorPrivKeys)
	if err != nil {
		t.Fatal(err)
	}
	profile.Prof.Profile("main_seal_block", time.Now().Sub(start))

	////////////////////////////////////////////////////////////////////////////////
	// 15. Insert the block into the blockchain
	start = time.Now()
	n, err := bc.InsertChain(types.Blocks{b})
	if err != nil {
		t.Fatal(err)
	}
	profile.Prof.Profile("main_insert_blockchain", time.Now().Sub(start))

	if n != 0 {
		fmt.Printf("N should be zero! (%d)\n", n)
		t.Fatal("N should be zero!")
	}

	////////////////////////////////////////////////////////////////////////////////
	// 16. apply reward
	//state.AddBalance(common.HexToAddress(contract.RNRewardAddr), rewardcontract)
	//state.AddBalance(common.HexToAddress(contract.CommitteeRewardAddr), rewardcontract)
	//state.AddBalance(common.HexToAddress(contract.PIReserveAddr), rewardcontract)
	addressBalanceMap.Add(&genesis_addr, big.NewInt(1000000000000000000))

	////////////////////////////////////////////////////////////////////////////////
	// 17. verification with addressBalanceMap
	VerifyValueTransfer(addressBalanceMap, statedb, t)

	if testing.Verbose() {
		profile.Prof.PrintProfileInfo()
	}
}
