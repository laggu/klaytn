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
	"crypto/rand"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/hexutil"
	"github.com/ground-x/klaytn/common/profile"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/kerrors"
	"github.com/ground-x/klaytn/params"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

// TestHumanReadableAddrValidationInTxPool tests human-readable address validation logic in TxPool.
// A human-readable account should have an alphanumeric address string with an alphabet in the current convention.
// Also, the address should consist of 5 ~ 13 characters.
func TestHumanReadableAddrValidationInTxPool(t *testing.T) {
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

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)
	amount := big.NewInt(100)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// creating account setting
	hex := make([]byte, 32)
	rand.Read(hex)
	EOAKey, _ := crypto.HexToECDSA(hexutil.Encode(hex)[2:])
	assert.Equal(t, nil, err)

	EOAAccKey := accountkey.NewAccountKeyPublicWithValue(&EOAKey.PublicKey)
	code := "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	invalidAddress := [][]byte{
		[]byte("ABCD.klaytn"),                // short address (length 4)
		[]byte("A23456789012341.klaytn"),     // long address (length 14)
		[]byte("1234567890.klaytn"),          // invalid address (number first)
		[]byte("ABCDE..klaytn"),              // invalid address (special character, dot)
		[]byte("ABCDEF"),                     // invalid address (no suffix)
		{0x66, 0x66, 0x66, 0x66, 0x88, 0x66}, // invalid address (non-alphanumeric)
	}

	// invalid address test
	for _, addr := range invalidAddress {
		var humanReadbleAddr common.Address
		humanReadbleAddr.SetBytesFromFront(addr)

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            humanReadbleAddr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    EOAAccKey,
		})
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrNotHumanReadableAddress, err)

		reservoir.Nonce += 1

		tx2, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            &humanReadbleAddr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)

		err = tx2.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx2)
		assert.Equal(t, kerrors.ErrNotHumanReadableAddress, err)

		reservoir.Nonce += 1

		tx3, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            &humanReadbleAddr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
			types.TxValueKeyFeePayer:      reservoir.Addr,
		})
		assert.Equal(t, nil, err)

		err = tx3.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx3.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx3)
		assert.Equal(t, kerrors.ErrNotHumanReadableAddress, err)

		reservoir.Nonce += 1

		tx4, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeployWithRatio, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 &humanReadbleAddr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyHumanReadable:      true,
			types.TxValueKeyData:               common.FromHex(code),
			types.TxValueKeyCodeFormat:         params.CodeFormatEVM,
			types.TxValueKeyFeePayer:           reservoir.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(99),
		})
		assert.Equal(t, nil, err)

		err = tx4.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx4.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx4)
		assert.Equal(t, kerrors.ErrNotHumanReadableAddress, err)

		reservoir.Nonce += 1
	}

	// valid address test
	{
		var humanReadbleAddr common.Address
		humanReadbleAddr.SetBytesFromFront([]byte("aidan.klaytn"))

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            humanReadbleAddr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    EOAAccKey,
		})
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)

		reservoir.Nonce += 1

		tx2, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            &humanReadbleAddr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)

		err = tx2.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx2)
		assert.Equal(t, nil, err)

		reservoir.Nonce += 1

		tx3, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            &humanReadbleAddr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
			types.TxValueKeyFeePayer:      reservoir.Addr,
		})
		assert.Equal(t, nil, err)

		err = tx3.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx3.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx3)
		assert.Equal(t, nil, err)

		reservoir.Nonce += 1

		tx4, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeployWithRatio, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 &humanReadbleAddr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyHumanReadable:      true,
			types.TxValueKeyData:               common.FromHex(code),
			types.TxValueKeyCodeFormat:         params.CodeFormatEVM,
			types.TxValueKeyFeePayer:           reservoir.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(99),
		})
		assert.Equal(t, nil, err)

		err = tx4.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx4.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx4)
		assert.Equal(t, nil, err)

		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestNonHumanReadableAddrValidationInTxPool tests non-human-readable address validation logic in TxPool.
