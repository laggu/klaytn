package blockchain

import (
	"github.com/ground-x/go-gxplatform/metrics"
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

	cacheGetRecentReceiptsMissMeter = metrics.NewRegisteredMeter("klay/cache/get/receipts/miss", nil)
	cacheGetRecentReceiptsHitMeter  = metrics.NewRegisteredMeter("klay/cache/get/receipts/hit", nil)

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
