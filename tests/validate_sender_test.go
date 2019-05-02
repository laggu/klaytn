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
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/profile"
	"github.com/ground-x/klaytn/crypto"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
	"time"
)

func TestValidateSenderContract(t *testing.T) {
	prof := profile.NewProfiler()

	if isCompilerAvailable() == false {
		if testing.Verbose() {
			fmt.Printf("TestFeePayerContract is skipped due to the lack of solc.")
		}
		return
	}

	if testing.Verbose() {
		enableLog()
	}

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
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	multisig, err := createMultisigAccount(uint(2),
		[]uint{1, 1, 1},
		[]string{"bb113e82881499a7a361e8354a5b68f6c6885c7bcba09ea2b0891480396c322e",
			"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae59438e989",
			"c32c471b732e2f56103e2f8e8cfd52792ef548f05f326e546a7d1fbf9d0419ec"},
		common.HexToAddress("0xbbfa38050bf3167c887c086758f448ce067ea8ea"))

	multisig2, err := createMultisigAccount(uint(2),
		[]uint{1, 1, 1},
		[]string{"bb113e82881499a7a361e8354a5b68f6c6885c7bcba09ea2b0891480396c322f",
			"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae59438e98a",
			"c32c471b732e2f56103e2f8e8cfd52792ef548f05f326e546a7d1fbf9d0419ed"},
		common.HexToAddress("0xbbfa38050bf3167c887c086758f448ce067ea8ec"))

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)
	gasPrice := new(big.Int).SetUint64(bcdata.bc.Config().UnitPrice)

	// 1. Create an account multisig using TxTypeAccountCreation.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            multisig.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    multisig.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 2. Create an account multisig2 using TxTypeAccountCreation.
	{
		var txs types.Transactions

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            multisig2.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    multisig2.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		txs = append(txs, tx)

		if err := bcdata.GenABlockWithTransactions(accountMap, txs, prof); err != nil {
			t.Fatal(err)
		}
		reservoir.Nonce += 1
	}

	// 3. Deploy the contract `ValidateSender`.
	start = time.Now()
	filepath := "../contracts/validatesender/validate_sender.sol"
	contracts, err := deployContract(filepath, bcdata, accountMap, prof)
	if err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_deployContract", time.Now().Sub(start))

	c := contracts["../contracts/validatesender/validate_sender.sol:ValidateSenderContract"]
	abii, err := abi.JSON(strings.NewReader(c.abi))
	assert.Equal(t, nil, err)

	n, err := accountMap.GetNonce(*bcdata.addrs[0])
	assert.Equal(t, nil, err)

	// 4. Check if the validation is successful with valid parameters of multisig.
	{
		msg := crypto.Keccak256Hash([]byte{0x1})
		sigs := make([]byte, 65*2)
		s1, err := crypto.Sign(msg[:], multisig.Keys[0])
		assert.Equal(t, nil, err)
		s2, err := crypto.Sign(msg[:], multisig.Keys[1])
		assert.Equal(t, nil, err)

		copy(sigs[0:65], s1[0:65])
		copy(sigs[65:130], s2[0:65])

		data, err := abii.Pack("ValidateSender", multisig.Addr, msg, sigs)
		if err != nil {
			t.Fatal(err)
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    n,
			types.TxValueKeyGasPrice: big.NewInt(0),
			types.TxValueKeyGasLimit: uint64(5000000),
			types.TxValueKeyFrom:     *bcdata.addrs[0],
			types.TxValueKeyAmount:   big.NewInt(0),
			types.TxValueKeyTo:       c.address,
			types.TxValueKeyData:     data,
		})
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, bcdata.privKeys[0])
		assert.Equal(t, nil, err)

		// 3. Call the given function `ValidateSender`.
		ret, err := callContract(bcdata, tx)
		assert.Equal(t, nil, err)

		// 4. Check the returned value.
		var validated bool
		if err := abii.Unpack(&validated, "ValidateSender", ret); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, validated)
	}

	// 5. Check if the validation is successful with valid parameters of reservoir.
	{
		msg := crypto.Keccak256Hash([]byte{0x1})
		sigs := make([]byte, 65)
		s1, err := crypto.Sign(msg[:], reservoir.Keys[0])
		assert.Equal(t, nil, err)

		copy(sigs[0:65], s1[0:65])

		data, err := abii.Pack("ValidateSender", reservoir.Addr, msg, sigs)
		if err != nil {
			t.Fatal(err)
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    n,
			types.TxValueKeyGasPrice: big.NewInt(0),
			types.TxValueKeyGasLimit: uint64(5000000),
			types.TxValueKeyFrom:     *bcdata.addrs[0],
			types.TxValueKeyAmount:   big.NewInt(0),
			types.TxValueKeyTo:       c.address,
			types.TxValueKeyData:     data,
		})
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, bcdata.privKeys[0])
		assert.Equal(t, nil, err)

		// 3. Call the given function `ValidateSender`.
		ret, err := callContract(bcdata, tx)
		assert.Equal(t, nil, err)

		// 4. Check the returned value.
		var validated bool
		if err := abii.Unpack(&validated, "ValidateSender", ret); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, validated)
	}

	// 5. Check if the validation is failed with wrong signature.
	{
		msg := crypto.Keccak256Hash([]byte{0x1})
		sigs := make([]byte, 65)
		s1, err := crypto.Sign(msg[:], multisig.Keys[0])
		assert.Equal(t, nil, err)

		copy(sigs[0:65], s1[0:65])

		data, err := abii.Pack("ValidateSender", multisig.Addr, msg, sigs)
		if err != nil {
			t.Fatal(err)
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    n,
			types.TxValueKeyGasPrice: big.NewInt(0),
			types.TxValueKeyGasLimit: uint64(5000000),
			types.TxValueKeyFrom:     *bcdata.addrs[0],
			types.TxValueKeyAmount:   big.NewInt(0),
			types.TxValueKeyTo:       c.address,
			types.TxValueKeyData:     data,
		})
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, bcdata.privKeys[0])
		assert.Equal(t, nil, err)

		// 3. Call the given function `ValidateSender`.
		ret, err := callContract(bcdata, tx)
		assert.Equal(t, nil, err)

		// 4. Check the returned value.
		var validated bool
		if err := abii.Unpack(&validated, "ValidateSender", ret); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, validated)
	}

	// 6. Check if the validation is failed with signed by multisig but the address was set to mulsig2.
	{
		msg := crypto.Keccak256Hash([]byte{0x1})
		sigs := make([]byte, 65*2)
		s1, err := crypto.Sign(msg[:], multisig.Keys[0])
		assert.Equal(t, nil, err)
		s2, err := crypto.Sign(msg[:], multisig.Keys[1])
		assert.Equal(t, nil, err)

		copy(sigs[0:65], s1[0:65])
		copy(sigs[65:130], s2[0:65])

		data, err := abii.Pack("ValidateSender", multisig2.Addr, msg, sigs)
		if err != nil {
			t.Fatal(err)
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    n,
			types.TxValueKeyGasPrice: big.NewInt(0),
			types.TxValueKeyGasLimit: uint64(5000000),
			types.TxValueKeyFrom:     *bcdata.addrs[0],
			types.TxValueKeyAmount:   big.NewInt(0),
			types.TxValueKeyTo:       c.address,
			types.TxValueKeyData:     data,
		})
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, bcdata.privKeys[0])
		assert.Equal(t, nil, err)

		// 3. Call the given function `ValidateSender`.
		ret, err := callContract(bcdata, tx)
		assert.Equal(t, nil, err)

		// 4. Check the returned value.
		var validated bool
		if err := abii.Unpack(&validated, "ValidateSender", ret); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, validated)
	}

	// 7. Check if the validation is failed with an unknown address.
	{
		msg := crypto.Keccak256Hash([]byte{0x1})
		sigs := make([]byte, 65*2)
		s1, err := crypto.Sign(msg[:], multisig.Keys[0])
		assert.Equal(t, nil, err)
		s2, err := crypto.Sign(msg[:], multisig.Keys[1])
		assert.Equal(t, nil, err)

		copy(sigs[0:65], s1[0:65])
		copy(sigs[65:130], s2[0:65])

		addr, err := common.FromHumanReadableAddress("colin" + ".klaytn")
		assert.Equal(t, nil, err)

		data, err := abii.Pack("ValidateSender", addr, msg, sigs)
		if err != nil {
			t.Fatal(err)
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeSmartContractExecution, map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:    n,
			types.TxValueKeyGasPrice: big.NewInt(0),
			types.TxValueKeyGasLimit: uint64(5000000),
			types.TxValueKeyFrom:     *bcdata.addrs[0],
			types.TxValueKeyAmount:   big.NewInt(0),
			types.TxValueKeyTo:       c.address,
			types.TxValueKeyData:     data,
		})
		assert.Equal(t, nil, err)

		err = tx.Sign(signer, bcdata.privKeys[0])
		assert.Equal(t, nil, err)

		// 3. Call the given function `ValidateSender`.
		ret, err := callContract(bcdata, tx)
		assert.Equal(t, nil, err)

		// 4. Check the returned value.
		var validated bool
		if err := abii.Unpack(&validated, "ValidateSender", ret); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, validated)
	}
}
