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

//+build RLPEncodeTest

package types

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
)

var payerKey = defaultFeePayerKey()

// TestTxRLPEncode tests encoding transactions.
func TestTxRLPEncode(t *testing.T) {
	funcs := []testingF{
		testTxRLPEncodeLegacy,

		testTxRLPEncodeValueTransfer,
		testTxRLPEncodeFeeDelegatedValueTransfer,
		testTxRLPEncodeFeeDelegatedValueTransferWithRatio,

		testTxRLPEncodeValueTransferMemo,
		testTxRLPEncodeFeeDelegatedValueTransferMemo,
		testTxRLPEncodeFeeDelegatedValueTransferMemoWithRatio,

		testTxRLPEncodeAccountCreation,

		testTxRLPEncodeAccountUpdate,
		testTxRLPEncodeFeeDelegatedAccountUpdate,
		testTxRLPEncodeFeeDelegatedAccountUpdateWithRatio,

		testTxRLPEncodeSmartContractDeploy,
		testTxRLPEncodeFeeDelegatedSmartContractDeploy,
		testTxRLPEncodeFeeDelegatedSmartContractDeployWithRatio,

		testTxRLPEncodeSmartContractExecution,
		testTxRLPEncodeFeeDelegatedSmartContractExecution,
		testTxRLPEncodeFeeDelegatedSmartContractExecutionWithRatio,

		testTxRLPEncodeCancel,
		testTxRLPEncodeFeeDelegatedCancel,
		testTxRLPEncodeFeeDelegatedCancelWithRatio,

		testTxRLPEncodeChainDataAnchoring,
	}

	for _, f := range funcs {
		fnname := getFunctionName(f)
		fnname = fnname[strings.LastIndex(fnname, ".")+1:]
		t.Run(fnname, func(t *testing.T) {
			f(t)
		})
	}
}

func printRLPEncode(chainId *big.Int, signer Signer, sigRLP *bytes.Buffer, txHashRLP *bytes.Buffer, rawTx *Transaction) {
	privateKey := crypto.FromECDSA(key)

	fmt.Printf("chainid %#x\n", chainId)
	fmt.Printf("prvkey %#x\n", privateKey)
	fmt.Printf("pubkeyX %#x\n", key.X)
	fmt.Printf("pubkeyY %#x\n", key.Y)
	fmt.Printf("sigRLP %#x\n", sigRLP.Bytes())
	fmt.Printf("sigHash %s\n", signer.Hash(rawTx).String())
	fmt.Printf("TxHash %#x\n", txHashRLP.Bytes())
	fmt.Println(rawTx)
}

func printFeeDelegatedRLPEncode(t *testing.T, chainId *big.Int, signer Signer, sigRLP *bytes.Buffer, feePayerSigRLP *bytes.Buffer, txHashRLP *bytes.Buffer, rawTx *Transaction) {
	privateKey := crypto.FromECDSA(key)

	fmt.Printf("chainid %#x\n", chainId)
	// Sender
	fmt.Printf("prvkey %#x\n", privateKey)
	fmt.Printf("pubkeyX %#x\n", key.X)
	fmt.Printf("pubkeyY %#x\n", key.Y)
	fmt.Printf("sigRLP %#x\n", sigRLP.Bytes())
	fmt.Printf("sigHash %s\n", signer.Hash(rawTx).String())

	// FeePayer
	feePayerPrivateKey := crypto.FromECDSA(payerKey)

	feePyaerHash, err := signer.HashFeePayer(rawTx)
	assert.Equal(t, nil, err)

	fmt.Printf("FeePayerPrvkey %#x\n", feePayerPrivateKey)
	fmt.Printf("FeePayerPubkeyX %#x\n", payerKey.X)
	fmt.Printf("FeePayerPubkeyY %#x\n", payerKey.Y)
	fmt.Printf("FeePayerRLP %#x\n", feePayerSigRLP.Bytes())
	fmt.Printf("FeePayerHash %s\n", feePyaerHash.String())

	fmt.Printf("TxHash %#x\n", txHashRLP.Bytes())
	fmt.Println(rawTx)
}

