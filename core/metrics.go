package core

import (
	"github.com/ground-x/go-gxplatform/metrics"
)

var (
	cacheGetBlockBodyTryMeter = metrics.NewRegisteredMeter("klay/cache/get/blockbody/try", nil)
	cacheGetBlockBodyHitMeter = metrics.NewRegisteredMeter("klay/cache/get/blockbody/hit", nil)

	cacheGetBlockBodyRLPTryMeter = metrics.NewRegisteredMeter("klay/cache/get/blockbodyrlp/try", nil)
	cacheGetBlockBodyRLPHitMeter = metrics.NewRegisteredMeter("klay/cache/get/blockbodyrlp/hit", nil)

	cacheGetBlockTryMeter = metrics.NewRegisteredMeter("klay/cache/get/block/try", nil)
	cacheGetBlockHitMeter = metrics.NewRegisteredMeter("klay/cache/get/block/hit", nil)

	cacheGetFutureBlockTryMeter = metrics.NewRegisteredMeter("klay/cache/get/futureblock/try", nil)
	cacheGetFutureBlockHitMeter = metrics.NewRegisteredMeter("klay/cache/get/futureblock/hit", nil)

	cacheGetBadBlockTryMeter = metrics.NewRegisteredMeter("klay/cache/get/badblock/try", nil)
	cacheGetBadBlockHitMeter = metrics.NewRegisteredMeter("klay/cache/get/badblock/hit", nil)

	cacheGetRecentTransactionsTryMeter = metrics.NewRegisteredMeter("klay/cache/get/transactions/try", nil)
	cacheGetRecentTransactionsHitMeter = metrics.NewRegisteredMeter("klay/cache/get/transactions/hit", nil)

	cacheGetRecentReceiptsTryMeter = metrics.NewRegisteredMeter("klay/cache/get/receipts/try", nil)
	cacheGetRecentReceiptsHitMeter = metrics.NewRegisteredMeter("klay/cache/get/receipts/hit", nil)

	cacheGetHeaderTryMeter = metrics.NewRegisteredMeter("klay/cache/get/header/try", nil)
	cacheGetHeaderHitMeter = metrics.NewRegisteredMeter("klay/cache/get/header/hit", nil)

	cacheGetTDTryMeter = metrics.NewRegisteredMeter("klay/cache/get/td/try", nil)
	cacheGetTDHitMeter = metrics.NewRegisteredMeter("klay/cache/get/td/hit", nil)

	cacheGetBlockNumberTryMeter = metrics.NewRegisteredMeter("klay/cache/get/blocknumber/try", nil)
	cacheGetBlockNumberHitMeter = metrics.NewRegisteredMeter("klay/cache/get/blocknumber/hit", nil)

	headBlockNumberGauge = metrics.NewRegisteredGauge("blockchain/head/blocknumber", nil)
	blockTxCountsMeter   = metrics.NewRegisteredMeter("blockchain/block/tx/rate", nil)
	blockTxCountsCounter = metrics.NewRegisteredCounter("blockchain/block/tx/counter", nil)

	txPoolPendingGauge = metrics.NewRegisteredGauge("tx/pool/pending/gauge", nil)
	txPoolQueueGauge   = metrics.NewRegisteredGauge("tx/pool/queue/gauge", nil)
)