// Non-human-readable account cannot have alphanumeric address w/ or w/o dot in the current convention.
func TestNonHumanReadableAddrValidationInTxPool(t *testing.T) {
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

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)
	amount := big.NewInt(100)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// creating account setting
	hex := make([]byte, 32)
	rand.Read(hex)
	EOAKey, _ := crypto.HexToECDSA(hexutil.Encode(hex)[2:])
	assert.Equal(t, nil, err)

	EOAAccKey := accountkey.NewAccountKeyPublicWithValue(&EOAKey.PublicKey)
	code := "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	invalidAddress := [][]byte{
		[]byte("ABC"),               // short alphanumeric
		[]byte("ABCDDE"),            // alphanumeric
		[]byte("ABCCDE12345"),       // alphanumeric
		[]byte("A2345678901234567"), // long alphanumeric
		[]byte("1234567890"),        // numeric
		[]byte("......."),           // dot
	}

	validAddress := [][]byte{
		{0x88, 0x88, 0x88, 0x88, 0x88, 0x88}, // non-alphanumeric
		{0x66, 0x66, 0x66, 0x66, 0x66, 0x88}, // alphanumeric + non-alphanumeric
		{0x77, 0x33, 0x22},                   // short non-alphanumeric
		{0x77, 0x33, 0x22, 0x33, 0x22, 0x33, 0x22, 0x33, 0x22, 0x33, 0x22, 0x33, 0x22, 0x33, 0x22}, // long non-alphanumeric
	}

	for _, addr := range invalidAddress {
		var humanReadbleAddr common.Address
		humanReadbleAddr.SetBytesFromFront(addr)

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            humanReadbleAddr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    EOAAccKey,
		})
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrNotNonHumanReadableAddress, err)

		reservoir.Nonce += 1

		tx2, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            &humanReadbleAddr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)

		err = tx2.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx2)
		assert.Equal(t, kerrors.ErrNotNonHumanReadableAddress, err)

		reservoir.Nonce += 1

		tx3, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            &humanReadbleAddr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
			types.TxValueKeyFeePayer:      reservoir.Addr,
		})
		assert.Equal(t, nil, err)

		err = tx3.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx3.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx3)
		assert.Equal(t, kerrors.ErrNotNonHumanReadableAddress, err)

		reservoir.Nonce += 1

		tx4, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeployWithRatio, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 &humanReadbleAddr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyHumanReadable:      false,
			types.TxValueKeyData:               common.FromHex(code),
			types.TxValueKeyCodeFormat:         params.CodeFormatEVM,
			types.TxValueKeyFeePayer:           reservoir.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(99),
		})
		assert.Equal(t, nil, err)

		err = tx4.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx4.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx4)
		assert.Equal(t, kerrors.ErrNotNonHumanReadableAddress, err)

		reservoir.Nonce += 1
	}

	for _, addr := range validAddress {
		var humanReadbleAddr common.Address
		humanReadbleAddr.SetBytesFromFront(addr)

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            humanReadbleAddr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    EOAAccKey,
		})
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)

		reservoir.Nonce += 1

		tx2, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            &humanReadbleAddr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)

		err = tx2.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx2)
		assert.Equal(t, nil, err)

		reservoir.Nonce += 1

		tx3, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            &humanReadbleAddr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
			types.TxValueKeyFeePayer:      reservoir.Addr,
		})
		assert.Equal(t, nil, err)

		err = tx3.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx3.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx3)
		assert.Equal(t, nil, err)

		reservoir.Nonce += 1

		tx4, err := types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeployWithRatio, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.Nonce,
			types.TxValueKeyFrom:               reservoir.Addr,
			types.TxValueKeyTo:                 &humanReadbleAddr,
			types.TxValueKeyAmount:             amount,
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyHumanReadable:      false,
			types.TxValueKeyData:               common.FromHex(code),
			types.TxValueKeyCodeFormat:         params.CodeFormatEVM,
			types.TxValueKeyFeePayer:           reservoir.Addr,
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(99),
		})
		assert.Equal(t, nil, err)

		err = tx4.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = tx4.SignFeePayerWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx4)
		assert.Equal(t, nil, err)

		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestValidatingUnavailableContractExecution tests validation logic of invalid contract execution transaction.
