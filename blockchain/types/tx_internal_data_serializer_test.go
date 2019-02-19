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
	"encoding/json"
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
	"testing"
)

var (
	to        = common.HexToAddress("7b65B75d204aBed71587c9E519a89277766EE1d0")
	key, from = defaultTestKey()
	feePayer  = common.HexToAddress("5A0043070275d9f6054307Ee7348bD660849D90f")
	nonce     = uint64(1234)
	amount    = big.NewInt(10)
	gasLimit  = uint64(999999999)
	gasPrice  = big.NewInt(25)
)

// TestTransactionSerialization tests RLP/JSON serialization for TxInternalData
func TestTransactionSerialization(t *testing.T) {
	var txs = []struct {
		Name string
		tx   TxInternalData
	}{
		{"OriginalTx", genLegacyTransaction()},
		{"SmartContractDeploy", genSmartContractDeployTransaction()},
		{"ValueTransfer", genValueTransferTransaction()},
		{"ChainDataTx", genChainDataTransaction()},
		{"AccountCreation", genAccountCreationTransaction()},
		{"AccountUpdate", genAccountUpdateTransaction()},
		{"FeeDelegatedValueTransfer", genFeeDelegatedValueTransferTransaction()},
		{"SmartContractExecution", genSmartContractExecutionTransaction()},
		{"FeeDelegatedValueTransferWithRatio", genFeeDelegatedValueTransferWithRatioTransaction()},
		{"Cancel", genCancelTransaction()},
	}

	var testcases = []struct {
		Name string
		fn   func(t *testing.T, tx TxInternalData)
	}{
		{"RLP", testTransactionRLP},
		{"JSON", testTransactionJSON},
	}

	for _, test := range testcases {
		for _, tx := range txs {
			Name := test.Name + "/" + tx.Name
			t.Run(Name, func(t *testing.T) {
				test.fn(t, tx.tx)
			})
		}
	}
}

func testTransactionRLP(t *testing.T, tx TxInternalData) {
	enc := newTxInternalDataSerializerWithValues(tx)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	b, err := rlp.EncodeToBytes(enc)
	if err != nil {
		panic(err)
	}

	dec := newTxInternalDataSerializer()

	if err := rlp.DecodeBytes(b, &dec); err != nil {
		panic(err)
	}

	if !tx.Equal(dec.tx) {
		t.Fatalf("tx != dec.tx\ntx=%v\ndec.tx=%v", tx, dec.tx)
	}
}

func testTransactionJSON(t *testing.T, tx TxInternalData) {
	enc := newTxInternalDataSerializerWithValues(tx)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	b, err := json.Marshal(enc)
	if err != nil {
		panic(err)
	}

	dec := newTxInternalDataSerializer()

	if err := json.Unmarshal(b, &dec); err != nil {
		panic(err)
	}

	if !tx.Equal(dec.tx) {
		t.Fatalf("tx != dec.tx\ntx=%v\ndec.tx=%v", tx, dec.tx)
	}
}

func genLegacyTransaction() TxInternalData {
	txdata, err := NewTxInternalDataWithMap(TxTypeLegacyTransaction, map[TxValueKeyType]interface{}{
		TxValueKeyNonce:    nonce,
		TxValueKeyTo:       to,
		TxValueKeyAmount:   amount,
		TxValueKeyGasLimit: gasLimit,
		TxValueKeyGasPrice: gasPrice,
		TxValueKeyData:     []byte("1234"),
	})

	if err != nil {
		// Since we do not have testing.T here, call panic() instead of t.Fatal().
		panic(err)
	}

	return txdata
}

func genValueTransferTransaction() TxInternalData {
	d, err := NewTxInternalDataWithMap(TxTypeValueTransfer, map[TxValueKeyType]interface{}{
		TxValueKeyNonce:    nonce,
		TxValueKeyTo:       to,
		TxValueKeyAmount:   amount,
		TxValueKeyGasLimit: gasLimit,
		TxValueKeyGasPrice: gasPrice,
		TxValueKeyFrom:     from,
	})

	if err != nil {
		// Since we do not have testing.T here, call panic() instead of t.Fatal().
		panic(err)
	}

	return d
}

func genSmartContractDeployTransaction() TxInternalData {
	d, err := NewTxInternalDataWithMap(TxTypeSmartContractDeploy, map[TxValueKeyType]interface{}{
		TxValueKeyNonce:         nonce,
		TxValueKeyAmount:        amount,
		TxValueKeyGasLimit:      gasLimit,
		TxValueKeyGasPrice:      gasPrice,
		TxValueKeyHumanReadable: true,
		TxValueKeyTo:            to,
		TxValueKeyFrom:          from,
		// The binary below is a compiled binary of contracts/reward/contract/KlaytnReward.sol.
		TxValueKeyData: common.Hex2Bytes("608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820627ca46bb09478a015762806cc00c431230501118c7c26c30ac58c4e09e51c4f0029"),
	})

	if err != nil {
		// Since we do not have testing.T here, call panic() instead of t.Fatal().
		panic(err)
	}

	return d
}

