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

package tests

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ground-x/klaytn/accounts/abi"
	"github.com/ground-x/klaytn/blockchain/state"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/profile"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/crypto/sha3"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/ground-x/klaytn/storage/database"
	"github.com/stretchr/testify/assert"
	"math/big"
	"math/rand"
	"strings"
	"testing"
	"time"
)

var (
	to       = common.HexToAddress("7b65B75d204aBed71587c9E519a89277766EE1d0")
	feePayer = common.HexToAddress("5A0043070275d9f6054307Ee7348bD660849D90f")
	nonce    = uint64(1234)
	amount   = big.NewInt(10)
	gasLimit = uint64(999999999)

	// TODO-Klaytn-Gas: When we have a configuration of Baobab or something, use gasPrice in that configuration.
	gasPrice = big.NewInt(25)
)

type TestAccountType struct {
	Addr   common.Address
	Key    *ecdsa.PrivateKey
	Nonce  uint64
	AccKey types.AccountKey
}

func genRandomHash() (h common.Hash) {
	hasher := sha3.NewKeccak256()

	r := rand.Uint64()
	rlp.Encode(hasher, r)
	hasher.Sum(h[:0])

	return h
}

// createAnonymousAccount creates an account whose address is derived from the private key.
func createAnonymousAccount(prvKeyHex string) (*TestAccountType, error) {
	key, err := crypto.HexToECDSA(prvKeyHex)
	if err != nil {
		return nil, err
	}

	addr := crypto.PubkeyToAddress(key.PublicKey)

	return &TestAccountType{
		Addr:   addr,
		Key:    key,
		Nonce:  uint64(0),
		AccKey: types.NewAccountKeyNil(),
	}, nil
}

// createDecoupledAccount creates an account whose address is decoupled with its private key.
func createDecoupledAccount(prvKeyHex string, addr common.Address) (*TestAccountType, error) {
	key, err := crypto.HexToECDSA(prvKeyHex)
	if err != nil {
		return nil, err
	}

	return &TestAccountType{
		Addr:   addr,
		Key:    key,
		Nonce:  uint64(0),
		AccKey: types.NewAccountKeyPublicWithValue(&key.PublicKey),
	}, nil
}

func createHumanReadableAccount(prvKeyHex string, humanReadableAddr string) (*TestAccountType, error) {
	key, err := crypto.HexToECDSA(prvKeyHex)
	if err != nil {
		return nil, err
	}

	addr, err := common.FromHumanReadableAddress(humanReadableAddr)
	if err != nil {
		return nil, err
	}

	return &TestAccountType{
		Addr:   addr,
		Key:    key,
		Nonce:  uint64(0),
		AccKey: types.NewAccountKeyPublicWithValue(&key.PublicKey),
	}, nil
}

// TODO-Klaytn-Account: Need to find a test mechanism to check failed transactions.
// Below code failed since it is trying to create an account with a human-readable address "1colin".
//func TestHumanReadableAddress(t *testing.T) {
//	if testing.Verbose() {
//		enableLog()
//	}
//	prof := profile.NewProfiler()
//
//	// Initialize blockchain
//	start := time.Now()
//	bcdata, err := NewBCData(6, 4)
//	if err != nil {
//		t.Fatal(err)
//	}
//	prof.Profile("main_init_blockchain", time.Now().Sub(start))
//	defer bcdata.Shutdown()
//
//	// Initialize address-balance map for verification
//	start = time.Now()
//	accountMap := NewAccountMap()
//	if err := accountMap.Initialize(bcdata); err != nil {
//		t.Fatal(err)
//	}
//	prof.Profile("main_init_accountMap", time.Now().Sub(start))
//
//	// reservoir account
//	reservoir := &TestAccountType{
//		Addr:  *bcdata.addrs[0],
//		Key:   bcdata.privKeys[0],
//		Nonce: uint64(0),
//	}
//	var colinAddr common.Address
//	colinAddr.SetBytesFromFront([]byte("1colin"))
//
//	colin, err := createDecoupledAccount("ed580f5bd71a2ee4dae5cb43e331b7d0318596e561e6add7844271ed94156b20", colinAddr)
//	assert.Equal(t, nil, err)
//
//	if testing.Verbose() {
//		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
//		fmt.Println("colinAddr = ", colin.Addr.String())
//	}
//
//	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
//
//	// Create an account "1colin" using TxTypeAccountCreation.
//	{
//		var txs types.Transactions
//
//		amount := new(big.Int).SetUint64(1000000000000)
//		values := map[types.TxValueKeyType]interface{}{
//			types.TxValueKeyNonce:         reservoir.Nonce,
//			types.TxValueKeyFrom:          reservoir.Addr,
//			types.TxValueKeyTo:            colin.Addr,
//			types.TxValueKeyAmount:        amount,
//			types.TxValueKeyGasLimit:      gasLimit,
//			types.TxValueKeyGasPrice:      gasPrice,
//			types.TxValueKeyHumanReadable: true,
//			types.TxValueKeyAccountKey:    colin.AccKey,
//		}
//		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
//		assert.Equal(t, nil, err)
//
//		err = tx.Sign(signer, reservoir.Key)
//		assert.Equal(t, nil, err)
//
//		txs = append(txs, tx)
//
//		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
//			t.Fatal(err)
//		}
//		reservoir.Nonce += 1
//	}
//
//	if testing.Verbose() {
//		prof.PrintProfileInfo()
//	}
//}

