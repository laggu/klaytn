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

// +build ServiceChainMultiNodeTest

package tests

import (
	"context"
	"fmt"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"time"

	"github.com/ground-x/klaytn/client"
	"testing"
)

// TestAnchoringServiceChainDataWithAPIs tests main scenario of anchoring chain data in service chain.
// - checking if service chain execution options are set.
// - testing of connection between main chain and service chain.
// - checking if the anchoring tx is in main chain block.
// - checking if the tx is indexed by the block hash of child chain.
// - checking is the receipt of the tx is stored in child chain.
func TestAnchoringServiceChainDataWithAPIs(t *testing.T) {
	// If you want to test this, you should run klaytn mainnet and service chain on your local
	// and set the below parameter.
	// And then run this test manually.
	// For now, this test is excluded in CI test.
	// TODO-Klaytn-ServiceChain This test will be included in CI test as a multi-node test.

	//Parent node's RPC endpoint.
	pEndPoint := "http://localhost:8546"

	//Parent node's KNI for child node to connect with parent node.
	parentKNI := "kni://fee1d3c14875c6b623a5d8d2cc0a83ed8d37207f20247fd238ab44ae8a1538e047280dc15bd86787dc3ca424bd798e9dad5a3047b5328ee9b4f764fcc971cd83@[::]:30304?discport=0"

	//Child node's RPC endpoint.
	cEndPoint := "http://localhost:7545"

	//Child chain address can be assumed by the address which have much klay on parent chain.
	expectedChainAddr := common.HexToAddress("0x569FeB666C99da7D7DD31e3EE7ACcf4c8fe9c054")

	//Default values for option parameters.
	expectedAnchoringPeriod := uint64(1)
	expectedChainTxLimit := uint64(100)

	//Get RPC client of Parent/Child nodes.
	pClient, err := client.Dial(pEndPoint)
	if err != nil {
		t.Fatalf("parent Dial Err : %v", err)
	}
	cClient, err := client.Dial(cEndPoint)
	if err != nil {
		t.Fatalf("child Dial Err : %v", err)
	}

	ctx := context.Background()

	//Remove the parent peer on the child node to disconnect each other of abnormal previous test.
	result, err := cClient.RemovePeer(ctx, parentKNI)
	if err != nil || !result {
		t.Fatalf("child RemovePeer Err %v", err)
	}

	//Add parent peer on child node to connect each other.
	result, err = cClient.AddPeerOnParentChain(ctx, parentKNI)
	if err != nil || !result {
		t.Fatalf("child AddPeerOnParentChain Err %v", err)
	}

	//Check if parent peer is enabled to index chain tx.
	result, err = pClient.GetChildChainIndexingEnabled(ctx)
	if err != nil {
		t.Fatalf("parent GetChildChainIndexingEnabled Err %v", err)
	}
	if result == false {
		t.Fatalf("parent GetChildChainIndexingEnabled Result %v", result)
	}

	//Check if mainChainAccountAddr is ok.
	mainChainAccountAddr, err := cClient.GetMainChainAccountAddr(ctx)
	if err != nil {
		t.Fatalf("parent GetMainChainAccountAddr Err %v", err)
	}
	if mainChainAccountAddr != expectedChainAddr {
		t.Fatalf("parent GetMainChainAccountAddr mainChainAccountAddr %v", mainChainAccountAddr.String())
	}

	//Check if GetAnchoringPeriod is ok.
	anchoringPeriod, err := cClient.GetAnchoringPeriod(ctx)
	if err != nil {
		t.Fatalf("parent GetAnchoringPeriod Err %v", err)
	}
	if anchoringPeriod != expectedAnchoringPeriod {
		t.Fatalf("parent GetAnchoringPeriod anchoringPeriod %v", anchoringPeriod)
	}

	//Check if getSentChainTxsLimit is ok.
	chainTxLimit, err := cClient.GetSentChainTxsLimit(ctx)
	if err != nil {
		t.Fatalf("parent GetSentChainTxsLimit Err %v", err)
	}
	if chainTxLimit != expectedChainTxLimit {
		t.Fatalf("parent GetSentChainTxsLimit chainTxLimit %v", chainTxLimit)
	}

	//Delay for connecting each other and starting anchoring feature.
	time.Sleep(5 * time.Second)

	//Get the block number of the child chain.
	cBlkNumber, err := cClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("child BlockNumber Err %v", err)
	}

	fmt.Printf("child block number = %v\n", cBlkNumber.String())

	//Get the block of the child chain.
	cBlock, err := cClient.BlockByNumber(ctx, cBlkNumber)
	if err != nil {
		t.Fatalf("child BlockByNumber Err %v", err)
	}
	fmt.Printf("child Block hash = %v\n", cBlock.Hash().String())

	//Delay for creating/propagating/executing the tx, and storing the receipt of the tx on child chain.
	time.Sleep(10 * time.Second)

	//Get the anchoring tx hash of the block hash.
	txHash, err := pClient.ConvertChildChainBlockHashToParentChainTxHash(ctx, cBlock.Hash())
	if err != nil {
		t.Fatalf("parent anchoring tx hash Err %v", err)
	}
	fmt.Printf("parent anchoring tx hash = %v\n", txHash.String())

	//Get the receipt of the anchoring tx.
	receipt, err := cClient.GetReceiptFromParentChain(ctx, cBlock.Hash())
	if err != nil {
		t.Fatalf("child GetReceiptFromParentChain Err %v", err)
	}
	fmt.Printf("child GetReceiptFromParentChain = %v\n", receipt.TxHash.String())

	//Check if tx hash is same with the receipt tx hash.
	if receipt.TxHash != txHash || txHash == types.EmptyRootHash {
		t.Fatalf("parent/child tx has is not matched. parent:%v, child:%v", txHash.String(), receipt.TxHash.String())
	}

	//Remove parent peer on child node to connect each other after completing this test.
	result, err = cClient.RemovePeer(ctx, parentKNI)
	if err != nil || !result {
		t.Fatalf("child RemovePeer Err %v", err)
	}
}