// TxPool will invalidate contract execution transactions sending to un-executable account even though the recipient is a contract account.
func TestValidatingUnavailableContractExecution(t *testing.T) {
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

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// various accounts preparation
	contract, err := createDefaultAccount(accountkey.AccountKeyTypePublic)
	assert.Equal(t, nil, err)

	contractInvalid, err := createDefaultAccount(accountkey.AccountKeyTypePublic)
	assert.Equal(t, nil, err)

	legacyEOA, err := createDefaultAccount(accountkey.AccountKeyTypeLegacy)
	assert.Equal(t, nil, err)

	EOA, err := createDefaultAccount(accountkey.AccountKeyTypePublic)
	assert.Equal(t, nil, err)

	code = "0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"

	// Prepare. creates various types of accounts tried to be executed
	{
		var txs types.Transactions
		amount := big.NewInt(100000)

		// tx to create a contract account
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyAmount:        new(big.Int).SetUint64(0),
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyTo:            &contract.Addr,
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyData:          common.FromHex(code),
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)
		reservoir.Nonce += 1

		// tx2 to create a invalid contract account
		values = map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyAmount:        new(big.Int).SetUint64(0),
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyTo:            &contractInvalid.Addr,
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyData:          []byte{}, // the invalid contract doesn't have contract code
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		}
		tx2, err := types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, values)
		assert.Equal(t, nil, err)

		err = tx2.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx2)
		reservoir.Nonce += 1

		// tx3 to create a legacy EOA account
		tx3 := types.NewTransaction(reservoir.GetNonce(), legacyEOA.GetAddr(), amount, gasLimit, gasPrice, []byte{})
		assert.Equal(t, nil, err)

		err = tx3.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx3)
		reservoir.Nonce += 1

		// tx4 to create an EOA account
		tx4, _, err := generateDefaultTx(reservoir, EOA, types.TxTypeAccountCreation)
		assert.Equal(t, nil, err)

		err = tx4.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx4)
		reservoir.Nonce += 1

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
	}

	// make TxPool to test validation in 'TxPool add' process
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// 1. contract execution transaction to the contract account.
	{
		tx, _ := genSmartContractExecution(t, signer, reservoir, contract, nil, gasPrice)

		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
	}

	// 2. contract execution transaction to the invalid contract account.
	{
		tx, _ := genSmartContractExecution(t, signer, reservoir, contractInvalid, nil, gasPrice)

		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrNotProgramAccount, err)
	}

	// 3. contract execution transaction to the Legacy EOA account.
	{
		tx, _ := genSmartContractExecution(t, signer, reservoir, legacyEOA, nil, gasPrice)

		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrNotProgramAccount, err)
	}

	// 4. contract execution transaction to the EOA account.
	{
		tx, _ := genSmartContractExecution(t, signer, reservoir, EOA, nil, gasPrice)

		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrNotProgramAccount, err)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestValidatingTxToPrecompiledContractAddress tests validation logic for the tx sending to reserved addresses.
