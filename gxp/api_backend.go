package gxp

import (
	"context"
	"github.com/ground-x/go-gxplatform/accounts"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/common/math"
	"github.com/ground-x/go-gxplatform/core"
	"github.com/ground-x/go-gxplatform/core/bloombits"
	"github.com/ground-x/go-gxplatform/core/rawdb"
	"github.com/ground-x/go-gxplatform/core/state"
	"github.com/ground-x/go-gxplatform/core/types"
	"github.com/ground-x/go-gxplatform/core/vm"
	"github.com/ground-x/go-gxplatform/event"
	"github.com/ground-x/go-gxplatform/gxdb"
	"github.com/ground-x/go-gxplatform/gxp/downloader"
	"github.com/ground-x/go-gxplatform/gxp/gasprice"
	"github.com/ground-x/go-gxplatform/params"
	"github.com/ground-x/go-gxplatform/rpc"
	"math/big"
	"github.com/ground-x/go-gxplatform/log"
)

// GxpAPIBackend implements gxpapi.Backend for full nodes
type GxpAPIBackend struct {
	gxp *GXP
	gpo *gasprice.Oracle
}

func (b *GxpAPIBackend) ChainConfig() *params.ChainConfig {
	return b.gxp.chainConfig
}

func (b *GxpAPIBackend) CurrentBlock() *types.Block {
	return b.gxp.blockchain.CurrentBlock()
}

func (b *GxpAPIBackend) SetHead(number uint64) {
	//b.gxp.protocolManager.downloader.Cancel()
	b.gxp.blockchain.SetHead(number)
}

func (b *GxpAPIBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	log.Info("GxpAPIBackend.HeaderByNumber")
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.gxp.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.gxp.blockchain.CurrentBlock().Header(), nil
	}
	return b.gxp.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *GxpAPIBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	log.Info("GxpAPIBackend.BlockByNumber")
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.gxp.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.gxp.blockchain.CurrentBlock(), nil
	}
	return b.gxp.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *GxpAPIBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.gxp.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.gxp.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *GxpAPIBackend) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	log.Info("GxpAPIBackend.GetBlock")
	return b.gxp.blockchain.GetBlockByHash(hash), nil
}

func (b *GxpAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	log.Info("GxpAPIBackend.GetReceipts")
	if number := rawdb.ReadHeaderNumber(b.gxp.chainDb, hash); number != nil {
		return rawdb.ReadReceipts(b.gxp.chainDb, hash, *number), nil
	}
	return nil, nil
}

func (b *GxpAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	log.Info("GxpAPIBackend.GetTd")
	number := rawdb.ReadHeaderNumber(b.gxp.chainDb, hash)
	if number == nil {
		return nil, nil
	}
	receipts := rawdb.ReadReceipts(b.gxp.chainDb, hash, *number)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *GxpAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	log.Info("GxpAPIBackend.GetTd")
	return b.gxp.blockchain.GetTdByHash(blockHash)
}

func (b *GxpAPIBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	log.Info("GxpAPIBackend.GetEVM")
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.gxp.BlockChain(), nil)
	return vm.NewEVM(context, state, b.gxp.chainConfig, vmCfg), vmError, nil
}

func (b *GxpAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	log.Info("GxpAPIBackend.SubscribeRemovedLogsEvent")
	return b.gxp.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *GxpAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	log.Info("GxpAPIBackend.SubscribeChainEvent")
	return b.gxp.BlockChain().SubscribeChainEvent(ch)
}

func (b *GxpAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	log.Info("GxpAPIBackend.SubscribeChainHeadEvent")
	return b.gxp.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *GxpAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	log.Info("GxpAPIBackend.SubscribeChainSideEvent")
	return b.gxp.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *GxpAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	log.Info("GxpAPIBackend.SubscribeLogsEvent")
	return b.gxp.BlockChain().SubscribeLogsEvent(ch)
}

func (b *GxpAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	log.Info("GxpAPIBackend.SendTx")
	return b.gxp.txPool.AddLocal(signedTx)
}

func (b *GxpAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	log.Info("GxpAPIBackend.GetPoolTransactions")
	pending, err := b.gxp.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *GxpAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	log.Info("GxpAPIBackend.GetPoolTransaction")
	return b.gxp.txPool.Get(hash)
}

func (b *GxpAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	log.Info("GxpAPIBackend.GetPoolNonce")
	return b.gxp.txPool.State().GetNonce(addr), nil
}

func (b *GxpAPIBackend) Stats() (pending int, queued int) {
	log.Info("GxpAPIBackend.Stats")
	return b.gxp.txPool.Stats()
}

func (b *GxpAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	log.Info("GxpAPIBackend.TxPoolContent")
	return b.gxp.TxPool().Content()
}

func (b *GxpAPIBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	log.Info("GxpAPIBackend.SubscribeNewTxsEvent")
	return b.gxp.TxPool().SubscribeNewTxsEvent(ch)
}

func (b *GxpAPIBackend) Downloader() *downloader.Downloader {
	return b.gxp.Downloader()
}

func (b *GxpAPIBackend) ProtocolVersion() int {
	return b.gxp.GxpVersion()
}

func (b *GxpAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *GxpAPIBackend) ChainDb() gxdb.Database {
	log.Info("GxpAPIBackend.ChainDb")
	return b.gxp.ChainDb()
}

func (b *GxpAPIBackend) EventMux() *event.TypeMux {
	log.Info("GxpAPIBackend.EventMux")
	return b.gxp.EventMux()
}

func (b *GxpAPIBackend) AccountManager() *accounts.Manager {
	log.Info("GxpAPIBackend.AccountManager")
	return b.gxp.AccountManager()
}

func (b *GxpAPIBackend) BloomStatus() (uint64, uint64) {
	log.Info("GxpAPIBackend.BloomStatus")
	sections, _, _ := b.gxp.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *GxpAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.gxp.bloomRequests)
	}
}