func testTxRLPEncodeLegacy(t *testing.T) {
	tx := genLegacyTransaction().(*TxInternalDataLegacy)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)
	err := rlp.Encode(sigRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.Payload,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.Payload,
		tx.V,
		tx.R,
		tx.S,
	})
	assert.Equal(t, nil, err)

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeValueTransfer(t *testing.T) {
	tx := genValueTransferTransaction().(*TxInternalDataValueTransfer)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.TxSignatures,
	})

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeValueTransferMemo(t *testing.T) {
	tx := genValueTransferMemoTransaction().(*TxInternalDataValueTransferMemo)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.TxSignatures,
	})

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeAccountCreation(t *testing.T) {
	tx := genAccountCreationTransaction().(*TxInternalDataAccountCreation)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	serializer := accountkey.NewAccountKeySerializerWithAccountKey(tx.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.HumanReadable,
		keyEnc,
		tx.TxSignatures,
	})

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeAccountUpdate(t *testing.T) {
	tx := genAccountUpdateTransaction().(*TxInternalDataAccountUpdate)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	serializer := accountkey.NewAccountKeySerializerWithAccountKey(tx.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		keyEnc,
		tx.TxSignatures,
	})

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeSmartContractDeploy(t *testing.T) {
	tx := genSmartContractDeployTransaction().(*TxInternalDataSmartContractDeploy)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.HumanReadable,
		tx.TxSignatures,
	})

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeSmartContractExecution(t *testing.T) {
	tx := genSmartContractExecutionTransaction().(*TxInternalDataSmartContractExecution)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.TxSignatures,
	})

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeCancel(t *testing.T) {
	tx := genCancelTransaction().(*TxInternalDataCancel)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		tx.TxSignatures,
	})

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeChainDataAnchoring(t *testing.T) {
	tx := genChainDataTransaction().(*TxInternalDataChainDataAnchoring)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		tx.Payload,
		tx.TxSignatures,
	})

	printRLPEncode(chainId, signer, sigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedValueTransfer(t *testing.T) {
	tx := genFeeDelegatedValueTransferTransaction().(*TxInternalDataFeeDelegatedValueTransfer)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedValueTransferMemo(t *testing.T) {
	tx := genFeeDelegatedValueTransferMemoTransaction().(*TxInternalDataFeeDelegatedValueTransferMemo)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedAccountUpdate(t *testing.T) {
	tx := genFeeDelegatedAccountUpdateTransaction().(*TxInternalDataFeeDelegatedAccountUpdate)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	serializer := accountkey.NewAccountKeySerializerWithAccountKey(tx.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		keyEnc,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedSmartContractDeploy(t *testing.T) {
	tx := genFeeDelegatedSmartContractDeployTransaction().(*TxInternalDataFeeDelegatedSmartContractDeploy)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.HumanReadable,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedSmartContractExecution(t *testing.T) {
	tx := genFeeDelegatedSmartContractExecutionTransaction().(*TxInternalDataFeeDelegatedSmartContractExecution)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedCancel(t *testing.T) {
	tx := genFeeDelegatedCancelTransaction().(*TxInternalDataFeeDelegatedCancel)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedValueTransferWithRatio(t *testing.T) {
	tx := genFeeDelegatedValueTransferWithRatioTransaction().(*TxInternalDataFeeDelegatedValueTransferWithRatio)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.FeeRatio,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedValueTransferMemoWithRatio(t *testing.T) {
	tx := genFeeDelegatedValueTransferMemoWithRatioTransaction().(*TxInternalDataFeeDelegatedValueTransferMemoWithRatio)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.FeeRatio,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedAccountUpdateWithRatio(t *testing.T) {
	tx := genFeeDelegatedAccountUpdateWithRatioTransaction().(*TxInternalDataFeeDelegatedAccountUpdateWithRatio)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	serializer := accountkey.NewAccountKeySerializerWithAccountKey(tx.Key)
	keyEnc, _ := rlp.EncodeToBytes(serializer)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		keyEnc,
		tx.FeeRatio,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedSmartContractDeployWithRatio(t *testing.T) {
	tx := genFeeDelegatedSmartContractDeployWithRatioTransaction().(*TxInternalDataFeeDelegatedSmartContractDeployWithRatio)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.HumanReadable,
		tx.FeeRatio,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedSmartContractExecutionWithRatio(t *testing.T) {
	tx := genFeeDelegatedSmartContractExecutionWithRatioTransaction().(*TxInternalDataFeeDelegatedSmartContractExecutionWithRatio)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.From,
		tx.Payload,
		tx.FeeRatio,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, rawTx)
}

func testTxRLPEncodeFeeDelegatedCancelWithRatio(t *testing.T) {
	tx := genFeeDelegatedCancelWithRatioTransaction().(*TxInternalDataFeeDelegatedCancelWithRatio)

	signer := MakeSigner(params.BFTTestChainConfig, big.NewInt(2))
	chainId := params.BFTTestChainConfig.ChainID
	rawTx := &Transaction{data: tx}
	rawTx.Sign(signer, key)
	rawTx.SignFeePayer(signer, payerKey)

	sigRLP := new(bytes.Buffer)

	err := rlp.Encode(sigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	feePayerSigRLP := new(bytes.Buffer)

	err = rlp.Encode(feePayerSigRLP, []interface{}{
		tx.SerializeForSignToBytes(),
		tx.FeePayer,
		chainId,
		uint(0),
		uint(0),
	})
	assert.Equal(t, nil, err)

	txHashRLP := new(bytes.Buffer)
	err = rlp.Encode(txHashRLP, tx.Type())
	assert.Equal(t, nil, err)

	err = rlp.Encode(txHashRLP, []interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.From,
		tx.FeeRatio,
		tx.TxSignatures,
		tx.FeePayer,
		tx.FeePayerSignatures,
	})

	printFeeDelegatedRLPEncode(t, chainId, signer, sigRLP, feePayerSigRLP, txHashRLP, rawTx)
}

func defaultFeePayerKey() *ecdsa.PrivateKey {
	key, _ := crypto.HexToECDSA("b9d5558443585bca6f225b935950e3f6e69f9da8a5809a83f51c3365dff53936")
	return key
}