// TestTransactionScenario tests a following scenario:
// 1. Transfer (reservoir -> anon) using a legacy transaction.
// 2. Create an account decoupled using TxTypeAccountCreation.
// 3. Transfer (reservoir -> decoupled) using TxTypeValueTransfer.
// 4. Transfer (decoupled -> reservoir) using TxTypeValueTransfer.
// 5. Create an account colin using TxTypeAccountCreation.
// 6. Transfer (colin-> reservoir) using TxTypeValueTransfer.
// 7. ChainDataAnchoring (reservoir -> reservoir) using TxTypeChainDataAnchoring.
// 8. Transfer (colin-> reservoir) using TxTypeFeeDelegatedValueTransfer with a fee payer (reservoir).
// 9. Transfer (colin-> reservoir) using TxTypeFeeDelegatedValueTransferWithRatio with a fee payer (reservoir) and a ratio of 30.
func TestTransactionScenario(t *testing.T) {
	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(6, 4)
	if err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer bcdata.Shutdown()

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Key:   bcdata.privKeys[0],
		Nonce: uint64(0),
	}

	// anonymous account
	anon, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	// decoupled account
	decoupled, err := createDecoupledAccount("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
		common.HexToAddress("0x75c3098be5e4b63fbac05838daaee378dd48098d"))
	assert.Equal(t, nil, err)

	colin, err := createHumanReadableAccount("ed580f5bd71a2ee4dae5cb43e331b7d0318596e561e6add7844271ed94156b20", "colin")
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
		fmt.Println("decoupledAddr = ", decoupled.Addr.String())
		fmt.Println("colinAddr = ", colin.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. Transfer (reservoir -> anon) using a legacy transaction.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(100000000000)
		tx := types.NewTransaction(reservoir.Nonce,
			anon.Addr, amount, gasLimit, gasPrice, []byte{})

		err := tx.Sign(signer, reservoir.Key)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 2. Create an account decoupled using TxTypeAccountCreation.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            decoupled.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    decoupled.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, reservoir.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// Create the same account decoupled. This should be failed.
	// TODO-Klaytn: make a test case for this error case.
	//{
	//	var txs types.Transactions
	//
	//	amount := new(big.Int).SetUint64(100000000)
	//	values := map[types.TxValueKeyType]interface{}{
	//		types.TxValueKeyNonce:      reservoir.Nonce,
	//		types.TxValueKeyFrom:       reservoir.Addr,
	//		types.TxValueKeyTo:         decoupled.Addr,
	//		types.TxValueKeyAmount:     amount,
	//		types.TxValueKeyGasLimit:   gasLimit,
	//		types.TxValueKeyGasPrice:   gasPrice,
	//		types.TxValueKeyAccountKey: decoupled.AccKey
	//	}
	//	tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
	//	assert.Equal(t, nil, err)
	//
	//	signedTx, err := types.SignTx(tx, signer, reservoir.Key)
	//	assert.Equal(t, nil, err)
	//
	//	txs = append(txs, signedTx)
	//
	//	if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
	//		t.Fatal(err)
	//	}
	//	reservoir.Nonce += 1
	//}

	// 3. Transfer (reservoir -> decoupled) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    reservoir.Nonce,
			types.TxValueKeyFrom:     reservoir.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, reservoir.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 4. Transfer (decoupled -> reservoir) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    decoupled.Nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, decoupled.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		decoupled.Nonce += 1
	}

	// 5. Create an account colin using TxTypeAccountCreation.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            colin.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    colin.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, reservoir.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 6. Transfer (colin-> reservoir) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    colin.Nonce,
			types.TxValueKeyFrom:     colin.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, colin.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1
	}

	// 7. ChainDataAnchoring (reservoir -> reservoir) using TxTypeChainDataAnchoring.
	{
		scData := types.NewChainHashes(bcdata.bc.CurrentBlock())
		dataAnchoredRLP, _ := rlp.EncodeToBytes(scData)

		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:        reservoir.Nonce,
			types.TxValueKeyFrom:         reservoir.Addr,
			types.TxValueKeyTo:           reservoir.Addr,
			types.TxValueKeyAmount:       amount,
			types.TxValueKeyGasLimit:     gasLimit,
			types.TxValueKeyGasPrice:     gasPrice,
			types.TxValueKeyAnchoredData: dataAnchoredRLP,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeChainDataAnchoring, values)
		if err != nil {
			t.Fatal(err)
		}
		signedTx, err := types.SignTx(tx, signer, reservoir.Key)
		if err != nil {
			t.Fatal(err)
		}
		txs = append(txs, signedTx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 8. Transfer (colin-> reservoir) using TxTypeFeeDelegatedValueTransfer with a fee payer (reservoir).
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    colin.Nonce,
			types.TxValueKeyFrom:     colin.Addr,
			types.TxValueKeyFeePayer: reservoir.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, colin.Key)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayer(signer, reservoir.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1
	}

	// 9. Transfer (colin-> reservoir) using TxTypeFeeDelegatedValueTransferWithRatio with a fee payer (reservoir) and a ratio of 30.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              colin.Nonce,
			types.TxValueKeyFrom:               colin.Addr,
			types.TxValueKeyFeePayer:           reservoir.Addr,
			types.TxValueKeyTo:                 reservoir.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyFeeRatioOfFeePayer: uint8(30),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferWithRatio, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, colin.Key)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayer(signer, reservoir.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestSmartContractScenario tests the following scenario:
// 1. Deploy smart contract (reservoir -> contract)
// 2. Check the the smart contract is deployed well.
// 3. Execute "reward" function with amountToSend
// 4. Validate "reward" function is executed correctly by executing "balanceOf".
func TestSmartContractScenario(t *testing.T) {
	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(6, 4)
	if err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer bcdata.Shutdown()

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Key:   bcdata.privKeys[0],
		Nonce: uint64(0),
	}

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
	}

	contract, err := createHumanReadableAccount("ed34b0cf47a0021e9897760f0a904a69260c2f638e0bcc805facb745ec3ff9ab",
		"contract")
	assert.Equal(t, nil, err)

	gasPrice := new(big.Int).SetUint64(0)
	gasLimit := uint64(250000000)

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. Deploy smart contract (reservoir -> contract)
	// Since Circle-CI does not have "solc" now, skip compiling the source code.
	//{
	//	var txs types.Transactions
	//
	//	filename := string("../contracts/reward/contract/KlaytnReward.sol")
	//
	//	contracts, err := compiler.CompileSolidity("", filename)
	//	assert.Equal(t, nil, err)
	//
	//
	//	amount := new(big.Int).SetUint64(0)
	//
	//	for _, c := range contracts {
	//		fmt.Printf("%s", c.Code)
	//		values := map[types.TxValueKeyType]interface{}{
	//			types.TxValueKeyNonce:         reservoir.Nonce,
	//			types.TxValueKeyFrom:          reservoir.Addr,
	//			types.TxValueKeyTo:            contract.Addr,
	//			types.TxValueKeyAmount:        amount,
	//			types.TxValueKeyGasLimit:      gasLimit,
	//			types.TxValueKeyGasPrice:      gasPrice,
	//			types.TxValueKeyHumanReadable: true,
	//			types.TxValueKeyData:          common.FromHex(c.Code),
	//		}
	//		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
	//		assert.Equal(t, nil, err)
	//
	//		err = tx.Sign(signer, reservoir.Key)
	//		assert.Equal(t, nil, err)
	//
	//		txs = append(txs, tx)
	//
	//		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
	//			t.Fatal(err)
	//		}
	//		break
	//	}
	//	reservoir.Nonce += 1
	//}

	// TODO-Klaytn-RemoveLater: When Circle-CI is ready to use "solc", remove the below code and uncomment the above code.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(0)
		code := common.Hex2Bytes("608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029")

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            contract.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          code,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, reservoir.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 2. Check the the smart contract is deployed well.
	{
		statedb, err := bcdata.bc.State()
		if err != nil {
			t.Fatal(err)
		}
		code := statedb.GetCode(contract.Addr)
		assert.Equal(t, 478, len(code))
	}

	// 3. Execute "reward" function with amountToSend
	amountToSend := new(big.Int).SetUint64(10)
	{
		var txs types.Transactions

		abiStr := `[{"constant":true,"inputs":[],"name":"totalAmount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"receiver","type":"address"}],"name":"reward","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"safeWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"}]`
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("reward", reservoir.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            contract.Addr,
			types.TxValueKeyAmount:        amountToSend,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, reservoir.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 4. Validate "reward" function is executed correctly by executing "balanceOf".
	{
		amount := new(big.Int).SetUint64(0)

		abiStr := `[{"constant":true,"inputs":[],"name":"totalAmount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"receiver","type":"address"}],"name":"reward","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"safeWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"}]`
		abii, err := abi.JSON(strings.NewReader(string(abiStr)))
		assert.Equal(t, nil, err)

		data, err := abii.Pack("balanceOf", reservoir.Addr)
		assert.Equal(t, nil, err)

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            contract.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          data,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, reservoir.Key)
		assert.Equal(t, nil, err)

		ret, err := callContract(bcdata, tx)
		assert.Equal(t, nil, err)

		balance := new(big.Int)
		abii.Unpack(&balance, "balanceOf", ret)
		fmt.Println("balance", balance)

		assert.Equal(t, amountToSend, balance)
		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountUpdate tests a following scenario:
// 1. Transfer (reservoir -> anon) using a legacy transaction.
// 2. Key update of anon using AccountUpdate
// 3. Create an account decoupled using TxTypeAccountCreation.
// 4. Key update of decoupled using AccountUpdate
// 5. Transfer (decoupled -> reservoir) using TxTypeValueTransfer.
// 6. Create an account colin using TxTypeAccountCreation.
// 7. Key update of colin using AccountUpdate
// 8. Transfer (colin-> reservoir) using TxTypeValueTransfer.
func TestAccountUpdate(t *testing.T) {
	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(6, 4)
	if err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer bcdata.Shutdown()

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Key:   bcdata.privKeys[0],
		Nonce: uint64(0),
	}

	// anonymous account
	anon, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	// decoupled account
	decoupled, err := createDecoupledAccount("c64f2cd1196e2a1791365b00c4bc07ab8f047b73152e4617c6ed06ac221a4b0c",
		common.HexToAddress("0x75c3098be5e4b63fbac05838daaee378dd48098d"))
	assert.Equal(t, nil, err)

	colin, err := createHumanReadableAccount("ed580f5bd71a2ee4dae5cb43e331b7d0318596e561e6add7844271ed94156b20", "colin")
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
		fmt.Println("decoupledAddr = ", decoupled.Addr.String())
		fmt.Println("colinAddr = ", colin.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. Transfer (reservoir -> anon) using a legacy transaction.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(100000000000)
		tx := types.NewTransaction(reservoir.Nonce,
			anon.Addr, amount, gasLimit, gasPrice, []byte{})

		err := tx.Sign(signer, reservoir.Key)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 2. Key update of anon using AccountUpdate
	// This test should be failed because a legacy account does not have attribute `key`.
	//{
	// var txs types.Transactions
	//
	// newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
	// if err != nil {
	//   t.Fatal(err)
	// }
	//
	// values := map[types.TxValueKeyType]interface{}{
	//   types.TxValueKeyNonce:         anon.Nonce,
	//   types.TxValueKeyFrom:          anon.Addr,
	//   types.TxValueKeyGasLimit:      gasLimit,
	//   types.TxValueKeyGasPrice:      gasPrice,
	//   types.TxValueKeyAccountKey:    types.NewAccountKeyPublicWithValue(&newKey.PublicKey),
	// }
	// tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
	// assert.Equal(t, nil, err)
	//
	// err = tx.Sign(signer, anon.Key)
	// assert.Equal(t, nil, err)
	//
	// txs = append(txs, tx)
	//
	// if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
	//   t.Fatal(err)
	// }
	// anon.Nonce += 1
	//
	// anon.Key = newKey
	// // This should be failed.
	//}

	// 3. Create an account decoupled using TxTypeAccountCreation.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            decoupled.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    decoupled.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, reservoir.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 4. Key update of decoupled using AccountUpdate
	{
		var txs types.Transactions

		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		if err != nil {
			t.Fatal(err)
		}

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      decoupled.Nonce,
			types.TxValueKeyFrom:       decoupled.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: types.NewAccountKeyPublicWithValue(&newKey.PublicKey),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, decoupled.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		decoupled.Nonce += 1

		decoupled.Key = newKey
	}

	// 5. Transfer (decoupled -> reservoir) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    decoupled.Nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, decoupled.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		decoupled.Nonce += 1
	}

	// 6. Create an account colin using TxTypeAccountCreation.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            colin.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    colin.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, reservoir.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 7. Key update of colin using AccountUpdate
	{
		var txs types.Transactions

		newKey, err := crypto.HexToECDSA("41bd2b972564206658eab115f26ff4db617e6eb39c81a557adc18d8305d2f867")
		if err != nil {
			t.Fatal(err)
		}

		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:      colin.Nonce,
			types.TxValueKeyFrom:       colin.Addr,
			types.TxValueKeyGasLimit:   gasLimit,
			types.TxValueKeyGasPrice:   gasPrice,
			types.TxValueKeyAccountKey: types.NewAccountKeyPublicWithValue(&newKey.PublicKey),
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountUpdate, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, colin.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1

		colin.Key = newKey
	}

	// 8. Transfer (colin-> reservoir) using TxTypeValueTransfer.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(10000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    colin.Nonce,
			types.TxValueKeyFrom:     colin.Addr,
			types.TxValueKeyTo:       reservoir.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, values)
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, colin.Key)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		colin.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestValidateSender tests ValidateSender with all transaction types.
func TestValidateSender(t *testing.T) {
	// anonymous account
	anon, err := createAnonymousAccount("1da6dfcb52128060cdd2108edb786ca0aff4ef1fa537574286eeabe5c2ebd5ca")
	assert.Equal(t, nil, err)

	// decoupled account
	decoupled, err := createDecoupledAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab",
		common.HexToAddress("0x5104711f7faa9e2dadf593e43db1577a2887636f"))
	assert.Equal(t, nil, err)

	initialBalance := big.NewInt(1000000)

	statedb, _ := state.New(common.Hash{}, state.NewDatabase(database.NewMemoryDBManager()))
	statedb.CreateAccountWithMap(anon.Addr, state.ExternallyOwnedAccountType,
		map[state.AccountValueKeyType]interface{}{
			state.AccountValueKeyNonce:         nonce,
			state.AccountValueKeyBalance:       initialBalance,
			state.AccountValueKeyHumanReadable: false,
			state.AccountValueKeyAccountKey:    anon.AccKey,
		})

	statedb.CreateAccountWithMap(decoupled.Addr, state.ExternallyOwnedAccountType,
		map[state.AccountValueKeyType]interface{}{
			state.AccountValueKeyNonce:         rand.Uint64(),
			state.AccountValueKeyBalance:       initialBalance,
			state.AccountValueKeyHumanReadable: false,
			state.AccountValueKeyAccountKey:    decoupled.AccKey,
		})

	signer := types.MakeSigner(params.BFTTestChainConfig, big.NewInt(32))
	gasPrice := new(big.Int).SetUint64(0)
	gasLimit := uint64(2500000)
	amount := new(big.Int).SetUint64(10000)

	// LegacyTransaction
	{
		amount := new(big.Int).SetUint64(100000000000)
		tx := types.NewTransaction(anon.Nonce,
			decoupled.Addr, amount, gasLimit, gasPrice, []byte{})

		err := tx.Sign(signer, anon.Key)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, anon.Addr, actualFrom)
	}

	// TxTypeValueTransfer
	{
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    nonce,
			types.TxValueKeyFrom:     anon.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		})
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, anon.Key)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, anon.Addr, actualFrom)
	}

	// TxTypeValueTransfer
	{
		tx, err := types.NewTransactionWithMap(types.TxTypeValueTransfer, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		})
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, decoupled.Key)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, decoupled.Addr, actualFrom)
	}

	// TxTypeSmartContractDeploy
	{
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         nonce,
			types.TxValueKeyFrom:          decoupled.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			// The binary below is a compiled binary of contracts/reward/contract/KlaytnReward.sol.
			types.TxValueKeyData: common.Hex2Bytes("608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"),
		})
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, decoupled.Key)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, decoupled.Addr, actualFrom)
	}

	// TxTypeSmartContractExecution
	{
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyTo:       anon.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
			// A abi-packed bytes calling "reward" of contracts/reward/contract/KlaytnReward.sol with an address "bc5951f055a85f41a3b62fd6f68ab7de76d299b2".
			types.TxValueKeyData: common.Hex2Bytes("6353586b000000000000000000000000bc5951f055a85f41a3b62fd6f68ab7de76d299b2"),
		})
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, decoupled.Key)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, decoupled.Addr, actualFrom)
	}

	// TxTypeChainDataAnchoring
	{
		dummyBlock := types.NewBlock(&types.Header{
			GasLimit: gasLimit,
		}, nil, nil, nil)

		scData := types.NewChainHashes(dummyBlock)
		dataAnchoredRLP, _ := rlp.EncodeToBytes(scData)

		var txs types.Transactions

		amount := new(big.Int).SetUint64(0)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:        anon.Nonce,
			types.TxValueKeyFrom:         anon.Addr,
			types.TxValueKeyTo:           anon.Addr,
			types.TxValueKeyAmount:       amount,
			types.TxValueKeyGasLimit:     gasLimit,
			types.TxValueKeyGasPrice:     gasPrice,
			types.TxValueKeyAnchoredData: dataAnchoredRLP,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeChainDataAnchoring, values)
		if err != nil {
			t.Fatal(err)
		}
		signedTx, err := types.SignTx(tx, signer, anon.Key)
		if err != nil {
			t.Fatal(err)
		}
		txs = append(txs, signedTx)

		actualFrom, _, err := types.ValidateSender(signer, signedTx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, anon.Addr, actualFrom)
	}

	// TxTypeFeeDelegatedValueTransfer
	{
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransfer, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    nonce,
			types.TxValueKeyFrom:     decoupled.Addr,
			types.TxValueKeyFeePayer: anon.Addr,
			types.TxValueKeyTo:       decoupled.Addr,
			types.TxValueKeyAmount:   amount,
			types.TxValueKeyGasLimit: gasLimit,
			types.TxValueKeyGasPrice: gasPrice,
		})
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, decoupled.Key)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayer(signer, anon.Key)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, decoupled.Addr, actualFrom)

		actualFeePayer, _, err := types.ValidateFeePayer(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, anon.Addr, actualFeePayer)
	}

	// TxTypeFeeDelegatedValueTransferWithRatio
	{
		tx, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedValueTransferWithRatio, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              nonce,
			types.TxValueKeyFrom:               decoupled.Addr,
			types.TxValueKeyFeePayer:           anon.Addr,
			types.TxValueKeyTo:                 decoupled.Addr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyFeeRatioOfFeePayer: uint(30),
		})
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, decoupled.Key)
		assert.Equal(t, nil, err)

		err = tx.SignFeePayer(signer, anon.Key)
		assert.Equal(t, nil, err)

		actualFrom, _, err := types.ValidateSender(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, decoupled.Addr, actualFrom)

		actualFeePayer, _, err := types.ValidateFeePayer(signer, tx, statedb)
		assert.Equal(t, nil, err)
		assert.Equal(t, anon.Addr, actualFeePayer)
	}
}
