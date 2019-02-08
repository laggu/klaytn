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
// This file is derived from eth/metrics.go (2018/06/04).
// Modified and improved for the klaytn development.

package blockchain

import (
	"github.com/ground-x/klaytn/metrics"
)

var (
	cacheGetBlockBodyMissMeter = metrics.NewRegisteredMeter("klay/cache/get/blockbody/miss", nil)
	cacheGetBlockBodyHitMeter  = metrics.NewRegisteredMeter("klay/cache/get/blockbody/hit", nil)

	cacheGetBlockBodyRLPMissMeter = metrics.NewRegisteredMeter("klay/cache/get/blockbodyrlp/miss", nil)
	cacheGetBlockBodyRLPHitMeter  = metrics.NewRegisteredMeter("klay/cache/get/blockbodyrlp/hit", nil)

	cacheGetBlockMissMeter = metrics.NewRegisteredMeter("klay/cache/get/block/miss", nil)
	cacheGetBlockHitMeter  = metrics.NewRegisteredMeter("klay/cache/get/block/hit", nil)

	cacheGetFutureBlockMissMeter = metrics.NewRegisteredMeter("klay/cache/get/futureblock/miss", nil)
	cacheGetFutureBlockHitMeter  = metrics.NewRegisteredMeter("klay/cache/get/futureblock/hit", nil)

	cacheGetBadBlockMissMeter = metrics.NewRegisteredMeter("klay/cache/get/badblock/miss", nil)
	cacheGetBadBlockHitMeter  = metrics.NewRegisteredMeter("klay/cache/get/badblock/hit", nil)

	cacheGetRecentTransactionsMissMeter = metrics.NewRegisteredMeter("klay/cache/get/transactions/miss", nil)
	cacheGetRecentTransactionsHitMeter  = metrics.NewRegisteredMeter("klay/cache/get/transactions/hit", nil)

	cacheGetRecentBlockReceiptsMissMeter = metrics.NewRegisteredMeter("klay/cache/get/blockreceipts/miss", nil)
	cacheGetRecentBlockReceiptsHitMeter  = metrics.NewRegisteredMeter("klay/cache/get/blockreceipts/hit", nil)

	cacheGetRecentTxReceiptMissMeter = metrics.NewRegisteredMeter("klay/cache/get/txreceipt/miss", nil)
	cacheGetRecentTxReceiptHitMeter  = metrics.NewRegisteredMeter("klay/cache/get/txreceipt/hit", nil)

	cacheGetHeaderMissMeter = metrics.NewRegisteredMeter("klay/cache/get/header/miss", nil)
	cacheGetHeaderHitMeter  = metrics.NewRegisteredMeter("klay/cache/get/header/hit", nil)

	cacheGetTDMissMeter = metrics.NewRegisteredMeter("klay/cache/get/td/miss", nil)
	cacheGetTDHitMeter  = metrics.NewRegisteredMeter("klay/cache/get/td/hit", nil)

	cacheGetBlockNumberMissMeter = metrics.NewRegisteredMeter("klay/cache/get/blocknumber/miss", nil)
	cacheGetBlockNumberHitMeter  = metrics.NewRegisteredMeter("klay/cache/get/blocknumber/hit", nil)

	headBlockNumberGauge = metrics.NewRegisteredGauge("blockchain/head/blocknumber", nil)
	blockTxCountsMeter   = metrics.NewRegisteredMeter("blockchain/block/tx/rate", nil)
	blockTxCountsCounter = metrics.NewRegisteredCounter("blockchain/block/tx/counter", nil)

	txPoolPendingGauge = metrics.NewRegisteredGauge("tx/pool/pending/gauge", nil)
	txPoolQueueGauge   = metrics.NewRegisteredGauge("tx/pool/queue/gauge", nil)
)
