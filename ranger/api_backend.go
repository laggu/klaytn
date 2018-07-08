package ranger

import (
	"math/big"
	"github.com/ground-x/go-gxplatform/params"
	"github.com/ground-x/go-gxplatform/core/types"
	"context"
	"github.com/ground-x/go-gxplatform/rpc"
	"github.com/ground-x/go-gxplatform/core/state"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/core/rawdb"
	"github.com/ground-x/go-gxplatform/core"
	"github.com/ground-x/go-gxplatform/event"
	"github.com/ground-x/go-gxplatform/core/vm"
	"github.com/ground-x/go-gxplatform/common/math"
	"github.com/ground-x/go-gxplatform/gxp/downloader"
	"github.com/ground-x/go-gxplatform/gxdb"
	"github.com/ground-x/go-gxplatform/accounts"
	"github.com/ground-x/go-gxplatform/core/bloombits"
	"time"
)

// RangerAPIBackend implements gxpapi.Backend for ranger nodes
type RangerAPIBackend struct {
	ranger *Ranger
}

func (b *RangerAPIBackend) ChainConfig() *params.ChainConfig {
	return b.ranger.chainConfig
}

func (b *RangerAPIBackend) CurrentBlock() *types.Block {
	return b.ranger.blockchain.CurrentBlock()
}

func (b *RangerAPIBackend) SetHead(number uint64) {
	//b.gxp.protocolManager.downloader.Cancel()
	b.ranger.blockchain.SetHead(number)
}

func (b *RangerAPIBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.ranger.blockchain.CurrentBlock().Header(), nil
	}
	return b.ranger.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *RangerAPIBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.ranger.blockchain.CurrentBlock(), nil
	}
	return b.ranger.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *RangerAPIBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.ranger.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *RangerAPIBackend) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.ranger.blockchain.GetBlockByHash(hash), nil
}

func (b *RangerAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	if number := rawdb.ReadHeaderNumber(b.ranger.chainDb, hash); number != nil {
		return rawdb.ReadReceipts(b.ranger.chainDb, hash, *number), nil
	}
	return nil, nil
}

func (b *RangerAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	number := rawdb.ReadHeaderNumber(b.ranger.chainDb, hash)
	if number == nil {
		return nil, nil
	}
	receipts := rawdb.ReadReceipts(b.ranger.chainDb, hash, *number)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *RangerAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.ranger.blockchain.GetTdByHash(blockHash)
}

func (b *RangerAPIBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.ranger.BlockChain(), nil)
	return vm.NewEVM(context, state, b.ranger.chainConfig, vmCfg), vmError, nil
}

func (b *RangerAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.ranger.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *RangerAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.ranger.BlockChain().SubscribeChainEvent(ch)
}

func (b *RangerAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.ranger.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *RangerAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.ranger.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *RangerAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.ranger.BlockChain().SubscribeLogsEvent(ch)
}

func (b *RangerAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return nil
}

func (b *RangerAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.ranger.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *RangerAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.ranger.txPool.Get(hash)
}

func (b *RangerAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.ranger.txPool.State().GetNonce(addr), nil
}

func (b *RangerAPIBackend) Stats() (pending int, queued int) {
	return b.ranger.txPool.Stats()
}

func (b *RangerAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.ranger.TxPool().Content()
}

func (b *RangerAPIBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return b.ranger.TxPool().SubscribeNewTxsEvent(ch)
}

func (b *RangerAPIBackend) Downloader() *downloader.Downloader {
	return b.ranger.Downloader()
}

func (b *RangerAPIBackend) ProtocolVersion() int {
	return b.ranger.GxpVersion()
}

func (b *RangerAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return common.Big0, nil
}

func (b *RangerAPIBackend) ChainDb() gxdb.Database {
	return b.ranger.ChainDb()
}

func (b *RangerAPIBackend) EventMux() *event.TypeMux {
	return b.ranger.EventMux()
}

func (b *RangerAPIBackend) AccountManager() *accounts.Manager {
	return b.ranger.AccountManager()
}

func (b *RangerAPIBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.ranger.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

const (
	// bloomFilterThreads is the number of goroutines used locally per filter to
	// multiplex requests onto the global servicing goroutines.
	bloomFilterThreads = 3

	// bloomRetrievalBatch is the maximum number of bloom bit retrievals to service
	// in a single batch.
	bloomRetrievalBatch = 16

	// bloomRetrievalWait is the maximum time to wait for enough bloom bit requests
	// to accumulate request an entire batch (avoiding hysteresis).
	bloomRetrievalWait = time.Duration(0)
)

func (b *RangerAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.ranger.bloomRequests)
	}
}