func genChainDataTransaction() TxInternalData {
	blockTxData := &ChainHashes{
		common.HexToHash("0"),
		common.HexToHash("1"),
		common.HexToHash("2"),
		common.HexToHash("3"),
		common.HexToHash("4"),
		big.NewInt(5)}

	anchoredData, err := rlp.EncodeToBytes(blockTxData)
	if err != nil {
		panic(err)
	}

	txdata, err := NewTxInternalDataWithMap(TxTypeChainDataAnchoring, map[TxValueKeyType]interface{}{
		TxValueKeyNonce:        nonce,
		TxValueKeyTo:           to,
		TxValueKeyFrom:         from,
		TxValueKeyAmount:       amount,
		TxValueKeyGasLimit:     gasLimit,
		TxValueKeyGasPrice:     gasPrice,
		TxValueKeyAnchoredData: anchoredData,
	})

	if err != nil {
		// Since we do not have testing.T here, call panic() instead of t.Fatal().
		panic(err)
	}

	return txdata
}

func genAccountCreationTransaction() TxInternalData {
	d, err := NewTxInternalDataWithMap(TxTypeAccountCreation, map[TxValueKeyType]interface{}{
		TxValueKeyNonce:         nonce,
		TxValueKeyTo:            to,
		TxValueKeyAmount:        amount,
		TxValueKeyGasLimit:      gasLimit,
		TxValueKeyGasPrice:      gasPrice,
		TxValueKeyFrom:          from,
		TxValueKeyHumanReadable: false,
		TxValueKeyAccountKey:    accountkey.NewAccountKeyPublicWithValue(&key.PublicKey),
	})

	if err != nil {
		// Since we do not have testing.T here, call panic() instead of t.Fatal().
		panic(err)
	}

	return d
}

func genFeeDelegatedValueTransferTransaction() TxInternalData {
	d, err := NewTxInternalDataWithMap(TxTypeFeeDelegatedValueTransfer, map[TxValueKeyType]interface{}{
		TxValueKeyNonce:    nonce,
		TxValueKeyTo:       to,
		TxValueKeyAmount:   amount,
		TxValueKeyGasLimit: gasLimit,
		TxValueKeyGasPrice: gasPrice,
		TxValueKeyFrom:     from,
		TxValueKeyFeePayer: feePayer,
	})

	if err != nil {
		// Since we do not have testing.T here, call panic() instead of t.Fatal().
		panic(err)
	}

	return d
}

func genFeeDelegatedValueTransferWithRatioTransaction() TxInternalData {
	d, err := NewTxInternalDataWithMap(TxTypeFeeDelegatedValueTransferWithRatio, map[TxValueKeyType]interface{}{
		TxValueKeyNonce:              nonce,
		TxValueKeyTo:                 to,
		TxValueKeyAmount:             amount,
		TxValueKeyGasLimit:           gasLimit,
		TxValueKeyGasPrice:           gasPrice,
		TxValueKeyFrom:               from,
		TxValueKeyFeePayer:           feePayer,
		TxValueKeyFeeRatioOfFeePayer: uint8(30),
	})

	if err != nil {
		// Since we do not have testing.T here, call panic() instead of t.Fatal().
		panic(err)
	}

	return d
}

func genSmartContractExecutionTransaction() TxInternalData {
	d, err := NewTxInternalDataWithMap(TxTypeSmartContractExecution, map[TxValueKeyType]interface{}{
		TxValueKeyNonce:    nonce,
		TxValueKeyTo:       to,
		TxValueKeyAmount:   amount,
		TxValueKeyGasLimit: gasLimit,
		TxValueKeyGasPrice: gasPrice,
		TxValueKeyFrom:     from,
		// A abi-packed bytes calling "reward" of contracts/reward/contract/KlaytnReward.sol with an address "bc5951f055a85f41a3b62fd6f68ab7de76d299b2".
		TxValueKeyData: common.Hex2Bytes("6353586b000000000000000000000000bc5951f055a85f41a3b62fd6f68ab7de76d299b2"),
	})

	if err != nil {
		// Since we do not have testing.T here, call panic() instead of t.Fatal().
		panic(err)
	}

	return d
}

func genAccountUpdateTransaction() TxInternalData {
	d, err := NewTxInternalDataWithMap(TxTypeAccountUpdate, map[TxValueKeyType]interface{}{
		TxValueKeyNonce:      nonce,
		TxValueKeyAmount:     amount,
		TxValueKeyGasLimit:   gasLimit,
		TxValueKeyGasPrice:   gasPrice,
		TxValueKeyFrom:       from,
		TxValueKeyAccountKey: accountkey.NewAccountKeyPublicWithValue(&key.PublicKey),
	})

	if err != nil {
		// Since we do not have testing.T here, call panic() instead of t.Fatal().
		panic(err)
	}

	return d
}

func genCancelTransaction() TxInternalData {
	d, err := NewTxInternalDataWithMap(TxTypeCancel, map[TxValueKeyType]interface{}{
		TxValueKeyNonce:    nonce,
		TxValueKeyAmount:   amount,
		TxValueKeyGasLimit: gasLimit,
		TxValueKeyGasPrice: gasPrice,
		TxValueKeyFrom:     from,
	})

	if err != nil {
		// Since we do not have testing.T here, call panic() instead of t.Fatal().
		panic(err)
	}

	return d
}