// PrecompiledContractAddresses (0x0001 ~ 0x03FF) are used for system services such as pre-compiled contracts.
// Therefore, the addresses should be called only by the internal call.
func TestValidatingTxToPrecompiledContractAddress(t *testing.T) {
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

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// make txpool
	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	account, err := createDefaultAccount(accountkey.AccountKeyTypePublic)
	assert.Equal(t, nil, err)

	// 1. tx (reservoir -> 0x03FF)
	{
		// The test address is 0x00000000000000000000000000000000000003FF
		account.Addr = common.BytesToAddress(hexutil.MustDecode("0x00000000000000000000000000000000000003FF"))

		tx := genLegacyValueTransfer(signer, reservoir, account)
		// txpool insert check
		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		// block tx check
		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)

		tx, _ = genValueTransfer(t, signer, reservoir, account, reservoir, gasPrice)
		// txpool insert check
		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		// block tx check
		receipt, _, err = applyTransaction(t, bcdata, tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)

		tx, _ = genFeeDelegatedValueTransfer(t, signer, reservoir, account, reservoir, gasPrice)
		// txpool insert check
		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		// block tx check
		receipt, _, err = applyTransaction(t, bcdata, tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)

		tx, _ = genFeeDelegatedWithRatioValueTransfer(t, signer, reservoir, account, reservoir, gasPrice)
		// txpool insert check
		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		// block tx check
		receipt, _, err = applyTransaction(t, bcdata, tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)

		tx, _ = genValueTransferWithMemo(t, signer, reservoir, account, reservoir, gasPrice)
		// txpool insert check
		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		// block tx check
		receipt, _, err = applyTransaction(t, bcdata, tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)

		tx, _ = genFeeDelegatedValueTransferWithMemo(t, signer, reservoir, account, reservoir, gasPrice)
		// txpool insert check
		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		// block tx check
		receipt, _, err = applyTransaction(t, bcdata, tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)

		tx, _ = genFeeDelegatedWithRatioValueTransferWithMemo(t, signer, reservoir, account, reservoir, gasPrice)
		// txpool insert check
		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		// block tx check
		receipt, _, err = applyTransaction(t, bcdata, tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)

		tx, err = types.NewTransactionWithMap(types.TxTypeAccountCreation, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            account.GetAddr(),
			types.TxValueKeyAmount:        big.NewInt(0),
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    account.GetAccKey(),
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		// txpool insert check
		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		// block tx check
		receipt, _, err = applyTransaction(t, bcdata, tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)

		tx, err = types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            &account.Addr,
			types.TxValueKeyAmount:        big.NewInt(0),
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          []byte{},
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		// txpool insert check
		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		// block tx check
		receipt, _, err = applyTransaction(t, bcdata, tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)

		tx, err = types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            &account.Addr,
			types.TxValueKeyAmount:        big.NewInt(0),
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          []byte{},
			types.TxValueKeyFeePayer:      reservoir.GetAddr(),
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		err = tx.SignFeePayerWithKeys(signer, reservoir.GetFeeKeys())
		assert.Equal(t, nil, err)
		// txpool insert check
		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		// block tx check
		receipt, _, err = applyTransaction(t, bcdata, tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)

		tx, err = types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeployWithRatio, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.GetNonce(),
			types.TxValueKeyFrom:               reservoir.GetAddr(),
			types.TxValueKeyTo:                 &account.Addr,
			types.TxValueKeyAmount:             big.NewInt(0),
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyHumanReadable:      false,
			types.TxValueKeyData:               []byte{},
			types.TxValueKeyFeePayer:           reservoir.GetAddr(),
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
			types.TxValueKeyCodeFormat:         params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		err = tx.SignFeePayerWithKeys(signer, reservoir.GetFeeKeys())
		assert.Equal(t, nil, err)
		// txpool insert check
		err = txpool.AddRemote(tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		// block tx check
		receipt, _, err = applyTransaction(t, bcdata, tx)
		assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)

		// Contract Execution Txs is filtered out before the validation of the recipient address.
		// It should be tested after the installation of a smart contract on the target address.
		//tx, _ = genSmartContractExecution(t, signer, reservoir, account, reservoir, gasPrice)
		//// txpool insert check
		//err = txpool.AddRemote(tx)
		//assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		//// block tx check
		//receipt, _, err = applyTransaction(t, bcdata, tx)
		//assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		//assert.Equal(t, (*types.Receipt)(nil), receipt)
		//
		//tx, _ = genFeeDelegatedSmartContractExecution(t, signer, reservoir, account, reservoir, gasPrice)
		//// txpool insert check
		//err = txpool.AddRemote(tx)
		//assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		//// block tx check
		//receipt, _, err = applyTransaction(t, bcdata, tx)
		//assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		//assert.Equal(t, (*types.Receipt)(nil), receipt)
		//
		//tx, _ = genFeeDelegatedWithRatioSmartContractExecution(t, signer, reservoir, account, reservoir, gasPrice)
		//// txpool insert check
		//err = txpool.AddRemote(tx)
		//assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		//// block tx check
		//receipt, _, err = applyTransaction(t, bcdata, tx)
		//assert.Equal(t, kerrors.ErrPrecompiledContractAddress, err)
		//assert.Equal(t, (*types.Receipt)(nil), receipt)
	}

	// 2. tx (reservoir -> 0x0400)
	{
		// The test address is 0x400
		account.Addr = common.BytesToAddress(hexutil.MustDecode("0x0000000000000000000000000000000000000400"))

		tx := genLegacyValueTransfer(signer, reservoir, account)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genValueTransfer(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genFeeDelegatedValueTransfer(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genFeeDelegatedWithRatioValueTransfer(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genValueTransferWithMemo(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genFeeDelegatedValueTransferWithMemo(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genFeeDelegatedWithRatioValueTransferWithMemo(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, err = types.NewTransactionWithMap(types.TxTypeAccountCreation, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            account.GetAddr(),
			types.TxValueKeyAmount:        big.NewInt(0),
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    account.GetAccKey(),
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, err = types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            &account.Addr,
			types.TxValueKeyAmount:        big.NewInt(0),
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          []byte{},
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, err = types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            &account.Addr,
			types.TxValueKeyAmount:        big.NewInt(0),
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          []byte{},
			types.TxValueKeyFeePayer:      reservoir.GetAddr(),
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		err = tx.SignFeePayerWithKeys(signer, reservoir.GetFeeKeys())
		assert.Equal(t, nil, err)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, err = types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeployWithRatio, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.GetNonce(),
			types.TxValueKeyFrom:               reservoir.GetAddr(),
			types.TxValueKeyTo:                 &account.Addr,
			types.TxValueKeyAmount:             big.NewInt(0),
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyHumanReadable:      false,
			types.TxValueKeyData:               []byte{},
			types.TxValueKeyFeePayer:           reservoir.GetAddr(),
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
			types.TxValueKeyCodeFormat:         params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		err = tx.SignFeePayerWithKeys(signer, reservoir.GetFeeKeys())
		assert.Equal(t, nil, err)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		// Contract Execution Txs is filtered out before the validation of the recipient address.
		// It should be tested after the installation of a smart contract on the target address.
		//tx, _ = genSmartContractExecution(t, signer, reservoir, account, reservoir, gasPrice)
		//err = txpool.AddRemote(tx)
		//assert.Equal(t, nil, err)
		//reservoir.Nonce++
		//
		//tx, _ = genFeeDelegatedSmartContractExecution(t, signer, reservoir, account, reservoir, gasPrice)
		//err = txpool.AddRemote(tx)
		//assert.Equal(t, nil, err)
		//reservoir.Nonce++
		//
		//tx, _ = genFeeDelegatedWithRatioSmartContractExecution(t, signer, reservoir, account, reservoir, gasPrice)
		//err = txpool.AddRemote(tx)
		//assert.Equal(t, nil, err)
		//reservoir.Nonce++
	}

	// 3. tx (reservoir -> 0x00)
	{
		// The test address is 0x00
		account.Addr = common.BytesToAddress(hexutil.MustDecode("0x0000000000000000000000000000000000000000"))

		tx := genLegacyValueTransfer(signer, reservoir, account)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genValueTransfer(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genFeeDelegatedValueTransfer(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genFeeDelegatedWithRatioValueTransfer(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genValueTransferWithMemo(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genFeeDelegatedValueTransferWithMemo(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genFeeDelegatedWithRatioValueTransferWithMemo(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, err = types.NewTransactionWithMap(types.TxTypeAccountCreation, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            account.GetAddr(),
			types.TxValueKeyAmount:        big.NewInt(0),
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    account.GetAccKey(),
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, err = types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            &account.Addr,
			types.TxValueKeyAmount:        big.NewInt(0),
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          []byte{},
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, err = types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            &account.Addr,
			types.TxValueKeyAmount:        big.NewInt(0),
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          []byte{},
			types.TxValueKeyFeePayer:      reservoir.GetAddr(),
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		err = tx.SignFeePayerWithKeys(signer, reservoir.GetFeeKeys())
		assert.Equal(t, nil, err)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, err = types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeployWithRatio, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.GetNonce(),
			types.TxValueKeyFrom:               reservoir.GetAddr(),
			types.TxValueKeyTo:                 &account.Addr,
			types.TxValueKeyAmount:             big.NewInt(0),
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyHumanReadable:      false,
			types.TxValueKeyData:               []byte{},
			types.TxValueKeyFeePayer:           reservoir.GetAddr(),
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
			types.TxValueKeyCodeFormat:         params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		err = tx.SignFeePayerWithKeys(signer, reservoir.GetFeeKeys())
		assert.Equal(t, nil, err)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		// Contract Execution Txs is filtered out before the validation of the recipient address.
		// It should be tested after the installation of a smart contract on the target address.
		//tx, _ = genSmartContractExecution(t, signer, reservoir, account, reservoir, gasPrice)
		//err = txpool.AddRemote(tx)
		//assert.Equal(t, nil, err)
		//reservoir.Nonce++
		//
		//tx, _ = genFeeDelegatedSmartContractExecution(t, signer, reservoir, account, reservoir, gasPrice)
		//err = txpool.AddRemote(tx)
		//assert.Equal(t, nil, err)
		//reservoir.Nonce++
		//
		//tx, _ = genFeeDelegatedWithRatioSmartContractExecution(t, signer, reservoir, account, reservoir, gasPrice)
		//err = txpool.AddRemote(tx)
		//assert.Equal(t, nil, err)
		//reservoir.Nonce++
	}

	// 4. tx (reservoir -> 0x0101000000000000000000000000000000000000)
	{
		// The test address is 0x0101000000000000000000000000000000000000
		account.Addr = common.BytesToAddress(hexutil.MustDecode("0x0101000000000000000000000000000000000000"))

		tx := genLegacyValueTransfer(signer, reservoir, account)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genValueTransfer(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genFeeDelegatedValueTransfer(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genFeeDelegatedWithRatioValueTransfer(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genValueTransferWithMemo(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genFeeDelegatedValueTransferWithMemo(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, _ = genFeeDelegatedWithRatioValueTransferWithMemo(t, signer, reservoir, account, reservoir, gasPrice)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, err = types.NewTransactionWithMap(types.TxTypeAccountCreation, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            account.GetAddr(),
			types.TxValueKeyAmount:        big.NewInt(0),
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    account.GetAccKey(),
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, err = types.NewTransactionWithMap(types.TxTypeSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            &account.Addr,
			types.TxValueKeyAmount:        big.NewInt(0),
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          []byte{},
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, err = types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeploy, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            &account.Addr,
			types.TxValueKeyAmount:        big.NewInt(0),
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyData:          []byte{},
			types.TxValueKeyFeePayer:      reservoir.GetAddr(),
			types.TxValueKeyCodeFormat:    params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		err = tx.SignFeePayerWithKeys(signer, reservoir.GetFeeKeys())
		assert.Equal(t, nil, err)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		tx, err = types.NewTransactionWithMap(types.TxTypeFeeDelegatedSmartContractDeployWithRatio, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:              reservoir.GetNonce(),
			types.TxValueKeyFrom:               reservoir.GetAddr(),
			types.TxValueKeyTo:                 &account.Addr,
			types.TxValueKeyAmount:             big.NewInt(0),
			types.TxValueKeyGasLimit:           gasLimit,
			types.TxValueKeyGasPrice:           gasPrice,
			types.TxValueKeyHumanReadable:      false,
			types.TxValueKeyData:               []byte{},
			types.TxValueKeyFeePayer:           reservoir.GetAddr(),
			types.TxValueKeyFeeRatioOfFeePayer: types.FeeRatio(30),
			types.TxValueKeyCodeFormat:         params.CodeFormatEVM,
		})
		assert.Equal(t, nil, err)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)
		err = tx.SignFeePayerWithKeys(signer, reservoir.GetFeeKeys())
		assert.Equal(t, nil, err)
		err = txpool.AddRemote(tx)
		assert.Equal(t, nil, err)
		reservoir.Nonce++

		// Contract Execution Txs is filtered out before the validation of the recipient address.
		// It should be tested after the installation of a smart contract on the target address.
		//tx, _ = genSmartContractExecution(t, signer, reservoir, account, reservoir, gasPrice)
		//err = txpool.AddRemote(tx)
		//assert.Equal(t, nil, err)
		//reservoir.Nonce++
		//
		//tx, _ = genFeeDelegatedSmartContractExecution(t, signer, reservoir, account, reservoir, gasPrice)
		//err = txpool.AddRemote(tx)
		//assert.Equal(t, nil, err)
		//reservoir.Nonce++
		//
		//tx, _ = genFeeDelegatedWithRatioSmartContractExecution(t, signer, reservoir, account, reservoir, gasPrice)
		//err = txpool.AddRemote(tx)
		//assert.Equal(t, nil, err)
		//reservoir.Nonce++
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestEmptyAccountGeneration tries to generate the same account twice with zero value.
// The second try will return `kerrors.ErrAccountAlreadyExists` since the account is generated at the first time.
// The stateDB does not remove an empty account which has zero value and zero nonce.
func TestEmptyAccountGeneration(t *testing.T) {
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

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	poolSlots := 1000
	txpoolconfig := blockchain.DefaultTxPoolConfig
	txpoolconfig.Journal = ""
	txpoolconfig.ExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.NonExecSlotsAccount = uint64(poolSlots)
	txpoolconfig.ExecSlotsAll = 2 * uint64(poolSlots)
	txpoolconfig.NonExecSlotsAll = 2 * uint64(poolSlots)
	txpool := blockchain.NewTxPool(txpoolconfig, bcdata.bc.Config(), bcdata.bc)

	EOA, err := createDefaultAccount(accountkey.AccountKeyTypePublic)
	assert.Equal(t, nil, err)

	legacy, err := createDefaultAccount(accountkey.AccountKeyTypeLegacy)
	assert.Equal(t, nil, err)

	legacyForLegacyTx, err := createDefaultAccount(accountkey.AccountKeyTypeLegacy)
	assert.Equal(t, nil, err)

	amount := big.NewInt(0)

	// zero value legacy transaction to non-existing address (result: no creation)
	{
		var txs types.Transactions

		tx := types.NewTransaction(reservoir.Nonce, legacyForLegacyTx.Addr, big.NewInt(0), gasLimit, gasPrice, nil)
		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)

		txs = append(txs, tx)
		reservoir.Nonce += 1

		// one more creation
		tx2 := types.NewTransaction(reservoir.Nonce, legacyForLegacyTx.Addr, big.NewInt(0), gasLimit, gasPrice, nil)
		err = tx2.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)

		txs = append(txs, tx2)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// zero value account creation with legacy key (result: no creation)
	{
		var txs types.Transactions

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            legacy.GetAddr(),
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    legacy.GetAccKey(),
		})
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)

		txs = append(txs, tx)
		reservoir.Nonce += 1

		// one more creation
		tx2, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            legacy.GetAddr(),
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    legacy.GetAccKey(),
		})
		assert.Equal(t, nil, err)

		err = tx2.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)

		txs = append(txs, tx2)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// zero value account creation with public key (result: creation)
	{
		var txs types.Transactions

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            EOA.GetAddr(),
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    EOA.GetAccKey(),
		})
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1

		// one more creation
		tx2, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.GetNonce(),
			types.TxValueKeyFrom:          reservoir.GetAddr(),
			types.TxValueKeyTo:            EOA.GetAddr(),
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    EOA.GetAccKey(),
		})
		assert.Equal(t, nil, err)

		err = tx2.SignWithKeys(signer, reservoir.GetTxKeys())
		assert.Equal(t, nil, err)

		err = txpool.AddRemote(tx2)
		assert.Equal(t, kerrors.ErrAccountAlreadyExists, err)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}
